// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	strings "dappco.re/go/api/internal/stdcompat/corestrings"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// ── Expvar runtime metrics endpoint ─────────────────────────────────

func TestWithExpvar_Good_EndpointReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithExpvar())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/vars")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /debug/vars, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("expected application/json content type, got %q", ct)
	}
}

func TestWithExpvar_Good_ContainsMemstats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithExpvar())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/vars")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if !strings.Contains(string(body), "memstats") {
		t.Fatal("expected response body to contain \"memstats\"")
	}
}

func TestWithExpvar_Good_ContainsCmdline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithExpvar())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/vars")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if !strings.Contains(string(body), "cmdline") {
		t.Fatal("expected response body to contain \"cmdline\"")
	}
}

func TestWithExpvar_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithRequestID(), api.WithExpvar())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/vars")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /debug/vars with middleware, got %d", resp.StatusCode)
	}

	// Verify the request ID middleware is still active.
	rid := resp.Header.Get("X-Request-ID")
	if rid == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID middleware")
	}
}

func TestWithExpvar_Bad_NotMountedWithoutOption(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, _ := api.New()

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/debug/vars", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /debug/vars without WithExpvar, got %d", w.Code)
	}
}
