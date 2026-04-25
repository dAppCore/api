// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"iter"

	"github.com/gin-gonic/gin"
)

// RouteGroup registers API routes onto a Gin router group.
// Subsystems implement this interface to declare their endpoints.
//
// Example:
//
//	var g api.RouteGroup = &myGroup{}
type RouteGroup interface {
	// Name returns a human-readable identifier for the group.
	Name() string

	// BasePath returns the URL prefix for all routes in this group.
	BasePath() string

	// RegisterRoutes mounts handlers onto the provided router group.
	RegisterRoutes(rg *gin.RouterGroup)
}

// Describable allows a route handler or controller to expose OpenAPI metadata
// without coupling callers to RouteDescription construction details.
//
// Example:
//
//	var d api.Describable = &myHandler{}
type Describable interface {
	// Describe returns the handler's request/response description.
	Describe() RouteDescription
	// OperationID returns the OpenAPI operation identifier.
	OperationID() string
	// Tags returns the OpenAPI tags associated with the operation.
	Tags() []string
	// Summary returns a short operation summary.
	Summary() string
	// Description returns a longer operation description.
	Description() string
}

// Renderable allows a route handler or controller to expose UI rendering
// hints that spec consumers can surface via vendor extensions.
//
// Example:
//
//	var r api.Renderable = &myHandler{}
type Renderable interface {
	// Render returns UI hints for the operation.
	Render() RenderHints
}

// RenderHints describes how a UI may present an operation.
type RenderHints struct {
	Kind    string       `json:"kind,omitempty"`    // "form" | "table" | "modal" | "grid"
	Fields  []FieldHint  `json:"fields,omitempty"`  // Form fields with validation hints.
	Actions []ActionHint `json:"actions,omitempty"` // Inline action buttons.
}

// FieldHint describes an input field for UI rendering.
type FieldHint struct {
	Name       string         `json:"name,omitempty"`
	Label      string         `json:"label,omitempty"`
	Type       string         `json:"type,omitempty"`
	Required   bool           `json:"required,omitempty"`
	Validation map[string]any `json:"validation,omitempty"`
}

// ActionHint describes an inline action a UI can render for an operation.
type ActionHint struct {
	Name    string `json:"name,omitempty"`
	Label   string `json:"label,omitempty"`
	Method  string `json:"method,omitempty"`
	Variant string `json:"variant,omitempty"`
}

// StreamGroup optionally declares WebSocket channels a subsystem publishes to.
//
// Example:
//
//	var sg api.StreamGroup = &myStreamGroup{}
type StreamGroup interface {
	// Channels returns the list of channel names this group streams on.
	Channels() []string
}

// DescribableGroup extends RouteGroup with OpenAPI metadata.
// RouteGroups that implement this will have their endpoints
// included in the generated OpenAPI specification.
//
// Example:
//
//	var dg api.DescribableGroup = &myDescribableGroup{}
type DescribableGroup interface {
	RouteGroup
	// Describe returns endpoint descriptions for OpenAPI generation.
	Describe() []RouteDescription
}

// DescribableGroupIter extends DescribableGroup with an iterator-based
// description source for callers that want to avoid slice allocation.
//
// Example:
//
//	var dg api.DescribableGroupIter = &myDescribableGroup{}
type DescribableGroupIter interface {
	DescribableGroup
	// DescribeIter returns endpoint descriptions for OpenAPI generation.
	DescribeIter() iter.Seq[RouteDescription]
}

// RouteDescription describes a single endpoint for OpenAPI generation.
//
// Example:
//
//	rd := api.RouteDescription{
//		Method:      "POST",
//		Path:        "/users",
//		Summary:     "Create a user",
//		Description: "Creates a new user account.",
//		Tags:        []string{"users"},
//		StatusCode:  201,
//		RequestBody: map[string]any{"type": "object"},
//		Response:    map[string]any{"type": "object"},
//	}
type RouteDescription struct {
	Method      string   // HTTP method: GET, POST, PUT, DELETE, PATCH
	Path        string   // Path relative to BasePath, e.g. "/generate"
	Summary     string   // Short summary
	Description string   // Long description
	Tags        []string // OpenAPI tags for grouping
	// Handler optionally points at the route handler/controller that implements
	// Describable and/or Renderable. RegisterRoutes still owns actual Gin
	// wiring; this field is metadata-only for spec generation.
	Handler any
	// CacheControl hints the framework that successful responses for this
	// operation should advertise the given Cache-Control policy in docs.
	CacheControl string
	// Hidden omits the route from generated documentation.
	Hidden bool
	// Deprecated marks the operation as deprecated in OpenAPI.
	Deprecated bool
	// SunsetDate marks when a deprecated operation will be removed.
	// Use YYYY-MM-DD or an RFC 7231 HTTP date string.
	SunsetDate string
	// ReplacementURL points to the successor endpoint URL, when known.
	// Replacement is kept as a legacy alias for existing call sites.
	ReplacementURL string
	Replacement    string
	// NoticeURL points to a detailed deprecation notice or migration guide,
	// surfaced as the API-Deprecation-Notice-URL response header per spec §8.
	NoticeURL string
	// StatusCode is the documented 2xx success status code.
	// Zero defaults to 200.
	StatusCode int
	// Security overrides the default bearerAuth requirement when non-nil.
	// Use an empty, non-nil slice to mark the route as public.
	Security        []map[string][]string
	Parameters      []ParameterDescription
	RequestBody     map[string]any // JSON Schema for request body (nil for GET)
	RequestExample  any            // Optional example payload for the request body.
	Response        map[string]any // JSON Schema for success response data
	ResponseExample any            // Optional example payload for the success response.
	ResponseHeaders map[string]string
}

// ParameterDescription describes an OpenAPI parameter for a route.
//
// Example:
//
//	param := api.ParameterDescription{
//		Name:        "id",
//		In:          "path",
//		Description: "User identifier",
//		Required:    true,
//		Schema:      map[string]any{"type": "string"},
//		Example:     "usr_123",
//	}
type ParameterDescription struct {
	Name        string         // Parameter name.
	In          string         // Parameter location: path, query, header, or cookie.
	Description string         // Human-readable parameter description.
	Required    bool           // Whether the parameter is required.
	Deprecated  bool           // Whether the parameter is deprecated.
	Schema      map[string]any // JSON Schema for the parameter value.
	Example     any            // Optional example value.
}
