// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
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

type sunsetHeaderStubGroup struct{}

func (sunsetHeaderStubGroup) Name() string     { return "legacy-headers" }
func (sunsetHeaderStubGroup) BasePath() string { return "/legacy-headers" }
func (sunsetHeaderStubGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", func(c *gin.Context) {
		c.Header("Deprecation", "false")
		c.Header("Sunset", "Wed, 01 Jan 2025 00:00:00 GMT")
		c.Header("X-API-Warn", "Existing warning")
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
	if got := w.Header().Get("API-Suggested-Replacement"); got != "/api/v2/status" {
		t.Fatalf("expected API-Suggested-Replacement to mirror replacement URL, got %q", got)
	}
	if got := w.Header().Get("X-API-Warn"); got != "This endpoint is deprecated and will be removed on 2025-06-01." {
		t.Fatalf("expected deprecation warning, got %q", got)
	}
}

func TestApiSunsetWith_Good_FormatsCommonSunsetDateForms(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		sunsetDate string
		want       string
	}{
		{
			name:       "rfc3339",
			sunsetDate: "2026-04-30T23:59:59Z",
			want:       "Thu, 30 Apr 2026 23:59:59 GMT",
		},
		{
			name:       "rfc7231",
			sunsetDate: "Thu, 30 Apr 2026 23:59:59 GMT",
			want:       "Thu, 30 Apr 2026 23:59:59 GMT",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mw := api.ApiSunsetWith(tc.sunsetDate, "POST /api/v2/billing/invoices")

			r := gin.New()
			r.Use(mw)
			r.GET("/billing", func(c *gin.Context) { c.JSON(http.StatusOK, api.OK("ok")) })

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/billing", nil)
			r.ServeHTTP(w, req)

			if got := w.Header().Get("Sunset"); got != tc.want {
				t.Fatalf("expected Sunset=%q, got %q", tc.want, got)
			}
			if got := w.Header().Get("Link"); got != "</api/v2/billing/invoices>; rel=\"successor-version\"" {
				t.Fatalf("expected successor Link header, got %q", got)
			}
			if got := w.Header().Get("API-Suggested-Replacement"); got != "POST /api/v2/billing/invoices" {
				t.Fatalf("expected API-Suggested-Replacement to preserve the original replacement, got %q", got)
			}
		})
	}
}

func TestApiSunsetWith_Good_StripsMethodFromSuccessorLink(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw := api.ApiSunsetWith("2026-04-30", "PATCH /api/v2/billing/invoices")

	r := gin.New()
	r.Use(mw)
	r.GET("/billing", func(c *gin.Context) { c.JSON(http.StatusOK, api.OK("ok")) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/billing", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Link"); got != "</api/v2/billing/invoices>; rel=\"successor-version\"" {
		t.Fatalf("expected method prefix to be stripped from successor Link, got %q", got)
	}
	if got := w.Header().Get("API-Suggested-Replacement"); got != "PATCH /api/v2/billing/invoices" {
		t.Fatalf("expected API-Suggested-Replacement to preserve the full replacement, got %q", got)
	}
}

func TestApiSunsetWith_Good_PreservesRawSuccessorTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw := api.ApiSunsetWith("2026-04-30", "/api/v2/billing/invoices")

	r := gin.New()
	r.Use(mw)
	r.GET("/billing", func(c *gin.Context) { c.JSON(http.StatusOK, api.OK("ok")) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/billing", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Link"); got != "</api/v2/billing/invoices>; rel=\"successor-version\"" {
		t.Fatalf("expected raw replacement path to be preserved in Link header, got %q", got)
	}
	if got := w.Header().Get("API-Suggested-Replacement"); got != "/api/v2/billing/invoices" {
		t.Fatalf("expected API-Suggested-Replacement to preserve the raw replacement, got %q", got)
	}
}

func TestApiSunsetWith_Ugly_PreservesUnknownMethodPrefixAsRawTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw := api.ApiSunsetWith("2026-04-30", "PURGE /api/v2/billing/invoices")

	r := gin.New()
	r.Use(mw)
	r.GET("/billing", func(c *gin.Context) { c.JSON(http.StatusOK, api.OK("ok")) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/billing", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Link"); got != "<PURGE /api/v2/billing/invoices>; rel=\"successor-version\"" {
		t.Fatalf("expected unknown method prefix to be preserved, got %q", got)
	}
	if got := w.Header().Get("API-Suggested-Replacement"); got != "PURGE /api/v2/billing/invoices" {
		t.Fatalf("expected API-Suggested-Replacement to preserve the raw replacement, got %q", got)
	}
}

func TestApiSunsetWith_Ugly_PreservesBareReplacementToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw := api.ApiSunsetWith("2026-04-30", "POST")

	r := gin.New()
	r.Use(mw)
	r.GET("/billing", func(c *gin.Context) { c.JSON(http.StatusOK, api.OK("ok")) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/billing", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Link"); got != "<POST>; rel=\"successor-version\"" {
		t.Fatalf("expected bare replacement token to be preserved, got %q", got)
	}
	if got := w.Header().Get("API-Suggested-Replacement"); got != "POST" {
		t.Fatalf("expected API-Suggested-Replacement to preserve the bare token, got %q", got)
	}
}

// TestApiSunsetWith_Good_AddsNoticeURLHeader exercises ApiSunsetWith with the
// WithSunsetNoticeURL option to verify the spec §8 notice header is emitted.
func TestApiSunsetWith_Good_AddsNoticeURLHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw := api.ApiSunsetWith(
		"2026-04-30",
		"POST /api/v2/billing/invoices",
		api.WithSunsetNoticeURL("https://docs.api.dappco.re/deprecation/billing"),
	)

	r := gin.New()
	r.Use(mw)
	r.GET("/billing", func(c *gin.Context) { c.JSON(http.StatusOK, api.OK("ok")) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/billing", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("API-Deprecation-Notice-URL"); got != "https://docs.api.dappco.re/deprecation/billing" {
		t.Fatalf("expected API-Deprecation-Notice-URL header, got %q", got)
	}
	if got := w.Header().Get("API-Suggested-Replacement"); got != "POST /api/v2/billing/invoices" {
		t.Fatalf("expected API-Suggested-Replacement to mirror replacement, got %q", got)
	}
}

// TestApiSunsetWith_Bad_OmitsEmptyOptionalHeaders ensures empty option values
// do not emit blank headers, keeping the response surface clean.
func TestApiSunsetWith_Bad_OmitsEmptyOptionalHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw := api.ApiSunsetWith("", "", api.WithSunsetNoticeURL("   "))

	r := gin.New()
	r.Use(mw)
	r.GET("/x", func(c *gin.Context) { c.JSON(http.StatusOK, api.OK("ok")) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Sunset"); got != "" {
		t.Fatalf("expected no Sunset header for empty date, got %q", got)
	}
	if got := w.Header().Get("Link"); got != "" {
		t.Fatalf("expected no Link header for empty replacement, got %q", got)
	}
	if got := w.Header().Get("API-Suggested-Replacement"); got != "" {
		t.Fatalf("expected no API-Suggested-Replacement for empty replacement, got %q", got)
	}
	if got := w.Header().Get("API-Deprecation-Notice-URL"); got != "" {
		t.Fatalf("expected no API-Deprecation-Notice-URL for blank URL, got %q", got)
	}
	if got := w.Header().Get("Deprecation"); got != "true" {
		t.Fatalf("expected Deprecation=true even with no metadata, got %q", got)
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

func TestWithSunset_Good_PreservesExistingDeprecationHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSunset("2025-06-01", "/api/v2/status"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e.Register(sunsetHeaderStubGroup{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/legacy-headers/status", nil)
	e.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if got := w.Header().Values("Deprecation"); len(got) != 2 {
		t.Fatalf("expected 2 Deprecation header values, got %v", got)
	}
	if got := w.Header().Values("Sunset"); len(got) != 2 {
		t.Fatalf("expected 2 Sunset header values, got %v", got)
	}
	if got := w.Header().Values("X-API-Warn"); len(got) != 2 {
		t.Fatalf("expected 2 X-API-Warn header values, got %v", got)
	}
	if got := w.Header().Values("Link"); len(got) != 2 {
		t.Fatalf("expected 2 Link header values, got %v", got)
	}
}

func TestApiSunsetWith_Ugly_PreservesInvalidSunsetValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw := api.ApiSunsetWith("not-a-date", "")

	r := gin.New()
	r.Use(mw)
	r.GET("/broken", func(c *gin.Context) { c.JSON(http.StatusOK, api.OK("ok")) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/broken", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Sunset"); got != "not-a-date" {
		t.Fatalf("expected invalid sunset value to be preserved, got %q", got)
	}
	if got := w.Header().Get("Deprecation"); got != "true" {
		t.Fatalf("expected Deprecation=true even for invalid date, got %q", got)
	}
}
