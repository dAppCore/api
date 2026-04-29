// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"dappco.re/go/api/internal/stdcompat/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// skipIfRaceDetector skips the test when the race detector is enabled.
// gin-contrib/timeout@v1.1.0 has a known data race on Context.index
// between the timeout goroutine and the handler goroutine.
func skipIfRaceDetector(t *testing.T) {
	t.Helper()
	if raceDetectorEnabled {
		t.Skip("skipping: gin-contrib/timeout has known data race (upstream bug)")
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────

// slowGroup provides a route that sleeps longer than the test timeout.
type slowGroup struct{}

func (s *slowGroup) Name() string     { return "slow" }
func (s *slowGroup) BasePath() string { return "/v1/slow" }
func (s *slowGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/wait", func(c *gin.Context) {
		time.Sleep(200 * time.Millisecond)
		c.JSON(http.StatusOK, api.OK("done"))
	})
}

// ── WithTimeout ─────────────────────────────────────────────────────────

func TestWithTimeout_Good_FastRequestSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithTimeout(500 * time.Millisecond))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success=true")
	}
	if resp.Data != "pong" {
		t.Fatalf("expected Data=%q, got %q", "pong", resp.Data)
	}
}

func TestWithTimeout_Good_SlowRequestTimesOut(t *testing.T) {
	skipIfRaceDetector(t)
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithTimeout(50 * time.Millisecond))
	e.Register(&slowGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/slow/wait", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", w.Code)
	}
}

func TestWithTimeout_Good_TimeoutResponseEnvelope(t *testing.T) {
	skipIfRaceDetector(t)
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithTimeout(50 * time.Millisecond))
	e.Register(&slowGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/slow/wait", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected Success=false")
	}
	if resp.Error == nil {
		t.Fatal("expected Error to be non-nil")
	}
	if resp.Error.Code != "timeout" {
		t.Fatalf("expected error code=%q, got %q", "timeout", resp.Error.Code)
	}
	if resp.Error.Message != "Request timed out" {
		t.Fatalf("expected error message=%q, got %q", "Request timed out", resp.Error.Message)
	}
}

func TestWithTimeout_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithRequestID(),
		api.WithTimeout(500*time.Millisecond),
	)
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// WithRequestID should still set the header.
	id := w.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}

	var resp api.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data != "pong" {
		t.Fatalf("expected Data=%q, got %q", "pong", resp.Data)
	}
}

func TestWithTimeout_Ugly_ZeroDurationDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithTimeout(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with zero timeout disabled, got %d", w.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data != "pong" {
		t.Fatalf("expected Data=%q, got %q", "pong", resp.Data)
	}
}
