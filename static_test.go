// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"dappco.re/go/api/internal/stdcompat/filepath"
	"dappco.re/go/api/internal/stdcompat/os"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// ── WithStatic ──────────────────────────────────────────────────────────

func TestWithStatic_Good_ServesFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	e, _ := api.New(api.WithStatic("/assets", dir))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/assets/hello.txt", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "hello world" {
		t.Fatalf("expected body=%q, got %q", "hello world", body)
	}
}

func TestWithStatic_Good_Returns404ForMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()

	e, _ := api.New(api.WithStatic("/assets", dir))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/assets/nonexistent.txt", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestWithStatic_Good_ServesIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<h1>Welcome</h1>"), 0644); err != nil {
		t.Fatalf("failed to write index.html: %v", err)
	}

	e, _ := api.New(api.WithStatic("/docs", dir))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/docs/", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "<h1>Welcome</h1>" {
		t.Fatalf("expected body=%q, got %q", "<h1>Welcome</h1>", body)
	}
}

func TestWithStatic_Good_CombinesWithRouteGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log('ok')"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	e, _ := api.New(api.WithStatic("/static", dir))
	e.Register(&stubGroup{})

	h := e.Handler()

	// Static file should be served.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/static/app.js", nil)
	h.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("static: expected 200, got %d", w1.Code)
	}
	if w1.Body.String() != "console.log('ok')" {
		t.Fatalf("static: unexpected body %q", w1.Body.String())
	}

	// API route should also work.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("api: expected 200, got %d", w2.Code)
	}
}

func TestWithStatic_Good_MultipleStaticDirs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir1 := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir1, "sdk.zip"), []byte("sdk-data"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	dir2 := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir2, "style.css"), []byte("body{}"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	e, _ := api.New(
		api.WithStatic("/downloads", dir1),
		api.WithStatic("/css", dir2),
	)

	h := e.Handler()

	// First static directory.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/downloads/sdk.zip", nil)
	h.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("downloads: expected 200, got %d", w1.Code)
	}
	if w1.Body.String() != "sdk-data" {
		t.Fatalf("downloads: expected body=%q, got %q", "sdk-data", w1.Body.String())
	}

	// Second static directory.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/css/style.css", nil)
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("css: expected 200, got %d", w2.Code)
	}
	if w2.Body.String() != "body{}" {
		t.Fatalf("css: expected body=%q, got %q", "body{}", w2.Body.String())
	}
}
