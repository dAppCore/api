// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// cacheCounterGroup registers routes that increment a counter on each call,
// allowing tests to distinguish cached from uncached responses.
type cacheCounterGroup struct {
	counter atomic.Int64
}

func (g *cacheCounterGroup) Name() string     { return "cache-test" }
func (g *cacheCounterGroup) BasePath() string { return "/cache" }
func (g *cacheCounterGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/counter", func(c *gin.Context) {
		n := g.counter.Add(1)
		c.JSON(http.StatusOK, api.OK(fmt.Sprintf("call-%d", n)))
	})
	rg.GET("/other", func(c *gin.Context) {
		n := g.counter.Add(1)
		c.JSON(http.StatusOK, api.OK(fmt.Sprintf("other-%d", n)))
	})
	rg.POST("/counter", func(c *gin.Context) {
		n := g.counter.Add(1)
		c.JSON(http.StatusOK, api.OK(fmt.Sprintf("post-%d", n)))
	})
}

// ── WithCache ───────────────────────────────────────────────────────────

func TestWithCache_Good_CachesGETResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	grp := &cacheCounterGroup{}
	e, _ := api.New(api.WithCache(5 * time.Second))
	e.Register(grp)

	h := e.Handler()

	// First request — cache MISS.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	h.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	body1 := w1.Body.String()
	if !strings.Contains(body1, "call-1") {
		t.Fatalf("expected body to contain %q, got %q", "call-1", body1)
	}

	// Second request — should be a cache HIT returning the same body.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	body2 := w2.Body.String()
	if body1 != body2 {
		t.Fatalf("expected cached body %q, got %q", body1, body2)
	}

	cacheHeader := w2.Header().Get("X-Cache")
	if cacheHeader != "HIT" {
		t.Fatalf("expected X-Cache=HIT, got %q", cacheHeader)
	}

	// Counter should still be 1 (handler was not called again).
	if grp.counter.Load() != 1 {
		t.Fatalf("expected counter=1 (cached), got %d", grp.counter.Load())
	}
}

func TestWithCache_Good_POSTNotCached(t *testing.T) {
	gin.SetMode(gin.TestMode)
	grp := &cacheCounterGroup{}
	e, _ := api.New(api.WithCache(5 * time.Second))
	e.Register(grp)

	h := e.Handler()

	// First POST request.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/cache/counter", nil)
	h.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	var resp1 api.Response[string]
	if err := json.Unmarshal(w1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp1.Data != "post-1" {
		t.Fatalf("expected Data=%q, got %q", "post-1", resp1.Data)
	}

	// Second POST request — should NOT be cached, counter increments.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/cache/counter", nil)
	h.ServeHTTP(w2, req2)

	var resp2 api.Response[string]
	if err := json.Unmarshal(w2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp2.Data != "post-2" {
		t.Fatalf("expected Data=%q, got %q", "post-2", resp2.Data)
	}

	// Counter should be 2 — both POST requests hit the handler.
	if grp.counter.Load() != 2 {
		t.Fatalf("expected counter=2, got %d", grp.counter.Load())
	}
}

func TestWithCache_Good_DifferentPathsSeparatelyCached(t *testing.T) {
	gin.SetMode(gin.TestMode)
	grp := &cacheCounterGroup{}
	e, _ := api.New(api.WithCache(5 * time.Second))
	e.Register(grp)

	h := e.Handler()

	// Request to /cache/counter.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	h.ServeHTTP(w1, req1)

	body1 := w1.Body.String()
	if !strings.Contains(body1, "call-1") {
		t.Fatalf("expected body to contain %q, got %q", "call-1", body1)
	}

	// Request to /cache/other — different path, should miss cache.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/cache/other", nil)
	h.ServeHTTP(w2, req2)

	body2 := w2.Body.String()
	if !strings.Contains(body2, "other-2") {
		t.Fatalf("expected body to contain %q, got %q", "other-2", body2)
	}

	// Counter is 2 — both paths hit the handler.
	if grp.counter.Load() != 2 {
		t.Fatalf("expected counter=2, got %d", grp.counter.Load())
	}

	// Re-request /cache/counter — should serve cached "call-1".
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	h.ServeHTTP(w3, req3)

	body3 := w3.Body.String()
	if body1 != body3 {
		t.Fatalf("expected cached body %q, got %q", body1, body3)
	}

	// Counter unchanged — served from cache.
	if grp.counter.Load() != 2 {
		t.Fatalf("expected counter=2 (cached), got %d", grp.counter.Load())
	}
}

func TestWithCache_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	grp := &cacheCounterGroup{}
	e, _ := api.New(
		api.WithRequestID(),
		api.WithCache(5*time.Second),
	)
	e.Register(grp)

	h := e.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// RequestID middleware should still set X-Request-ID.
	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}

	// Body should contain the expected response.
	body := w.Body.String()
	if !strings.Contains(body, "call-1") {
		t.Fatalf("expected body to contain %q, got %q", "call-1", body)
	}
}

func TestWithCache_Good_ExpiredCacheMisses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	grp := &cacheCounterGroup{}
	e, _ := api.New(api.WithCache(50 * time.Millisecond))
	e.Register(grp)

	h := e.Handler()

	// First request — populates cache.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	h.ServeHTTP(w1, req1)

	body1 := w1.Body.String()
	if !strings.Contains(body1, "call-1") {
		t.Fatalf("expected body to contain %q, got %q", "call-1", body1)
	}

	// Wait for cache to expire.
	time.Sleep(100 * time.Millisecond)

	// Second request — cache expired, handler called again.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	h.ServeHTTP(w2, req2)

	body2 := w2.Body.String()
	if !strings.Contains(body2, "call-2") {
		t.Fatalf("expected body to contain %q after expiry, got %q", "call-2", body2)
	}

	// Counter should be 2 — both requests hit the handler.
	if grp.counter.Load() != 2 {
		t.Fatalf("expected counter=2, got %d", grp.counter.Load())
	}
}

func TestWithCache_Good_EvictsWhenCapacityReached(t *testing.T) {
	gin.SetMode(gin.TestMode)
	grp := &cacheCounterGroup{}
	e, _ := api.New(api.WithCache(5*time.Second, 1))
	e.Register(grp)

	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	h.ServeHTTP(w1, req1)
	if !strings.Contains(w1.Body.String(), "call-1") {
		t.Fatalf("expected first response to contain %q, got %q", "call-1", w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/cache/other", nil)
	h.ServeHTTP(w2, req2)
	if !strings.Contains(w2.Body.String(), "other-2") {
		t.Fatalf("expected second response to contain %q, got %q", "other-2", w2.Body.String())
	}

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	h.ServeHTTP(w3, req3)
	if !strings.Contains(w3.Body.String(), "call-3") {
		t.Fatalf("expected evicted response to contain %q, got %q", "call-3", w3.Body.String())
	}

	if grp.counter.Load() != 3 {
		t.Fatalf("expected counter=3 after eviction, got %d", grp.counter.Load())
	}
}
