// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── Test helpers ────────────────────────────────────────────────────────

// healthGroup is a minimal RouteGroup for testing Engine integration.
type healthGroup struct{}

func (h *healthGroup) Name() string     { return "health-extra" }
func (h *healthGroup) BasePath() string { return "/v1" }
func (h *healthGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("echo"))
	})
}

type panicGroup struct{}

func (p *panicGroup) Name() string     { return "panic" }
func (p *panicGroup) BasePath() string { return "/panic" }
func (p *panicGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/boom", func(c *gin.Context) {
		panic("boom")
	})
}

// ── New ─────────────────────────────────────────────────────────────────

func TestNew_Good(t *testing.T) {
	e, err := api.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil Engine")
	}
}

func TestNew_Good_WithAddr(t *testing.T) {
	e, err := api.New(api.WithAddr(":9090"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Addr() != ":9090" {
		t.Fatalf("expected addr=%q, got %q", ":9090", e.Addr())
	}
}

// ── Default address ─────────────────────────────────────────────────────

func TestAddr_Good_Default(t *testing.T) {
	e, _ := api.New()
	if e.Addr() != ":8080" {
		t.Fatalf("expected default addr=%q, got %q", ":8080", e.Addr())
	}
}

// ── Register + Groups ───────────────────────────────────────────────────

func TestRegister_Good(t *testing.T) {
	e, _ := api.New()
	e.Register(&healthGroup{})

	groups := e.Groups()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Name() != "health-extra" {
		t.Fatalf("expected group name=%q, got %q", "health-extra", groups[0].Name())
	}
}

func TestRegister_Good_MultipleGroups(t *testing.T) {
	e, _ := api.New()
	e.Register(&healthGroup{})
	e.Register(&stubGroup{})

	if len(e.Groups()) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(e.Groups()))
	}
}

// ── Handler ─────────────────────────────────────────────────────────────

func TestHandler_Good_HealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New()

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
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
	if resp.Data != "healthy" {
		t.Fatalf("expected Data=%q, got %q", "healthy", resp.Data)
	}
}

func TestHandler_Good_RegisteredRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New()
	e.Register(&healthGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/echo", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data != "echo" {
		t.Fatalf("expected Data=%q, got %q", "echo", resp.Data)
	}
}

func TestHandler_Bad_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New()

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/nonexistent", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_Bad_PanicReturnsEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRequestID())
	e.Register(&panicGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/panic/boom", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
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
	if resp.Error.Code != "internal_server_error" {
		t.Fatalf("expected error code=%q, got %q", "internal_server_error", resp.Error.Code)
	}
	if resp.Error.Message != "Internal server error" {
		t.Fatalf("expected error message=%q, got %q", "Internal server error", resp.Error.Message)
	}
	if got := w.Header().Get("X-Request-ID"); got == "" {
		t.Fatal("expected X-Request-ID header to survive panic recovery")
	}
}

// ── Serve + graceful shutdown ───────────────────────────────────────────

func TestServe_Good_GracefulShutdown(t *testing.T) {
	// Pick a random free port to avoid conflicts.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	e, _ := api.New(api.WithAddr(addr))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() {
		errCh <- e.Serve(ctx)
	}()

	// Wait for server to be ready.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Verify the server responds.
	resp, err := http.Get("http://" + addr + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Cancel context to trigger graceful shutdown.
	cancel()

	select {
	case serveErr := <-errCh:
		if serveErr != nil {
			t.Fatalf("Serve returned unexpected error: %v", serveErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return within 5 seconds after context cancellation")
	}
}

func TestServe_Bad_ReturnsListenErrorBeforeCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve port: %v", err)
	}
	addr := ln.Addr().String()
	defer ln.Close()

	e, _ := api.New(api.WithAddr(addr))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- e.Serve(ctx)
	}()

	select {
	case serveErr := <-errCh:
		if serveErr == nil {
			t.Fatal("expected Serve to return a listen error, got nil")
		}
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("Serve did not return promptly after listener failure")
	}
}
