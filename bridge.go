// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"iter"

	"github.com/gin-gonic/gin"
)

// ToolDescriptor describes a tool that can be exposed as a REST endpoint.
type ToolDescriptor struct {
	Name         string         // Tool name, e.g. "file_read" (becomes POST path segment)
	Description  string         // Human-readable description
	Group        string         // OpenAPI tag group, e.g. "files"
	InputSchema  map[string]any // JSON Schema for request body
	OutputSchema map[string]any // JSON Schema for response data (optional)
}

// ToolBridge converts tool descriptors into REST endpoints and OpenAPI paths.
// It implements both RouteGroup and DescribableGroup.
type ToolBridge struct {
	basePath string
	name     string
	tools    []boundTool
}

type boundTool struct {
	descriptor ToolDescriptor
	handler    gin.HandlerFunc
}

// NewToolBridge creates a bridge that mounts tool endpoints at basePath.
func NewToolBridge(basePath string) *ToolBridge {
	return &ToolBridge{
		basePath: basePath,
		name:     "tools",
	}
}

// Add registers a tool with its HTTP handler.
func (b *ToolBridge) Add(desc ToolDescriptor, handler gin.HandlerFunc) {
	b.tools = append(b.tools, boundTool{descriptor: desc, handler: handler})
}

// Name returns the bridge identifier.
func (b *ToolBridge) Name() string { return b.name }

// BasePath returns the URL prefix for all tool endpoints.
func (b *ToolBridge) BasePath() string { return b.basePath }

// RegisterRoutes mounts POST /{tool_name} for each registered tool.
func (b *ToolBridge) RegisterRoutes(rg *gin.RouterGroup) {
	for _, t := range b.tools {
		rg.POST("/"+t.descriptor.Name, t.handler)
	}
}

// Describe returns OpenAPI route descriptions for all registered tools.
func (b *ToolBridge) Describe() []RouteDescription {
	descs := make([]RouteDescription, 0, len(b.tools))
	for _, t := range b.tools {
		tags := []string{t.descriptor.Group}
		if t.descriptor.Group == "" {
			tags = []string{b.name}
		}
		descs = append(descs, RouteDescription{
			Method:      "POST",
			Path:        "/" + t.descriptor.Name,
			Summary:     t.descriptor.Description,
			Description: t.descriptor.Description,
			Tags:        tags,
			RequestBody: t.descriptor.InputSchema,
			Response:    t.descriptor.OutputSchema,
		})
	}
	return descs
}

// DescribeIter returns an iterator over OpenAPI route descriptions for all registered tools.
func (b *ToolBridge) DescribeIter() iter.Seq[RouteDescription] {
	return func(yield func(RouteDescription) bool) {
		for _, t := range b.tools {
			tags := []string{t.descriptor.Group}
			if t.descriptor.Group == "" {
				tags = []string{b.name}
			}
			rd := RouteDescription{
				Method:      "POST",
				Path:        "/" + t.descriptor.Name,
				Summary:     t.descriptor.Description,
				Description: t.descriptor.Description,
				Tags:        tags,
				RequestBody: t.descriptor.InputSchema,
				Response:    t.descriptor.OutputSchema,
			}
			if !yield(rd) {
				return
			}
		}
	}
}

// Tools returns all registered tool descriptors.
func (b *ToolBridge) Tools() []ToolDescriptor {
	descs := make([]ToolDescriptor, len(b.tools))
	for i, t := range b.tools {
		descs[i] = t.descriptor
	}
	return descs
}

// ToolsIter returns an iterator over all registered tool descriptors.
func (b *ToolBridge) ToolsIter() iter.Seq[ToolDescriptor] {
	return func(yield func(ToolDescriptor) bool) {
		for _, t := range b.tools {
			if !yield(t.descriptor) {
				return
			}
		}
	}
}
