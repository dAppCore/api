// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	api "dappco.re/go/api"
	"github.com/gin-gonic/gin"
)

// serve builds a test engine with the upstream router mounted and returns a live server.
func serve(t *testing.T, reg *api.UpstreamRegistry, opts ...api.UpstreamRouterOption) *httptest.Server {
	t.Helper()
	e, err := api.New(api.WithUpstreamRouter(reg, opts...))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return httptest.NewServer(e.Handler())
}

func post(t *testing.T, base, path, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(base+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func TestUpstreamRouter_RoutesByModel_Good(t *testing.T) {
	upA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"upstream":"A"}`)
	}))
	defer upA.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.Set("lemma", api.Upstream{URL: upA.URL}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	srv := serve(t, reg)
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"lemma"}`)
	defer resp.Body.Close()
	got, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(got), `"upstream":"A"`) {
		t.Fatalf("body = %s, want routed to A", got)
	}
}

func TestUpstreamRouter_MissingModel_Bad(t *testing.T) {
	reg := api.NewUpstreamRegistry()
	_ = reg.SetDefault(api.Upstream{URL: "https://example.com"})
	srv := serve(t, reg)
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestUpstreamRouter_Failover_Good(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer dead.Close()
	live := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer live.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.Set("m", api.Upstream{URL: dead.URL}, api.Upstream{URL: live.URL}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	srv := serve(t, reg)
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m"}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (failed over to live)", resp.StatusCode)
	}
}

func TestUpstreamRouter_AllDown_503_Ugly(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer dead.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("m", api.Upstream{URL: dead.URL})
	srv := serve(t, reg)
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m"}`)
	defer resp.Body.Close()
	got, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Error("missing Retry-After header on 503")
	}
	if strings.Contains(string(got), dead.URL) {
		t.Error("upstream URL leaked into client response body")
	}
}

func TestUpstreamRouter_StreamingPassthrough_Good(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f, _ := w.(http.Flusher)
		for _, chunk := range []string{"data: a\n\n", "data: b\n\n", "data: [DONE]\n\n"} {
			_, _ = io.WriteString(w, chunk)
			if f != nil {
				f.Flush()
			}
		}
	}))
	defer up.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("m", api.Upstream{URL: up.URL})
	// Out transformer present to prove it is NOT applied to streams.
	srv := serve(t, reg, api.WithUpstreamTransformerOut(api.RenameFields(map[string]string{"x": "y"})))
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m"}`)
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}
	sc := bufio.NewScanner(resp.Body)
	var lines int
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "data:") {
			lines++
		}
	}
	if lines != 3 {
		t.Fatalf("got %d data lines, want 3 (stream byte-preserved)", lines)
	}
}

func TestUpstreamRouter_TransformInOut_Good(t *testing.T) {
	var gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = io.WriteString(w, `{"internal_id":42}`)
	}))
	defer up.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("m", api.Upstream{URL: up.URL})
	srv := serve(t, reg,
		api.WithUpstreamTransformerIn(api.RenameFields(map[string]string{"q": "prompt"})),
		api.WithUpstreamTransformerOut(api.RenameFields(map[string]string{"internal_id": "id"})),
	)
	defer srv.Close()

	// Selector reads "model" from the original body; the in-transform then renames
	// q->prompt before dispatch so the upstream sees the translated shape.
	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m","q":"hello"}`)
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if !strings.Contains(gotBody, `"prompt"`) {
		t.Errorf("upstream body = %s, want renamed q->prompt", gotBody)
	}
	if !strings.Contains(string(out), `"id":42`) {
		t.Errorf("client body = %s, want renamed internal_id->id", out)
	}
}

func TestUpstreamRouter_RouteHookOverride_Good(t *testing.T) {
	upB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"pool":"B"}`)
	}))
	defer upB.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("b", api.Upstream{URL: upB.URL})
	srv := serve(t, reg, api.WithRouteHook(func(_ *gin.Context, _ string, _ []byte) (string, error) {
		return "b", nil
	}))
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"anything"}`)
	defer resp.Body.Close()
	got, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(got), `"pool":"B"`) {
		t.Fatalf("body = %s, want hook-overridden pool B", got)
	}
}

func TestUpstreamRouter_SSRFPosture_Bad(t *testing.T) {
	reg := api.NewUpstreamRegistry() // no allow-list
	if err := reg.Set("m", api.Upstream{URL: "http://127.0.0.1:11434"}); err == nil {
		t.Fatal("loopback accepted without AllowPrivateUpstreams, want rejection")
	}
}

func TestUpstreamRouter_Composition_PreNextMiddleware_Good(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer up.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.SetDefault(api.Upstream{URL: up.URL})
	// WithRateLimit runs BEFORE the handler (pre-c.Next()), annotating passing
	// requests with X-RateLimit-Limit. A response written by the proxy during the
	// handler therefore still carries this header — proving engine middleware
	// wraps (gates) the mounted router. Post-Next response-header middleware (e.g.
	// ApiSunset) cannot apply here because the proxy commits the response during
	// the handler; see the WithUpstreamRouter docs.
	e, _ := api.New(api.WithRateLimit(100), api.WithUpstreamRouter(reg))
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m"}`)
	defer resp.Body.Close()
	got, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (request proxied through)", resp.StatusCode)
	}
	if !strings.Contains(string(got), `"ok":true`) {
		t.Fatalf("body = %s, want proxied upstream body", got)
	}
	if resp.Header.Get("X-RateLimit-Limit") == "" {
		t.Fatal("X-RateLimit-Limit absent — engine middleware did not wrap the mounted router")
	}
}
