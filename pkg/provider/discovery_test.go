// SPDX-License-Identifier: EUPL-1.2

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
	canonicalSpecPath, err := filepath.EvalSymlinks(specPath)
	require.NoError(t, err)
	assert.Equal(t, canonicalSpecPath, specProvider.SpecFile())

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

func TestDiscover_Good_LoadsYAMLProvidersFromCleanDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".core", "providers")
	require.NoError(t, os.MkdirAll(dir, 0755))
	upstream := newDiscoveryUpstream(t)

	writeProviderManifest(t, dir, "alpha", upstream)
	writeProviderManifest(t, dir, "beta", upstream)

	providers, err := provider.Discover(dir)
	require.NoError(t, err)
	require.Len(t, providers, 2)

	names := []string{providers[0].Name(), providers[1].Name()}
	assert.ElementsMatch(t, []string{"alpha", "beta"}, names)
}

func TestDiscover_Good_DirWithDotDotSegmentResolves(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "providers")
	require.NoError(t, os.MkdirAll(dir, 0755))
	writeProviderManifest(t, dir, "dotdot", newDiscoveryUpstream(t))

	providers, err := provider.Discover(filepath.Join(root, "other", "..", "providers"))
	require.NoError(t, err)
	require.Len(t, providers, 1)
	assert.Equal(t, "dotdot", providers[0].Name())
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

func TestDiscover_Bad_SymlinkedDirRefused(t *testing.T) {
	root := t.TempDir()
	realDir := filepath.Join(root, "real-providers")
	linkDir := filepath.Join(root, "providers")
	require.NoError(t, os.MkdirAll(realDir, 0755))
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	providers, err := provider.Discover(linkDir)
	require.Error(t, err)
	assert.Nil(t, providers)
	assert.Contains(t, err.Error(), "symlinked provider directory rejected")
}

func TestDiscover_Bad_SymlinkManifestOutsideDirRefused(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "providers")
	require.NoError(t, os.MkdirAll(dir, 0755))
	outside := filepath.Join(root, "outside.yaml")
	require.NoError(t, os.WriteFile(outside, []byte("not: loaded\n"), 0644))
	if err := os.Symlink(outside, filepath.Join(dir, "leak.yaml")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	providers, err := provider.Discover(dir)
	require.Error(t, err)
	assert.Nil(t, providers)
	assert.Contains(t, err.Error(), "symlinked provider manifest rejected")
}

func TestDiscover_Bad_SymlinkManifestWithinDirRefused(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "providers")
	require.NoError(t, os.MkdirAll(dir, 0755))
	realManifest := writeProviderManifest(t, dir, "real", newDiscoveryUpstream(t))
	if err := os.Symlink(realManifest, filepath.Join(dir, "alias.yaml")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	providers, err := provider.Discover(dir)
	require.Error(t, err)
	assert.Nil(t, providers)
	assert.Contains(t, err.Error(), "symlinked provider manifest rejected")
}

func newDiscoveryUpstream(t *testing.T) string {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)
	return upstream.URL
}

func writeProviderManifest(t *testing.T, dir, name, upstream string) string {
	t.Helper()
	path := filepath.Join(dir, name+".yaml")
	require.NoError(t, os.WriteFile(path, []byte(`
name: `+name+`
basePath: /api/`+name+`
upstream: `+upstream+`
`), 0644))
	return path
}
