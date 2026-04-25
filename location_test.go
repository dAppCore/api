// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/location/v2"
	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// ── Helpers ─────────────────────────────────────────────────────────────

// locationTestGroup exposes a route that returns the detected location.
type locationTestGroup struct{}

func (l *locationTestGroup) Name() string     { return "loc" }
func (l *locationTestGroup) BasePath() string { return "/loc" }
func (l *locationTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/info", func(c *gin.Context) {
		url := location.Get(c)
		c.JSON(http.StatusOK, api.OK(map[string]string{
			"scheme": url.Scheme,
			"host":   url.Host,
		}))
	})
}

// locationResponse is the typed response envelope for location info tests.
type locationResponse struct {
	Success bool              `json:"success"`
	Data    map[string]string `json:"data"`
}

// ── WithLocation ────────────────────────────────────────────────────────

func TestWithLocation_Good_DetectsForwardedHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithLocation())
	e.Register(&locationTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/loc/info", nil)
	req.Header.Set("X-Forwarded-Host", "api.example.com")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp locationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data["host"] != "api.example.com" {
		t.Fatalf("expected host=%q, got %q", "api.example.com", resp.Data["host"])
	}
}

func TestWithLocation_Good_DetectsForwardedProto(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithLocation())
	e.Register(&locationTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/loc/info", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp locationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data["scheme"] != "https" {
		t.Fatalf("expected scheme=%q, got %q", "https", resp.Data["scheme"])
	}
}

func TestWithLocation_Good_FallsBackToRequestHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithLocation())
	e.Register(&locationTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/loc/info", nil)
	// No X-Forwarded-* headers — middleware should fall back to defaults.
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp locationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	// Without forwarded headers the middleware falls back to its default
	// scheme ("http"). The host will be either the request Host header
	// value or the configured default; either way it must not be empty.
	if resp.Data["scheme"] != "http" {
		t.Fatalf("expected fallback scheme=%q, got %q", "http", resp.Data["scheme"])
	}
	if resp.Data["host"] == "" {
		t.Fatal("expected a non-empty host in fallback mode")
	}
}

func TestWithLocation_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithLocation(),
		api.WithRequestID(),
	)
	e.Register(&locationTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/loc/info", nil)
	req.Header.Set("X-Forwarded-Host", "proxy.example.com")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Location middleware should populate the detected host.
	var resp locationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data["host"] != "proxy.example.com" {
		t.Fatalf("expected host=%q, got %q", "proxy.example.com", resp.Data["host"])
	}

	// RequestID middleware should also have run.
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}

func TestWithLocation_Good_BothHeadersCombined(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithLocation())
	e.Register(&locationTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/loc/info", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "secure.example.com")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp locationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data["scheme"] != "https" {
		t.Fatalf("expected scheme=%q, got %q", "https", resp.Data["scheme"])
	}
	if resp.Data["host"] != "secure.example.com" {
		t.Fatalf("expected host=%q, got %q", "secure.example.com", resp.Data["host"])
	}
}
