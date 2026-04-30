// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// ── Stub implementations ────────────────────────────────────────────────

type stubGroup struct{}

func (s *stubGroup) Name() string     { return "stub" }
func (s *stubGroup) BasePath() string { return "/stub" }
func (s *stubGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("pong"))
	})
}

type stubStreamGroup struct {
	stubGroup
}

func (s *stubStreamGroup) Channels() []string {
	return []string{"events", "updates"}
}

// ── RouteGroup interface ────────────────────────────────────────────────

func TestRouteGroup_Good_InterfaceSatisfaction(t *testing.T) {
	var g api.RouteGroup = &stubGroup{}

	if g.Name() != "stub" {
		t.Fatalf("expected Name=%q, got %q", "stub", g.Name())
	}
	if g.BasePath() != "/stub" {
		t.Fatalf("expected BasePath=%q, got %q", "/stub", g.BasePath())
	}
}

func TestRouteGroup_Good_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	g := &stubGroup{}
	rg := engine.Group(g.BasePath())
	g.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

// ── StreamGroup interface ───────────────────────────────────────────────

func TestStreamGroup_Good_InterfaceSatisfaction(t *testing.T) {
	var g api.StreamGroup = &stubStreamGroup{}

	channels := g.Channels()
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
	if channels[0] != "events" {
		t.Fatalf("expected channels[0]=%q, got %q", "events", channels[0])
	}
	if channels[1] != "updates" {
		t.Fatalf("expected channels[1]=%q, got %q", "updates", channels[1])
	}
}

func TestStreamGroup_Good_AlsoSatisfiesRouteGroup(t *testing.T) {
	sg := &stubStreamGroup{}

	// A StreamGroup's embedded stubGroup should also satisfy RouteGroup.
	var rg api.RouteGroup = sg
	if rg.Name() != "stub" {
		t.Fatalf("expected Name=%q, got %q", "stub", rg.Name())
	}
}

// ── DescribableGroup interface ────────────────────────────────────────

// describableStub implements DescribableGroup for testing.
type describableStub struct {
	stubGroup
	descriptions []api.RouteDescription
}

func (d *describableStub) Describe() []api.RouteDescription {
	return d.descriptions
}

func TestDescribableGroup_Good_ImplementsRouteGroup(t *testing.T) {
	stub := &describableStub{}

	// Must satisfy DescribableGroup.
	var dg api.DescribableGroup = stub
	if dg.Name() != "stub" {
		t.Fatalf("expected Name=%q, got %q", "stub", dg.Name())
	}

	// Must also satisfy RouteGroup since DescribableGroup embeds it.
	var rg api.RouteGroup = stub
	if rg.BasePath() != "/stub" {
		t.Fatalf("expected BasePath=%q, got %q", "/stub", rg.BasePath())
	}
}

func TestDescribableGroup_Good_DescribeReturnsRoutes(t *testing.T) {
	stub := &describableStub{
		descriptions: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/items",
				Summary: "List items",
				Tags:    []string{"items"},
			},
			{
				Method:  "POST",
				Path:    "/items",
				Summary: "Create item",
				Tags:    []string{"items"},
				RequestBody: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
					},
				},
			},
		},
	}

	var dg api.DescribableGroup = stub
	descs := dg.Describe()

	if len(descs) != 2 {
		t.Fatalf("expected 2 descriptions, got %d", len(descs))
	}
	if descs[0].Method != "GET" {
		t.Fatalf("expected descs[0].Method=%q, got %q", "GET", descs[0].Method)
	}
	if descs[0].Summary != "List items" {
		t.Fatalf("expected descs[0].Summary=%q, got %q", "List items", descs[0].Summary)
	}
	if descs[1].Method != "POST" {
		t.Fatalf("expected descs[1].Method=%q, got %q", "POST", descs[1].Method)
	}
	if descs[1].RequestBody == nil {
		t.Fatal("expected descs[1].RequestBody to be non-nil")
	}
}

func TestDescribableGroup_Good_EmptyDescribe(t *testing.T) {
	stub := &describableStub{
		descriptions: nil,
	}

	var dg api.DescribableGroup = stub
	descs := dg.Describe()

	if descs != nil {
		t.Fatalf("expected nil descriptions, got %v", descs)
	}
}

func TestDescribableGroup_Good_MultipleVerbs(t *testing.T) {
	stub := &describableStub{
		descriptions: []api.RouteDescription{
			{Method: "GET", Path: "/resources", Summary: "List resources"},
			{Method: "POST", Path: "/resources", Summary: "Create resource"},
			{Method: "DELETE", Path: "/resources/:id", Summary: "Delete resource"},
		},
	}

	var dg api.DescribableGroup = stub
	descs := dg.Describe()

	if len(descs) != 3 {
		t.Fatalf("expected 3 descriptions, got %d", len(descs))
	}

	expected := []string{"GET", "POST", "DELETE"}
	for i, want := range expected {
		if descs[i].Method != want {
			t.Fatalf("expected descs[%d].Method=%q, got %q", i, want, descs[i].Method)
		}
	}
}

func TestDescribableGroup_Bad_NilSchemas(t *testing.T) {
	stub := &describableStub{
		descriptions: []api.RouteDescription{
			{
				Method:      "GET",
				Path:        "/health",
				Summary:     "Health check",
				RequestBody: nil,
				Response:    nil,
			},
		},
	}

	var dg api.DescribableGroup = stub
	descs := dg.Describe()

	if len(descs) != 1 {
		t.Fatalf("expected 1 description, got %d", len(descs))
	}
	if descs[0].RequestBody != nil {
		t.Fatalf("expected nil RequestBody, got %v", descs[0].RequestBody)
	}
	if descs[0].Response != nil {
		t.Fatalf("expected nil Response, got %v", descs[0].Response)
	}
}
