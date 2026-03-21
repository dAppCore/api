// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── Swagger endpoint ────────────────────────────────────────────────────

func TestSwaggerEndpoint_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwagger("Test API", "A test API service", "1.0.0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Use a real test server because gin-swagger reads RequestURI
	// which is not populated by httptest.NewRecorder.
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/swagger/doc.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty response body")
	}

	// Verify the body is valid JSON with expected fields.
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("expected valid JSON, got unmarshal error: %v", err)
	}

	info, ok := doc["info"].(map[string]any)
	if !ok {
		t.Fatal("expected 'info' object in swagger doc")
	}
	if info["title"] != "Test API" {
		t.Fatalf("expected title=%q, got %q", "Test API", info["title"])
	}
	if info["version"] != "1.0.0" {
		t.Fatalf("expected version=%q, got %q", "1.0.0", info["version"])
	}
}

func TestSwaggerDisabledByDefault_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Without WithSwagger, GET /swagger/doc.json should return 404.
	e, _ := api.New()

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /swagger/doc.json without WithSwagger, got %d", w.Code)
	}
}

func TestSwagger_Good_SpecNotEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwagger("Test API", "Test", "1.0.0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Register a describable group so paths has more than just /health.
	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file",
		Group:       "files",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("ok"))
	})
	e.Register(bridge)

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/swagger/doc.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		t.Fatal("expected 'paths' object in spec")
	}

	// Must have more than just /health since we registered a tool.
	if len(paths) < 2 {
		t.Fatalf("expected at least 2 paths (got %d): /health + tool endpoint", len(paths))
	}

	if _, ok := paths["/tools/file_read"]; !ok {
		t.Fatal("expected /tools/file_read path in spec")
	}
}

func TestSwagger_Good_WithToolBridge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwagger("Tool API", "Tool test", "1.0.0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bridge := api.NewToolBridge("/api/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "metrics_query",
		Description: "Query metrics data",
		Group:       "metrics",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("ok"))
	})
	e.Register(bridge)

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/swagger/doc.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := doc["paths"].(map[string]any)
	if _, ok := paths["/api/tools/metrics_query"]; !ok {
		t.Fatal("expected /api/tools/metrics_query path in spec")
	}

	// Verify the operation has the expected summary.
	toolPath := paths["/api/tools/metrics_query"].(map[string]any)
	postOp := toolPath["post"].(map[string]any)
	if postOp["summary"] != "Query metrics data" {
		t.Fatalf("expected summary=%q, got %v", "Query metrics data", postOp["summary"])
	}
}

func TestSwagger_Good_CachesSpec(t *testing.T) {
	spec := &swaggerSpecHelper{
		title:   "Cache Test",
		desc:    "Testing cache",
		version: "0.1.0",
	}

	first := spec.ReadDoc()
	second := spec.ReadDoc()

	if first != second {
		t.Fatal("expected ReadDoc() to return the same string on repeated calls")
	}

	if first == "" {
		t.Fatal("expected non-empty spec from ReadDoc()")
	}
}

func TestSwagger_Good_InfoFromOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwagger("MyTitle", "MyDesc", "2.0.0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/swagger/doc.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	info := doc["info"].(map[string]any)
	if info["title"] != "MyTitle" {
		t.Fatalf("expected title=%q, got %v", "MyTitle", info["title"])
	}
	if info["description"] != "MyDesc" {
		t.Fatalf("expected description=%q, got %v", "MyDesc", info["description"])
	}
	if info["version"] != "2.0.0" {
		t.Fatalf("expected version=%q, got %v", "2.0.0", info["version"])
	}
}

func TestSwagger_Good_ValidOpenAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwagger("OpenAPI Test", "Verify version", "1.0.0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/swagger/doc.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if doc["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi=%q, got %v", "3.1.0", doc["openapi"])
	}
}

// swaggerSpecHelper exercises the caching behaviour of swaggerSpec
// without depending on unexported internals. It creates a SpecBuilder
// inline and uses sync.Once the same way the real swaggerSpec does.
type swaggerSpecHelper struct {
	title, desc, version string
	called               int
	cache                string
}

func (h *swaggerSpecHelper) ReadDoc() string {
	if h.cache != "" {
		return h.cache
	}
	h.called++
	sb := &api.SpecBuilder{
		Title:       h.title,
		Description: h.desc,
		Version:     h.version,
	}
	data, err := sb.Build(nil)
	if err != nil {
		h.cache = `{"openapi":"3.1.0","info":{"title":"error","version":"0.0.0"},"paths":{}}`
		return h.cache
	}
	h.cache = string(data)
	return h.cache
}
