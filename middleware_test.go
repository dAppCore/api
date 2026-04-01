// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── Helpers ─────────────────────────────────────────────────────────────

// mwTestGroup provides a simple /v1/secret endpoint for middleware tests.
type mwTestGroup struct{}

func (m *mwTestGroup) Name() string     { return "mw-test" }
func (m *mwTestGroup) BasePath() string { return "/v1" }
func (m *mwTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/secret", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("classified"))
	})
}

type swaggerLikeGroup struct{}

func (g *swaggerLikeGroup) Name() string     { return "swagger-like" }
func (g *swaggerLikeGroup) BasePath() string { return "/swaggerx" }
func (g *swaggerLikeGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/secret", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("classified"))
	})
}

type requestIDTestGroup struct {
	gotID *string
}

func (g requestIDTestGroup) Name() string     { return "request-id" }
func (g requestIDTestGroup) BasePath() string { return "/v1" }
func (g requestIDTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/secret", func(c *gin.Context) {
		*g.gotID = api.GetRequestID(c)
		c.JSON(http.StatusOK, api.OK("classified"))
	})
}

type requestMetaTestGroup struct{}

func (g requestMetaTestGroup) Name() string     { return "request-meta" }
func (g requestMetaTestGroup) BasePath() string { return "/v1" }
func (g requestMetaTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/meta", func(c *gin.Context) {
		time.Sleep(2 * time.Millisecond)
		resp := api.AttachRequestMeta(c, api.Paginated("classified", 1, 25, 100))
		c.JSON(http.StatusOK, resp)
	})
}

type autoResponseMetaTestGroup struct{}

func (g autoResponseMetaTestGroup) Name() string     { return "auto-response-meta" }
func (g autoResponseMetaTestGroup) BasePath() string { return "/v1" }
func (g autoResponseMetaTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/meta", func(c *gin.Context) {
		time.Sleep(2 * time.Millisecond)
		c.JSON(http.StatusOK, api.Paginated("classified", 1, 25, 100))
	})
}

// ── Bearer auth ─────────────────────────────────────────────────────────

func TestBearerAuth_Bad_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBearerAuth("s3cret"))
	e.Register(&mwTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/secret", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != "unauthorised" {
		t.Fatalf("expected error code=%q, got %+v", "unauthorised", resp.Error)
	}
}

func TestBearerAuth_Bad_WrongToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBearerAuth("s3cret"))
	e.Register(&mwTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/secret", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != "unauthorised" {
		t.Fatalf("expected error code=%q, got %+v", "unauthorised", resp.Error)
	}
}

func TestBearerAuth_Good_CorrectToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBearerAuth("s3cret"))
	e.Register(&mwTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/secret", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data != "classified" {
		t.Fatalf("expected Data=%q, got %q", "classified", resp.Data)
	}
}

func TestBearerAuth_Good_HealthBypassesAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBearerAuth("s3cret"))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	// No Authorization header.
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /health, got %d", w.Code)
	}
}

func TestBearerAuth_Bad_SimilarPrefixDoesNotBypassAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBearerAuth("s3cret"))
	e.Register(&swaggerLikeGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/swaggerx/secret", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for /swaggerx/secret, got %d", w.Code)
	}
}

// ── Request ID ──────────────────────────────────────────────────────────

func TestRequestID_Good_GeneratedWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRequestID())

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
	// 16 bytes = 32 hex characters.
	if len(id) != 32 {
		t.Fatalf("expected 32-char hex ID, got %d chars: %q", len(id), id)
	}
}

func TestRequestID_Good_PreservesClientID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRequestID())

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "client-id-abc")
	h.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-ID")
	if id != "client-id-abc" {
		t.Fatalf("expected X-Request-ID=%q, got %q", "client-id-abc", id)
	}
}

func TestRequestID_Good_ContextAccessor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRequestID())

	var gotID string
	e.Register(requestIDTestGroup{gotID: &gotID})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/secret", nil)
	req.Header.Set("X-Request-ID", "client-id-xyz")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if gotID == "" {
		t.Fatal("expected GetRequestID to return the request ID inside the handler")
	}
	if gotID != "client-id-xyz" {
		t.Fatalf("expected GetRequestID=%q, got %q", "client-id-xyz", gotID)
	}
}

func TestRequestID_Good_RequestMetaHelper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRequestID())
	e.Register(requestMetaTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set("X-Request-ID", "client-id-meta")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Meta == nil {
		t.Fatal("expected Meta to be present")
	}
	if resp.Meta.RequestID != "client-id-meta" {
		t.Fatalf("expected request_id=%q, got %q", "client-id-meta", resp.Meta.RequestID)
	}
	if resp.Meta.Duration == "" {
		t.Fatal("expected duration to be populated")
	}
	if resp.Meta.Page != 1 || resp.Meta.PerPage != 25 || resp.Meta.Total != 100 {
		t.Fatalf("expected pagination metadata to be preserved, got %+v", resp.Meta)
	}
}

func TestResponseMeta_Good_AttachesMetaAutomatically(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithRequestID(),
		api.WithResponseMeta(),
	)
	e.Register(autoResponseMetaTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set("X-Request-ID", "client-id-auto-meta")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Meta == nil {
		t.Fatal("expected Meta to be present")
	}
	if resp.Meta.RequestID != "client-id-auto-meta" {
		t.Fatalf("expected request_id=%q, got %q", "client-id-auto-meta", resp.Meta.RequestID)
	}
	if resp.Meta.Duration == "" {
		t.Fatal("expected duration to be populated")
	}
	if resp.Meta.Page != 1 || resp.Meta.PerPage != 25 || resp.Meta.Total != 100 {
		t.Fatalf("expected pagination metadata to be preserved, got %+v", resp.Meta)
	}
	if got := w.Header().Get("X-Request-ID"); got != "client-id-auto-meta" {
		t.Fatalf("expected response header X-Request-ID=%q, got %q", "client-id-auto-meta", got)
	}
}

// ── CORS ────────────────────────────────────────────────────────────────

func TestCORS_Good_PreflightAllOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithCORS("*"))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("expected 200 or 204 for preflight, got %d", w.Code)
	}

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin=%q, got %q", "*", origin)
	}

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Fatal("expected Access-Control-Allow-Methods to be set")
	}

	headers := w.Header().Get("Access-Control-Allow-Headers")
	if headers == "" {
		t.Fatal("expected Access-Control-Allow-Headers to be set")
	}
}

func TestCORS_Good_SpecificOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithCORS("https://app.example.com"))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	h.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://app.example.com" {
		t.Fatalf("expected origin=%q, got %q", "https://app.example.com", origin)
	}
}

func TestCORS_Bad_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithCORS("https://allowed.example.com"))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	h.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "" {
		t.Fatalf("expected no Access-Control-Allow-Origin for disallowed origin, got %q", origin)
	}
}
