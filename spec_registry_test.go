// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

type specRegistryStubGroup struct {
	name     string
	basePath string
}

func (g *specRegistryStubGroup) Name() string                       { return g.name }
func (g *specRegistryStubGroup) BasePath() string                   { return g.basePath }
func (g *specRegistryStubGroup) RegisterRoutes(rg *gin.RouterGroup) {}

func TestRegisterSpecGroups_Good_DeduplicatesByIdentity(t *testing.T) {
	snapshot := api.RegisteredSpecGroups()
	api.ResetSpecGroups()
	t.Cleanup(func() {
		api.ResetSpecGroups()
		api.RegisterSpecGroups(snapshot...)
	})

	first := &specRegistryStubGroup{name: "alpha", basePath: "/alpha"}
	second := &specRegistryStubGroup{name: "alpha", basePath: "/alpha"}
	third := &specRegistryStubGroup{name: "beta", basePath: "/beta"}

	api.RegisterSpecGroups(nil, first, second, third, first)

	groups := api.RegisteredSpecGroups()
	if len(groups) != 2 {
		t.Fatalf("expected 2 unique groups, got %d", len(groups))
	}

	if groups[0].Name() != "alpha" || groups[0].BasePath() != "/alpha" {
		t.Fatalf("expected first group to be alpha at /alpha, got %s at %s", groups[0].Name(), groups[0].BasePath())
	}
	if groups[1].Name() != "beta" || groups[1].BasePath() != "/beta" {
		t.Fatalf("expected second group to be beta at /beta, got %s at %s", groups[1].Name(), groups[1].BasePath())
	}
}

func TestRegisterSpecGroups_Good_IteratorReturnsSnapshot(t *testing.T) {
	snapshot := api.RegisteredSpecGroups()
	api.ResetSpecGroups()
	t.Cleanup(func() {
		api.ResetSpecGroups()
		api.RegisterSpecGroups(snapshot...)
	})

	first := &specRegistryStubGroup{name: "alpha", basePath: "/alpha"}
	second := &specRegistryStubGroup{name: "beta", basePath: "/beta"}

	api.RegisterSpecGroups(first)

	iter := api.RegisteredSpecGroupsIter()

	api.RegisterSpecGroups(second)

	var groups []api.RouteGroup
	for group := range iter {
		groups = append(groups, group)
	}

	if len(groups) != 1 {
		t.Fatalf("expected iterator snapshot to contain 1 group, got %d", len(groups))
	}
	if groups[0].Name() != "alpha" || groups[0].BasePath() != "/alpha" {
		t.Fatalf("expected iterator snapshot to preserve alpha at /alpha, got %s at %s", groups[0].Name(), groups[0].BasePath())
	}
}
