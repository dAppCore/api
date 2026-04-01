// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── Test helpers ──────────────────────────────────────────────────────────

type specStubGroup struct {
	name     string
	basePath string
	descs    []api.RouteDescription
}

func (s *specStubGroup) Name() string                       { return s.name }
func (s *specStubGroup) BasePath() string                   { return s.basePath }
func (s *specStubGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (s *specStubGroup) Describe() []api.RouteDescription   { return s.descs }

type plainStubGroup struct{}

func (plainStubGroup) Name() string                       { return "plain" }
func (plainStubGroup) BasePath() string                   { return "/plain" }
func (plainStubGroup) RegisterRoutes(rg *gin.RouterGroup) {}

// ── SpecBuilder tests ─────────────────────────────────────────────────────

func TestSpecBuilder_Good_EmptyGroups(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:       "Test",
		Description: "Empty test",
		Version:     "0.0.1",
	}

	data, err := sb.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify OpenAPI version.
	if spec["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi=3.1.0, got %v", spec["openapi"])
	}

	// Verify /health path exists.
	paths := spec["paths"].(map[string]any)
	if _, ok := paths["/health"]; !ok {
		t.Fatal("expected /health path in spec")
	}

	// Verify system tag exists.
	tags := spec["tags"].([]any)
	found := false
	for _, tag := range tags {
		tm := tag.(map[string]any)
		if tm["name"] == "system" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected system tag in spec")
	}

	components := spec["components"].(map[string]any)
	securitySchemes := components["securitySchemes"].(map[string]any)
	bearerAuth := securitySchemes["bearerAuth"].(map[string]any)
	if bearerAuth["type"] != "http" {
		t.Fatalf("expected bearerAuth.type=http, got %v", bearerAuth["type"])
	}
	if bearerAuth["scheme"] != "bearer" {
		t.Fatalf("expected bearerAuth.scheme=bearer, got %v", bearerAuth["scheme"])
	}

	security := spec["security"].([]any)
	if len(security) != 1 {
		t.Fatalf("expected one default security requirement, got %d", len(security))
	}
	req := security[0].(map[string]any)
	if _, ok := req["bearerAuth"]; !ok {
		t.Fatal("expected default bearerAuth security requirement")
	}
}

func TestSpecBuilder_Good_WithDescribableGroup(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:       "Test",
		Description: "Test API",
		Version:     "1.0.0",
	}

	group := &specStubGroup{
		name:     "items",
		basePath: "/api/items",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/list",
				Summary: "List items",
				Tags:    []string{"items"},
				Response: map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
				},
			},
			{
				Method:      "POST",
				Path:        "/create",
				Summary:     "Create item",
				Description: "Creates a new item",
				Tags:        []string{"items"},
				RequestBody: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
					},
				},
				Response: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{"type": "integer"},
					},
				},
			},
		},
	}

	data, err := sb.Build([]api.RouteGroup{group})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := spec["paths"].(map[string]any)

	// Verify GET /api/items/list exists.
	listPath, ok := paths["/api/items/list"]
	if !ok {
		t.Fatal("expected /api/items/list path in spec")
	}
	getOp := listPath.(map[string]any)["get"]
	if getOp == nil {
		t.Fatal("expected GET operation on /api/items/list")
	}
	if getOp.(map[string]any)["summary"] != "List items" {
		t.Fatalf("expected summary='List items', got %v", getOp.(map[string]any)["summary"])
	}
	if getOp.(map[string]any)["operationId"] != "get_api_items_list" {
		t.Fatalf("expected operationId='get_api_items_list', got %v", getOp.(map[string]any)["operationId"])
	}

	// Verify POST /api/items/create exists with request body.
	createPath, ok := paths["/api/items/create"]
	if !ok {
		t.Fatal("expected /api/items/create path in spec")
	}
	postOp := createPath.(map[string]any)["post"]
	if postOp == nil {
		t.Fatal("expected POST operation on /api/items/create")
	}
	if postOp.(map[string]any)["summary"] != "Create item" {
		t.Fatalf("expected summary='Create item', got %v", postOp.(map[string]any)["summary"])
	}
	if postOp.(map[string]any)["operationId"] != "post_api_items_create" {
		t.Fatalf("expected operationId='post_api_items_create', got %v", postOp.(map[string]any)["operationId"])
	}
	if postOp.(map[string]any)["requestBody"] == nil {
		t.Fatal("expected requestBody on POST /api/items/create")
	}
}

func TestSpecBuilder_Good_EnvelopeWrapping(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "data",
		basePath: "/data",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/fetch",
				Summary: "Fetch data",
				Tags:    []string{"data"},
				Response: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"value": map[string]any{"type": "string"},
					},
				},
			},
		},
	}

	data, err := sb.Build([]api.RouteGroup{group})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := spec["paths"].(map[string]any)
	fetchPath := paths["/data/fetch"].(map[string]any)
	getOp := fetchPath["get"].(map[string]any)
	responses := getOp["responses"].(map[string]any)
	resp200 := responses["200"].(map[string]any)
	content := resp200["content"].(map[string]any)
	appJSON := content["application/json"].(map[string]any)
	schema := appJSON["schema"].(map[string]any)
	if getOp["operationId"] != "get_data_fetch" {
		t.Fatalf("expected operationId='get_data_fetch', got %v", getOp["operationId"])
	}

	// Verify envelope structure.
	if schema["type"] != "object" {
		t.Fatalf("expected schema type=object, got %v", schema["type"])
	}

	properties := schema["properties"].(map[string]any)

	// Verify success field.
	success := properties["success"].(map[string]any)
	if success["type"] != "boolean" {
		t.Fatalf("expected success.type=boolean, got %v", success["type"])
	}

	// Verify data field contains the original response schema.
	dataField := properties["data"].(map[string]any)
	if dataField["type"] != "object" {
		t.Fatalf("expected data.type=object, got %v", dataField["type"])
	}
	dataProps := dataField["properties"].(map[string]any)
	if dataProps["value"] == nil {
		t.Fatal("expected data.properties.value to exist")
	}

	// Verify required contains "success".
	required := schema["required"].([]any)
	foundSuccess := false
	for _, r := range required {
		if r == "success" {
			foundSuccess = true
			break
		}
	}
	if !foundSuccess {
		t.Fatal("expected 'success' in required array")
	}
}

func TestSpecBuilder_Good_OperationIDPreservesPathParams(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "users",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/users/{id}",
				Summary: "Get user by id",
				Tags:    []string{"users"},
				Response: map[string]any{
					"type": "object",
				},
			},
			{
				Method:  "GET",
				Path:    "/users/{name}",
				Summary: "Get user by name",
				Tags:    []string{"users"},
				Response: map[string]any{
					"type": "object",
				},
			},
		},
	}

	data, err := sb.Build([]api.RouteGroup{group})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := spec["paths"].(map[string]any)
	byID := paths["/api/users/{id}"].(map[string]any)["get"].(map[string]any)
	byName := paths["/api/users/{name}"].(map[string]any)["get"].(map[string]any)

	if byID["operationId"] != "get_api_users_id" {
		t.Fatalf("expected operationId='get_api_users_id', got %v", byID["operationId"])
	}
	if byName["operationId"] != "get_api_users_name" {
		t.Fatalf("expected operationId='get_api_users_name', got %v", byName["operationId"])
	}
	if byID["operationId"] == byName["operationId"] {
		t.Fatal("expected unique operationId values for distinct path parameters")
	}
}

func TestSpecBuilder_Good_RequestBodyOnDelete(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "resources",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:  "DELETE",
				Path:    "/resources/{id}",
				Summary: "Delete resource",
				Tags:    []string{"resources"},
				RequestBody: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"reason": map[string]any{"type": "string"},
					},
				},
				Response: map[string]any{
					"type": "object",
				},
			},
		},
	}

	data, err := sb.Build([]api.RouteGroup{group})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := spec["paths"].(map[string]any)
	deleteOp := paths["/api/resources/{id}"].(map[string]any)["delete"].(map[string]any)
	if deleteOp["requestBody"] == nil {
		t.Fatal("expected requestBody on DELETE /api/resources/{id}")
	}
}

func TestSpecBuilder_Good_RequestBodyOnHead(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "resources",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:  "HEAD",
				Path:    "/resources/{id}",
				Summary: "Check resource",
				Tags:    []string{"resources"},
				RequestBody: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"include": map[string]any{"type": "string"},
					},
				},
				Response: map[string]any{
					"type": "object",
				},
			},
		},
	}

	data, err := sb.Build([]api.RouteGroup{group})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := spec["paths"].(map[string]any)
	headOp := paths["/api/resources/{id}"].(map[string]any)["head"].(map[string]any)
	if headOp["requestBody"] == nil {
		t.Fatal("expected requestBody on HEAD /api/resources/{id}")
	}
}

func TestSpecBuilder_Good_NonDescribableGroup(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	data, err := sb.Build([]api.RouteGroup{plainStubGroup{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify plainStubGroup appears in tags.
	tags := spec["tags"].([]any)
	foundPlain := false
	for _, tag := range tags {
		tm := tag.(map[string]any)
		if tm["name"] == "plain" {
			foundPlain = true
			break
		}
	}
	if !foundPlain {
		t.Fatal("expected 'plain' tag in spec for non-describable group")
	}

	// Verify only /health exists in paths (plain group adds no paths).
	paths := spec["paths"].(map[string]any)
	if len(paths) != 1 {
		t.Fatalf("expected 1 path (/health only), got %d", len(paths))
	}
	if _, ok := paths["/health"]; !ok {
		t.Fatal("expected /health path in spec")
	}
	health := paths["/health"].(map[string]any)["get"].(map[string]any)
	if health["operationId"] != "get_health" {
		t.Fatalf("expected operationId='get_health', got %v", health["operationId"])
	}
	if security := health["security"]; security == nil {
		t.Fatal("expected explicit public security override on /health")
	}
	if len(health["security"].([]any)) != 0 {
		t.Fatalf("expected /health security to be empty, got %v", health["security"])
	}
}

func TestSpecBuilder_Good_DefaultTagsFromGroupName(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "fallback",
		basePath: "/api/fallback",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/status",
				Summary: "Check status",
				Response: map[string]any{
					"type": "object",
				},
			},
		},
	}

	data, err := sb.Build([]api.RouteGroup{group})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	operation := spec["paths"].(map[string]any)["/api/fallback/status"].(map[string]any)["get"].(map[string]any)
	tags, ok := operation["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags array, got %T", operation["tags"])
	}
	if len(tags) != 1 || tags[0] != "fallback" {
		t.Fatalf("expected fallback tag from group name, got %v", operation["tags"])
	}
}

func TestSpecBuilder_Good_ToolBridgeIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sb := &api.SpecBuilder{
		Title:   "Tool API",
		Version: "1.0.0",
	}

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file from disk",
		Group:       "files",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"content": map[string]any{"type": "string"},
			},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("ok"))
	})
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

	data, err := sb.Build([]api.RouteGroup{bridge})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := spec["paths"].(map[string]any)

	// Verify POST /tools/file_read exists.
	fileReadPath, ok := paths["/tools/file_read"]
	if !ok {
		t.Fatal("expected /tools/file_read path in spec")
	}
	postOp := fileReadPath.(map[string]any)["post"]
	if postOp == nil {
		t.Fatal("expected POST operation on /tools/file_read")
	}
	if postOp.(map[string]any)["summary"] != "Read a file from disk" {
		t.Fatalf("expected summary='Read a file from disk', got %v", postOp.(map[string]any)["summary"])
	}
	if postOp.(map[string]any)["operationId"] != "post_tools_file_read" {
		t.Fatalf("expected operationId='post_tools_file_read', got %v", postOp.(map[string]any)["operationId"])
	}

	// Verify POST /tools/metrics_query exists.
	metricsPath, ok := paths["/tools/metrics_query"]
	if !ok {
		t.Fatal("expected /tools/metrics_query path in spec")
	}
	metricsOp := metricsPath.(map[string]any)["post"]
	if metricsOp == nil {
		t.Fatal("expected POST operation on /tools/metrics_query")
	}
	if metricsOp.(map[string]any)["summary"] != "Query metrics data" {
		t.Fatalf("expected summary='Query metrics data', got %v", metricsOp.(map[string]any)["summary"])
	}
	if metricsOp.(map[string]any)["operationId"] != "post_tools_metrics_query" {
		t.Fatalf("expected operationId='post_tools_metrics_query', got %v", metricsOp.(map[string]any)["operationId"])
	}

	// Verify request body is present on both (both are POST with InputSchema).
	if postOp.(map[string]any)["requestBody"] == nil {
		t.Fatal("expected requestBody on POST /tools/file_read")
	}
	if metricsOp.(map[string]any)["requestBody"] == nil {
		t.Fatal("expected requestBody on POST /tools/metrics_query")
	}
}

func TestSpecBuilder_Bad_InfoFields(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:       "MyAPI",
		Description: "Test API",
		Version:     "1.0.0",
	}

	data, err := sb.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	info := spec["info"].(map[string]any)
	if info["title"] != "MyAPI" {
		t.Fatalf("expected title=MyAPI, got %v", info["title"])
	}
	if info["description"] != "Test API" {
		t.Fatalf("expected description='Test API', got %v", info["description"])
	}
	if info["version"] != "1.0.0" {
		t.Fatalf("expected version=1.0.0, got %v", info["version"])
	}
}
