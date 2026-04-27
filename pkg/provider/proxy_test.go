// SPDX-License-Identifier: EUPL-1.2

package provider_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"dappco.re/go/api"
	"dappco.re/go/api/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestProxyProvider_Name_Good(t *testing.T) {
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "cool-widget",
		BasePath: "/api/v1/cool-widget",
		Upstream: "http://127.0.0.1:9999",
	})
	assert.Equal(t, "cool-widget", p.Name())
}

func TestProxyProvider_BasePath_Good(t *testing.T) {
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "cool-widget",
		BasePath: "/api/v1/cool-widget",
		Upstream: "http://127.0.0.1:9999",
	})
	assert.Equal(t, "/api/v1/cool-widget", p.BasePath())
}

func TestProxyProvider_Element_Good(t *testing.T) {
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
	assert.Equal(t, "core-cool-widget", p.Element().Tag)
	assert.Equal(t, "/assets/cool-widget.js", p.Element().Source)
}

func TestProxyProvider_SpecFile_Good(t *testing.T) {
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "cool-widget",
		BasePath: "/api/v1/cool-widget",
		Upstream: "http://127.0.0.1:9999",
		SpecFile: "/tmp/openapi.json",
	})
	assert.Equal(t, "/tmp/openapi.json", p.SpecFile())
}

func TestProxyProvider_Proxy_Good(t *testing.T) {
	// Start a test upstream server.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"path":   r.URL.Path,
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
	require.NoError(t, err)
	engine.Register(p)

	handler := engine.Handler()

	// Send a request through the proxy.
	req := httptest.NewRequest("GET", "/api/v1/test-proxy/items", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)

	// The upstream should see the path with base path stripped.
	assert.Equal(t, "/items", body["path"])
	assert.Equal(t, "GET", body["method"])
}

func TestProxyProvider_ProxyRoot_Good(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"path": r.URL.Path}
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
	require.NoError(t, err)
	engine.Register(p)

	handler := engine.Handler()

	// Request to the base path itself (root of the provider).
	req := httptest.NewRequest("GET", "/api/v1/test-proxy/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "/", body["path"])
}

func TestProxyProvider_HealthPassthrough_Good(t *testing.T) {
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
	require.NoError(t, err)
	engine.Register(p)

	handler := engine.Handler()

	req := httptest.NewRequest("GET", "/api/v1/health-test/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

func TestProxyProvider_Renderable_Good(t *testing.T) {
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
	require.Len(t, renderables, 1)
	assert.Equal(t, "core-test-panel", renderables[0].Element().Tag)
}

func TestProxyProvider_Ugly_InvalidUpstream(t *testing.T) {
	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "bad",
		BasePath: "/api/v1/bad",
		Upstream: "://not-a-url",
	})

	require.NotNil(t, p)
	assert.Error(t, p.Err())

	engine, err := api.New()
	require.NoError(t, err)
	engine.Register(p)

	handler := engine.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bad/items", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.Equal(t, false, body["success"])
	errObj, ok := body["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "invalid_provider_configuration", errObj["code"])
}

func TestProxyProvider_NewProxy_Good_PublicUpstream(t *testing.T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "public",
		BasePath: "/api/v1/public",
		Upstream: "http://1.1.1.1/x",
	})

	require.NotNil(t, p)
	assert.NoError(t, p.Err())
}

func TestProxyProvider_NewProxy_Bad_BlocksMetadataIP(t *testing.T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	assertProviderUpstreamBlocked(t, "http://169.254.169.254/x")
}

func TestProxyProvider_NewProxy_Bad_BlocksLoopback(t *testing.T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	assertProviderUpstreamBlocked(t, "http://127.0.0.1:5432/")
}

func TestProxyProvider_NewProxy_Bad_BlocksRFC1918(t *testing.T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	assertProviderUpstreamBlocked(t, "http://10.0.0.1/x")
}

func TestProxyProvider_NewProxy_Good_AllowListPermitsLoopback(t *testing.T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "127.0.0.0/8")

	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "allowed-loopback",
		BasePath: "/api/v1/allowed-loopback",
		Upstream: "http://127.0.0.1:5432/",
	})

	require.NotNil(t, p)
	assert.NoError(t, p.Err())
}

func TestProxyProvider_NewProxy_Bad_AllowListDoesNotPermitOtherPrivateCIDRs(t *testing.T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "127.0.0.0/8")

	assertProviderUpstreamBlocked(t, "http://10.0.0.1/")
}

func TestProxyProvider_NewProxy_Bad_BlocksHostnameResolvingToLoopback(t *testing.T) {
	t.Setenv("CORE_PROVIDER_UPSTREAM_ALLOW", "")

	assertProviderUpstreamBlocked(t, "http://localhost:5432/")
}

func assertProviderUpstreamBlocked(t *testing.T, upstream string) {
	t.Helper()

	p := provider.NewProxy(provider.ProxyConfig{
		Name:     "blocked",
		BasePath: "/api/v1/blocked",
		Upstream: upstream,
	})

	require.NotNil(t, p)
	err := p.Err()
	require.Error(t, err)
	assert.True(t, errors.Is(err, provider.ErrProviderUpstreamBlocked), "expected ErrProviderUpstreamBlocked, got %v", err)

	var blocked *provider.ProviderUpstreamBlockedError
	require.True(t, errors.As(err, &blocked), "expected ProviderUpstreamBlockedError, got %T", err)
	assert.Equal(t, upstream, blocked.Upstream)
	assert.NotEmpty(t, blocked.Reason)
}
