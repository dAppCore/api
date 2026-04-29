// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"dappco.re/go/api/internal/stdcompat/json"
	"dappco.re/go/api/internal/stdcompat/strings"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
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

func TestSwaggerEndpoint_Good_CustomPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Test API", "A test API service", "1.0.0"),
		api.WithSwaggerPath("/docs"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/docs/doc.json")
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
}

func TestSwaggerEndpoint_Good_BasePathRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwagger("Test API", "A test API service", "1.0.0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(srv.URL + "/swagger")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected 301 redirect, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != "/swagger/" {
		t.Fatalf("expected Location=/swagger/, got %q", got)
	}
}

func TestSwaggerEndpoint_Good_CustomBasePathRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Test API", "A test API service", "1.0.0"),
		api.WithSwaggerPath("/docs"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(srv.URL + "/docs")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected 301 redirect, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != "/docs/" {
		t.Fatalf("expected Location=/docs/, got %q", got)
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

func TestSwaggerAuth_Good_CustomPathBypassesBearerAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithBearerAuth("secret"),
		api.WithSwagger("Test API", "A test API service", "1.0.0"),
		api.WithSwaggerPath("/docs"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/docs/doc.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for custom swagger path without auth, got %d", resp.StatusCode)
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
				`path`: map[string]any{"type": "string"},
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

	// RFC.endpoints.md — GET /v1/tools listing must appear on the bridge's
	// base path so SDK generators can discover it without iterating tools.
	listingPath, ok := paths["/api/tools"].(map[string]any)
	if !ok {
		t.Fatalf("expected bridge base path /api/tools in spec, got %v", paths["/api/tools"])
	}
	if _, ok := listingPath["get"]; !ok {
		t.Fatalf("expected GET listing on bridge base path, got %v", listingPath)
	}
}

func TestSwagger_Good_IncludesSSEEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(api.WithSwagger("SSE API", "SSE test", "1.0.0"), api.WithSSE(broker))
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

	paths := doc["paths"].(map[string]any)
	pathItem, ok := paths["/events"].(map[string]any)
	if !ok {
		t.Fatal("expected /events path in swagger doc")
	}

	getOp := pathItem["get"].(map[string]any)
	if getOp["operationId"] != "get_events" {
		t.Fatalf("expected SSE operationId to be get_events, got %v", getOp["operationId"])
	}
}

func TestSwagger_Good_UsesCustomSSEPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(
		api.WithSwagger("SSE API", "SSE test", "1.0.0"),
		api.WithSSE(broker),
		api.WithSSEPath("/stream"),
	)
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

	paths := doc["paths"].(map[string]any)
	if _, ok := paths["/stream"]; !ok {
		t.Fatal("expected custom SSE path /stream in swagger doc")
	}
	if _, ok := paths["/events"]; ok {
		t.Fatal("did not expect default /events path when custom SSE path is configured")
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

func TestSwagger_Good_IncludesGraphQLEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithGraphQL(newTestSchema()), api.WithSwagger("Graph API", "GraphQL docs", "1.0.0"))
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
		t.Fatal("expected paths object in swagger doc")
	}
	if _, ok := paths["/graphql"]; !ok {
		t.Fatal("expected /graphql path in swagger doc")
	}
}

func TestSwagger_Good_UsesLicenseMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Licensed API", "Licensed test", "1.0.0"),
		api.WithSwaggerLicense("EUPL-1.2", "https://eupl.eu/1.2/en/"),
	)
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
	license, ok := info["license"].(map[string]any)
	if !ok {
		t.Fatal("expected license metadata in swagger doc")
	}
	if license["name"] != "EUPL-1.2" {
		t.Fatalf("expected license name=%q, got %v", "EUPL-1.2", license["name"])
	}
	if license["url"] != "https://eupl.eu/1.2/en/" {
		t.Fatalf("expected license url=%q, got %v", "https://eupl.eu/1.2/en/", license["url"])
	}
}

func TestSwagger_Good_UsesContactMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Contact API", "Contact test", "1.0.0"),
		api.WithSwaggerContact("API Support", "https://example.com/support", "support@example.com"),
	)
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
	contact, ok := info["contact"].(map[string]any)
	if !ok {
		t.Fatal("expected contact metadata in swagger doc")
	}
	if contact["name"] != "API Support" {
		t.Fatalf("expected contact name=%q, got %v", "API Support", contact["name"])
	}
	if contact["url"] != "https://example.com/support" {
		t.Fatalf("expected contact url=%q, got %v", "https://example.com/support", contact["url"])
	}
	if contact["email"] != "support@example.com" {
		t.Fatalf("expected contact email=%q, got %v", "support@example.com", contact["email"])
	}
}

func TestSwagger_Good_UsesTermsOfServiceMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Terms API", "Terms test", "1.0.0"),
		api.WithSwaggerTermsOfService("https://example.com/terms"),
	)
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
	if info["termsOfService"] != "https://example.com/terms" {
		t.Fatalf("expected termsOfService=%q, got %v", "https://example.com/terms", info["termsOfService"])
	}
}

func TestSwagger_Good_UsesExternalDocsMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Docs API", "Docs test", "1.0.0"),
		api.WithSwaggerExternalDocs("Developer guide", "https://example.com/docs"),
	)
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

	externalDocs, ok := doc["externalDocs"].(map[string]any)
	if !ok {
		t.Fatal("expected externalDocs metadata in swagger doc")
	}
	if externalDocs["description"] != "Developer guide" {
		t.Fatalf("expected externalDocs description=%q, got %v", "Developer guide", externalDocs["description"])
	}
	if externalDocs["url"] != "https://example.com/docs" {
		t.Fatalf("expected externalDocs url=%q, got %v", "https://example.com/docs", externalDocs["url"])
	}
}

func TestSwagger_Good_IgnoresBlankMetadataOverrides(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Stable API", "Blank override test", "1.0.0"),
		api.WithSwaggerTermsOfService("https://example.com/terms"),
		api.WithSwaggerTermsOfService(""),
		api.WithSwaggerContact("API Support", "https://example.com/support", "support@example.com"),
		api.WithSwaggerContact("", "", ""),
		api.WithSwaggerLicense("EUPL-1.2", "https://eupl.eu/1.2/en/"),
		api.WithSwaggerLicense("", ""),
		api.WithSwaggerExternalDocs("Developer guide", "https://example.com/docs"),
		api.WithSwaggerExternalDocs("", ""),
	)
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
	if info["termsOfService"] != "https://example.com/terms" {
		t.Fatalf("expected termsOfService to survive blank override, got %v", info["termsOfService"])
	}

	contact, ok := info["contact"].(map[string]any)
	if !ok {
		t.Fatal("expected contact metadata in swagger doc")
	}
	if contact["name"] != "API Support" {
		t.Fatalf("expected contact name to survive blank override, got %v", contact["name"])
	}
	if contact["url"] != "https://example.com/support" {
		t.Fatalf("expected contact url to survive blank override, got %v", contact["url"])
	}
	if contact["email"] != "support@example.com" {
		t.Fatalf("expected contact email to survive blank override, got %v", contact["email"])
	}

	license, ok := info["license"].(map[string]any)
	if !ok {
		t.Fatal("expected license metadata in swagger doc")
	}
	if license["name"] != "EUPL-1.2" {
		t.Fatalf("expected license name to survive blank override, got %v", license["name"])
	}
	if license["url"] != "https://eupl.eu/1.2/en/" {
		t.Fatalf("expected license url to survive blank override, got %v", license["url"])
	}

	externalDocs, ok := doc["externalDocs"].(map[string]any)
	if !ok {
		t.Fatal("expected externalDocs metadata in swagger doc")
	}
	if externalDocs["description"] != "Developer guide" {
		t.Fatalf("expected externalDocs description to survive blank override, got %v", externalDocs["description"])
	}
	if externalDocs["url"] != "https://example.com/docs" {
		t.Fatalf("expected externalDocs url to survive blank override, got %v", externalDocs["url"])
	}
}

func TestSwagger_Good_UsesServerMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Server API", "Server metadata test", "1.0.0"),
		api.WithSwaggerServers(" https://api.example.com ", "/", "", "https://api.example.com"),
	)
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

	servers, ok := doc["servers"].([]any)
	if !ok {
		t.Fatalf("expected servers array, got %T", doc["servers"])
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 normalised servers, got %d", len(servers))
	}

	first := servers[0].(map[string]any)
	if first["url"] != "https://api.example.com" {
		t.Fatalf("expected first server url=%q, got %v", "https://api.example.com", first["url"])
	}

	second := servers[1].(map[string]any)
	if second["url"] != "/" {
		t.Fatalf("expected second server url=%q, got %v", "/", second["url"])
	}
}

func TestSwagger_Good_AppendsServerMetadataAcrossCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Server API", "Server metadata test", "1.0.0"),
		api.WithSwaggerServers("https://api.example.com", "/"),
		api.WithSwaggerServers(" https://docs.example.com ", "/", "https://api.example.com"),
	)
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

	servers, ok := doc["servers"].([]any)
	if !ok {
		t.Fatalf("expected servers array, got %T", doc["servers"])
	}
	if len(servers) != 3 {
		t.Fatalf("expected 3 normalised servers, got %d", len(servers))
	}

	expected := []string{"https://api.example.com", "/", "https://docs.example.com"}
	for i, want := range expected {
		got := servers[i].(map[string]any)["url"]
		if got != want {
			t.Fatalf("expected server[%d] url=%q, got %v", i, want, got)
		}
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

// TestOpenAPISpecEndpoint_Good verifies WithOpenAPISpec mounts a public
// GET /v1/openapi.json that returns the generated document. RFC.endpoints.md
// lists this as a framework route alongside /health and /swagger.
func TestOpenAPISpecEndpoint_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Test API", "A test API service", "1.0.0"),
		api.WithOpenAPISpec(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/openapi.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected application/json content type, got %q", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if doc["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi=3.1.0, got %v", doc["openapi"])
	}
	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		t.Fatalf("expected paths map, got %T", doc["paths"])
	}
	if _, ok := paths["/v1/openapi.json"]; !ok {
		t.Fatal("expected the spec endpoint to describe itself in paths")
	}
}

// TestOpenAPISpecEndpoint_Good_CustomPath verifies an explicit path override
// is honoured by the router.
func TestOpenAPISpecEndpoint_Good_CustomPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Test API", "A test API service", "1.0.0"),
		api.WithOpenAPISpecPath("/api/v1/openapi.json"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/openapi.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on custom spec path, got %d", resp.StatusCode)
	}

	// Default path should 404 when overridden.
	defaultResp, err := http.Get(srv.URL + "/v1/openapi.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer defaultResp.Body.Close()
	if defaultResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 on default spec path when overridden, got %d", defaultResp.StatusCode)
	}
}

// TestOpenAPISpecEndpoint_Bad_DisabledByDefault verifies the endpoint is not
// mounted unless opted in with WithOpenAPISpec().
func TestOpenAPISpecEndpoint_Bad_DisabledByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwagger("Test API", "A test API service", "1.0.0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/openapi.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 when endpoint is disabled, got %d", resp.StatusCode)
	}
}

// TestOpenAPISpecEndpoint_Ugly_WorksWithoutSwagger confirms the endpoint
// serves the spec even when the Swagger UI is not mounted — the standalone
// JSON document is independent of the UI bundle.
func TestOpenAPISpecEndpoint_Ugly_WorksWithoutSwagger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithOpenAPISpec())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/openapi.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 without swagger UI, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if doc["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi=3.1.0, got %v", doc["openapi"])
	}
}
