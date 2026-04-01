// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

type sunsetStubGroup struct{}

func (sunsetStubGroup) Name() string     { return "legacy" }
func (sunsetStubGroup) BasePath() string { return "/legacy" }
func (sunsetStubGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("ok"))
	})
}

type sunsetLinkStubGroup struct{}

func (sunsetLinkStubGroup) Name() string     { return "legacy-link" }
func (sunsetLinkStubGroup) BasePath() string { return "/legacy-link" }
func (sunsetLinkStubGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", func(c *gin.Context) {
		c.Header("Link", "<https://example.com/docs>; rel=\"help\"")
		c.JSON(http.StatusOK, api.OK("ok"))
	})
}

func TestWithSunset_Good_AddsDeprecationHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSunset("2025-06-01", "/api/v2/status"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e.Register(sunsetStubGroup{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/legacy/status", nil)
	e.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Deprecation"); got != "true" {
		t.Fatalf("expected Deprecation=true, got %q", got)
	}
	if got := w.Header().Get("Sunset"); got != "Sun, 01 Jun 2025 00:00:00 GMT" {
		t.Fatalf("expected formatted Sunset header, got %q", got)
	}
	if got := w.Header().Get("Link"); got != "</api/v2/status>; rel=\"successor-version\"" {
		t.Fatalf("expected successor Link header, got %q", got)
	}
	if got := w.Header().Get("X-API-Warn"); got != "This endpoint is deprecated and will be removed on 2025-06-01." {
		t.Fatalf("expected deprecation warning, got %q", got)
	}
}

func TestWithSunset_Good_PreservesExistingLinkHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSunset("2025-06-01", "/api/v2/status"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e.Register(sunsetLinkStubGroup{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/legacy-link/status", nil)
	e.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	links := w.Header().Values("Link")
	if len(links) != 2 {
		t.Fatalf("expected 2 Link header values, got %v", links)
	}
	if links[0] != "<https://example.com/docs>; rel=\"help\"" {
		t.Fatalf("expected existing Link header to be preserved first, got %q", links[0])
	}
	if links[1] != "</api/v2/status>; rel=\"successor-version\"" {
		t.Fatalf("expected successor Link header to be appended, got %q", links[1])
	}
}
