// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	core "dappco.re/go"
)

type fakeRoundTripper struct {
	fn func(*http.Request) (*http.Response, error)
}

func (f fakeRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f.fn(r) }

func newResp(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("ok")),
		Header:     http.Header{},
	}
}

func requestWithPool(pool []Upstream, key string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader("{}"))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("{}")), nil }
	ctx := context.WithValue(req.Context(), poolCtxKey, pool)
	ctx = context.WithValue(ctx, keyCtxKey, key)
	return req.WithContext(ctx)
}

func TestUpstreamTransport_FailoverThenSuccess_Good(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	var hits []string
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		hits = append(hits, r.URL.Host)
		if r.URL.Host == "a" {
			return newResp(http.StatusBadGateway), nil
		}
		return newResp(http.StatusOK), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 2, failover: defaultFailoverStatuses()}
	pool := []Upstream{{URL: "http://a", Weight: 1}, {URL: "http://b", Weight: 1}}

	resp, err := tr.RoundTrip(requestWithPool(pool, "k"))
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(hits) != 2 {
		t.Fatalf("attempts = %v, want 2 (a then b)", hits)
	}
}

func TestUpstreamTransport_HeaderInjection_Good(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	var gotAuth string
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		return newResp(http.StatusOK), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 1, failover: defaultFailoverStatuses()}
	pool := []Upstream{{URL: "http://a", Headers: map[string]string{"Authorization": "Bearer up-key"}}}

	if _, err := tr.RoundTrip(requestWithPool(pool, "k")); err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if gotAuth != "Bearer up-key" {
		t.Fatalf("injected auth = %q, want Bearer up-key", gotAuth)
	}
}

func TestUpstreamTransport_AllFail_Bad(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		return newResp(http.StatusServiceUnavailable), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 2, failover: defaultFailoverStatuses()}
	pool := []Upstream{{URL: "http://a"}, {URL: "http://b"}}

	_, err := tr.RoundTrip(requestWithPool(pool, "k"))
	var re *routerError
	if !core.As(err, &re) || re.status != http.StatusServiceUnavailable {
		t.Fatalf("err = %v, want *routerError status 503", err)
	}
}

func TestUpstreamTransport_FourxxPassthrough_Good(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	var hits int
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		hits++
		return newResp(http.StatusNotFound), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 2, failover: defaultFailoverStatuses()}
	pool := []Upstream{{URL: "http://a"}, {URL: "http://b"}}

	resp, err := tr.RoundTrip(requestWithPool(pool, "k"))
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (4xx is not a failover status)", resp.StatusCode)
	}
	if hits != 1 {
		t.Fatalf("attempts = %d, want 1 (no failover, no markFailed on 4xx)", hits)
	}
}

func TestUpstreamTransport_TransportErrorRetry_Good(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	var hits []string
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		hits = append(hits, r.URL.Host)
		if r.URL.Host == "a" {
			return nil, errors.New("dial tcp: connection refused")
		}
		return newResp(http.StatusOK), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 2, failover: defaultFailoverStatuses()}
	pool := []Upstream{{URL: "http://a", Weight: 1}, {URL: "http://b", Weight: 1}}

	resp, err := tr.RoundTrip(requestWithPool(pool, "k"))
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (retried past the transport error)", resp.StatusCode)
	}
	if len(hits) != 2 || hits[1] != "b" {
		t.Fatalf("attempts = %v, want [a b] (transport error on a, retried b)", hits)
	}
}

func TestUpstreamTransport_BodyReplayedOnRetry_Good(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	const payload = `{"model":"m","prompt":"the full body must survive failover"}`
	var seen []string
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		seen = append(seen, string(b))
		if r.URL.Host == "a" {
			return newResp(http.StatusBadGateway), nil
		}
		return newResp(http.StatusOK), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 2, failover: defaultFailoverStatuses()}
	pool := []Upstream{{URL: "http://a", Weight: 1}, {URL: "http://b", Weight: 1}}

	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(payload))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(payload)), nil }
	ctx := context.WithValue(req.Context(), poolCtxKey, pool)
	ctx = context.WithValue(ctx, keyCtxKey, "k")

	resp, err := tr.RoundTrip(req.WithContext(ctx))
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(seen) != 2 {
		t.Fatalf("got %d attempts, want 2", len(seen))
	}
	for i, body := range seen {
		if body != payload {
			t.Fatalf("attempt %d body = %q, want full payload %q", i, body, payload)
		}
	}
}

func TestUpstreamTransport_AllCooling503_Bad(t *testing.T) {
	now := time.Unix(1000, 0)
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return now })
	pool := []Upstream{{URL: "http://a"}, {URL: "http://b"}}
	// Pre-cool every member so each pick returns !ok and the loop exits with
	// lastErr == nil — the all-cooling path, distinct from the all-fail path.
	bal.markFailed("http://a")
	bal.markFailed("http://b")

	var hits int
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		hits++
		return newResp(http.StatusOK), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 2, failover: defaultFailoverStatuses()}

	_, err := tr.RoundTrip(requestWithPool(pool, "k"))
	var re *routerError
	if !core.As(err, &re) || re.status != http.StatusServiceUnavailable {
		t.Fatalf("err = %v, want *routerError status 503", err)
	}
	if re.cause != nil {
		t.Fatalf("all-cooling error carried a cause %v, want nil", re.cause)
	}
	if hits != 0 {
		t.Fatalf("base dispatched %d times, want 0 (every member cooling)", hits)
	}
}

func TestUpstreamTransport_ApplyUpstreamRewrite_Good(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	var gotScheme, gotHost, gotPath string
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		gotScheme, gotHost, gotPath = r.URL.Scheme, r.URL.Host, r.URL.Path
		return newResp(http.StatusOK), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 1, failover: defaultFailoverStatuses()}
	// Upstream carries a base path; the incoming /v1/chat/completions must be
	// prefixed with it after the rewrite.
	pool := []Upstream{{URL: "https://gw.example.com:8443/proxy"}}

	if _, err := tr.RoundTrip(requestWithPool(pool, "k")); err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if gotScheme != "https" {
		t.Fatalf("scheme = %q, want https", gotScheme)
	}
	if gotHost != "gw.example.com:8443" {
		t.Fatalf("host = %q, want gw.example.com:8443", gotHost)
	}
	if gotPath != "/proxy/v1/chat/completions" {
		t.Fatalf("path = %q, want /proxy/v1/chat/completions (base-path prefixed)", gotPath)
	}
}
