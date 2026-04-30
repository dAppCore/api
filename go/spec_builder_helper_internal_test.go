// SPDX-License-Identifier: EUPL-1.2

package api

import "testing"

func TestEngine_SwaggerConfig_Good_NormalisesPathAtSnapshot(t *testing.T) {
	e := &Engine{
		swaggerPath: "  /docs/ ",
	}

	cfg := e.SwaggerConfig()
	if cfg.Path != "/docs" {
		t.Fatalf("expected normalised Swagger path /docs, got %q", cfg.Path)
	}
}

func TestEngine_TransportConfig_Good_NormalisesGraphQLPathAtSnapshot(t *testing.T) {
	e := &Engine{
		graphql: &graphqlConfig{
			path: "  /gql/ ",
		},
	}

	cfg := e.TransportConfig()
	if cfg.GraphQLPath != "/gql" {
		t.Fatalf("expected normalised GraphQL path /gql, got %q", cfg.GraphQLPath)
	}
}
