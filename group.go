// SPDX-License-Identifier: EUPL-1.2

package api

import "github.com/gin-gonic/gin"

// RouteGroup registers API routes onto a Gin router group.
// Subsystems implement this interface to declare their endpoints.
type RouteGroup interface {
	// Name returns a human-readable identifier for the group.
	Name() string

	// BasePath returns the URL prefix for all routes in this group.
	BasePath() string

	// RegisterRoutes mounts handlers onto the provided router group.
	RegisterRoutes(rg *gin.RouterGroup)
}

// StreamGroup optionally declares WebSocket channels a subsystem publishes to.
type StreamGroup interface {
	// Channels returns the list of channel names this group streams on.
	Channels() []string
}

// DescribableGroup extends RouteGroup with OpenAPI metadata.
// RouteGroups that implement this will have their endpoints
// included in the generated OpenAPI specification.
type DescribableGroup interface {
	RouteGroup
	// Describe returns endpoint descriptions for OpenAPI generation.
	Describe() []RouteDescription
}

// RouteDescription describes a single endpoint for OpenAPI generation.
type RouteDescription struct {
	Method      string   // HTTP method: GET, POST, PUT, DELETE, PATCH
	Path        string   // Path relative to BasePath, e.g. "/generate"
	Summary     string   // Short summary
	Description string   // Long description
	Tags        []string // OpenAPI tags for grouping
	Parameters  []ParameterDescription
	RequestBody map[string]any // JSON Schema for request body (nil for GET)
	Response    map[string]any // JSON Schema for success response data
}

// ParameterDescription describes an OpenAPI parameter for a route.
type ParameterDescription struct {
	Name        string         // Parameter name.
	In          string         // Parameter location: path, query, header, or cookie.
	Description string         // Human-readable parameter description.
	Required    bool           // Whether the parameter is required.
	Deprecated  bool           // Whether the parameter is deprecated.
	Schema      map[string]any // JSON Schema for the parameter value.
	Example     any            // Optional example value.
}
