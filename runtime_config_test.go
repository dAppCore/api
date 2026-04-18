// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"slices"
	"testing"
	"time"

	api "dappco.re/go/core/api"
)

// TestEngine_RuntimeConfig_Good_SnapshotsCurrentSettings verifies the
// aggregate runtime snapshot mirrors the current engine configuration.
func TestEngine_RuntimeConfig_Good_SnapshotsCurrentSettings(t *testing.T) {
	broker := api.NewSSEBroker()
	e, err := api.New(
		api.WithSwagger("Runtime API", "Runtime snapshot", "1.2.3"),
		api.WithSwaggerPath("/docs"),
		api.WithCacheLimits(5*time.Minute, 10, 1024),
		api.WithGraphQL(newTestSchema(), api.WithPlayground()),
		api.WithI18n(api.I18nConfig{
			DefaultLocale: "en-GB",
			Supported:     []string{"en-GB", "fr"},
		}),
		api.WithWSPath("/socket"),
		api.WithSSE(broker),
		api.WithSSEPath("/events"),
		api.WithAuthentik(api.AuthentikConfig{
			Issuer:       "https://auth.example.com",
			ClientID:     "runtime-client",
			TrustedProxy: true,
			PublicPaths:  []string{"/public", "/docs"},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.RuntimeConfig()

	if !cfg.Swagger.Enabled {
		t.Fatal("expected swagger snapshot to be enabled")
	}
	if cfg.Swagger.Path != "/docs" {
		t.Fatalf("expected swagger path /docs, got %q", cfg.Swagger.Path)
	}
	if cfg.Transport.SwaggerPath != "/docs" {
		t.Fatalf("expected transport swagger path /docs, got %q", cfg.Transport.SwaggerPath)
	}
	if cfg.Transport.GraphQLPlaygroundPath != "/graphql/playground" {
		t.Fatalf("expected transport graphql playground path /graphql/playground, got %q", cfg.Transport.GraphQLPlaygroundPath)
	}
	if !cfg.Cache.Enabled || cfg.Cache.TTL != 5*time.Minute {
		t.Fatalf("expected cache snapshot to be populated, got %+v", cfg.Cache)
	}
	if !cfg.GraphQL.Enabled {
		t.Fatal("expected GraphQL snapshot to be enabled")
	}
	if cfg.GraphQL.Path != "/graphql" {
		t.Fatalf("expected GraphQL path /graphql, got %q", cfg.GraphQL.Path)
	}
	if !cfg.GraphQL.Playground {
		t.Fatal("expected GraphQL playground snapshot to be enabled")
	}
	if cfg.GraphQL.PlaygroundPath != "/graphql/playground" {
		t.Fatalf("expected GraphQL playground path /graphql/playground, got %q", cfg.GraphQL.PlaygroundPath)
	}
	if cfg.I18n.DefaultLocale != "en-GB" {
		t.Fatalf("expected default locale en-GB, got %q", cfg.I18n.DefaultLocale)
	}
	if !slices.Equal(cfg.I18n.Supported, []string{"en-GB", "fr"}) {
		t.Fatalf("expected supported locales [en-GB fr], got %v", cfg.I18n.Supported)
	}
	if cfg.Authentik.Issuer != "https://auth.example.com" {
		t.Fatalf("expected Authentik issuer https://auth.example.com, got %q", cfg.Authentik.Issuer)
	}
	if cfg.Authentik.ClientID != "runtime-client" {
		t.Fatalf("expected Authentik client ID runtime-client, got %q", cfg.Authentik.ClientID)
	}
	if !cfg.Authentik.TrustedProxy {
		t.Fatal("expected Authentik trusted proxy to be enabled")
	}
	if !slices.Equal(cfg.Authentik.PublicPaths, []string{"/public", "/docs"}) {
		t.Fatalf("expected Authentik public paths [/public /docs], got %v", cfg.Authentik.PublicPaths)
	}
}

// TestEngine_RuntimeConfig_Good_EmptyOnNilEngine verifies the nil receiver
// guard returns an empty runtime snapshot.
func TestEngine_RuntimeConfig_Good_EmptyOnNilEngine(t *testing.T) {
	var e *api.Engine

	cfg := e.RuntimeConfig()
	if cfg.Swagger.Enabled || cfg.Transport.SwaggerEnabled || cfg.GraphQL.Enabled || cfg.Cache.Enabled || cfg.I18n.DefaultLocale != "" || cfg.Authentik.Issuer != "" {
		t.Fatalf("expected zero-value runtime config, got %+v", cfg)
	}
}
