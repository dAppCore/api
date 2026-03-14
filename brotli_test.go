// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "forge.lthn.ai/core/api"
)

// ── WithBrotli ────────────────────────────────────────────────────────

func TestWithBrotli_Good_CompressesResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBrotli())
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "br" {
		t.Fatalf("expected Content-Encoding=%q, got %q", "br", ce)
	}
}

func TestWithBrotli_Good_NoCompressionWithoutAcceptHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBrotli())
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	// Deliberately not setting Accept-Encoding header.
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce == "br" {
		t.Fatal("expected no br Content-Encoding when client does not request it")
	}
}

func TestWithBrotli_Good_DefaultLevel(t *testing.T) {
	// Calling WithBrotli() with no arguments should use default compression
	// and not panic.
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBrotli())
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "br" {
		t.Fatalf("expected Content-Encoding=%q with default level, got %q", "br", ce)
	}
}

func TestWithBrotli_Good_CustomLevel(t *testing.T) {
	// WithBrotli(BrotliBestSpeed) should work without panicking and still compress.
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBrotli(api.BrotliBestSpeed))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "br" {
		t.Fatalf("expected Content-Encoding=%q with BestSpeed, got %q", "br", ce)
	}
}

func TestWithBrotli_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithBrotli(),
		api.WithRequestID(),
	)
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Both brotli compression and request ID should be present.
	ce := w.Header().Get("Content-Encoding")
	if ce != "br" {
		t.Fatalf("expected Content-Encoding=%q, got %q", "br", ce)
	}

	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}
