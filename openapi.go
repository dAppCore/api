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
	Servers     []string
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
		"security": []any{
			map[string]any{
				"bearerAuth": []any{},
			},
		},
	}

	if servers := normaliseServers(sb.Servers); len(servers) > 0 {
		out := make([]map[string]any, 0, len(servers))
		for _, server := range servers {
			out = append(out, map[string]any{"url": server})
		}
		spec["servers"] = out
	}

	// Add component schemas for the response envelope.
	spec["components"] = map[string]any{
		"schemas": map[string]any{
			"Response": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"success": map[string]any{"type": "boolean"},
					"data":    map[string]any{},
					"error": map[string]any{
						"$ref": "#/components/schemas/Error",
					},
					"meta": map[string]any{
						"$ref": "#/components/schemas/Meta",
					},
				},
				"required": []string{"success"},
			},
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
		"securitySchemes": map[string]any{
			"bearerAuth": map[string]any{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
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
				"responses":   healthResponses(),
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
				"tags":        resolvedOperationTags(g, rd),
				"operationId": operationID(method, fullPath, operationIDs),
				"security": []any{
					map[string]any{
						"bearerAuth": []any{},
					},
				},
				"responses": operationResponses(method, rd.Response),
			}

			// Add request body for methods that accept one.
			// The contract only excludes GET; other verbs may legitimately carry bodies.
			if rd.RequestBody != nil && method != "get" {
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

	// The built-in health check remains public, so override the inherited
	// default security requirement with an explicit empty array.
	if health, ok := paths["/health"].(map[string]any); ok {
		if op, ok := health["get"].(map[string]any); ok {
			op["security"] = []any{}
		}
	}

	return paths
}

// operationResponses builds the standard response set for a documented API
// operation. The framework always exposes the common envelope responses, plus
// middleware-driven 429 and 504 errors.
func operationResponses(method string, dataSchema map[string]any) map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	if method == "get" {
		successHeaders = mergeHeaders(successHeaders, cacheSuccessHeaders())
	}

	return map[string]any{
		"200": map[string]any{
			"description": "Successful response",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(dataSchema),
				},
			},
			"headers": successHeaders,
		},
		"400": map[string]any{
			"description": "Bad request",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders()),
		},
		"401": map[string]any{
			"description": "Unauthorised",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders()),
		},
		"403": map[string]any{
			"description": "Forbidden",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders()),
		},
		"429": map[string]any{
			"description": "Too many requests",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitHeaders()),
		},
		"504": map[string]any{
			"description": "Gateway timeout",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders()),
		},
		"500": map[string]any{
			"description": "Internal server error",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders()),
		},
	}
}

// healthResponses builds the response set for the built-in health endpoint.
// It stays public, but rate limiting and timeouts can still apply.
func healthResponses() map[string]any {
	return map[string]any{
		"200": map[string]any{
			"description": "Server is healthy",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(map[string]any{"type": "string"}),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders(), cacheSuccessHeaders()),
		},
		"429": map[string]any{
			"description": "Too many requests",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitHeaders()),
		},
		"504": map[string]any{
			"description": "Gateway timeout",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders()),
		},
		"500": map[string]any{
			"description": "Internal server error",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders()),
		},
	}
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

// resolvedOperationTags returns the explicit route tags when provided, or a
// stable fallback derived from the group's name when the route omits tags.
func resolvedOperationTags(g RouteGroup, rd RouteDescription) []string {
	if len(rd.Tags) > 0 {
		return rd.Tags
	}

	if name := g.Name(); name != "" {
		return []string{name}
	}

	return nil
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

// rateLimitHeaders documents the response headers emitted when rate limiting
// rejects a request.
func rateLimitHeaders() map[string]any {
	return map[string]any{
		"X-RateLimit-Limit": map[string]any{
			"description": "Maximum number of requests allowed in the current window",
			"schema": map[string]any{
				"type":    "integer",
				"minimum": 1,
			},
		},
		"X-RateLimit-Remaining": map[string]any{
			"description": "Number of requests remaining in the current window",
			"schema": map[string]any{
				"type":    "integer",
				"minimum": 0,
			},
		},
		"X-RateLimit-Reset": map[string]any{
			"description": "Unix timestamp when the rate limit window resets",
			"schema": map[string]any{
				"type":    "integer",
				"minimum": 1,
			},
		},
		"Retry-After": map[string]any{
			"description": "Seconds until the rate limit resets",
			"schema": map[string]any{
				"type":    "integer",
				"minimum": 1,
			},
		},
	}
}

// rateLimitSuccessHeaders documents the response headers emitted on
// successful requests when rate limiting is enabled.
func rateLimitSuccessHeaders() map[string]any {
	return map[string]any{
		"X-RateLimit-Limit": map[string]any{
			"description": "Maximum number of requests allowed in the current window",
			"schema": map[string]any{
				"type":    "integer",
				"minimum": 1,
			},
		},
		"X-RateLimit-Remaining": map[string]any{
			"description": "Number of requests remaining in the current window",
			"schema": map[string]any{
				"type":    "integer",
				"minimum": 0,
			},
		},
		"X-RateLimit-Reset": map[string]any{
			"description": "Unix timestamp when the rate limit window resets",
			"schema": map[string]any{
				"type":    "integer",
				"minimum": 1,
			},
		},
	}
}

// cacheSuccessHeaders documents the response header emitted on successful
// cache hits.
func cacheSuccessHeaders() map[string]any {
	return map[string]any{
		"X-Cache": map[string]any{
			"description": "Indicates the response was served from the in-memory cache",
			"schema": map[string]any{
				"type": "string",
			},
		},
	}
}

// standardResponseHeaders documents headers emitted by the response envelope
// middleware on all responses when request IDs are enabled.
func standardResponseHeaders() map[string]any {
	return map[string]any{
		"X-Request-ID": map[string]any{
			"description": "Request identifier propagated from the client or generated by the server",
			"schema": map[string]any{
				"type": "string",
			},
		},
	}
}

// mergeHeaders combines multiple OpenAPI header maps into one.
func mergeHeaders(sets ...map[string]any) map[string]any {
	merged := make(map[string]any)
	for _, set := range sets {
		for name, value := range set {
			merged[name] = value
		}
	}
	return merged
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
