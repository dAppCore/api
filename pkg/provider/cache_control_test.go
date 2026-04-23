// SPDX-License-Identifier: EUPL-1.2

package provider_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"dappco.re/go/core/api"
	"dappco.re/go/core/api/pkg/provider"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type cacheControlProvider struct {
	basePath             string
	withDescriptions     bool
	overrideCacheControl string
}

func (p *cacheControlProvider) Name() string     { return "cache-control" }
func (p *cacheControlProvider) BasePath() string { return p.basePath }

func (p *cacheControlProvider) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/items/:id", func(c *gin.Context) {
		if p.overrideCacheControl != "" {
			c.Header("Cache-Control", p.overrideCacheControl)
		}
		c.String(http.StatusOK, "ok")
	})
	rg.POST("/sessions", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})
}

func (p *cacheControlProvider) Describe() []api.RouteDescription {
	if !p.withDescriptions {
		return nil
	}

	return []api.RouteDescription{
		{
			Method:       http.MethodGet,
			Path:         "/items/{id}",
			Summary:      "Fetch an item",
			CacheControl: "public, max-age=300",
		},
		{
			Method:       http.MethodPost,
			Path:         "/sessions",
			Summary:      "Create a session",
			StatusCode:   http.StatusCreated,
			CacheControl: "no-store",
		},
	}
}

type undescribedCacheControlProvider struct {
	basePath string
}

func (p *undescribedCacheControlProvider) Name() string     { return "plain-cache-control" }
func (p *undescribedCacheControlProvider) BasePath() string { return p.basePath }

func (p *undescribedCacheControlProvider) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/items/:id", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
}

func mountProviderHandler(providers ...provider.Provider) http.Handler {
	reg := provider.NewRegistry()
	for _, p := range providers {
		reg.Add(p)
	}

	engine, err := api.New()
	if err != nil {
		panic(err)
	}
	reg.MountAll(engine)
	return engine.Handler()
}

func TestCacheControl_MountAll_Good_AppliesDescribedPolicies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := mountProviderHandler(&cacheControlProvider{
		basePath:         "/api/cache",
		withDescriptions: true,
	})

	getRec := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/api/cache/items/123", nil)
	handler.ServeHTTP(getRec, getReq)
	require.Equal(t, "public, max-age=300", getRec.Header().Get("Cache-Control"))

	postRec := httptest.NewRecorder()
	postReq := httptest.NewRequest(http.MethodPost, "/api/cache/sessions", nil)
	handler.ServeHTTP(postRec, postReq)
	require.Equal(t, "no-store", postRec.Header().Get("Cache-Control"))
}

func TestCacheControl_MountAll_Bad_SkipsProvidersWithoutDescriptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := mountProviderHandler(&undescribedCacheControlProvider{
		basePath: "/api/plain",
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/plain/items/123", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, "", rec.Header().Get("Cache-Control"))
}

func TestCacheControl_MountAll_Ugly_PreservesExplicitHandlerHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := mountProviderHandler(&cacheControlProvider{
		basePath:             "/api/override",
		withDescriptions:     true,
		overrideCacheControl: "private, no-store",
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/override/items/123", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, "private, no-store", rec.Header().Get("Cache-Control"))
}
