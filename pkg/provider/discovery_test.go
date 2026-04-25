// SPDX-Licence-Identifier: EUPL-1.2

package provider_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"dappco.re/go/api"
	"dappco.re/go/api/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscover_Good_LoadsYAMLProxyProvider(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"path": r.URL.Path})
	}))
	defer upstream.Close()

	dir := filepath.Join(t.TempDir(), ".core", "providers")
	require.NoError(t, os.MkdirAll(dir, 0755))
	specPath := filepath.Join(filepath.Dir(dir), "specs", "openapi.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(specPath), 0755))
	require.NoError(t, os.WriteFile(specPath, []byte("openapi: 3.1.0\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cool.yaml"), []byte(`
name: cool-widget
runtime: php
base_path: /api/v1/cool-widget/
upstream: `+upstream.URL+`
spec_file: ../specs/openapi.yaml
element:
  tag: core-cool-widget
  source: /assets/cool-widget.js
`), 0644))

	providers, err := provider.Discover(dir)
	require.NoError(t, err)
	require.Len(t, providers, 1)

	p := providers[0]
	assert.Equal(t, "cool-widget", p.Name())
	assert.Equal(t, "/api/v1/cool-widget", p.BasePath())

	specProvider, ok := p.(interface{ SpecFile() string })
	require.True(t, ok)
	assert.Equal(t, specPath, specProvider.SpecFile())

	renderable, ok := p.(provider.Renderable)
	require.True(t, ok)
	assert.Equal(t, "core-cool-widget", renderable.Element().Tag)

	engine, err := api.New()
	require.NoError(t, err)
	engine.Register(p)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cool-widget/ping", nil)
	engine.Handler().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "/ping", body["path"])
}

func TestDiscover_Good_MissingDirIsEmpty(t *testing.T) {
	providers, err := provider.Discover(filepath.Join(t.TempDir(), ".core", "providers"))
	require.NoError(t, err)
	assert.Empty(t, providers)
}

func TestDiscover_Bad_InvalidManifest(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".core", "providers")
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.yaml"), []byte(`
name: broken
basePath: /api/broken
`), 0644))

	providers, err := provider.Discover(dir)
	require.Error(t, err)
	assert.Nil(t, providers)
	assert.Contains(t, err.Error(), "upstream is required")
}
