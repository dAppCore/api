// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"iter"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── Test helpers ──────────────────────────────────────────────────────────

type specStubGroup struct {
	name     string
	basePath string
	hidden   bool
	descs    []api.RouteDescription
}

func (s *specStubGroup) Name() string                       { return s.name }
func (s *specStubGroup) BasePath() string                   { return s.basePath }
func (s *specStubGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (s *specStubGroup) Describe() []api.RouteDescription   { return s.descs }
func (s *specStubGroup) Hidden() bool                       { return s.hidden }

type plainStubGroup struct{}

func (plainStubGroup) Name() string                       { return "plain" }
func (plainStubGroup) BasePath() string                   { return "/plain" }
func (plainStubGroup) RegisterRoutes(rg *gin.RouterGroup) {}

type iterStubGroup struct {
	name     string
	basePath string
	descs    []api.RouteDescription
}

func (s *iterStubGroup) Name() string                       { return s.name }
func (s *iterStubGroup) BasePath() string                   { return s.basePath }
func (s *iterStubGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (s *iterStubGroup) Describe() []api.RouteDescription   { return nil }
func (s *iterStubGroup) DescribeIter() iter.Seq[api.RouteDescription] {
	return func(yield func(api.RouteDescription) bool) {
		for _, rd := range s.descs {
			if !yield(rd) {
				return
			}
		}
	}
}

type iterNilFallbackGroup struct {
	name     string
	basePath string
	descs    []api.RouteDescription
}

func (s *iterNilFallbackGroup) Name() string                       { return s.name }
func (s *iterNilFallbackGroup) BasePath() string                   { return s.basePath }
func (s *iterNilFallbackGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (s *iterNilFallbackGroup) Describe() []api.RouteDescription   { return s.descs }
func (s *iterNilFallbackGroup) DescribeIter() iter.Seq[api.RouteDescription] {
	return nil
}

type countingIterGroup struct {
	name          string
	basePath      string
	descs         []api.RouteDescription
	describeCalls int
}

func (s *countingIterGroup) Name() string                       { return s.name }
func (s *countingIterGroup) BasePath() string                   { return s.basePath }
func (s *countingIterGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (s *countingIterGroup) Describe() []api.RouteDescription   { return nil }
func (s *countingIterGroup) DescribeIter() iter.Seq[api.RouteDescription] {
	s.describeCalls++
	return func(yield func(api.RouteDescription) bool) {
		for _, rd := range s.descs {
			if !yield(rd) {
				return
			}
		}
	}
}

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
	health := paths["/health"].(map[string]any)["get"].(map[string]any)
	healthResponses := health["responses"].(map[string]any)
	if _, ok := healthResponses["429"]; !ok {
		t.Fatal("expected 429 response on /health")
	}
	if _, ok := healthResponses["504"]; !ok {
		t.Fatal("expected 504 response on /health")
	}
	if _, ok := healthResponses["500"]; !ok {
		t.Fatal("expected 500 response on /health")
	}
	rateLimit429 := healthResponses["429"].(map[string]any)
	headers := rateLimit429["headers"].(map[string]any)
	if _, ok := headers["Retry-After"]; !ok {
		t.Fatal("expected Retry-After header on /health 429 response")
	}
	if _, ok := headers["X-Request-ID"]; !ok {
		t.Fatal("expected X-Request-ID header on /health 429 response")
	}
	if _, ok := headers["X-RateLimit-Limit"]; !ok {
		t.Fatal("expected X-RateLimit-Limit header on /health 429 response")
	}
	if _, ok := headers["X-RateLimit-Remaining"]; !ok {
		t.Fatal("expected X-RateLimit-Remaining header on /health 429 response")
	}
	if _, ok := headers["X-RateLimit-Reset"]; !ok {
		t.Fatal("expected X-RateLimit-Reset header on /health 429 response")
	}
	health504 := healthResponses["504"].(map[string]any)
	health504Headers := health504["headers"].(map[string]any)
	if _, ok := health504Headers["X-Request-ID"]; !ok {
		t.Fatal("expected X-Request-ID header on /health 504 response")
	}
	if _, ok := health504Headers["X-RateLimit-Limit"]; !ok {
		t.Fatal("expected X-RateLimit-Limit header on /health 504 response")
	}
	if _, ok := health504Headers["X-RateLimit-Remaining"]; !ok {
		t.Fatal("expected X-RateLimit-Remaining header on /health 504 response")
	}
	if _, ok := health504Headers["X-RateLimit-Reset"]; !ok {
		t.Fatal("expected X-RateLimit-Reset header on /health 504 response")
	}
	health200 := health["responses"].(map[string]any)["200"].(map[string]any)
	health200Headers := health200["headers"].(map[string]any)
	if _, ok := health200Headers["X-Cache"]; !ok {
		t.Fatal("expected X-Cache header on /health 200 response")
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
	schemas := components["schemas"].(map[string]any)
	if _, ok := schemas["Response"]; !ok {
		t.Fatal("expected Response component schema in spec")
	}
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

func TestSpecBuilder_Good_GraphQLEndpoint(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:       "Test",
		Description: "GraphQL test",
		Version:     "1.0.0",
		GraphQLPath: "/graphql",
	}

	data, err := sb.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	tags := spec["tags"].([]any)
	found := false
	for _, tag := range tags {
		tm := tag.(map[string]any)
		if tm["name"] == "graphql" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected graphql tag in spec")
	}

	paths := spec["paths"].(map[string]any)
	pathItem, ok := paths["/graphql"].(map[string]any)
	if !ok {
		t.Fatal("expected /graphql path in spec")
	}

	postOp := pathItem["post"].(map[string]any)
	if postOp["operationId"] != "post_graphql" {
		t.Fatalf("expected GraphQL operationId to be post_graphql, got %v", postOp["operationId"])
	}

	requestBody := postOp["requestBody"].(map[string]any)
	schema := requestBody["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	properties := schema["properties"].(map[string]any)
	if _, ok := properties["query"]; !ok {
		t.Fatal("expected GraphQL request schema to include query field")
	}
	if _, ok := properties["variables"]; !ok {
		t.Fatal("expected GraphQL request schema to include variables field")
	}
	if _, ok := properties["operationName"]; !ok {
		t.Fatal("expected GraphQL request schema to include operationName field")
	}
}

func TestSpecBuilder_Good_InfoIncludesLicenseMetadata(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:       "Test",
		Description: "Licensed test API",
		Version:     "1.2.3",
		LicenseName: "EUPL-1.2",
		LicenseURL:  "https://eupl.eu/1.2/en/",
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
	license, ok := info["license"].(map[string]any)
	if !ok {
		t.Fatal("expected license metadata in spec info")
	}
	if license["name"] != "EUPL-1.2" {
		t.Fatalf("expected license name EUPL-1.2, got %v", license["name"])
	}
	if license["url"] != "https://eupl.eu/1.2/en/" {
		t.Fatalf("expected license url to be preserved, got %v", license["url"])
	}
}

func TestSpecBuilder_Good_InfoIncludesContactMetadata(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:        "Test",
		Description:  "Contact test API",
		Version:      "1.2.3",
		ContactName:  "API Support",
		ContactURL:   "https://example.com/support",
		ContactEmail: "support@example.com",
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
	contact, ok := info["contact"].(map[string]any)
	if !ok {
		t.Fatal("expected contact metadata in spec info")
	}
	if contact["name"] != "API Support" {
		t.Fatalf("expected contact name API Support, got %v", contact["name"])
	}
	if contact["url"] != "https://example.com/support" {
		t.Fatalf("expected contact url to be preserved, got %v", contact["url"])
	}
	if contact["email"] != "support@example.com" {
		t.Fatalf("expected contact email to be preserved, got %v", contact["email"])
	}
}

func TestSpecBuilder_Good_InfoIncludesTermsOfService(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:          "Test",
		Description:    "Terms test API",
		Version:        "1.2.3",
		TermsOfService: "https://example.com/terms",
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
	if info["termsOfService"] != "https://example.com/terms" {
		t.Fatalf("expected termsOfService to be preserved, got %v", info["termsOfService"])
	}
}

func TestSpecBuilder_Good_InfoIncludesExternalDocs(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:                   "Test",
		Description:             "External docs test API",
		Version:                 "1.2.3",
		ExternalDocsDescription: "Developer guide",
		ExternalDocsURL:         "https://example.com/docs",
	}

	data, err := sb.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	externalDocs, ok := spec["externalDocs"].(map[string]any)
	if !ok {
		t.Fatal("expected externalDocs metadata in spec")
	}
	if externalDocs["description"] != "Developer guide" {
		t.Fatalf("expected externalDocs description to be preserved, got %v", externalDocs["description"])
	}
	if externalDocs["url"] != "https://example.com/docs" {
		t.Fatalf("expected externalDocs url to be preserved, got %v", externalDocs["url"])
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
				RequestExample: map[string]any{
					"name": "Widget",
				},
				Response: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{"type": "integer"},
					},
				},
				ResponseExample: map[string]any{
					"id": 42,
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
	requestBody := postOp.(map[string]any)["requestBody"].(map[string]any)
	appJSON := requestBody["content"].(map[string]any)["application/json"].(map[string]any)
	if appJSON["example"].(map[string]any)["name"] != "Widget" {
		t.Fatalf("expected request example to be preserved, got %v", appJSON["example"])
	}

	responses := postOp.(map[string]any)["responses"].(map[string]any)
	created := responses["200"].(map[string]any)
	createdJSON := created["content"].(map[string]any)["application/json"].(map[string]any)
	if createdJSON["example"].(map[string]any)["id"] != float64(42) {
		t.Fatalf("expected response example to be preserved, got %v", createdJSON["example"])
	}
}

func TestSpecBuilder_Good_DescribeIterGroup(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &iterStubGroup{
		name:     "iter",
		basePath: "/api/iter",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/status",
				Summary: "Iter status",
				Tags:    []string{"iter"},
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

	op := spec["paths"].(map[string]any)["/api/iter/status"].(map[string]any)["get"].(map[string]any)
	if op["summary"] != "Iter status" {
		t.Fatalf("expected summary='Iter status', got %v", op["summary"])
	}
	tags, ok := op["tags"].([]any)
	if !ok || len(tags) != 1 || tags[0] != "iter" {
		t.Fatalf("expected tags to be populated from DescribeIter, got %v", op["tags"])
	}
}

func TestSpecBuilder_Good_DescribeIterSnapshotOnce(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &countingIterGroup{
		name:     "counted",
		basePath: "/api/count",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/status",
				Summary: "Counted status",
				Tags:    []string{"counted"},
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

	if group.describeCalls != 1 {
		t.Fatalf("expected DescribeIter to be called once, got %d", group.describeCalls)
	}

	op := spec["paths"].(map[string]any)["/api/count/status"].(map[string]any)["get"].(map[string]any)
	if op["summary"] != "Counted status" {
		t.Fatalf("expected summary='Counted status', got %v", op["summary"])
	}
}

func TestSpecBuilder_Good_DescribeIterNilFallsBackToDescribe(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &iterNilFallbackGroup{
		name:     "fallback-iter",
		basePath: "/api/fallback-iter",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/status",
				Summary: "Fallback status",
				Tags:    []string{"fallback-iter"},
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

	op := spec["paths"].(map[string]any)["/api/fallback-iter/status"].(map[string]any)["get"].(map[string]any)
	if op["summary"] != "Fallback status" {
		t.Fatalf("expected summary='Fallback status', got %v", op["summary"])
	}
}

func TestSpecBuilder_Good_SecuredResponses(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "secure",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/private",
				Summary: "Private endpoint",
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

	responses := spec["paths"].(map[string]any)["/api/private"].(map[string]any)["get"].(map[string]any)["responses"].(map[string]any)
	if _, ok := responses["401"]; !ok {
		t.Fatal("expected 401 response in secured operation")
	}
	if _, ok := responses["403"]; !ok {
		t.Fatal("expected 403 response in secured operation")
	}
	if _, ok := responses["429"]; !ok {
		t.Fatal("expected 429 response in secured operation")
	}
	if _, ok := responses["504"]; !ok {
		t.Fatal("expected 504 response in secured operation")
	}
	if _, ok := responses["500"]; !ok {
		t.Fatal("expected 500 response in secured operation")
	}
	rateLimit429 := responses["429"].(map[string]any)
	headers := rateLimit429["headers"].(map[string]any)
	if _, ok := headers["Retry-After"]; !ok {
		t.Fatal("expected Retry-After header in secured operation 429 response")
	}
	if _, ok := headers["X-Request-ID"]; !ok {
		t.Fatal("expected X-Request-ID header in secured operation 429 response")
	}
	if _, ok := headers["X-RateLimit-Limit"]; !ok {
		t.Fatal("expected X-RateLimit-Limit header in secured operation 429 response")
	}
	if _, ok := headers["X-RateLimit-Remaining"]; !ok {
		t.Fatal("expected X-RateLimit-Remaining header in secured operation 429 response")
	}
	if _, ok := headers["X-RateLimit-Reset"]; !ok {
		t.Fatal("expected X-RateLimit-Reset header in secured operation 429 response")
	}
	for _, code := range []string{"400", "401", "403", "504", "500"} {
		resp := responses[code].(map[string]any)
		respHeaders := resp["headers"].(map[string]any)
		if _, ok := respHeaders["X-Request-ID"]; !ok {
			t.Fatalf("expected X-Request-ID header in secured operation %s response", code)
		}
		if _, ok := respHeaders["X-RateLimit-Limit"]; !ok {
			t.Fatalf("expected X-RateLimit-Limit header in secured operation %s response", code)
		}
		if _, ok := respHeaders["X-RateLimit-Remaining"]; !ok {
			t.Fatalf("expected X-RateLimit-Remaining header in secured operation %s response", code)
		}
		if _, ok := respHeaders["X-RateLimit-Reset"]; !ok {
			t.Fatalf("expected X-RateLimit-Reset header in secured operation %s response", code)
		}
	}
}

func TestSpecBuilder_Good_CustomSuccessStatusCode(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "items",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:     "POST",
				Path:       "/items",
				Summary:    "Create item",
				StatusCode: http.StatusCreated,
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

	responses := spec["paths"].(map[string]any)["/api/items"].(map[string]any)["post"].(map[string]any)["responses"].(map[string]any)
	if _, ok := responses["201"]; !ok {
		t.Fatal("expected 201 response for created operation")
	}
	if _, ok := responses["200"]; ok {
		t.Fatal("expected 200 response to be omitted when a custom success status is declared")
	}

	created := responses["201"].(map[string]any)
	if created["description"] != "Created" {
		t.Fatalf("expected created description, got %v", created["description"])
	}
	if created["content"] == nil {
		t.Fatal("expected content for 201 response")
	}
}

func TestSpecBuilder_Good_NoContentSuccessStatusCode(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "items",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:     "DELETE",
				Path:       "/items/{id}",
				Summary:    "Delete item",
				StatusCode: http.StatusNoContent,
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

	responses := spec["paths"].(map[string]any)["/api/items/{id}"].(map[string]any)["delete"].(map[string]any)["responses"].(map[string]any)
	resp204 := responses["204"].(map[string]any)
	if resp204["description"] != "No content" {
		t.Fatalf("expected no-content description, got %v", resp204["description"])
	}
	if _, ok := resp204["content"]; ok {
		t.Fatal("expected no content block for 204 response")
	}
}

func TestSpecBuilder_Good_RouteSecurityOverrides(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "security",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:   "GET",
				Path:     "/public",
				Summary:  "Public endpoint",
				Security: []map[string][]string{},
				Response: map[string]any{
					"type": "object",
				},
			},
			{
				Method:  "GET",
				Path:    "/scoped",
				Summary: "Scoped endpoint",
				Security: []map[string][]string{
					{
						"bearerAuth": []string{},
					},
					{
						"oauth2": []string{"read:items"},
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

	publicOp := paths["/api/public"].(map[string]any)["get"].(map[string]any)
	publicSecurity, ok := publicOp["security"].([]any)
	if !ok {
		t.Fatalf("expected public security array, got %T", publicOp["security"])
	}
	if len(publicSecurity) != 0 {
		t.Fatalf("expected public route to have empty security requirement, got %v", publicSecurity)
	}
	publicResponses := publicOp["responses"].(map[string]any)
	if _, ok := publicResponses["401"]; ok {
		t.Fatal("expected public route to omit 401 response documentation")
	}
	if _, ok := publicResponses["403"]; ok {
		t.Fatal("expected public route to omit 403 response documentation")
	}

	scopedOp := paths["/api/scoped"].(map[string]any)["get"].(map[string]any)
	scopedSecurity, ok := scopedOp["security"].([]any)
	if !ok {
		t.Fatalf("expected scoped security array, got %T", scopedOp["security"])
	}
	if len(scopedSecurity) != 2 {
		t.Fatalf("expected 2 security requirements, got %d", len(scopedSecurity))
	}
	firstReq := scopedSecurity[0].(map[string]any)
	if _, ok := firstReq["bearerAuth"]; !ok {
		t.Fatalf("expected bearerAuth requirement, got %v", firstReq)
	}
	secondReq := scopedSecurity[1].(map[string]any)
	if scopes, ok := secondReq["oauth2"].([]any); !ok || len(scopes) != 1 || scopes[0] != "read:items" {
		t.Fatalf("expected oauth2 read:items requirement, got %v", secondReq["oauth2"])
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
	headers := resp200["headers"].(map[string]any)
	if _, ok := headers["X-Request-ID"]; !ok {
		t.Fatal("expected X-Request-ID header on 200 response")
	}
	if _, ok := headers["X-RateLimit-Limit"]; !ok {
		t.Fatal("expected X-RateLimit-Limit header on 200 response")
	}
	if _, ok := headers["X-RateLimit-Remaining"]; !ok {
		t.Fatal("expected X-RateLimit-Remaining header on 200 response")
	}
	if _, ok := headers["X-RateLimit-Reset"]; !ok {
		t.Fatal("expected X-RateLimit-Reset header on 200 response")
	}
	if _, ok := headers["X-Cache"]; !ok {
		t.Fatal("expected X-Cache header on 200 response")
	}
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

func TestSpecBuilder_Good_RequestExampleWithoutSchema(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "resources",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:  "POST",
				Path:    "/resources",
				Summary: "Create resource",
				Tags:    []string{"resources"},
				RequestExample: map[string]any{
					"name": "Example resource",
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

	postOp := spec["paths"].(map[string]any)["/api/resources"].(map[string]any)["post"].(map[string]any)
	requestBody := postOp["requestBody"].(map[string]any)
	appJSON := requestBody["content"].(map[string]any)["application/json"].(map[string]any)

	if appJSON["example"].(map[string]any)["name"] != "Example resource" {
		t.Fatalf("expected request example to be preserved, got %v", appJSON["example"])
	}

	schema := appJSON["schema"].(map[string]any)
	if len(schema) != 0 {
		t.Fatalf("expected example-only request body to use an empty schema, got %v", schema)
	}
}

func TestSpecBuilder_Good_ResponseHeaders(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "downloads",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/exports/{id}",
				Summary: "Download export",
				ResponseHeaders: map[string]string{
					"Content-Disposition": "Download filename suggested by the server",
					"X-Export-ID":         "Identifier for the generated export",
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

	responses := spec["paths"].(map[string]any)["/api/exports/{id}"].(map[string]any)["get"].(map[string]any)["responses"].(map[string]any)
	resp200 := responses["200"].(map[string]any)
	headers, ok := resp200["headers"].(map[string]any)
	if !ok {
		t.Fatalf("expected headers map, got %T", resp200["headers"])
	}

	header, ok := headers["Content-Disposition"].(map[string]any)
	if !ok {
		t.Fatal("expected Content-Disposition response header to be documented")
	}
	if header["description"] != "Download filename suggested by the server" {
		t.Fatalf("expected header description to be preserved, got %v", header["description"])
	}
	schema := header["schema"].(map[string]any)
	if schema["type"] != "string" {
		t.Fatalf("expected response header schema type string, got %v", schema["type"])
	}

	errorResp := responses["500"].(map[string]any)
	errorHeaders, ok := errorResp["headers"].(map[string]any)
	if !ok {
		t.Fatalf("expected 500 headers map, got %T", errorResp["headers"])
	}
	if _, ok := errorHeaders["Content-Disposition"]; !ok {
		t.Fatal("expected route-specific headers on error responses too")
	}
	if _, ok := errorHeaders["X-Export-ID"]; !ok {
		t.Fatal("expected route-specific headers on error responses too")
	}
}

func TestSpecBuilder_Good_PathParameters(t *testing.T) {
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
				Path:    "/users/{id}/{slug}",
				Summary: "Get user",
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

	op := spec["paths"].(map[string]any)["/api/users/{id}/{slug}"].(map[string]any)["get"].(map[string]any)
	params, ok := op["parameters"].([]any)
	if !ok {
		t.Fatalf("expected parameters array, got %T", op["parameters"])
	}
	if len(params) != 2 {
		t.Fatalf("expected 2 path parameters, got %d", len(params))
	}

	first := params[0].(map[string]any)
	if first["name"] != "id" {
		t.Fatalf("expected first parameter name=id, got %v", first["name"])
	}
	if first["in"] != "path" {
		t.Fatalf("expected first parameter in=path, got %v", first["in"])
	}
	if required, ok := first["required"].(bool); !ok || !required {
		t.Fatalf("expected first parameter to be required, got %v", first["required"])
	}

	second := params[1].(map[string]any)
	if second["name"] != "slug" {
		t.Fatalf("expected second parameter name=slug, got %v", second["name"])
	}
}

func TestSpecBuilder_Good_PathNormalisation(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "users",
		basePath: "/api/",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "users/{id}",
				Summary: "Get user",
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
	if _, ok := paths["/api/users/{id}"]; !ok {
		t.Fatalf("expected normalised path /api/users/{id}, got %v", paths)
	}
}

func TestSpecBuilder_Good_GinPathParameters(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "users",
		basePath: "/api/",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "users/:id",
				Summary: "Get user",
				Response: map[string]any{
					"type": "object",
				},
			},
			{
				Method:  "GET",
				Path:    "files/*path",
				Summary: "Get file",
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

	userOp := paths["/api/users/{id}"].(map[string]any)["get"].(map[string]any)
	userParams := userOp["parameters"].([]any)
	if len(userParams) != 1 {
		t.Fatalf("expected 1 parameter for gin path, got %d", len(userParams))
	}
	if userParams[0].(map[string]any)["name"] != "id" {
		t.Fatalf("expected gin path parameter name=id, got %v", userParams[0])
	}

	fileOp := paths["/api/files/{path}"].(map[string]any)["get"].(map[string]any)
	fileParams := fileOp["parameters"].([]any)
	if len(fileParams) != 1 {
		t.Fatalf("expected 1 parameter for wildcard path, got %d", len(fileParams))
	}
	if fileParams[0].(map[string]any)["name"] != "path" {
		t.Fatalf("expected wildcard parameter name=path, got %v", fileParams[0])
	}
}

func TestSpecBuilder_Good_ExplicitParameters(t *testing.T) {
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
				Summary: "Get user",
				Parameters: []api.ParameterDescription{
					{
						Name:        "id",
						In:          "path",
						Description: "User identifier",
						Schema: map[string]any{
							"type": "string",
						},
					},
					{
						Name:        "verbose",
						In:          "query",
						Description: "Include verbose details",
						Schema: map[string]any{
							"type": "boolean",
						},
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

	op := spec["paths"].(map[string]any)["/api/users/{id}"].(map[string]any)["get"].(map[string]any)
	params, ok := op["parameters"].([]any)
	if !ok {
		t.Fatalf("expected parameters array, got %T", op["parameters"])
	}
	if len(params) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(params))
	}

	pathParam := params[0].(map[string]any)
	if pathParam["name"] != "id" {
		t.Fatalf("expected path parameter name=id, got %v", pathParam["name"])
	}
	if pathParam["in"] != "path" {
		t.Fatalf("expected path parameter in=path, got %v", pathParam["in"])
	}
	if pathParam["description"] != "User identifier" {
		t.Fatalf("expected merged path parameter description, got %v", pathParam["description"])
	}

	queryParam := params[1].(map[string]any)
	if queryParam["name"] != "verbose" {
		t.Fatalf("expected query parameter name=verbose, got %v", queryParam["name"])
	}
	if queryParam["in"] != "query" {
		t.Fatalf("expected query parameter in=query, got %v", queryParam["in"])
	}
	if required, ok := queryParam["required"].(bool); !ok || required {
		t.Fatalf("expected query parameter to be optional, got %v", queryParam["required"])
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

func TestSpecBuilder_Good_EmptyDescribableGroupStillAddsTag(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "empty",
		basePath: "/api/empty",
		descs:    nil,
	}

	data, err := sb.Build([]api.RouteGroup{group})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	tags := spec["tags"].([]any)
	foundEmpty := false
	for _, tag := range tags {
		tm := tag.(map[string]any)
		if tm["name"] == "empty" {
			foundEmpty = true
			break
		}
	}
	if !foundEmpty {
		t.Fatal("expected empty describable group to appear in spec tags")
	}

	paths := spec["paths"].(map[string]any)
	if len(paths) != 1 {
		t.Fatalf("expected only /health path, got %d paths", len(paths))
	}
	if _, ok := paths["/health"]; !ok {
		t.Fatal("expected /health path in spec")
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

func TestSpecBuilder_Good_TagsAreSortedDeterministically(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "gamma",
		basePath: "/api/gamma",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/status",
				Summary: "Check status",
				Tags:    []string{"zeta", "alpha", "beta"},
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

	tags, ok := spec["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags array, got %T", spec["tags"])
	}

	names := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := raw.(map[string]any)
		name, _ := tag["name"].(string)
		names = append(names, name)
	}

	expected := []string{"system", "alpha", "beta", "gamma", "zeta"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d tags, got %d: %v", len(expected), len(names), names)
	}
	for i := range expected {
		if names[i] != expected[i] {
			t.Fatalf("expected tag order %v, got %v", expected, names)
		}
	}
}

func TestSpecBuilder_Good_DeprecatedOperation(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "legacy",
		basePath: "/api/legacy",
		descs: []api.RouteDescription{
			{
				Method:      "GET",
				Path:        "/status",
				Summary:     "Check legacy status",
				Deprecated:  true,
				SunsetDate:  "2025-06-01",
				Replacement: "/api/v2/status",
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

	op := spec["paths"].(map[string]any)["/api/legacy/status"].(map[string]any)["get"].(map[string]any)
	deprecated, ok := op["deprecated"].(bool)
	if !ok {
		t.Fatalf("expected deprecated boolean, got %T", op["deprecated"])
	}
	if !deprecated {
		t.Fatal("expected deprecated operation to be marked true")
	}

	responses := op["responses"].(map[string]any)
	success := responses["200"].(map[string]any)
	headers := success["headers"].(map[string]any)
	for _, name := range []string{"Deprecation", "Sunset", "Link", "X-API-Warn"} {
		if _, ok := headers[name]; !ok {
			t.Fatalf("expected deprecation header %q in operation response headers", name)
		}
	}

	components := spec["components"].(map[string]any)
	headerComponents := components["headers"].(map[string]any)
	for _, name := range []string{"deprecation", "sunset", "link", "xapiwarn"} {
		if _, ok := headerComponents[name]; !ok {
			t.Fatalf("expected reusable header component %q", name)
		}
	}

	deprecationHeader := headers["Deprecation"].(map[string]any)
	if got := deprecationHeader["$ref"]; got != "#/components/headers/deprecation" {
		t.Fatalf("expected Deprecation header to reference shared component, got %v", got)
	}
}

func TestSpecBuilder_Good_BlankTagsAreIgnored(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	group := &specStubGroup{
		name:     "   ",
		basePath: "/api/blank",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/status",
				Summary: "Check status",
				Tags:    []string{"", "  ", "data", "data"},
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

	tags := spec["tags"].([]any)
	var foundData bool
	for _, raw := range tags {
		tag := raw.(map[string]any)
		name, _ := tag["name"].(string)
		if name == "" {
			t.Fatal("expected blank tag names to be ignored")
		}
		if name == "data" {
			foundData = true
		}
	}
	if !foundData {
		t.Fatal("expected data tag to be retained")
	}

	op := spec["paths"].(map[string]any)["/api/blank/status"].(map[string]any)["get"].(map[string]any)
	opTags, ok := op["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags array, got %T", op["tags"])
	}
	if len(opTags) != 1 || opTags[0] != "data" {
		t.Fatalf("expected operation tags to be cleaned to [data], got %v", opTags)
	}
}

func TestSpecBuilder_Good_BlankRouteTagsFallBackToGroupName(t *testing.T) {
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
				Tags:    []string{"", "  "},
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

	op := spec["paths"].(map[string]any)["/api/fallback/status"].(map[string]any)["get"].(map[string]any)
	tags, ok := op["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags array, got %T", op["tags"])
	}
	if len(tags) != 1 || tags[0] != "fallback" {
		t.Fatalf("expected blank route tags to fall back to group name, got %v", tags)
	}
}

func TestSpecBuilder_Good_HiddenRoutesAreOmitted(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	visible := &specStubGroup{
		name:     "visible",
		basePath: "/api",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/public",
				Summary: "Public endpoint",
				Tags:    []string{"public"},
				Response: map[string]any{
					"type": "object",
				},
			},
			{
				Method:  "GET",
				Path:    "/internal",
				Summary: "Internal endpoint",
				Tags:    []string{"internal"},
				Hidden:  true,
				Response: map[string]any{
					"type": "object",
				},
			},
		},
	}

	hidden := &specStubGroup{
		name:     "hidden-group",
		basePath: "/api/internal",
		hidden:   true,
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/status",
				Summary: "Hidden group endpoint",
				Tags:    []string{"hidden"},
				Response: map[string]any{
					"type": "object",
				},
			},
		},
	}

	data, err := sb.Build([]api.RouteGroup{visible, hidden})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := spec["paths"].(map[string]any)
	if _, ok := paths["/api/public"]; !ok {
		t.Fatal("expected visible route to remain in the spec")
	}
	if _, ok := paths["/api/internal"]; ok {
		t.Fatal("did not expect hidden route to appear in the spec")
	}
	if _, ok := paths["/api/internal/status"]; ok {
		t.Fatal("did not expect hidden group routes to appear in the spec")
	}

	tags := spec["tags"].([]any)
	foundPublic := false
	foundInternal := false
	foundHidden := false
	foundVisibleGroup := false
	foundHiddenGroup := false
	for _, raw := range tags {
		tag := raw.(map[string]any)
		name, _ := tag["name"].(string)
		switch name {
		case "public":
			foundPublic = true
		case "internal":
			foundInternal = true
		case "hidden":
			foundHidden = true
		case "visible":
			foundVisibleGroup = true
		case "hidden-group":
			foundHiddenGroup = true
		}
	}

	if !foundPublic {
		t.Fatal("expected public tag to remain in the spec")
	}
	if !foundVisibleGroup {
		t.Fatal("expected visible group tag to remain in the spec")
	}
	if foundInternal {
		t.Fatal("did not expect hidden route tag to appear in the spec")
	}
	if foundHidden {
		t.Fatal("did not expect hidden group route tag to appear in the spec")
	}
	if foundHiddenGroup {
		t.Fatal("did not expect hidden group tag to appear in the spec")
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

	tags, ok := spec["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags array, got %T", spec["tags"])
	}
	expectedTags := map[string]bool{
		"system":  true,
		"tools":   true,
		"files":   true,
		"metrics": true,
	}
	for _, raw := range tags {
		tag := raw.(map[string]any)
		name, _ := tag["name"].(string)
		delete(expectedTags, name)
	}
	if len(expectedTags) != 0 {
		t.Fatalf("expected declared tags to include system, tools, files, and metrics, missing %v", expectedTags)
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

func TestSpecBuilder_Good_Servers(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
		Servers: []string{
			" https://api.example.com ",
			"/",
			"",
			"https://api.example.com",
		},
	}

	data, err := sb.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	servers, ok := spec["servers"].([]any)
	if !ok {
		t.Fatalf("expected servers array, got %T", spec["servers"])
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

func TestSpecBuilder_Good_ServersCollapseTrailingSlashes(t *testing.T) {
	sb := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
		Servers: []string{
			"https://api.example.com/",
			"https://api.example.com",
			"/api/",
			"/api",
		},
	}

	data, err := sb.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	servers, ok := spec["servers"].([]any)
	if !ok {
		t.Fatalf("expected servers array, got %T", spec["servers"])
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 collapsed servers, got %d", len(servers))
	}

	first := servers[0].(map[string]any)
	if first["url"] != "https://api.example.com" {
		t.Fatalf("expected first server url=%q, got %v", "https://api.example.com", first["url"])
	}
	second := servers[1].(map[string]any)
	if second["url"] != "/api" {
		t.Fatalf("expected second server url=%q, got %v", "/api", second["url"])
	}
}
