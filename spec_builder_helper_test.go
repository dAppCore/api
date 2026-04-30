// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"slices"

	api "dappco.re/go/api"
)

func TestEngine_Good_OpenAPISpecBuilderCarriesEngineMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(
		api.WithSwagger("Engine API", "Engine metadata", "2.0.0"),
		api.WithSwaggerSummary("Engine overview"),
		api.WithSwaggerPath("/docs"),
		api.WithSwaggerTermsOfService("https://example.com/terms"),
		api.WithSwaggerContact("API Support", "https://example.com/support", "support@example.com"),
		api.WithSwaggerServers("https://api.example.com", "/", "https://api.example.com"),
		api.WithSwaggerLicense("EUPL-1.2", "https://eupl.eu/1.2/en/"),
		api.WithSwaggerSecuritySchemes(map[string]any{
			"apiKeyAuth": map[string]any{
				"type": "apiKey",
				"in":   "header",
				"name": "X-API-Key",
			},
		}),
		api.WithSwaggerExternalDocs("Developer guide", "https://example.com/docs"),
		api.WithCacheLimits(5*time.Minute, 42, 8192),
		api.WithI18n(api.I18nConfig{
			DefaultLocale: "en-GB",
			Supported:     []string{"en-GB", "fr"},
		}),
		api.WithAuthentik(api.AuthentikConfig{
			Issuer:       "https://auth.example.com",
			ClientID:     "core-client",
			TrustedProxy: true,
			PublicPaths:  []string{" /public/ ", "docs", "/public"},
		}),
		api.WithWSPath("/socket"),
		api.WithWSHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
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
	if err := coreJSONUnmarshal(data, &spec); err != nil {
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
	if info["summary"] != "Engine overview" {
		t.Fatalf("expected summary Engine overview, got %v", info["summary"])
	}

	if got := spec["x-swagger-ui-path"]; got != "/docs" {
		t.Fatalf("expected x-swagger-ui-path=/docs, got %v", got)
	}
	if got := spec["x-swagger-enabled"]; got != true {
		t.Fatalf("expected x-swagger-enabled=true, got %v", got)
	}
	if got := spec["x-graphql-enabled"]; got != true {
		t.Fatalf("expected x-graphql-enabled=true, got %v", got)
	}
	if got := spec["x-graphql-path"]; got != "/gql" {
		t.Fatalf("expected x-graphql-path=/gql, got %v", got)
	}
	if got := spec["x-graphql-playground"]; got != true {
		t.Fatalf("expected x-graphql-playground=true, got %v", got)
	}
	if got := spec["x-graphql-playground-path"]; got != "/gql/playground" {
		t.Fatalf("expected x-graphql-playground-path=/gql/playground, got %v", got)
	}
	if got := spec["x-ws-path"]; got != "/socket" {
		t.Fatalf("expected x-ws-path=/socket, got %v", got)
	}
	if got := spec["x-ws-enabled"]; got != true {
		t.Fatalf("expected x-ws-enabled=true, got %v", got)
	}
	if got := spec["x-sse-path"]; got != "/events" {
		t.Fatalf("expected x-sse-path=/events, got %v", got)
	}
	if got := spec["x-sse-enabled"]; got != true {
		t.Fatalf("expected x-sse-enabled=true, got %v", got)
	}
	if got := spec["x-pprof-enabled"]; got != true {
		t.Fatalf("expected x-pprof-enabled=true, got %v", got)
	}
	if got := spec["x-expvar-enabled"]; got != true {
		t.Fatalf("expected x-expvar-enabled=true, got %v", got)
	}
	if got := spec["x-cache-enabled"]; got != true {
		t.Fatalf("expected x-cache-enabled=true, got %v", got)
	}
	if got := spec["x-cache-ttl"]; got != "5m0s" {
		t.Fatalf("expected x-cache-ttl=5m0s, got %v", got)
	}
	if got := spec["x-cache-max-entries"]; got != float64(42) {
		t.Fatalf("expected x-cache-max-entries=42, got %v", got)
	}
	if got := spec["x-cache-max-bytes"]; got != float64(8192) {
		t.Fatalf("expected x-cache-max-bytes=8192, got %v", got)
	}
	if got := spec["x-i18n-default-locale"]; got != "en-GB" {
		t.Fatalf("expected x-i18n-default-locale=en-GB, got %v", got)
	}
	locales, ok := spec["x-i18n-supported-locales"].([]any)
	if !ok {
		t.Fatalf("expected x-i18n-supported-locales array, got %T", spec["x-i18n-supported-locales"])
	}
	if len(locales) != 2 || locales[0] != "en-GB" || locales[1] != "fr" {
		t.Fatalf("expected supported locales [en-GB fr], got %v", locales)
	}
	if got := spec["x-authentik-issuer"]; got != "https://auth.example.com" {
		t.Fatalf("expected x-authentik-issuer=https://auth.example.com, got %v", got)
	}
	if got := spec["x-authentik-client-id"]; got != "core-client" {
		t.Fatalf("expected x-authentik-client-id=core-client, got %v", got)
	}
	if got := spec["x-authentik-trusted-proxy"]; got != true {
		t.Fatalf("expected x-authentik-trusted-proxy=true, got %v", got)
	}
	publicPaths, ok := spec["x-authentik-public-paths"].([]any)
	if !ok {
		t.Fatalf("expected x-authentik-public-paths array, got %T", spec["x-authentik-public-paths"])
	}
	if len(publicPaths) != 4 || publicPaths[0] != "/health" || publicPaths[1] != "/swagger" || publicPaths[2] != "/docs" || publicPaths[3] != "/public" {
		t.Fatalf("expected public paths [/health /swagger /docs /public], got %v", publicPaths)
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

	securitySchemes, ok := spec["components"].(map[string]any)["securitySchemes"].(map[string]any)
	if !ok {
		t.Fatal("expected securitySchemes metadata in generated spec")
	}
	apiKeyAuth, ok := securitySchemes["apiKeyAuth"].(map[string]any)
	if !ok {
		t.Fatal("expected apiKeyAuth security scheme in generated spec")
	}
	if apiKeyAuth["type"] != "apiKey" {
		t.Fatalf("expected apiKeyAuth.type=apiKey, got %v", apiKeyAuth["type"])
	}
	if apiKeyAuth["in"] != "header" {
		t.Fatalf("expected apiKeyAuth.in=header, got %v", apiKeyAuth["in"])
	}
	if apiKeyAuth["name"] != "X-API-Key" {
		t.Fatalf("expected apiKeyAuth.name=X-API-Key, got %v", apiKeyAuth["name"])
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
	if _, ok := paths["/socket"]; !ok {
		t.Fatal("expected custom WebSocket path from engine metadata in generated spec")
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

func TestEngine_Good_SwaggerConfigCarriesEngineMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("Engine API", "Engine metadata", "2.0.0"),
		api.WithSwaggerSummary("Engine overview"),
		api.WithSwaggerTermsOfService("https://example.com/terms"),
		api.WithSwaggerContact("API Support", "https://example.com/support", "support@example.com"),
		api.WithSwaggerServers("https://api.example.com", "/", "https://api.example.com"),
		api.WithSwaggerLicense("EUPL-1.2", "https://eupl.eu/1.2/en/"),
		api.WithSwaggerSecuritySchemes(map[string]any{
			"apiKeyAuth": map[string]any{
				"type": "apiKey",
				"in":   "header",
				"name": "X-API-Key",
			},
		}),
		api.WithSwaggerExternalDocs("Developer guide", "https://example.com/docs"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.SwaggerConfig()
	if !cfg.Enabled {
		t.Fatal("expected Swagger to be enabled")
	}
	if cfg.Path != "" {
		t.Fatalf("expected empty Swagger path when none is configured, got %q", cfg.Path)
	}
	if cfg.Title != "Engine API" {
		t.Fatalf("expected title Engine API, got %q", cfg.Title)
	}
	if cfg.Description != "Engine metadata" {
		t.Fatalf("expected description Engine metadata, got %q", cfg.Description)
	}
	if cfg.Version != "2.0.0" {
		t.Fatalf("expected version 2.0.0, got %q", cfg.Version)
	}
	if cfg.Summary != "Engine overview" {
		t.Fatalf("expected summary Engine overview, got %q", cfg.Summary)
	}
	if cfg.TermsOfService != "https://example.com/terms" {
		t.Fatalf("expected termsOfService to be preserved, got %q", cfg.TermsOfService)
	}
	if cfg.ContactName != "API Support" {
		t.Fatalf("expected contact name API Support, got %q", cfg.ContactName)
	}
	if cfg.LicenseName != "EUPL-1.2" {
		t.Fatalf("expected licence name EUPL-1.2, got %q", cfg.LicenseName)
	}
	if cfg.ExternalDocsURL != "https://example.com/docs" {
		t.Fatalf("expected external docs URL https://example.com/docs, got %q", cfg.ExternalDocsURL)
	}
	if len(cfg.Servers) != 2 {
		t.Fatalf("expected 2 normalised servers, got %d", len(cfg.Servers))
	}
	if cfg.Servers[0] != "https://api.example.com" {
		t.Fatalf("expected first server to be https://api.example.com, got %q", cfg.Servers[0])
	}
	if cfg.Servers[1] != "/" {
		t.Fatalf("expected second server to be /, got %q", cfg.Servers[1])
	}

	cfgWithPath, err := api.New(
		api.WithSwagger("Engine API", "Engine metadata", "2.0.0"),
		api.WithSwaggerPath("/docs"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := cfgWithPath.SwaggerConfig()
	if snap.Path != "/docs" {
		t.Fatalf("expected Swagger path /docs, got %q", snap.Path)
	}

	apiKeyAuth, ok := cfg.SecuritySchemes["apiKeyAuth"].(map[string]any)
	if !ok {
		t.Fatal("expected apiKeyAuth security scheme in Swagger config")
	}
	if apiKeyAuth["name"] != "X-API-Key" {
		t.Fatalf("expected apiKeyAuth.name=X-API-Key, got %v", apiKeyAuth["name"])
	}

	cfg.Servers[0] = "https://mutated.example.com"
	apiKeyAuth["name"] = "Changed"

	reshot := e.SwaggerConfig()
	if reshot.Servers[0] != "https://api.example.com" {
		t.Fatalf("expected engine servers to be cloned, got %q", reshot.Servers[0])
	}
	reshotScheme, ok := reshot.SecuritySchemes["apiKeyAuth"].(map[string]any)
	if !ok {
		t.Fatal("expected apiKeyAuth security scheme in cloned Swagger config")
	}
	if reshotScheme["name"] != "X-API-Key" {
		t.Fatalf("expected cloned security scheme name X-API-Key, got %v", reshotScheme["name"])
	}
}

func TestEngine_Good_SwaggerConfigTrimsRuntimeMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithSwagger("  Engine API  ", "  Engine metadata  ", "  2.0.0  "),
		api.WithSwaggerSummary("  Engine overview  "),
		api.WithSwaggerTermsOfService("  https://example.com/terms  "),
		api.WithSwaggerContact("  API Support  ", "  https://example.com/support  ", "  support@example.com  "),
		api.WithSwaggerLicense("  EUPL-1.2  ", "  https://eupl.eu/1.2/en/  "),
		api.WithSwaggerExternalDocs("  Developer guide  ", "  https://example.com/docs  "),
		api.WithAuthentik(api.AuthentikConfig{
			Issuer:       "  https://auth.example.com  ",
			ClientID:     "  core-client  ",
			TrustedProxy: true,
			PublicPaths:  []string{" /public/ ", " docs ", "/public"},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	swagger := e.SwaggerConfig()
	if swagger.Title != "Engine API" {
		t.Fatalf("expected trimmed title Engine API, got %q", swagger.Title)
	}
	if swagger.Description != "Engine metadata" {
		t.Fatalf("expected trimmed description Engine metadata, got %q", swagger.Description)
	}
	if swagger.Version != "2.0.0" {
		t.Fatalf("expected trimmed version 2.0.0, got %q", swagger.Version)
	}
	if swagger.Summary != "Engine overview" {
		t.Fatalf("expected trimmed summary Engine overview, got %q", swagger.Summary)
	}
	if swagger.TermsOfService != "https://example.com/terms" {
		t.Fatalf("expected trimmed termsOfService, got %q", swagger.TermsOfService)
	}
	if swagger.ContactName != "API Support" || swagger.ContactURL != "https://example.com/support" || swagger.ContactEmail != "support@example.com" {
		t.Fatalf("expected trimmed contact metadata, got %+v", swagger)
	}
	if swagger.LicenseName != "EUPL-1.2" || swagger.LicenseURL != "https://eupl.eu/1.2/en/" {
		t.Fatalf("expected trimmed licence metadata, got %+v", swagger)
	}
	if swagger.ExternalDocsDescription != "Developer guide" || swagger.ExternalDocsURL != "https://example.com/docs" {
		t.Fatalf("expected trimmed external docs metadata, got %+v", swagger)
	}

	auth := e.AuthentikConfig()
	if auth.Issuer != "https://auth.example.com" {
		t.Fatalf("expected trimmed issuer, got %q", auth.Issuer)
	}
	if auth.ClientID != "core-client" {
		t.Fatalf("expected trimmed client ID, got %q", auth.ClientID)
	}
	if want := []string{"/public", "/docs"}; !slices.Equal(auth.PublicPaths, want) {
		t.Fatalf("expected trimmed public paths %v, got %v", want, auth.PublicPaths)
	}

	builder := e.OpenAPISpecBuilder()
	data, err := builder.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := coreJSONUnmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in generated spec")
	}
	if info["title"] != "Engine API" || info["description"] != "Engine metadata" || info["version"] != "2.0.0" || info["summary"] != "Engine overview" {
		t.Fatalf("expected trimmed OpenAPI info block, got %+v", info)
	}
	if info["termsOfService"] != "https://example.com/terms" {
		t.Fatalf("expected trimmed termsOfService in spec, got %v", info["termsOfService"])
	}
}

func TestEngine_Good_TransportConfigCarriesEngineMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(
		api.WithSwagger("Engine API", "Engine metadata", "2.0.0"),
		api.WithSwaggerPath("/docs"),
		api.WithWSPath("/socket"),
		api.WithWSHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
		api.WithGraphQL(newTestSchema(), api.WithPlayground(), api.WithGraphQLPath("/gql")),
		api.WithSSE(broker),
		api.WithSSEPath("/events"),
		api.WithPprof(),
		api.WithExpvar(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.TransportConfig()
	if !cfg.SwaggerEnabled {
		t.Fatal("expected Swagger to be enabled")
	}
	if cfg.SwaggerPath != "/docs" {
		t.Fatalf("expected swagger path /docs, got %q", cfg.SwaggerPath)
	}
	if cfg.GraphQLPath != "/gql" {
		t.Fatalf("expected graphql path /gql, got %q", cfg.GraphQLPath)
	}
	if !cfg.GraphQLEnabled {
		t.Fatal("expected GraphQL to be enabled")
	}
	if !cfg.GraphQLPlayground {
		t.Fatal("expected GraphQL playground to be enabled")
	}
	if !cfg.WSEnabled {
		t.Fatal("expected WebSocket to be enabled")
	}
	if cfg.WSPath != "/socket" {
		t.Fatalf("expected ws path /socket, got %q", cfg.WSPath)
	}
	if !cfg.SSEEnabled {
		t.Fatal("expected SSE to be enabled")
	}
	if cfg.SSEPath != "/events" {
		t.Fatalf("expected sse path /events, got %q", cfg.SSEPath)
	}
	if !cfg.PprofEnabled {
		t.Fatal("expected pprof to be enabled")
	}
	if !cfg.ExpvarEnabled {
		t.Fatal("expected expvar to be enabled")
	}
}

func TestEngine_Good_TransportConfigReportsDisabledSwaggerWithoutUI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwaggerPath("/docs"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.TransportConfig()
	if cfg.SwaggerEnabled {
		t.Fatal("expected Swagger to remain disabled when only the path is configured")
	}
	if cfg.SwaggerPath != "/docs" {
		t.Fatalf("expected swagger path /docs, got %q", cfg.SwaggerPath)
	}
}

// TestEngine_Good_TransportConfigReportsChatCompletions verifies that the
// chat completions resolver surfaces through TransportConfig so callers can
// discover the RFC §11.1 endpoint without rebuilding the engine.
func TestEngine_Good_TransportConfigReportsChatCompletions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := api.NewModelResolver()
	e, err := api.New(api.WithChatCompletions(resolver))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.TransportConfig()
	if !cfg.ChatCompletionsEnabled {
		t.Fatal("expected chat completions to be enabled")
	}
	if cfg.ChatCompletionsPath != "/v1/chat/completions" {
		t.Fatalf("expected chat completions path /v1/chat/completions, got %q", cfg.ChatCompletionsPath)
	}
}

// TestEngine_Good_TransportConfigHonoursChatCompletionsPathOverride verifies
// that WithChatCompletionsPath surfaces through TransportConfig.
func TestEngine_Good_TransportConfigHonoursChatCompletionsPathOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := api.NewModelResolver()
	e, err := api.New(
		api.WithChatCompletions(resolver),
		api.WithChatCompletionsPath("/chat"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.TransportConfig()
	if cfg.ChatCompletionsPath != "/chat" {
		t.Fatalf("expected chat completions path /chat, got %q", cfg.ChatCompletionsPath)
	}
}

// TestEngine_Good_TransportConfigReportsOpenAPISpec verifies that the
// WithOpenAPISpec option surfaces the standalone JSON endpoint (RFC
// /v1/openapi.json) through TransportConfig so callers can discover it
// alongside the other framework routes.
func TestEngine_Good_TransportConfigReportsOpenAPISpec(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithOpenAPISpec())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.TransportConfig()
	if !cfg.OpenAPISpecEnabled {
		t.Fatal("expected OpenAPISpecEnabled=true")
	}
	if cfg.OpenAPISpecPath != "/v1/openapi.json" {
		t.Fatalf("expected OpenAPISpecPath=/v1/openapi.json, got %q", cfg.OpenAPISpecPath)
	}
}

// TestEngine_Good_TransportConfigHonoursOpenAPISpecPathOverride verifies
// that WithOpenAPISpecPath surfaces through TransportConfig.
func TestEngine_Good_TransportConfigHonoursOpenAPISpecPathOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithOpenAPISpecPath("/api/v1/openapi.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.TransportConfig()
	if !cfg.OpenAPISpecEnabled {
		t.Fatal("expected OpenAPISpecEnabled inferred from path override")
	}
	if cfg.OpenAPISpecPath != "/api/v1/openapi.json" {
		t.Fatalf("expected custom path, got %q", cfg.OpenAPISpecPath)
	}
}

// TestEngine_Bad_TransportConfigOmitsOpenAPISpecWhenDisabled confirms the
// standalone OpenAPI endpoint reports as disabled when neither WithOpenAPISpec
// nor WithOpenAPISpecPath has been invoked.
func TestEngine_Bad_TransportConfigOmitsOpenAPISpecWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.TransportConfig()
	if cfg.OpenAPISpecEnabled {
		t.Fatal("expected OpenAPISpecEnabled=false when not configured")
	}
	if cfg.OpenAPISpecPath != "" {
		t.Fatalf("expected empty OpenAPISpecPath, got %q", cfg.OpenAPISpecPath)
	}
}

// TestEngine_Bad_TransportConfigFallsBackToDefaultOpenAPISpecPathWhenBlank
// verifies that a blank override still enables the endpoint and resolves to
// the RFC default path.
func TestEngine_Bad_TransportConfigFallsBackToDefaultOpenAPISpecPathWhenBlank(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithOpenAPISpecPath("   "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.TransportConfig()
	if !cfg.OpenAPISpecEnabled {
		t.Fatal("expected OpenAPISpecEnabled=true from blank override")
	}
	if cfg.OpenAPISpecPath != "/v1/openapi.json" {
		t.Fatalf("expected default OpenAPISpecPath=/v1/openapi.json, got %q", cfg.OpenAPISpecPath)
	}
}

// TestEngine_Ugly_TransportConfigNormalisesOpenAPISpecPathOverride verifies
// that the custom path override is trimmed and promoted to an absolute
// route path.
func TestEngine_Ugly_TransportConfigNormalisesOpenAPISpecPathOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithOpenAPISpecPath("  api/v1/openapi.json  "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.TransportConfig()
	if !cfg.OpenAPISpecEnabled {
		t.Fatal("expected OpenAPISpecEnabled=true from path override")
	}
	if cfg.OpenAPISpecPath != "/api/v1/openapi.json" {
		t.Fatalf("expected normalised OpenAPISpecPath=/api/v1/openapi.json, got %q", cfg.OpenAPISpecPath)
	}
}

func TestEngine_Good_OpenAPISpecBuilderExportsDefaultSwaggerPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwagger("Engine API", "Engine metadata", "2.0.0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := e.OpenAPISpecBuilder()
	data, err := builder.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := coreJSONUnmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got := spec["x-swagger-ui-path"]; got != "/swagger" {
		t.Fatalf("expected default x-swagger-ui-path=/swagger, got %v", got)
	}
}

func TestEngine_Good_OpenAPISpecBuilderCarriesExplicitSwaggerPathWithoutUI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwaggerPath("/docs"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := e.OpenAPISpecBuilder()
	data, err := builder.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := coreJSONUnmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got := spec["x-swagger-ui-path"]; got != "/docs" {
		t.Fatalf("expected explicit x-swagger-ui-path=/docs, got %v", got)
	}
}

func TestEngine_Good_OpenAPISpecBuilderCarriesConfiguredWSPathWithoutHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithWSPath("/socket"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := e.OpenAPISpecBuilder()
	data, err := builder.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := coreJSONUnmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got := spec["x-ws-path"]; got != "/socket" {
		t.Fatalf("expected x-ws-path=/socket, got %v", got)
	}
}

func TestEngine_Good_OpenAPISpecBuilderCarriesConfiguredSSEPathWithoutBroker(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSSE(nil), api.WithSSEPath("/events"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := e.OpenAPISpecBuilder()
	data, err := builder.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := coreJSONUnmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got := spec["x-sse-path"]; got != "/events" {
		t.Fatalf("expected x-sse-path=/events, got %v", got)
	}
}

func TestEngine_Good_OpenAPISpecBuilderClonesSecuritySchemes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	securityScheme := map[string]any{
		"type": "oauth2",
		"flows": map[string]any{
			"clientCredentials": map[string]any{
				"tokenUrl": "https://auth.example.com/token",
			},
		},
	}
	schemes := map[string]any{
		"oauth2": securityScheme,
	}

	e, err := api.New(
		api.WithSwagger("Engine API", "Engine metadata", "2.0.0"),
		api.WithSwaggerSecuritySchemes(schemes),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mutate the original input after configuration. The builder snapshot should
	// remain stable and keep the original token URL.
	securityScheme["type"] = "mutated"
	securityScheme["flows"].(map[string]any)["clientCredentials"].(map[string]any)["tokenUrl"] = "https://mutated.example.com/token"

	data, err := e.OpenAPISpecBuilder().Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := coreJSONUnmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	securitySchemes := spec["components"].(map[string]any)["securitySchemes"].(map[string]any)
	oauth2, ok := securitySchemes["oauth2"].(map[string]any)
	if !ok {
		t.Fatal("expected oauth2 security scheme in generated spec")
	}
	if oauth2["type"] != "oauth2" {
		t.Fatalf("expected cloned oauth2.type=oauth2, got %v", oauth2["type"])
	}
	flows := oauth2["flows"].(map[string]any)
	clientCredentials := flows["clientCredentials"].(map[string]any)
	if clientCredentials["tokenUrl"] != "https://auth.example.com/token" {
		t.Fatalf("expected original tokenUrl to be preserved, got %v", clientCredentials["tokenUrl"])
	}
}

// TestEngine_Ugly_OpenAPISpecBuilderSkipsBlankSecuritySchemeEntries verifies
// that empty or nil security scheme entries are ignored while valid entries
// are cloned into the generated spec.
func TestEngine_Ugly_OpenAPISpecBuilderSkipsBlankSecuritySchemeEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithSwagger("Engine API", "Engine metadata", "2.0.0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api.WithSwaggerSecuritySchemes(nil)(e)
	api.WithSwaggerSecuritySchemes(map[string]any{
		"":     nil,
		"skip": nil,
		"apiKeyAuth": map[string]any{
			"type": "apiKey",
			"in":   "header",
			"name": "X-API-Key",
		},
	})(e)

	data, err := e.OpenAPISpecBuilder().Build(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := coreJSONUnmarshal(data, &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	securitySchemes := spec["components"].(map[string]any)["securitySchemes"].(map[string]any)
	if _, ok := securitySchemes["apiKeyAuth"]; !ok {
		t.Fatalf("expected apiKeyAuth security scheme, got %v", securitySchemes)
	}
	if _, ok := securitySchemes[""]; ok {
		t.Fatalf("expected blank security scheme key to be ignored, got %v", securitySchemes)
	}
	if _, ok := securitySchemes["skip"]; ok {
		t.Fatalf("expected nil security scheme to be ignored, got %v", securitySchemes)
	}
}
