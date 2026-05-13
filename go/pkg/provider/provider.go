// SPDX-License-Identifier: EUPL-1.2

// Package provider defines the Service Provider Framework interfaces.
//
// A Provider extends api.RouteGroup with a provider identity. Providers
// register through the existing api.Engine.Register() method, inheriting
// middleware, CORS, Swagger, and OpenAPI generation automatically.
//
// Optional interfaces (Streamable, Describable, Renderable) declare
// additional capabilities that consumers (GUI, MCP, WS hub) can discover
// via type assertion.
package provider

import (
	"dappco.re/go/api"
)

// Provider extends RouteGroup with a provider identity.
// Every Provider is a RouteGroup and registers through api.Engine.Register().
type Provider interface {
	api.RouteGroup // Name(), BasePath(), RegisterRoutes(*gin.RouterGroup)
}

// Streamable providers emit real-time events via WebSocket.
// The hub is injected at construction time. Channels() declares the
// event prefixes this provider will emit (e.g. "brain.*", "process.*").
type Streamable interface {
	Provider
	Channels() []string
}

// Describable providers expose structured route descriptions for OpenAPI.
// This extends the existing DescribableGroup interface from go-api.
type Describable interface {
	Provider
	api.DescribableGroup // Describe() []RouteDescription
}

// Renderable providers declare a custom element for GUI display.
type Renderable interface {
	Provider
	Element() ElementSpec
}

// ElementSpec describes a web component for GUI rendering.
type ElementSpec struct {
	// Tag is the custom element tag name, e.g. "core-brain-panel".
	Tag string `json:"tag"`

	// Source is the URL or embedded path to the JS bundle.
	Source string `json:"source"`
}
