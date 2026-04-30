// SPDX-License-Identifier: EUPL-1.2

package provider_test

import (
	"net/http"
	"net/http/httptest"

	. "dappco.re/go"
	"dappco.re/go/api"
	"dappco.re/go/api/pkg/provider"
)

func TestDiscover_Good_LoadsYAMLProxyProvider(t *T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		coreJSONEncode(w, map[string]string{`path`: r.URL.Path})
	}))
	defer upstream.Close()

	dir := PathJoin(t.TempDir(), ".core", "providers")
	RequireNoError(t, coreMkdirAll(dir, 0755))
	specPath := PathJoin(PathDir(dir), "specs", "openapi.yaml")
	RequireNoError(t, coreMkdirAll(PathDir(specPath), 0755))
	RequireNoError(t, coreWriteFile(specPath, []byte("openapi: 3.1.0\n"), 0644))
	RequireNoError(t, coreWriteFile(PathJoin(dir, "cool.yaml"), []byte(`
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
	RequireNoError(t, err)
	AssertLen(t, providers, 1)

	p := providers[0]
	AssertEqual(t, "cool-widget", p.Name())
	AssertEqual(t, "/api/v1/cool-widget", p.BasePath())

	specProvider, ok := p.(interface{ SpecFile() string })
	RequireTrue(t, ok)
	canonicalSpecPath, err := corePathEvalSymlinks(specPath)
	RequireNoError(t, err)
	AssertEqual(t, canonicalSpecPath, specProvider.SpecFile())

	renderable, ok := p.(provider.Renderable)
	RequireTrue(t, ok)
	AssertEqual(t, "core-cool-widget", renderable.Element().Tag)

	engine, err := api.New()
	RequireNoError(t, err)
	engine.Register(p)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cool-widget/ping", nil)
	engine.Handler().ServeHTTP(w, req)

	AssertEqual(t, http.StatusOK, w.Code)
	var body map[string]string
	RequireNoError(t, coreJSONUnmarshal(w.Body.Bytes(), &body))
	AssertEqual(t, "/ping", body[`path`])
}

func TestDiscover_Good_MissingDirIsEmpty(t *T) {
	providers, err := provider.Discover(PathJoin(t.TempDir(), ".core", "providers"))
	RequireNoError(t, err)
	AssertEmpty(t, providers)
}

func TestDiscover_Good_LoadsYAMLProvidersFromCleanDir(t *T) {
	dir := PathJoin(t.TempDir(), ".core", "providers")
	RequireNoError(t, coreMkdirAll(dir, 0755))
	upstream := newDiscoveryUpstream(t)

	writeProviderManifest(t, dir, "alpha", upstream)
	writeProviderManifest(t, dir, "beta", upstream)

	providers, err := provider.Discover(dir)
	RequireNoError(t, err)
	AssertLen(t, providers, 2)

	names := []string{providers[0].Name(), providers[1].Name()}
	AssertElementsMatch(t, []string{"alpha", "beta"}, names)
}

func TestDiscover_Good_DirWithDotDotSegmentResolves(t *T) {
	root := t.TempDir()
	dir := PathJoin(root, "providers")
	RequireNoError(t, coreMkdirAll(dir, 0755))
	writeProviderManifest(t, dir, "dotdot", newDiscoveryUpstream(t))

	providers, err := provider.Discover(PathJoin(root, "other", "..", "providers"))
	RequireNoError(t, err)
	AssertLen(t, providers, 1)
	AssertEqual(t, "dotdot", providers[0].Name())
}

func TestDiscover_Bad_InvalidManifest(t *T) {
	dir := PathJoin(t.TempDir(), ".core", "providers")
	RequireNoError(t, coreMkdirAll(dir, 0755))
	RequireNoError(t, coreWriteFile(PathJoin(dir, "broken.yaml"), []byte(`
name: broken
basePath: /api/broken
`), 0644))

	providers, err := provider.Discover(dir)
	AssertError(t, err)
	AssertNil(t, providers)
	AssertContains(t, err.Error(), "upstream is required")
}

func TestDiscover_Bad_SymlinkedDirRefused(t *T) {
	root := t.TempDir()
	realDir := PathJoin(root, "real-providers")
	linkDir := PathJoin(root, "providers")
	RequireNoError(t, coreMkdirAll(realDir, 0755))
	if err := coreSymlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	providers, err := provider.Discover(linkDir)
	AssertError(t, err)
	AssertNil(t, providers)
	AssertContains(t, err.Error(), "symlinked provider directory rejected")
}

func TestDiscover_Bad_SymlinkManifestOutsideDirRefused(t *T) {
	root := t.TempDir()
	dir := PathJoin(root, "providers")
	RequireNoError(t, coreMkdirAll(dir, 0755))
	outside := PathJoin(root, "outside.yaml")
	RequireNoError(t, coreWriteFile(outside, []byte("not: loaded\n"), 0644))
	if err := coreSymlink(outside, PathJoin(dir, "leak.yaml")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	providers, err := provider.Discover(dir)
	AssertError(t, err)
	AssertNil(t, providers)
	AssertContains(t, err.Error(), "symlinked provider manifest rejected")
}

func TestDiscover_Bad_SymlinkManifestWithinDirRefused(t *T) {
	dir := PathJoin(t.TempDir(), "providers")
	RequireNoError(t, coreMkdirAll(dir, 0755))
	realManifest := writeProviderManifest(t, dir, "real", newDiscoveryUpstream(t))
	if err := coreSymlink(realManifest, PathJoin(dir, "alias.yaml")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	providers, err := provider.Discover(dir)
	AssertError(t, err)
	AssertNil(t, providers)
	AssertContains(t, err.Error(), "symlinked provider manifest rejected")
}

func newDiscoveryUpstream(t *T) string {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)
	return upstream.URL
}

func writeProviderManifest(t *T, dir, name, upstream string) string {
	t.Helper()
	path := PathJoin(dir, name+".yaml")
	RequireNoError(t, coreWriteFile(path, []byte(`
name: `+name+`
basePath: /api/`+name+`
upstream: `+upstream+`
`), 0644))
	return path
}
