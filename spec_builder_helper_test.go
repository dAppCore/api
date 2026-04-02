// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
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
	if err := json.Unmarshal(data, &spec); err != nil {
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
	if err := json.Unmarshal(data, &spec); err != nil {
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
	if err := json.Unmarshal(data, &spec); err != nil {
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
	if err := json.Unmarshal(data, &spec); err != nil {
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
	if err := json.Unmarshal(data, &spec); err != nil {
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
