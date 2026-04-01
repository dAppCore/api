// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strconv"

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
	if validator := newToolInputValidator(desc.InputSchema); validator != nil {
		handler = wrapToolHandler(handler, validator)
	}
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

func wrapToolHandler(handler gin.HandlerFunc, validator *toolInputValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, FailWithDetails(
				"invalid_request_body",
				"Unable to read request body",
				map[string]any{"error": err.Error()},
			))
			return
		}

		if err := validator.Validate(body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, FailWithDetails(
				"invalid_request_body",
				"Request body does not match the declared tool schema",
				map[string]any{"error": err.Error()},
			))
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		handler(c)
	}
}

type toolInputValidator struct {
	schema map[string]any
}

func newToolInputValidator(schema map[string]any) *toolInputValidator {
	if len(schema) == 0 {
		return nil
	}
	return &toolInputValidator{schema: schema}
}

func (v *toolInputValidator) Validate(body []byte) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return fmt.Errorf("request body is required")
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()

	var payload any
	if err := dec.Decode(&payload); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("request body must contain a single JSON value")
	}

	return validateSchemaNode(payload, v.schema, "")
}

func validateSchemaNode(value any, schema map[string]any, path string) error {
	if len(schema) == 0 {
		return nil
	}

	if schemaType, _ := schema["type"].(string); schemaType != "" {
		switch schemaType {
		case "object":
			obj, ok := value.(map[string]any)
			if !ok {
				return typeError(path, "object", value)
			}

			for _, name := range stringList(schema["required"]) {
				if _, ok := obj[name]; !ok {
					return fmt.Errorf("%s is missing required field %q", displayPath(path), name)
				}
			}

			for name, rawChild := range schemaMap(schema["properties"]) {
				childSchema, ok := rawChild.(map[string]any)
				if !ok {
					continue
				}
				childValue, ok := obj[name]
				if !ok {
					continue
				}
				if err := validateSchemaNode(childValue, childSchema, joinPath(path, name)); err != nil {
					return err
				}
			}
			return nil
		case "array":
			arr, ok := value.([]any)
			if !ok {
				return typeError(path, "array", value)
			}
			if items := schemaMap(schema["items"]); len(items) > 0 {
				for i, item := range arr {
					if err := validateSchemaNode(item, items, joinPath(path, strconv.Itoa(i))); err != nil {
						return err
					}
				}
			}
			return nil
		case "string":
			if _, ok := value.(string); !ok {
				return typeError(path, "string", value)
			}
			return nil
		case "boolean":
			if _, ok := value.(bool); !ok {
				return typeError(path, "boolean", value)
			}
			return nil
		case "integer":
			if !isIntegerValue(value) {
				return typeError(path, "integer", value)
			}
			return nil
		case "number":
			if !isNumberValue(value) {
				return typeError(path, "number", value)
			}
			return nil
		}
	}

	if props := schemaMap(schema["properties"]); len(props) > 0 {
		return validateSchemaNode(value, map[string]any{
			"type":       "object",
			"properties": props,
			"required":   schema["required"],
		}, path)
	}

	return nil
}

func typeError(path, want string, value any) error {
	return fmt.Errorf("%s must be %s, got %s", displayPath(path), want, describeJSONValue(value))
}

func displayPath(path string) string {
	if path == "" {
		return "request body"
	}
	return "request body." + path
}

func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func schemaMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	m, _ := value.(map[string]any)
	return m
}

func stringList(value any) []string {
	switch raw := value.(type) {
	case []any:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			name, ok := item.(string)
			if !ok {
				continue
			}
			out = append(out, name)
		}
		return out
	case []string:
		return append([]string(nil), raw...)
	default:
		return nil
	}
}

func isIntegerValue(value any) bool {
	switch v := value.(type) {
	case json.Number:
		_, err := v.Int64()
		return err == nil
	case float64:
		return v == float64(int64(v))
	default:
		return false
	}
}

func isNumberValue(value any) bool {
	switch value.(type) {
	case json.Number, float64:
		return true
	default:
		return false
	}
}

func describeJSONValue(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case bool:
		return "boolean"
	case json.Number, float64:
		return "number"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return fmt.Sprintf("%T", value)
	}
}
