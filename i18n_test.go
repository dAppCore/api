// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── Helpers ─────────────────────────────────────────────────────────────

// i18nTestGroup provides routes that expose locale detection results.
type i18nTestGroup struct{}

func (i *i18nTestGroup) Name() string     { return "i18n" }
func (i *i18nTestGroup) BasePath() string { return "/i18n" }
func (i *i18nTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/locale", func(c *gin.Context) {
		locale := api.GetLocale(c)
		c.JSON(http.StatusOK, api.OK(map[string]string{"locale": locale}))
	})
	rg.GET("/greeting", func(c *gin.Context) {
		msg, ok := api.GetMessage(c, "greeting")
		c.JSON(http.StatusOK, api.OK(map[string]any{
			"locale":  api.GetLocale(c),
			"message": msg,
			"found":   ok,
		}))
	})
}

// i18nLocaleResponse is the typed response for locale detection tests.
type i18nLocaleResponse struct {
	Success bool              `json:"success"`
	Data    map[string]string `json:"data"`
}

// i18nMessageResponse is the typed response for message lookup tests.
type i18nMessageResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Locale  string `json:"locale"`
		Message string `json:"message"`
		Found   bool   `json:"found"`
	} `json:"data"`
}

// ── Tests ───────────────────────────────────────────────────────────────

func TestWithI18n_Good_DetectsLocaleFromHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithI18n(api.I18nConfig{
		Supported: []string{"en", "fr", "de"},
	}))
	e.Register(&i18nTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/i18n/locale", nil)
	req.Header.Set("Accept-Language", "fr")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp i18nLocaleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data["locale"] != "fr" {
		t.Fatalf("expected locale=%q, got %q", "fr", resp.Data["locale"])
	}
}

func TestWithI18n_Good_FallsBackToDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithI18n(api.I18nConfig{
		DefaultLocale: "en",
		Supported:     []string{"en", "fr"},
	}))
	e.Register(&i18nTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/i18n/locale", nil)
	// No Accept-Language header.
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp i18nLocaleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data["locale"] != "en" {
		t.Fatalf("expected locale=%q, got %q", "en", resp.Data["locale"])
	}
}

func TestWithI18n_Good_QualityWeighting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithI18n(api.I18nConfig{
		Supported: []string{"en", "fr", "de"},
	}))
	e.Register(&i18nTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/i18n/locale", nil)
	// French has higher quality weight than German.
	req.Header.Set("Accept-Language", "de;q=0.5, fr;q=0.9")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp i18nLocaleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data["locale"] != "fr" {
		t.Fatalf("expected locale=%q, got %q", "fr", resp.Data["locale"])
	}
}

func TestWithI18n_Good_PreservesMatchedLocaleTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithI18n(api.I18nConfig{
		DefaultLocale: "en",
		Supported:     []string{"en", "fr", "fr-CA"},
	}))
	e.Register(&i18nTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/i18n/locale", nil)
	req.Header.Set("Accept-Language", "fr-CA, fr;q=0.8")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp i18nLocaleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data["locale"] != "fr-CA" {
		t.Fatalf("expected locale=%q, got %q", "fr-CA", resp.Data["locale"])
	}
}

func TestWithI18n_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithI18n(api.I18nConfig{
			Supported: []string{"en", "fr"},
		}),
		api.WithRequestID(),
	)
	e.Register(&i18nTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/i18n/locale", nil)
	req.Header.Set("Accept-Language", "fr")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// i18n middleware should detect French.
	var resp i18nLocaleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data["locale"] != "fr" {
		t.Fatalf("expected locale=%q, got %q", "fr", resp.Data["locale"])
	}

	// RequestID middleware should also have run.
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}

func TestWithI18n_Good_LooksUpMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithI18n(api.I18nConfig{
		DefaultLocale: "en",
		Supported:     []string{"en", "fr"},
		Messages: map[string]map[string]string{
			"en": {"greeting": "Hello"},
			"fr": {"greeting": "Bonjour"},
		},
	}))
	e.Register(&i18nTestGroup{})

	h := e.Handler()

	// Test French message lookup.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/i18n/greeting", nil)
	req.Header.Set("Accept-Language", "fr")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp i18nMessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data.Locale != "fr" {
		t.Fatalf("expected locale=%q, got %q", "fr", resp.Data.Locale)
	}
	if resp.Data.Message != "Bonjour" {
		t.Fatalf("expected message=%q, got %q", "Bonjour", resp.Data.Message)
	}
	if !resp.Data.Found {
		t.Fatal("expected found=true")
	}

	// Test English message lookup.
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/i18n/greeting", nil)
	req.Header.Set("Accept-Language", "en")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var respEn i18nMessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &respEn); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if respEn.Data.Message != "Hello" {
		t.Fatalf("expected message=%q, got %q", "Hello", respEn.Data.Message)
	}
}
