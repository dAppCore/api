// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// ── WithGzip ──────────────────────────────────────────────────────────

func TestWithGzip_Good_CompressesResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithGzip())
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "gzip" {
		t.Fatalf("expected Content-Encoding=%q, got %q", "gzip", ce)
	}
}

func TestWithGzip_Good_NoCompressionWithoutAcceptHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithGzip())
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
	if ce == "gzip" {
		t.Fatal("expected no gzip Content-Encoding when client does not request it")
	}
}

func TestWithGzip_Good_DefaultLevel(t *testing.T) {
	// Calling WithGzip() with no arguments should use default compression
	// and not panic.
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithGzip())
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "gzip" {
		t.Fatalf("expected Content-Encoding=%q with default level, got %q", "gzip", ce)
	}
}

func TestWithGzip_Good_CustomLevel(t *testing.T) {
	// WithGzip(gzip.BestSpeed) should work without panicking and still compress.
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithGzip(gzip.BestSpeed))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "gzip" {
		t.Fatalf("expected Content-Encoding=%q with BestSpeed, got %q", "gzip", ce)
	}
}

func TestWithGzip_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithGzip(),
		api.WithRequestID(),
	)
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Both gzip compression and request ID should be present.
	ce := w.Header().Get("Content-Encoding")
	if ce != "gzip" {
		t.Fatalf("expected Content-Encoding=%q, got %q", "gzip", ce)
	}

	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}
