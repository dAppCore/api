// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"dappco.re/go/api/internal/stdcompat/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

type describableSpecGroup struct {
	name     string
	basePath string
	descs    []api.RouteDescription
}

func (g *describableSpecGroup) Name() string                       { return g.name }
func (g *describableSpecGroup) BasePath() string                   { return g.basePath }
func (g *describableSpecGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (g *describableSpecGroup) Describe() []api.RouteDescription   { return g.descs }

type describableHandler struct {
	desc            api.RouteDescription
	operationID     string
	tags            []string
	summary         string
	longDescription string
}

func (h *describableHandler) Describe() api.RouteDescription {
	if h == nil {
		return api.RouteDescription{}
	}
	return h.desc
}

func (h *describableHandler) OperationID() string {
	if h == nil {
		return ""
	}
	return h.operationID
}

func (h *describableHandler) Tags() []string {
	if h == nil {
		return nil
	}
	return h.tags
}

func (h *describableHandler) Summary() string {
	if h == nil {
		return ""
	}
	return h.summary
}

func (h *describableHandler) Description() string {
	if h == nil {
		return ""
	}
	return h.longDescription
}

func buildDescribableOperation(t *testing.T, group api.RouteGroup, path, method string) map[string]any {
	t.Helper()

	builder := &api.SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}

	data, err := builder.Build([]api.RouteGroup{group})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := spec["paths"].(map[string]any)
	pathItem, ok := paths[path].(map[string]any)
	if !ok {
		t.Fatalf("expected path %q in spec", path)
	}

	operation, ok := pathItem[method].(map[string]any)
	if !ok {
		t.Fatalf("expected %s operation on %q", method, path)
	}

	return operation
}

func TestDescribable_Good_HandlerMetadataFlowsToOpenAPI(t *testing.T) {
	handler := &describableHandler{
		desc: api.RouteDescription{
			StatusCode: http.StatusCreated,
			RequestBody: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
			Response: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "string"},
				},
			},
		},
		operationID:     "widgets_create",
		tags:            []string{"widgets", "catalog"},
		summary:         "Create widget",
		longDescription: "Creates a widget and returns the stored record.",
	}

	group := &describableSpecGroup{
		name:     "widgets",
		basePath: "/api/widgets",
		descs: []api.RouteDescription{
			{
				Method:  http.MethodPost,
				Path:    "/",
				Handler: handler,
			},
		},
	}

	operation := buildDescribableOperation(t, group, "/api/widgets", "post")

	if got := operation["operationId"]; got != "widgets_create" {
		t.Fatalf("expected handler operationId, got %v", got)
	}
	if got := operation["summary"]; got != "Create widget" {
		t.Fatalf("expected handler summary, got %v", got)
	}
	if got := operation["description"]; got != "Creates a widget and returns the stored record." {
		t.Fatalf("expected handler description, got %v", got)
	}

	tags, ok := operation["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags array, got %T", operation["tags"])
	}
	if len(tags) != 2 || tags[0] != "widgets" || tags[1] != "catalog" {
		t.Fatalf("expected handler tags, got %v", tags)
	}

	requestBody := operation["requestBody"].(map[string]any)
	content := requestBody["content"].(map[string]any)
	schema := content["application/json"].(map[string]any)["schema"].(map[string]any)
	properties := schema["properties"].(map[string]any)
	if _, ok := properties["name"]; !ok {
		t.Fatal("expected request body schema from handler Describe")
	}

	responses := operation["responses"].(map[string]any)
	if _, ok := responses["201"]; !ok {
		t.Fatal("expected status code from handler Describe")
	}
}

func TestDescribable_Bad_MissingHandlerMetadataFallsBackSafely(t *testing.T) {
	group := &describableSpecGroup{
		name:     "widgets",
		basePath: "/api/widgets",
		descs: []api.RouteDescription{
			{
				Method:      http.MethodGet,
				Path:        "/status",
				Summary:     "Widget status",
				Description: "Returns widget availability.",
				Tags:        []string{"status"},
				Handler:     &describableHandler{},
			},
		},
	}

	operation := buildDescribableOperation(t, group, "/api/widgets/status", "get")

	if got := operation["operationId"]; got != "get_api_widgets_status" {
		t.Fatalf("expected generated operationId fallback, got %v", got)
	}
	if got := operation["summary"]; got != "Widget status" {
		t.Fatalf("expected route summary fallback, got %v", got)
	}
	if got := operation["description"]; got != "Returns widget availability." {
		t.Fatalf("expected route description fallback, got %v", got)
	}

	tags, ok := operation["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags array, got %T", operation["tags"])
	}
	if len(tags) != 1 || tags[0] != "status" {
		t.Fatalf("expected route tag fallback, got %v", tags)
	}
}

func TestDescribable_Ugly_NilHandlerIsIgnored(t *testing.T) {
	group := &describableSpecGroup{
		name:     "widgets",
		basePath: "/api/widgets",
		descs: []api.RouteDescription{
			{
				Method:  http.MethodGet,
				Path:    "/status",
				Handler: (*describableHandler)(nil),
			},
		},
	}

	operation := buildDescribableOperation(t, group, "/api/widgets/status", "get")

	if got := operation["operationId"]; got != "get_api_widgets_status" {
		t.Fatalf("expected generated operationId with nil handler, got %v", got)
	}

	tags, ok := operation["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags array, got %T", operation["tags"])
	}
	if len(tags) != 1 || tags[0] != "widgets" {
		t.Fatalf("expected group-name tag fallback, got %v", tags)
	}
}
