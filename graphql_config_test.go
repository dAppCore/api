// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

func TestEngine_GraphQLConfig_Good_SnapshotsCurrentSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithGraphQL(newTestSchema(), api.WithPlayground(), api.WithGraphQLPath(" /gql/ ")),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.GraphQLConfig()
	if !cfg.Enabled {
		t.Fatal("expected GraphQL to be enabled")
	}
	if cfg.Path != "/gql" {
		t.Fatalf("expected GraphQL path /gql, got %q", cfg.Path)
	}
	if !cfg.Playground {
		t.Fatal("expected GraphQL playground to be enabled")
	}
}

func TestEngine_GraphQLConfig_Good_EmptyOnNilEngine(t *testing.T) {
	var e *api.Engine

	cfg := e.GraphQLConfig()
	if cfg.Enabled || cfg.Path != "" || cfg.Playground {
		t.Fatalf("expected zero-value GraphQL config, got %+v", cfg)
	}
}
