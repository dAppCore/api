// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"testing"

	"github.com/gin-gonic/gin"
)

type swaggerSnapshotGroup struct {
	name     string
	basePath string
	descs    []RouteDescription
}

func (g *swaggerSnapshotGroup) Name() string                      { return g.name }
func (g *swaggerSnapshotGroup) BasePath() string                  { return g.basePath }
func (g *swaggerSnapshotGroup) RegisterRoutes(_ *gin.RouterGroup) {}
func (g *swaggerSnapshotGroup) Describe() []RouteDescription {
	return g.descs
}

func TestSwaggerSpec_ReadDoc_Good_SnapshotsGroups(t *testing.T) {
	original := &swaggerSnapshotGroup{
		name:     "first",
		basePath: "/first",
		descs: []RouteDescription{
			{
				Method:  "GET",
				Path:    "/ping",
				Summary: "First group",
				Response: map[string]any{
					"type": "string",
				},
			},
		},
	}
	replacement := &swaggerSnapshotGroup{
		name:     "second",
		basePath: "/second",
		descs: []RouteDescription{
			{
				Method:  "GET",
				Path:    "/pong",
				Summary: "Second group",
				Response: map[string]any{
					"type": "string",
				},
			},
		},
	}

	groups := []RouteGroup{original}
	spec := newSwaggerSpec(&SpecBuilder{
		Title:   "Test",
		Version: "1.0.0",
	}, groups)

	groups[0] = replacement

	var doc map[string]any
	if err := coreJSONUnmarshal([]byte(spec.ReadDoc()), &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	paths := doc["paths"].(map[string]any)
	if _, ok := paths["/first/ping"]; !ok {
		t.Fatal("expected original group path to remain in the spec")
	}
	if _, ok := paths["/second/pong"]; ok {
		t.Fatal("did not expect mutated group path to leak into the spec")
	}
}
