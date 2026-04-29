// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"dappco.re/go/api/internal/stdcompat/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

type attachRequestMetaTestGroup struct {
	handler gin.HandlerFunc
}

func (g attachRequestMetaTestGroup) Name() string     { return "attach-request-meta" }
func (g attachRequestMetaTestGroup) BasePath() string { return "/v1" }
func (g attachRequestMetaTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/meta", g.handler)
}

// ── OK ──────────────────────────────────────────────────────────────────

func TestOK_Good(t *testing.T) {
	r := api.OK("hello")

	if !r.Success {
		t.Fatal("expected Success=true")
	}
	if r.Data != "hello" {
		t.Fatalf("expected Data=%q, got %q", "hello", r.Data)
	}
	if r.Error != nil {
		t.Fatal("expected Error to be nil")
	}
	if r.Meta != nil {
		t.Fatal("expected Meta to be nil")
	}
}

func TestOK_Good_StructData(t *testing.T) {
	type user struct {
		Name string `json:"name"`
	}
	r := api.OK(user{Name: "Ada"})

	if !r.Success {
		t.Fatal("expected Success=true")
	}
	if r.Data.Name != "Ada" {
		t.Fatalf("expected Data.Name=%q, got %q", "Ada", r.Data.Name)
	}
}

func TestOK_Good_JSONOmitsErrorAndMeta(t *testing.T) {
	r := api.OK("data")
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, ok := raw["error"]; ok {
		t.Fatal("expected 'error' field to be omitted from JSON")
	}
	if _, ok := raw["meta"]; ok {
		t.Fatal("expected 'meta' field to be omitted from JSON")
	}
	if _, ok := raw["success"]; !ok {
		t.Fatal("expected 'success' field to be present")
	}
	if _, ok := raw["data"]; !ok {
		t.Fatal("expected 'data' field to be present")
	}
}

// ── Fail ────────────────────────────────────────────────────────────────

func TestFail_Good(t *testing.T) {
	r := api.Fail("NOT_FOUND", "resource not found")

	if r.Success {
		t.Fatal("expected Success=false")
	}
	if r.Error == nil {
		t.Fatal("expected Error to be non-nil")
	}
	if r.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected Code=%q, got %q", "NOT_FOUND", r.Error.Code)
	}
	if r.Error.Message != "resource not found" {
		t.Fatalf("expected Message=%q, got %q", "resource not found", r.Error.Message)
	}
	if r.Error.Details != nil {
		t.Fatal("expected Details to be nil")
	}
}

func TestFail_Good_JSONOmitsData(t *testing.T) {
	r := api.Fail("ERR", "something went wrong")
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, ok := raw["data"]; ok {
		t.Fatal("expected 'data' field to be omitted from JSON")
	}
	if _, ok := raw["error"]; !ok {
		t.Fatal("expected 'error' field to be present")
	}
}

// ── FailWithDetails ─────────────────────────────────────────────────────

func TestFailWithDetails_Good(t *testing.T) {
	details := map[string]string{"field": "email", "reason": "invalid format"}
	r := api.FailWithDetails("VALIDATION", "validation failed", details)

	if r.Success {
		t.Fatal("expected Success=false")
	}
	if r.Error == nil {
		t.Fatal("expected Error to be non-nil")
	}
	if r.Error.Code != "VALIDATION" {
		t.Fatalf("expected Code=%q, got %q", "VALIDATION", r.Error.Code)
	}
	if r.Error.Details == nil {
		t.Fatal("expected Details to be non-nil")
	}
}

func TestFailWithDetails_Good_JSONIncludesDetails(t *testing.T) {
	r := api.FailWithDetails("ERR", "bad", "extra info")
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	errObj, ok := raw["error"].(map[string]any)
	if !ok {
		t.Fatal("expected 'error' to be an object")
	}
	if _, ok := errObj["details"]; !ok {
		t.Fatal("expected 'details' field to be present in error")
	}
}

// ── Paginated ───────────────────────────────────────────────────────────

func TestPaginated_Good(t *testing.T) {
	items := []string{"a", "b", "c"}
	r := api.Paginated(items, 2, 25, 100)

	if !r.Success {
		t.Fatal("expected Success=true")
	}
	if len(r.Data) != 3 {
		t.Fatalf("expected 3 items, got %d", len(r.Data))
	}
	if r.Meta == nil {
		t.Fatal("expected Meta to be non-nil")
	}
	if r.Meta.Page != 2 {
		t.Fatalf("expected Page=2, got %d", r.Meta.Page)
	}
	if r.Meta.PerPage != 25 {
		t.Fatalf("expected PerPage=25, got %d", r.Meta.PerPage)
	}
	if r.Meta.Total != 100 {
		t.Fatalf("expected Total=100, got %d", r.Meta.Total)
	}
}

func TestPaginated_Good_JSONIncludesMeta(t *testing.T) {
	r := api.Paginated([]int{1}, 1, 10, 50)
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, ok := raw["meta"]; !ok {
		t.Fatal("expected 'meta' field to be present")
	}
	meta := raw["meta"].(map[string]any)
	if meta["page"].(float64) != 1 {
		t.Fatalf("expected page=1, got %v", meta["page"])
	}
	if meta["per_page"].(float64) != 10 {
		t.Fatalf("expected per_page=10, got %v", meta["per_page"])
	}
	if meta["total"].(float64) != 50 {
		t.Fatalf("expected total=50, got %v", meta["total"])
	}
}

func TestResponse_AttachRequestMeta_Good_FillsMetaFromRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithRequestID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e.Register(attachRequestMetaTestGroup{
		handler: func(c *gin.Context) {
			resp := api.AttachRequestMeta(c, api.OK("classified"))
			c.JSON(http.StatusOK, resp)
		},
	})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set("X-Request-ID", "client-id-meta")
	e.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
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
	if resp.Meta.Page != 0 || resp.Meta.PerPage != 0 || resp.Meta.Total != 0 {
		t.Fatalf("expected empty pagination metadata when none was provided, got %+v", resp.Meta)
	}
}

func TestResponse_AttachRequestMeta_Bad_ReturnsResponseUnchangedWithoutRequestMeta(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e.Register(attachRequestMetaTestGroup{
		handler: func(c *gin.Context) {
			resp := api.AttachRequestMeta(c, api.OK("plain"))
			c.JSON(http.StatusOK, resp)
		},
	})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/meta", nil)
	e.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Meta != nil {
		t.Fatalf("expected Meta to remain nil, got %+v", resp.Meta)
	}
}

func TestResponse_AttachRequestMeta_Ugly_PreservesExistingMetaFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithRequestID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e.Register(attachRequestMetaTestGroup{
		handler: func(c *gin.Context) {
			resp := api.Paginated("classified", 7, 25, 100)
			resp = api.AttachRequestMeta(c, resp)
			c.JSON(http.StatusOK, resp)
		},
	})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set("X-Request-ID", "client-id-meta")
	e.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
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
	if resp.Meta.Page != 7 || resp.Meta.PerPage != 25 || resp.Meta.Total != 100 {
		t.Fatalf("expected existing pagination metadata to be preserved, got %+v", resp.Meta)
	}
}
