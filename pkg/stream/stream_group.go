// SPDX-License-Identifier: EUPL-1.2

// Package stream defines declarative SSE and WebSocket endpoint groups that
// can be mounted onto an api.Engine.
package stream

import (
	"net/http"
	"slices"
	"strings"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

// Protocol identifies the wire protocol a stream handler serves.
type Protocol string

const (
	// ProtocolSSE identifies a Server-Sent Events endpoint.
	ProtocolSSE Protocol = "sse"
	// ProtocolWebSocket identifies a WebSocket endpoint.
	ProtocolWebSocket Protocol = "websocket"
)

// Handler describes a single stream-capable route.
//
// The protocol and path are retained as declarative metadata so callers can
// inspect mounted stream surfaces and future OpenAPI hooks can consume them.
type Handler struct {
	Protocol Protocol
	Method   string
	Path     string
	Handle   gin.HandlerFunc
}

// Registrar is the minimal Gin registration surface required by StreamGroup.
// Both *gin.Engine and *gin.RouterGroup satisfy this contract.
type Registrar interface {
	Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
}

// StreamGroup declares a named set of SSE/WebSocket handlers.
//
// Example:
//
//	var group stream.StreamGroup = stream.NewGroup(
//		"system",
//		stream.SSE("/events", func(c *gin.Context) {}),
//	)
type StreamGroup interface {
	// Register mounts all handlers onto the supplied registrar.
	Register(reg Registrar)

	// Name returns a human-readable identifier for the group.
	Name() string

	// Handlers returns the group's declared handler metadata.
	Handlers() []Handler
}

// Group is a small concrete StreamGroup implementation backed by a handler
// slice. It is suitable for most SSE/WebSocket endpoint declarations.
type Group struct {
	name     string
	handlers []Handler
}

// NewGroup creates a StreamGroup with normalised handler metadata.
func NewGroup(name string, handlers ...Handler) *Group {
	return &Group{
		name:     core.Trim(name),
		handlers: normaliseHandlers(handlers),
	}
}

// Name returns the group's identifier.
func (g *Group) Name() string {
	if g == nil {
		return ""
	}
	return g.name
}

// Handlers returns a defensive copy of the group's handler metadata.
func (g *Group) Handlers() []Handler {
	if g == nil || len(g.handlers) == 0 {
		return nil
	}
	return slices.Clone(g.handlers)
}

// Register mounts all valid handlers onto the supplied registrar.
func (g *Group) Register(reg Registrar) {
	if g == nil || reg == nil {
		return
	}

	for _, handler := range g.handlers {
		reg.Handle(handler.Method, handler.Path, handler.Handle)
	}
}

// SSE creates a GET Server-Sent Events handler descriptor.
func SSE(path string, handle gin.HandlerFunc) Handler {
	return Handler{
		Protocol: ProtocolSSE,
		Method:   http.MethodGet,
		Path:     path,
		Handle:   handle,
	}
}

// WebSocket creates a GET WebSocket handler descriptor.
func WebSocket(path string, handle gin.HandlerFunc) Handler {
	return Handler{
		Protocol: ProtocolWebSocket,
		Method:   http.MethodGet,
		Path:     path,
		Handle:   handle,
	}
}

func normaliseHandlers(handlers []Handler) []Handler {
	if len(handlers) == 0 {
		return nil
	}

	out := make([]Handler, 0, len(handlers))
	for _, handler := range handlers {
		handler = normaliseHandler(handler)
		if !handler.valid() {
			continue
		}
		out = append(out, handler)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

func normaliseHandler(handler Handler) Handler {
	handler.Protocol = normaliseProtocol(handler.Protocol)

	method := strings.ToUpper(core.Trim(handler.Method))
	if method == "" {
		method = http.MethodGet
	}
	handler.Method = method
	handler.Path = normalisePath(handler.Path)

	return handler
}

func (h Handler) valid() bool {
	return h.Protocol != "" && h.Path != "" && h.Handle != nil
}

func normaliseProtocol(protocol Protocol) Protocol {
	switch strings.ToLower(core.Trim(string(protocol))) {
	case "event-stream", "eventstream", "sse":
		return ProtocolSSE
	case "websocket", "ws":
		return ProtocolWebSocket
	default:
		return ""
	}
}

func normalisePath(path string) string {
	path = core.Trim(path)
	if path == "" {
		return ""
	}

	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "/"
	}

	return "/" + trimmed
}
