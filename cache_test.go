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

type cacheSizedGroup struct {
	counter atomic.Int64
}

func (g *cacheSizedGroup) Name() string     { return "cache-sized" }
func (g *cacheSizedGroup) BasePath() string { return "/cache" }
func (g *cacheSizedGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/small", func(c *gin.Context) {
		n := g.counter.Add(1)
		c.JSON(http.StatusOK, api.OK(fmt.Sprintf("small-%d-%s", n, strings.Repeat("a", 96))))
	})
	rg.GET("/large", func(c *gin.Context) {
		n := g.counter.Add(1)
		c.JSON(http.StatusOK, api.OK(fmt.Sprintf("large-%d-%s", n, strings.Repeat("b", 96))))
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

func TestWithCache_Good_PreservesCurrentRequestIDOnHit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	grp := &cacheCounterGroup{}
	e, _ := api.New(
		api.WithRequestID(),
		api.WithCache(5*time.Second),
	)
	e.Register(grp)

	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	req1.Header.Set("X-Request-ID", "first-request-id")
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}
	if got := w1.Header().Get("X-Request-ID"); got != "first-request-id" {
		t.Fatalf("expected first response request ID %q, got %q", "first-request-id", got)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
	req2.Header.Set("X-Request-ID", "second-request-id")
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	if got := w2.Header().Get("X-Request-ID"); got != "second-request-id" {
		t.Fatalf("expected cached response to preserve current request ID %q, got %q", "second-request-id", got)
	}
	if got := w2.Header().Get("X-Cache"); got != "HIT" {
		t.Fatalf("expected X-Cache=HIT, got %q", got)
	}

	var resp2 api.Response[string]
	if err := json.Unmarshal(w2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp2.Data != "call-1" {
		t.Fatalf("expected cached response data %q, got %q", "call-1", resp2.Data)
	}
	if resp2.Meta == nil {
		t.Fatal("expected cached response meta to be attached")
	}
	if resp2.Meta.RequestID != "second-request-id" {
		t.Fatalf("expected cached response request_id=%q, got %q", "second-request-id", resp2.Meta.RequestID)
	}
	if resp2.Meta.Duration == "" {
		t.Fatal("expected cached response duration to be refreshed")
	}
}

func TestWithCache_Good_PreservesCurrentRequestMetaOnHit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithRequestID(),
		api.WithCache(5*time.Second),
	)
	e.Register(requestMetaTestGroup{})

	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/v1/meta", nil)
	req1.Header.Set("X-Request-ID", "first-request-id")
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	var resp1 api.Response[string]
	if err := json.Unmarshal(w1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp1.Meta == nil {
		t.Fatal("expected meta on first response")
	}
	if resp1.Meta.RequestID != "first-request-id" {
		t.Fatalf("expected first response request_id=%q, got %q", "first-request-id", resp1.Meta.RequestID)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/v1/meta", nil)
	req2.Header.Set("X-Request-ID", "second-request-id")
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	var resp2 api.Response[string]
	if err := json.Unmarshal(w2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp2.Meta == nil {
		t.Fatal("expected meta on cached response")
	}
	if resp2.Meta.RequestID != "second-request-id" {
		t.Fatalf("expected cached response request_id=%q, got %q", "second-request-id", resp2.Meta.RequestID)
	}
	if resp2.Meta.Duration == "" {
		t.Fatal("expected cached response duration to be refreshed")
	}
	if resp2.Meta.Page != 1 || resp2.Meta.PerPage != 25 || resp2.Meta.Total != 100 {
		t.Fatalf("expected pagination metadata to remain intact, got %+v", resp2.Meta)
	}
	if got := w2.Header().Get("X-Request-ID"); got != "second-request-id" {
		t.Fatalf("expected response header X-Request-ID=%q, got %q", "second-request-id", got)
	}
}

type cacheHeaderGroup struct{}

func (cacheHeaderGroup) Name() string     { return "cache-headers" }
func (cacheHeaderGroup) BasePath() string { return "/cache" }
func (cacheHeaderGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/multi", func(c *gin.Context) {
		c.Writer.Header().Add("Link", "</next?page=2>; rel=\"next\"")
		c.Writer.Header().Add("Link", "</prev?page=0>; rel=\"prev\"")
		c.JSON(http.StatusOK, api.OK("cached"))
	})
}

func TestWithCache_Good_PreservesMultiValueHeadersOnHit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithCache(5 * time.Second))
	e.Register(cacheHeaderGroup{})

	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/cache/multi", nil)
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/cache/multi", nil)
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on cache hit, got %d", w2.Code)
	}

	linkHeaders := w2.Header().Values("Link")
	if len(linkHeaders) != 2 {
		t.Fatalf("expected 2 Link headers on cache hit, got %v", linkHeaders)
	}
	if linkHeaders[0] != "</next?page=2>; rel=\"next\"" {
		t.Fatalf("expected first Link header to be preserved, got %q", linkHeaders[0])
	}
	if linkHeaders[1] != "</prev?page=0>; rel=\"prev\"" {
		t.Fatalf("expected second Link header to be preserved, got %q", linkHeaders[1])
	}
}

func TestWithCache_Ugly_NonPositiveTTLDisablesMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	grp := &cacheCounterGroup{}
	e, _ := api.New(api.WithCache(0))
	e.Register(grp)

	h := e.Handler()

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/cache/counter", nil)
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected request %d to succeed with disabled cache, got %d", i+1, w.Code)
		}
		if got := w.Header().Get("X-Cache"); got != "" {
			t.Fatalf("expected no X-Cache header with disabled cache, got %q", got)
		}
	}

	if grp.counter.Load() != 2 {
		t.Fatalf("expected counter=2 with disabled cache, got %d", grp.counter.Load())
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

func TestWithCache_Good_EvictsWhenSizeLimitReached(t *testing.T) {
	gin.SetMode(gin.TestMode)
	grp := &cacheSizedGroup{}
	e, _ := api.New(api.WithCacheLimits(5*time.Second, 10, 250))
	e.Register(grp)

	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/cache/small", nil)
	h.ServeHTTP(w1, req1)
	if !strings.Contains(w1.Body.String(), "small-1") {
		t.Fatalf("expected first response to contain %q, got %q", "small-1", w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/cache/large", nil)
	h.ServeHTTP(w2, req2)
	if !strings.Contains(w2.Body.String(), "large-2") {
		t.Fatalf("expected second response to contain %q, got %q", "large-2", w2.Body.String())
	}

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/cache/small", nil)
	h.ServeHTTP(w3, req3)
	if !strings.Contains(w3.Body.String(), "small-3") {
		t.Fatalf("expected size-limited cache to evict the oldest entry, got %q", w3.Body.String())
	}

	if got := w3.Header().Get("X-Cache"); got != "" {
		t.Fatalf("expected re-executed response to miss the cache, got X-Cache=%q", got)
	}

	if grp.counter.Load() != 3 {
		t.Fatalf("expected counter=3 after size-based eviction, got %d", grp.counter.Load())
	}
}
