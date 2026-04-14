// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"iter"
	"os"
	"testing"

	core "dappco.re/go/core"
	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

type specCmdStubGroup struct{}

func (specCmdStubGroup) Name() string                       { return "registered" }
func (specCmdStubGroup) BasePath() string                   { return "/registered" }
func (specCmdStubGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (specCmdStubGroup) Describe() []api.RouteDescription {
	return []api.RouteDescription{
		{
			Method:  "GET",
			Path:    "/ping",
			Summary: "Ping registered group",
			Tags:    []string{"registered"},
			Response: map[string]any{
				"type": "string",
			},
		},
	}
}

func collectRouteGroups(groups iter.Seq[api.RouteGroup]) []api.RouteGroup {
	out := make([]api.RouteGroup, 0)
	for group := range groups {
		out = append(out, group)
	}
	return out
}

// TestCmdSpec_AddSpecCommand_Good verifies the spec command registers under
// the expected api/spec path with an executable Action.
func TestCmdSpec_AddSpecCommand_Good(t *testing.T) {
	c := core.New()
	addSpecCommand(c)

	r := c.Command("api/spec")
	if !r.OK {
		t.Fatalf("expected api/spec command to be registered")
	}
	cmd, ok := r.Value.(*core.Command)
	if !ok {
		t.Fatalf("expected *core.Command, got %T", r.Value)
	}
	if cmd.Action == nil {
		t.Fatal("expected non-nil Action on api/spec")
	}
	if cmd.Description == "" {
		t.Fatal("expected Description on api/spec")
	}
}

// TestCmdSpec_SpecAction_Good_WritesJSONToFile exercises the spec action with
// an output file flag and verifies the resulting OpenAPI document parses.
func TestCmdSpec_SpecAction_Good_WritesJSONToFile(t *testing.T) {
	outputFile := t.TempDir() + "/spec.json"
	opts := core.NewOptions(
		core.Option{Key: "output", Value: outputFile},
		core.Option{Key: "format", Value: "json"},
	)

	r := specAction(opts)
	if !r.OK {
		t.Fatalf("expected OK result, got %v", r.Value)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("expected spec file to be written: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("expected valid JSON spec, got error: %v", err)
	}
	if spec["openapi"] == nil {
		t.Fatal("expected openapi field in generated spec")
	}
}

// TestCmdSpec_SpecConfigFromOptions_Good_FlagsArePreserved ensures that flag
// keys from the CLI parser populate the spec builder configuration.
func TestCmdSpec_SpecConfigFromOptions_Good_FlagsArePreserved(t *testing.T) {
	opts := core.NewOptions(
		core.Option{Key: "title", Value: "Custom API"},
		core.Option{Key: "summary", Value: "Brief summary"},
		core.Option{Key: "description", Value: "Long description"},
		core.Option{Key: "version", Value: "9.9.9"},
		core.Option{Key: "swagger-path", Value: "/docs"},
		core.Option{Key: "graphql-playground", Value: true},
		core.Option{Key: "cache", Value: true},
		core.Option{Key: "cache-max-entries", Value: 100},
		core.Option{Key: "i18n-supported-locales", Value: "en-GB,fr"},
	)

	cfg := specConfigFromOptions(opts)

	if cfg.title != "Custom API" {
		t.Fatalf("expected title=Custom API, got %q", cfg.title)
	}
	if cfg.summary != "Brief summary" {
		t.Fatalf("expected summary preserved, got %q", cfg.summary)
	}
	if cfg.version != "9.9.9" {
		t.Fatalf("expected version=9.9.9, got %q", cfg.version)
	}
	if cfg.swaggerPath != "/docs" {
		t.Fatalf("expected swagger path, got %q", cfg.swaggerPath)
	}
	if !cfg.graphqlPlayground {
		t.Fatal("expected graphql playground enabled")
	}
	if !cfg.cacheEnabled {
		t.Fatal("expected cache enabled")
	}
	if cfg.cacheMaxEntries != 100 {
		t.Fatalf("expected cacheMaxEntries=100, got %d", cfg.cacheMaxEntries)
	}
	if cfg.i18nSupportedLocales != "en-GB,fr" {
		t.Fatalf("expected i18n supported locales, got %q", cfg.i18nSupportedLocales)
	}
}

// TestCmdSpec_SpecConfigFromOptions_Good_OpenAPIAndChatFlagsPreserved
// verifies the new spec-level flags for the standalone OpenAPI JSON and
// chat completions endpoints round-trip through the CLI parser.
func TestCmdSpec_SpecConfigFromOptions_Good_OpenAPIAndChatFlagsPreserved(t *testing.T) {
	opts := core.NewOptions(
		core.Option{Key: "openapi-spec", Value: true},
		core.Option{Key: "openapi-spec-path", Value: "/api/v1/openapi.json"},
		core.Option{Key: "chat-completions", Value: true},
		core.Option{Key: "chat-completions-path", Value: "/api/v1/chat/completions"},
	)

	cfg := specConfigFromOptions(opts)

	if !cfg.openAPISpecEnabled {
		t.Fatal("expected openAPISpecEnabled=true")
	}
	if cfg.openAPISpecPath != "/api/v1/openapi.json" {
		t.Fatalf("expected openAPISpecPath=%q, got %q", "/api/v1/openapi.json", cfg.openAPISpecPath)
	}
	if !cfg.chatCompletionsEnabled {
		t.Fatal("expected chatCompletionsEnabled=true")
	}
	if cfg.chatCompletionsPath != "/api/v1/chat/completions" {
		t.Fatalf("expected chatCompletionsPath=%q, got %q", "/api/v1/chat/completions", cfg.chatCompletionsPath)
	}
}

// TestCmdSpec_NewSpecBuilder_Good_PropagatesNewFlags verifies that the
// spec builder respects the new OpenAPI and ChatCompletions flags.
func TestCmdSpec_NewSpecBuilder_Good_PropagatesNewFlags(t *testing.T) {
	cfg := specBuilderConfig{
		title:                  "Test",
		version:                "1.0.0",
		openAPISpecEnabled:     true,
		openAPISpecPath:        "/api/v1/openapi.json",
		chatCompletionsEnabled: true,
		chatCompletionsPath:    "/api/v1/chat/completions",
	}

	builder, err := newSpecBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !builder.OpenAPISpecEnabled {
		t.Fatal("expected OpenAPISpecEnabled=true on builder")
	}
	if builder.OpenAPISpecPath != "/api/v1/openapi.json" {
		t.Fatalf("expected OpenAPISpecPath=%q, got %q", "/api/v1/openapi.json", builder.OpenAPISpecPath)
	}
	if !builder.ChatCompletionsEnabled {
		t.Fatal("expected ChatCompletionsEnabled=true on builder")
	}
	if builder.ChatCompletionsPath != "/api/v1/chat/completions" {
		t.Fatalf("expected ChatCompletionsPath=%q, got %q", "/api/v1/chat/completions", builder.ChatCompletionsPath)
	}
}

// TestCmdSpec_NewSpecBuilder_Ugly_PathImpliesEnabled verifies that supplying
// only a path override turns the endpoint on automatically so callers need
// not pass both flags in CI scripts.
func TestCmdSpec_NewSpecBuilder_Ugly_PathImpliesEnabled(t *testing.T) {
	cfg := specBuilderConfig{
		title:               "Test",
		version:             "1.0.0",
		openAPISpecPath:     "/api/v1/openapi.json",
		chatCompletionsPath: "/api/v1/chat/completions",
	}

	builder, err := newSpecBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !builder.OpenAPISpecEnabled {
		t.Fatal("expected OpenAPISpecEnabled to be inferred from path override")
	}
	if !builder.ChatCompletionsEnabled {
		t.Fatal("expected ChatCompletionsEnabled to be inferred from path override")
	}
}

// TestCmdSpec_SpecConfigFromOptions_Bad_DefaultsApplied ensures empty values
// do not blank out required defaults like title, description, version.
func TestCmdSpec_SpecConfigFromOptions_Bad_DefaultsApplied(t *testing.T) {
	opts := core.NewOptions()
	cfg := specConfigFromOptions(opts)

	if cfg.title != "Lethean Core API" {
		t.Fatalf("expected default title, got %q", cfg.title)
	}
	if cfg.description != "Lethean Core API" {
		t.Fatalf("expected default description, got %q", cfg.description)
	}
	if cfg.version != "1.0.0" {
		t.Fatalf("expected default version, got %q", cfg.version)
	}
}

// TestCmdSpec_StringOr_Ugly_TrimsWhitespaceFallback covers the awkward
// whitespace path where an option is set to whitespace but should still
// fall back to the supplied default.
func TestCmdSpec_StringOr_Ugly_TrimsWhitespaceFallback(t *testing.T) {
	if got := stringOr("   ", "fallback"); got != "fallback" {
		t.Fatalf("expected whitespace to fall back to default, got %q", got)
	}
	if got := stringOr("value", "fallback"); got != "value" {
		t.Fatalf("expected explicit value to win, got %q", got)
	}
	if got := stringOr("", ""); got != "" {
		t.Fatalf("expected empty/empty to remain empty, got %q", got)
	}
}

// TestSpecGroupsIter_Good_DeduplicatesExtraBridge verifies the iterator does
// not emit a duplicate when the registered groups already contain a tool
// bridge with the same base path.
func TestSpecGroupsIter_Good_DeduplicatesExtraBridge(t *testing.T) {
	snapshot := api.RegisteredSpecGroups()
	api.ResetSpecGroups()
	t.Cleanup(func() {
		api.ResetSpecGroups()
		api.RegisterSpecGroups(snapshot...)
	})

	group := specCmdStubGroup{}
	api.RegisterSpecGroups(group)

	groups := collectRouteGroups(specGroupsIter(group))

	if len(groups) != 1 {
		t.Fatalf("expected duplicate extra group to be skipped, got %d groups", len(groups))
	}
	if groups[0].Name() != group.Name() || groups[0].BasePath() != group.BasePath() {
		t.Fatalf("expected original group to be preserved, got %s at %s", groups[0].Name(), groups[0].BasePath())
	}
}
