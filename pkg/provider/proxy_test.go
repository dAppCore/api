// SPDX-License-Identifier: EUPL-1.2

package provider_test

import (
	"dappco.re/go/api/internal/stdcompat/errors"
	"dappco.re/go/api/internal/stdcompat/json"
	"dappco.re/go/api/internal/stdcompat/os"
	"net/http"
	"net/http/httptest"
	"testing"

	. "dappco.re/go"
	"dappco.re/go/api"
	"dappco.re/go/api/pkg/provider"
)

func TestMain(m *testing.M) {
	const env = "CORE_PROVIDER_UPSTREAM_ALLOW"

	previous, hadPrevious := os.LookupEnv(env)
	_ = os.Setenv(env, "127.0.0.0/8,::1/128")
	code := m.Run()
	if hadPrevious {
		_ = os.Setenv(env, previous)
	} else {
		_ = os.Unsetenv(env)
	}
	os.Exit(code)
}

// -- ProxyProvider tests ------------------------------------------------------

func TestProxyProvider_Name_Good(t *T) {
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "cool-widget",
		BasePath: "/api/v1/cool-widget",
		Upstream: "http://127.0.0.1:9999",
	})
	AssertEqual(t, "cool-widget", p.Name())
}

func TestProxyProvider_BasePath_Good(t *T) {
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "cool-widget",
		BasePath: "/api/v1/cool-widget",
		Upstream: "http://127.0.0.1:9999",
	})
	AssertEqual(t, "/api/v1/cool-widget", p.BasePath())
}

func TestProxyProvider_Element_Good(t *T) {
	elem := provider.ElementSpec{
		Tag:    "core-cool-widget",
		Source: "/assets/cool-widget.js",
	}
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "cool-widget",
		BasePath: "/api/v1/cool-widget",
		Upstream: "http://127.0.0.1:9999",
		Element:  elem,
	})
	AssertEqual(t, "core-cool-widget", p.Element().Tag)
	AssertEqual(t, "/assets/cool-widget.js", p.Element().Source)
}

func TestProxyProvider_SpecFile_Good(t *T) {
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "cool-widget",
		BasePath: "/api/v1/cool-widget",
		Upstream: "http://127.0.0.1:9999",
		SpecFile: "/tmp/openapi.json",
	})
	AssertEqual(t, "/tmp/openapi.json", p.SpecFile())
}

func TestProxyProviderProxyForwards(t *T) {
	// Start a test upstream server.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			`path`:   r.URL.Path,
			"method": r.Method,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer upstream.Close()

	// Create proxy provider pointing to the test server.
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "test-proxy",
		BasePath: "/api/v1/test-proxy",
		Upstream: upstream.URL,
	})

	// Mount on an api.Engine.
	engine, err := api.New()
	RequireNoError(t, err)
	engine.Register(p)

	handler := engine.Handler()

	// Send a request through the proxy.
	req := httptest.NewRequest("GET", "/api/v1/test-proxy/items", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	AssertEqual(t, http.StatusOK, w.Code)

	var body map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &body)
	RequireNoError(t, err)

	// The upstream should see the path with base path stripped.
	AssertEqual(t, "/items", body[`path`])
	AssertEqual(t, "GET", body["method"])
}

func TestProxyProviderProxyRootForwards(t *T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{`path`: r.URL.Path}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer upstream.Close()

	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "test-proxy",
		BasePath: "/api/v1/test-proxy",
		Upstream: upstream.URL,
	})

	engine, err := api.New()
	RequireNoError(t, err)
	engine.Register(p)

	handler := engine.Handler()

	// Request to the base path itself (root of the provider).
	req := httptest.NewRequest("GET", "/api/v1/test-proxy/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	AssertEqual(t, http.StatusOK, w.Code)

	var body map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &body)
	RequireNoError(t, err)
	AssertEqual(t, "/", body[`path`])
}

func TestProxyProviderHealthPassthrough(t *T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer upstream.Close()

	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "health-test",
		BasePath: "/api/v1/health-test",
		Upstream: upstream.URL,
	})

	engine, err := api.New()
	RequireNoError(t, err)
	engine.Register(p)

	handler := engine.Handler()

	req := httptest.NewRequest("GET", "/api/v1/health-test/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	AssertEqual(t, http.StatusOK, w.Code)
	AssertContains(t, w.Body.String(), `"status":"ok"`)
}

func TestProxyProvider_Renderable_Good(t *T) {
	// Verify ProxyProvider satisfies Renderable via the Registry.
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "renderable-proxy",
		BasePath: "/api/v1/renderable",
		Upstream: "http://127.0.0.1:9999",
		Element:  provider.ElementSpec{Tag: "core-test-panel", Source: "/assets/test.js"},
	})

	reg := provider.NewRegistry()
	reg.Add(p)

	renderables := reg.Renderable()
	AssertLen(t, renderables, 1)
	AssertEqual(t, "core-test-panel", renderables[0].Element().Tag)
}

func TestProxyProvider_Ugly_InvalidUpstream(t *T) {
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "bad",
		BasePath: "/api/v1/bad",
		Upstream: "://not-a-url",
	})

	AssertNotNil(t, p)
	AssertError(t, p.Err())

	engine, err := api.New()
	RequireNoError(t, err)
	engine.Register(p)

	handler := engine.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bad/items", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	AssertEqual(t, http.StatusInternalServerError, w.Code)

	var body map[string]any
	RequireNoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	AssertEqual(t, false, body["success"])
	errObj, ok := body["error"].(map[string]any)
	RequireTrue(t, ok)
	AssertEqual(t, "invalid_provider_configuration", errObj["code"])
}

func TestProxyProvider_NewProxy_Good_PublicUpstream(t *T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "public",
		BasePath: "/api/v1/public",
		Upstream: "http://1.1.1.1/x",
	})

	AssertNotNil(t, p)
	AssertNoError(t, p.Err())
}

func TestProxyProvider_NewProxy_Bad_BlocksMetadataIP(t *T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	err := assertProviderUpstreamBlocked(t, "http://169.254.169.254/x")
	AssertContains(t, err.Error(), "blocked")
}

func TestProxyProvider_NewProxy_Bad_BlocksLoopback(t *T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	err := assertProviderUpstreamBlocked(t, "http://127.0.0.1:5432/")
	AssertContains(t, err.Error(), "blocked")
}

func TestProxyProvider_NewProxy_Bad_BlocksRFC1918(t *T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	err := assertProviderUpstreamBlocked(t, "http://10.0.0.1/x")
	AssertContains(t, err.Error(), "blocked")
}

func TestProxyProvider_NewProxy_Good_AllowListPermitsLoopback(t *T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "127.0.0.0/8")

	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "allowed-loopback",
		BasePath: "/api/v1/allowed-loopback",
		Upstream: "http://127.0.0.1:5432/",
	})

	AssertNotNil(t, p)
	AssertNoError(t, p.Err())
}

func TestProxyProvider_NewProxy_Bad_AllowListDoesNotPermitOtherPrivateCIDRs(t *T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "127.0.0.0/8")

	err := assertProviderUpstreamBlocked(t, "http://10.0.0.1/")
	AssertContains(t, err.Error(), "blocked")
}

func TestProxyProvider_NewProxy_Bad_BlocksHostnameResolvingToLoopback(t *T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	err := assertProviderUpstreamBlocked(t, "http://localhost:5432/")
	AssertContains(t, err.Error(), "blocked")
}

func assertProviderUpstreamBlocked(t *T, upstream string) error {
	t.Helper()

	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "blocked",
		BasePath: "/api/v1/blocked",
		Upstream: upstream,
	})

	AssertNotNil(t, p)
	err := p.Err()
	AssertError(t, err)
	AssertTrue(t, errors.Is(err, provider.ErrProviderUpstreamBlocked), "expected ErrProviderUpstreamBlocked")

	var blocked *provider.ProviderUpstreamBlockedError
	RequireTrue(t, errors.As(err, &blocked), "expected ProviderUpstreamBlockedError")
	AssertEqual(t, upstream, blocked.Upstream)
	AssertNotEmpty(t, blocked.Reason)
	return err
}
