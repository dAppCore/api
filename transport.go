// SPDX-License-Identifier: EUPL-1.2

package api

import "strings"

// TransportConfig captures the configured transport endpoints and flags for an Engine.
//
// It is intentionally small and serialisable so callers can inspect the active HTTP
// surface without rebuilding an OpenAPI document.
//
// Example:
//
//	cfg := api.TransportConfig{SwaggerPath: "/swagger", WSPath: "/ws"}
type TransportConfig struct {
	SwaggerEnabled    bool
	SwaggerPath       string
	GraphQLPath       string
	GraphQLEnabled    bool
	GraphQLPlayground bool
	WSPath            string
	SSEPath           string
	PprofEnabled      bool
	ExpvarEnabled     bool
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
		SwaggerEnabled:    e.swaggerEnabled,
		GraphQLEnabled:    e.graphql != nil,
		GraphQLPlayground: e.graphql != nil && e.graphql.playground,
		PprofEnabled:      e.pprofEnabled,
		ExpvarEnabled:     e.expvarEnabled,
	}

	if e.swaggerEnabled || strings.TrimSpace(e.swaggerPath) != "" {
		cfg.SwaggerPath = resolveSwaggerPath(e.swaggerPath)
	}
	if e.graphql != nil {
		cfg.GraphQLPath = e.graphql.path
	}
	if e.wsHandler != nil || strings.TrimSpace(e.wsPath) != "" {
		cfg.WSPath = resolveWSPath(e.wsPath)
	}
	if e.sseBroker != nil || strings.TrimSpace(e.ssePath) != "" {
		cfg.SSEPath = resolveSSEPath(e.ssePath)
	}

	return cfg
}
