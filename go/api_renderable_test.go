// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

type renderableSpecGroup struct {
	name     string
	basePath string
	descs    []api.RouteDescription
}

func (g *renderableSpecGroup) Name() string                       { return g.name }
func (g *renderableSpecGroup) BasePath() string                   { return g.basePath }
func (g *renderableSpecGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (g *renderableSpecGroup) Describe() []api.RouteDescription   { return g.descs }

type renderableHandler struct {
	hints api.RenderHints
}

func (h *renderableHandler) Render() api.RenderHints {
	if h == nil {
		return api.RenderHints{}
	}
	return h.hints
}

func buildRenderableOperation(t *testing.T, group api.RouteGroup, path, method string) map[string]any {
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
	if err := coreJSONUnmarshal(data, &spec); err != nil {
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

func TestRenderable_Good_HandlerHintsFlowToOpenAPI(t *testing.T) {
	group := &renderableSpecGroup{
		name:     "widgets",
		basePath: "/api/widgets",
		descs: []api.RouteDescription{
			{
				Method: http.MethodPost,
				Path:   "/",
				Handler: &renderableHandler{
					hints: api.RenderHints{
						Kind: "form",
						Fields: []api.FieldHint{
							{
								Name:     "name",
								Label:    "Name",
								Type:     "text",
								Required: true,
								Validation: map[string]any{
									"minLength": 3,
								},
							},
						},
						Actions: []api.ActionHint{
							{
								Name:    "preview",
								Label:   "Preview",
								Method:  http.MethodGet,
								Variant: "secondary",
							},
						},
					},
				},
			},
		},
	}

	operation := buildRenderableOperation(t, group, "/api/widgets", "post")

	rawHints, ok := operation["x-render-hints"].(map[string]any)
	if !ok {
		t.Fatalf("expected x-render-hints extension, got %T", operation["x-render-hints"])
	}
	if got := rawHints["kind"]; got != "form" {
		t.Fatalf("expected render kind form, got %v", got)
	}

	fields, ok := rawHints["fields"].([]any)
	if !ok || len(fields) != 1 {
		t.Fatalf("expected one render field, got %v", rawHints["fields"])
	}
	field := fields[0].(map[string]any)
	if got := field["name"]; got != "name" {
		t.Fatalf("expected render field name, got %v", got)
	}
	if got := field["required"]; got != true {
		t.Fatalf("expected render field required=true, got %v", got)
	}
	validation := field["validation"].(map[string]any)
	if got := validation["minLength"]; got != float64(3) {
		t.Fatalf("expected validation minLength=3, got %v", got)
	}

	actions, ok := rawHints["actions"].([]any)
	if !ok || len(actions) != 1 {
		t.Fatalf("expected one render action, got %v", rawHints["actions"])
	}
	action := actions[0].(map[string]any)
	if got := action["name"]; got != "preview" {
		t.Fatalf("expected render action name, got %v", got)
	}
}

func TestRenderable_Bad_EmptyHintsAreOmittedSafely(t *testing.T) {
	group := &renderableSpecGroup{
		name:     "widgets",
		basePath: "/api/widgets",
		descs: []api.RouteDescription{
			{
				Method:  http.MethodGet,
				Path:    "/status",
				Handler: &renderableHandler{},
			},
		},
	}

	operation := buildRenderableOperation(t, group, "/api/widgets/status", "get")

	if _, ok := operation["x-render-hints"]; ok {
		t.Fatalf("expected empty render hints to be omitted, got %v", operation["x-render-hints"])
	}
}

func TestRenderable_Ugly_NilHandlerIsIgnored(t *testing.T) {
	group := &renderableSpecGroup{
		name:     "widgets",
		basePath: "/api/widgets",
		descs: []api.RouteDescription{
			{
				Method:  http.MethodGet,
				Path:    "/status",
				Handler: (*renderableHandler)(nil),
			},
		},
	}

	operation := buildRenderableOperation(t, group, "/api/widgets/status", "get")

	if _, ok := operation["x-render-hints"]; ok {
		t.Fatalf("expected nil renderable handler to be ignored, got %v", operation["x-render-hints"])
	}
}
