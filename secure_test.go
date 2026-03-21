// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── WithSecure ──────────────────────────────────────────────────────────

func TestWithSecure_Good_SetsHSTSHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithSecure())

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	sts := w.Header().Get("Strict-Transport-Security")
	if sts == "" {
		t.Fatal("expected Strict-Transport-Security header to be set")
	}
	if !strings.Contains(sts, "max-age=31536000") {
		t.Fatalf("expected max-age=31536000 in STS header, got %q", sts)
	}
	if !strings.Contains(strings.ToLower(sts), "includesubdomains") {
		t.Fatalf("expected includeSubdomains in STS header, got %q", sts)
	}
}

func TestWithSecure_Good_SetsFrameOptionsDeny(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithSecure())

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	xfo := w.Header().Get("X-Frame-Options")
	if xfo != "DENY" {
		t.Fatalf("expected X-Frame-Options=%q, got %q", "DENY", xfo)
	}
}

func TestWithSecure_Good_SetsContentTypeNosniff(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithSecure())

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	cto := w.Header().Get("X-Content-Type-Options")
	if cto != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options=%q, got %q", "nosniff", cto)
	}
}

func TestWithSecure_Good_SetsReferrerPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithSecure())

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	rp := w.Header().Get("Referrer-Policy")
	if rp != "strict-origin-when-cross-origin" {
		t.Fatalf("expected Referrer-Policy=%q, got %q", "strict-origin-when-cross-origin", rp)
	}
}

func TestWithSecure_Good_AllHeadersPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithSecure())
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify all security headers are present on a regular route.
	checks := map[string]string{
		"X-Frame-Options":        "DENY",
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, want := range checks {
		got := w.Header().Get(header)
		if got != want {
			t.Errorf("header %s: expected %q, got %q", header, want, got)
		}
	}

	sts := w.Header().Get("Strict-Transport-Security")
	if sts == "" {
		t.Error("expected Strict-Transport-Security header to be set")
	}
}

func TestWithSecure_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithSecure(),
		api.WithRequestID(),
	)

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Both secure headers and request ID should be present.
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("expected X-Frame-Options header from WithSecure")
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}

func TestWithSecure_Bad_NoSSLRedirect(t *testing.T) {
	// SSL redirect is not enabled — the middleware runs behind a TLS-terminating
	// reverse proxy. Verify plain HTTP requests are not redirected.
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithSecure())

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	// Should get 200, not a 301/302 redirect.
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (no SSL redirect), got %d", w.Code)
	}
}

func TestWithSecure_Ugly_DoubleSecureDoesNotPanic(t *testing.T) {
	// Applying WithSecure twice should not panic or cause issues.
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithSecure(),
		api.WithSecure(),
	)

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Headers should still be correctly set.
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("expected X-Frame-Options=DENY after double WithSecure")
	}
}
