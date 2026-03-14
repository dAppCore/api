// SPDX-Licence-Identifier: EUPL-1.2

package provider_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"forge.lthn.ai/core/api"
	"forge.lthn.ai/core/api/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	assert.Panics(t, func() {
		provider.NewProxy(provider.ProxyConfig{
			Name:     "bad",
			BasePath: "/api/v1/bad",
			Upstream: "://not-a-url",
		})
	})
}
