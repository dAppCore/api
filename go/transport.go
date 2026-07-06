// SPDX-License-Identifier: EUPL-1.2

package api

import core "dappco.re/go"

// TransportConfig captures the configured transport endpoints and flags for an Engine.
//
// It is intentionally small and serialisable so callers can inspect the active HTTP
// surface without rebuilding an OpenAPI document.
//
// Example:
//
//	cfg := api.TransportConfig{SwaggerPath: "/swagger", WSPath: "/ws"}
type TransportConfig struct {
	SwaggerEnabled        bool
	SwaggerPath           string
	GraphQLPath           string
	GraphQLEnabled        bool
	GraphQLPlayground     bool
	GraphQLPlaygroundPath string
	WSEnabled             bool
	WSPath                string
	SSEEnabled            bool
	SSEPath               string
	PprofEnabled          bool
	ExpvarEnabled         bool
	OpenAPISpecEnabled    bool
	OpenAPISpecPath       string
	UpstreamRouterPaths   []string
}

// TransportConfig returns the currently configured transport metadata for the engine.
//
// The result snapshots the Engine state at call time and normalises any configured
// URL paths using the same rules as the runtime handlers.
//
// Example:
//
//	cfg := engine.TransportConfig()
func (e *Engine) TransportConfig() TransportConfig {
	if e == nil {
		return TransportConfig{}
	}

	cfg := TransportConfig{
		SwaggerEnabled:     e.swaggerEnabled,
		WSEnabled:          e.wsHandler != nil || e.wsGinHandler != nil,
		SSEEnabled:         e.sseBroker != nil,
		PprofEnabled:       e.pprofEnabled,
		ExpvarEnabled:      e.expvarEnabled,
		OpenAPISpecEnabled: e.openAPISpecEnabled,
	}
	gql := e.GraphQLConfig()
	cfg.GraphQLEnabled = gql.Enabled
	cfg.GraphQLPlayground = gql.Playground
	cfg.GraphQLPlaygroundPath = gql.PlaygroundPath

	if e.swaggerEnabled || core.Trim(e.swaggerPath) != "" {
		cfg.SwaggerPath = resolveSwaggerPath(e.swaggerPath)
	}
	if gql.Path != "" {
		cfg.GraphQLPath = gql.Path
	}
	if e.wsHandler != nil || e.wsGinHandler != nil || core.Trim(e.wsPath) != "" {
		cfg.WSPath = resolveWSPath(e.wsPath)
	}
	if e.sseBroker != nil || core.Trim(e.ssePath) != "" {
		cfg.SSEPath = resolveSSEPath(e.ssePath)
	}
	if e.openAPISpecEnabled || core.Trim(e.openAPISpecPath) != "" {
		cfg.OpenAPISpecPath = resolveOpenAPISpecPath(e.openAPISpecPath)
	}
	if e.upstreamRouter != nil {
		cfg.UpstreamRouterPaths = append([]string(nil), e.upstreamRouter.paths...)
	}

	return cfg
}

func trimPathSlashes(path string) string {
	for core.HasPrefix(path, "/") {
		path = core.TrimPrefix(path, "/")
	}
	for core.HasSuffix(path, "/") {
		path = core.TrimSuffix(path, "/")
	}
	return path
}
