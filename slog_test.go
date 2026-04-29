// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	bytes "dappco.re/go/api/internal/stdcompat/corebytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// ── WithSlog ──────────────────────────────────────────────────────────

func TestWithSlog_Good_LogsRequestFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	e, _ := api.New(api.WithSlog(logger))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected slog output, got empty string")
	}

	// The structured log should contain request fields.
	for _, field := range []string{"status", "method", `path`, "latency", "ip"} {
		if !bytes.Contains(buf.Bytes(), []byte(field)) {
			t.Errorf("expected log output to contain field %q, got: %s", field, output)
		}
	}
}

func TestWithSlog_Good_NilLoggerUsesDefault(t *testing.T) {
	// Passing nil should not panic; it uses slog.Default().
	gin.SetMode(gin.TestMode)

	e, _ := api.New(api.WithSlog(nil))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestWithSlog_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	e, _ := api.New(
		api.WithSlog(logger),
		api.WithRequestID(),
	)

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Both slog output and request ID header should be present.
	if buf.Len() == 0 {
		t.Fatal("expected slog output from WithSlog")
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}

func TestWithSlog_Good_Logs404Status(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	e, _ := api.New(api.WithSlog(logger))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/nonexistent", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected slog output for 404 request")
	}

	// Should contain the 404 status.
	if !bytes.Contains(buf.Bytes(), []byte("404")) {
		t.Errorf("expected log to contain status 404, got: %s", output)
	}
}

func TestWithSlog_Bad_LogsMethodAndPath(t *testing.T) {
	// Verifies POST method and custom path appear in log output.
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	e, _ := api.New(api.WithSlog(logger))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("POST")) {
		t.Errorf("expected log to contain method POST, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("/stub/ping")) {
		t.Errorf("expected log to contain path /stub/ping, got: %s", output)
	}
}

func TestWithSlog_Ugly_DoubleSlogDoesNotPanic(t *testing.T) {
	// Applying WithSlog twice should not panic.
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	e, _ := api.New(
		api.WithSlog(logger),
		api.WithSlog(logger),
	)

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
