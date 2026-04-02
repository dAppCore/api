// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

// defaultGraphQLPath is the URL path where the GraphQL endpoint is mounted.
const defaultGraphQLPath = "/graphql"

// graphqlConfig holds configuration for the GraphQL endpoint.
type graphqlConfig struct {
	schema     graphql.ExecutableSchema
	path       string
	playground bool
}

// GraphQLOption configures a GraphQL endpoint.
//
// Example:
//
//	opts := []api.GraphQLOption{api.WithPlayground(), api.WithGraphQLPath("/gql")}
type GraphQLOption func(*graphqlConfig)

// WithPlayground enables the GraphQL Playground UI at {path}/playground.
//
// Example:
//
//	api.WithGraphQL(schema, api.WithPlayground())
func WithPlayground() GraphQLOption {
	return func(cfg *graphqlConfig) {
		cfg.playground = true
	}
}

// WithGraphQLPath sets a custom URL path for the GraphQL endpoint.
// The default path is "/graphql".
//
// Example:
//
//	api.WithGraphQL(schema, api.WithGraphQLPath("/gql"))
func WithGraphQLPath(path string) GraphQLOption {
	return func(cfg *graphqlConfig) {
		cfg.path = normaliseGraphQLPath(path)
	}
}

// mountGraphQL registers the GraphQL handler and optional playground on the Gin engine.
func mountGraphQL(r *gin.Engine, cfg *graphqlConfig) {
	srv := handler.NewDefaultServer(cfg.schema)
	graphqlHandler := gin.WrapH(srv)

	// Mount the GraphQL endpoint for all HTTP methods (POST for queries/mutations,
	// GET for playground redirects and introspection).
	r.Any(cfg.path, graphqlHandler)

	if cfg.playground {
		playgroundPath := cfg.path + "/playground"
		playgroundHandler := playground.Handler("GraphQL", cfg.path)
		r.GET(playgroundPath, wrapHTTPHandler(playgroundHandler))
	}
}

// normaliseGraphQLPath coerces custom GraphQL paths into a stable form.
// The path always begins with a single slash and never ends with one.
func normaliseGraphQLPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return defaultGraphQLPath
	}

	path = "/" + strings.Trim(path, "/")
	if path == "/" {
		return defaultGraphQLPath
	}

	return path
}

// wrapHTTPHandler adapts a standard http.Handler to a Gin handler function.
func wrapHTTPHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
