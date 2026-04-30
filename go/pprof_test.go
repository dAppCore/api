// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// ── Pprof profiling endpoints ─────────────────────────────────────────

func TestWithPprof_Good_IndexAccessible(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithPprof())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/pprof/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /debug/pprof/, got %d", resp.StatusCode)
	}
}

func TestWithPprof_Good_ProfileEndpointExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithPprof())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/pprof/heap")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /debug/pprof/heap, got %d", resp.StatusCode)
	}
}

func TestWithPprof_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithRequestID(), api.WithPprof())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/pprof/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /debug/pprof/ with middleware, got %d", resp.StatusCode)
	}

	// Verify the request ID middleware is still active.
	rid := resp.Header.Get("X-Request-ID")
	if rid == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID middleware")
	}
}

func TestWithPprof_Bad_NotMountedWithoutOption(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, _ := api.New()

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /debug/pprof/ without WithPprof, got %d", w.Code)
	}
}

func TestWithPprof_Good_CmdlineEndpointExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithPprof())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/pprof/cmdline")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /debug/pprof/cmdline, got %d", resp.StatusCode)
	}
}
