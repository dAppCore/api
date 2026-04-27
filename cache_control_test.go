// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type cacheControlStubGroup struct {
	descs []RouteDescription
}

func (g *cacheControlStubGroup) Name() string                       { return "cache-control" }
func (g *cacheControlStubGroup) BasePath() string                   { return "/v1" }
func (g *cacheControlStubGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (g *cacheControlStubGroup) Describe() []RouteDescription       { return g.descs }

func TestCacheControl_cacheControlPolicies_Good_MapsDescribedRoutes(t *testing.T) {
	group := &cacheControlStubGroup{
		descs: []RouteDescription{
			{
				Method:       http.MethodGet,
				Path:         "/items/{id}",
				CacheControl: "public, max-age=60",
				Summary:      "Fetch an item",
			},
		},
	}

	policies := cacheControlPolicies([]RouteGroup{group})
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(policies))
	}

	if got := policies["GET /v1/items/:id"]; got != "public, max-age=60" {
		t.Fatalf("expected policy for converted Gin path, got %q", got)
	}
}

func TestCacheControl_cacheControlPolicies_Bad_ReturnsNilWithoutPolicies(t *testing.T) {
	group := &cacheControlStubGroup{
		descs: []RouteDescription{
			{},
			{Method: http.MethodGet},
			{CacheControl: "public, max-age=60"},
		},
	}

	if got := cacheControlPolicies([]RouteGroup{group}); got != nil {
		t.Fatalf("expected nil policies when every description is incomplete, got %v", got)
	}
}

func TestCacheControl_cacheControlPolicies_Ugly_SkipsMalformedDescriptions(t *testing.T) {
	group := &cacheControlStubGroup{
		descs: []RouteDescription{
			{
				Method:       "   ",
				Path:         "/items/{id}",
				CacheControl: "public, max-age=60",
			},
			{
				Method:       http.MethodGet,
				Path:         "/items/{id}",
				CacheControl: "   ",
			},
		},
	}

	if got := cacheControlPolicies([]RouteGroup{group}); got != nil {
		t.Fatalf("expected malformed descriptions to be skipped, got %v", got)
	}
}

func TestCacheControl_openAPIPathToGinPath_Good_ConvertsParameters(t *testing.T) {
	cases := map[string]string{
		"/items/{id}":           "/items/:id",
		"/workspaces/{ws}/jobs": "/workspaces/:ws/jobs",
		"/":                     "/",
	}

	for input, want := range cases {
		if got := openAPIPathToGinPath(input); got != want {
			t.Fatalf("openAPIPathToGinPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCacheControl_openAPIPathToGinPath_Bad_PreservesInvalidParameterTokens(t *testing.T) {
	cases := map[string]string{
		"/items/{id/extra}":   "/items/{id/extra}",
		"/items/{}/details":   "/items/{}/details",
		"/items/{id}{suffix}": "/items/{id}{suffix}",
	}

	for input, want := range cases {
		if got := openAPIPathToGinPath(input); got != want {
			t.Fatalf("openAPIPathToGinPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCacheControl_openAPIPathToGinPath_Ugly_NormalisesWhitespaceAndRoot(t *testing.T) {
	if got := openAPIPathToGinPath("   "); got != "/" {
		t.Fatalf("expected whitespace-only path to normalize to /, got %q", got)
	}
	if got := openAPIPathToGinPath(""); got != "/" {
		t.Fatalf("expected empty path to normalize to /, got %q", got)
	}
}

func TestCacheControl_cacheControlMiddleware_Good_AddsHeaderOnSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(cacheControlMiddleware(map[string]string{
		"GET /v1/items/:id": "public, max-age=60",
	}))
	r.GET("/v1/items/:id", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/items/123", nil)
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=60" {
		t.Fatalf("expected Cache-Control header to be set, got %q", got)
	}
}

func TestCacheControl_cacheControlMiddleware_Bad_SkipsNonSuccessResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(cacheControlMiddleware(map[string]string{
		"GET /v1/items/:id": "public, max-age=60",
	}))
	r.GET("/v1/items/:id", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/items/123", nil)
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("expected no Cache-Control header on non-success responses, got %q", got)
	}
}

func TestCacheControl_cacheControlMiddleware_Ugly_PreservesExplicitHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(cacheControlMiddleware(map[string]string{
		"GET /v1/items/:id": "public, max-age=60",
	}))
	r.GET("/v1/items/:id", func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-store")
		c.String(http.StatusOK, "ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/items/123", nil)
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("expected explicit Cache-Control header to be preserved, got %q", got)
	}
}

func TestCacheControl_cacheControlMiddleware_Ugly_NoPoliciesIsNoop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(cacheControlMiddleware(nil))
	r.GET("/v1/items/:id", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/items/123", nil)
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("expected no Cache-Control header with empty policies, got %q", got)
	}
}

func TestCacheControl_cacheControlMiddleware_Ugly_SkipsWhenRouteTemplateMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(cacheControlMiddleware(map[string]string{
		"GET /v1/items/:id": "public, max-age=60",
	}))
	r.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("expected no Cache-Control header when FullPath is empty, got %q", got)
	}
}
