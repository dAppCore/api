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
