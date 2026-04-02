// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

func TestEngine_Good_OpenAPISpecBuilderCarriesEngineMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(
		api.WithSwagger("Engine API", "Engine metadata", "2.0.0"),
		api.WithSwaggerTermsOfService("https://example.com/terms"),
		api.WithSwaggerContact("API Support", "https://example.com/support", "support@example.com"),
		api.WithSwaggerServers("https://api.example.com", "/", "https://api.example.com"),
		api.WithSwaggerLicense("EUPL-1.2", "https://eupl.eu/1.2/en/"),
		api.WithSwaggerExternalDocs("Developer guide", "https://example.com/docs"),
		api.WithGraphQL(newTestSchema(), api.WithPlayground(), api.WithGraphQLPath("/gql")),
		api.WithSSE(broker),
		api.WithSSEPath("/events"),
		api.WithPprof(),
		api.WithExpvar(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := e.OpenAPISpecBuilder()
	data, err := builder.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in generated spec")
	}
	if info["title"] != "Engine API" {
		t.Fatalf("expected title Engine API, got %v", info["title"])
	}
	if info["description"] != "Engine metadata" {
		t.Fatalf("expected description Engine metadata, got %v", info["description"])
	}
	if info["version"] != "2.0.0" {
		t.Fatalf("expected version 2.0.0, got %v", info["version"])
	}

	contact, ok := info["contact"].(map[string]any)
	if !ok {
		t.Fatal("expected contact metadata in generated spec")
	}
	if contact["name"] != "API Support" {
		t.Fatalf("expected contact name API Support, got %v", contact["name"])
	}

	license, ok := info["license"].(map[string]any)
	if !ok {
		t.Fatal("expected licence metadata in generated spec")
	}
	if license["name"] != "EUPL-1.2" {
		t.Fatalf("expected licence name EUPL-1.2, got %v", license["name"])
	}

	if info["termsOfService"] != "https://example.com/terms" {
		t.Fatalf("expected termsOfService to be preserved, got %v", info["termsOfService"])
	}

	externalDocs, ok := spec["externalDocs"].(map[string]any)
	if !ok {
		t.Fatal("expected externalDocs metadata in generated spec")
	}
	if externalDocs["url"] != "https://example.com/docs" {
		t.Fatalf("expected externalDocs url to be preserved, got %v", externalDocs["url"])
	}

	servers, ok := spec["servers"].([]any)
	if !ok {
		t.Fatalf("expected servers array in generated spec, got %T", spec["servers"])
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 normalised servers, got %d", len(servers))
	}
	if servers[0].(map[string]any)["url"] != "https://api.example.com" {
		t.Fatalf("expected first server to be https://api.example.com, got %v", servers[0])
	}
	if servers[1].(map[string]any)["url"] != "/" {
		t.Fatalf("expected second server to be /, got %v", servers[1])
	}

	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatalf("expected paths object in generated spec, got %T", spec["paths"])
	}
	if _, ok := paths["/gql"]; !ok {
		t.Fatal("expected GraphQL path from engine metadata in generated spec")
	}
	if _, ok := paths["/gql/playground"]; !ok {
		t.Fatal("expected GraphQL playground path from engine metadata in generated spec")
	}
	if _, ok := paths["/events"]; !ok {
		t.Fatal("expected SSE path from engine metadata in generated spec")
	}
	if _, ok := paths["/debug/pprof"]; !ok {
		t.Fatal("expected pprof path from engine metadata in generated spec")
	}
	if _, ok := paths["/debug/vars"]; !ok {
		t.Fatal("expected expvar path from engine metadata in generated spec")
	}
}
