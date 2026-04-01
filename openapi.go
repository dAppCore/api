// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"
)

// SpecBuilder constructs an OpenAPI 3.1 specification from registered RouteGroups.
type SpecBuilder struct {
	Title       string
	Description string
	Version     string
}

// Build generates the complete OpenAPI 3.1 JSON spec.
// Groups implementing DescribableGroup contribute endpoint documentation.
// Other groups are listed as tags only.
func (sb *SpecBuilder) Build(groups []RouteGroup) ([]byte, error) {
	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       sb.Title,
			"description": sb.Description,
			"version":     sb.Version,
		},
		"paths": sb.buildPaths(groups),
		"tags":  sb.buildTags(groups),
	}

	// Add component schemas for the response envelope.
	spec["components"] = map[string]any{
		"schemas": map[string]any{
			"Error": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code":    map[string]any{"type": "string"},
					"message": map[string]any{"type": "string"},
					"details": map[string]any{},
				},
				"required": []string{"code", "message"},
			},
			"Meta": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"request_id": map[string]any{"type": "string"},
					"duration":   map[string]any{"type": "string"},
					"page":       map[string]any{"type": "integer"},
					"per_page":   map[string]any{"type": "integer"},
					"total":      map[string]any{"type": "integer"},
				},
			},
		},
	}

	return json.MarshalIndent(spec, "", "  ")
}

// buildPaths generates the paths object from all DescribableGroups.
func (sb *SpecBuilder) buildPaths(groups []RouteGroup) map[string]any {
	operationIDs := map[string]int{}
	paths := map[string]any{
		// Built-in health endpoint.
		"/health": map[string]any{
			"get": map[string]any{
				"summary":     "Health check",
				"description": "Returns server health status",
				"tags":        []string{"system"},
				"operationId": operationID("get", "/health", operationIDs),
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Server is healthy",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": envelopeSchema(map[string]any{"type": "string"}),
							},
						},
					},
				},
			},
		},
	}

	for _, g := range groups {
		dg, ok := g.(DescribableGroup)
		if !ok {
			continue
		}
		for _, rd := range dg.Describe() {
			fullPath := g.BasePath() + rd.Path
			method := strings.ToLower(rd.Method)

			operation := map[string]any{
				"summary":     rd.Summary,
				"description": rd.Description,
				"tags":        rd.Tags,
				"operationId": operationID(method, fullPath, operationIDs),
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Successful response",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": envelopeSchema(rd.Response),
							},
						},
					},
					"400": map[string]any{
						"description": "Bad request",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": envelopeSchema(nil),
							},
						},
					},
				},
			}

			// Add request body for methods that accept one.
			if rd.RequestBody != nil && (method == "post" || method == "put" || method == "patch") {
				operation["requestBody"] = map[string]any{
					"required": true,
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": rd.RequestBody,
						},
					},
				}
			}

			// Create or extend path item.
			if existing, exists := paths[fullPath]; exists {
				existing.(map[string]any)[method] = operation
			} else {
				paths[fullPath] = map[string]any{
					method: operation,
				}
			}
		}
	}

	return paths
}

// buildTags generates the tags array from all RouteGroups.
func (sb *SpecBuilder) buildTags(groups []RouteGroup) []map[string]any {
	tags := []map[string]any{
		{"name": "system", "description": "System endpoints"},
	}
	seen := map[string]bool{"system": true}

	for _, g := range groups {
		name := g.Name()
		if !seen[name] {
			tags = append(tags, map[string]any{
				"name":        name,
				"description": name + " endpoints",
			})
			seen[name] = true
		}
	}

	return tags
}

// envelopeSchema wraps a data schema in the standard Response[T] envelope.
func envelopeSchema(dataSchema map[string]any) map[string]any {
	properties := map[string]any{
		"success": map[string]any{"type": "boolean"},
		"error": map[string]any{
			"$ref": "#/components/schemas/Error",
		},
		"meta": map[string]any{
			"$ref": "#/components/schemas/Meta",
		},
	}

	if dataSchema != nil {
		properties["data"] = dataSchema
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   []string{"success"},
	}
}

// operationID builds a stable OpenAPI operationId from the HTTP method and path.
// The generated identifier is lower snake_case and preserves path parameter names.
func operationID(method, path string, operationIDs map[string]int) string {
	var b strings.Builder
	b.Grow(len(method) + len(path) + 1)
	lastUnderscore := false

	writeUnderscore := func() {
		if b.Len() > 0 && !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}

	appendToken := func(r rune) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if unicode.IsUpper(r) {
				r = unicode.ToLower(r)
			}
			b.WriteRune(r)
			lastUnderscore = false
			return
		}
		writeUnderscore()
	}

	for _, r := range method {
		appendToken(r)
	}
	writeUnderscore()
	for _, r := range path {
		switch r {
		case '/':
			writeUnderscore()
		case '-':
			writeUnderscore()
		case '.':
			writeUnderscore()
		case ' ':
			writeUnderscore()
		default:
			appendToken(r)
		}
	}

	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "operation"
	}

	if operationIDs == nil {
		return out
	}

	count := operationIDs[out]
	operationIDs[out] = count + 1
	if count == 0 {
		return out
	}
	return out + "_" + strconv.Itoa(count+1)
}
