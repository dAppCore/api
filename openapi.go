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
			fullPath := joinOpenAPIPath(g.BasePath(), rd.Path)
			method := strings.ToLower(rd.Method)

			operation := map[string]any{
				"summary":     rd.Summary,
				"description": rd.Description,
				"operationId": operationID(method, fullPath, operationIDs),
				"responses":   operationResponses(method, rd.Response),
			}
			if rd.Security != nil {
				operation["security"] = rd.Security
			} else {
				operation["security"] = []any{
					map[string]any{
						"bearerAuth": []any{},
					},
				}
			}
			if tags := resolvedOperationTags(g, rd); len(tags) > 0 {
				operation["tags"] = tags
			}

			if params := pathParameters(fullPath); len(params) > 0 {
				operation["parameters"] = params
			}
			if explicit := operationParameters(rd.Parameters); len(explicit) > 0 {
				operation["parameters"] = mergeOperationParameters(operation["parameters"], explicit)
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

// joinOpenAPIPath normalises a base path and relative route path into a single
// OpenAPI path without duplicate or missing separators. Gin-style parameters
// such as :id and *path are converted to OpenAPI template parameters.
func joinOpenAPIPath(basePath, routePath string) string {
	basePath = strings.TrimSpace(basePath)
	routePath = strings.TrimSpace(routePath)

	if basePath == "" {
		basePath = "/"
	}
	if routePath == "" || routePath == "/" {
		return normaliseOpenAPIPath(basePath)
	}

	basePath = normaliseOpenAPIPath(basePath)
	routePath = normaliseOpenAPIPath(routePath)

	if basePath == "/" {
		return routePath
	}

	return strings.TrimRight(basePath, "/") + "/" + strings.TrimPrefix(routePath, "/")
}

// normaliseOpenAPIPath trims whitespace and collapses trailing separators
// while preserving the root path and converting Gin-style path parameters.
func normaliseOpenAPIPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}

	segments := strings.Split(path, "/")
	cleaned := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		switch {
		case strings.HasPrefix(segment, ":") && len(segment) > 1:
			segment = "{" + segment[1:] + "}"
		case strings.HasPrefix(segment, "*") && len(segment) > 1:
			segment = "{" + segment[1:] + "}"
		}
		cleaned = append(cleaned, segment)
	}

	if len(cleaned) == 0 {
		return "/"
	}

	return "/" + strings.Join(cleaned, "/")
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
		name := strings.TrimSpace(g.Name())
		if name != "" && !seen[name] {
			tags = append(tags, map[string]any{
				"name":        name,
				"description": name + " endpoints",
			})
			seen[name] = true
		}

		dg, ok := g.(DescribableGroup)
		if !ok {
			continue
		}

		for _, rd := range dg.Describe() {
			for _, tag := range rd.Tags {
				tag = strings.TrimSpace(tag)
				if tag == "" || seen[tag] {
					continue
				}
				tags = append(tags, map[string]any{
					"name":        tag,
					"description": tag + " endpoints",
				})
				seen[tag] = true
			}
		}
	}

	return tags
}

// pathParameters extracts unique OpenAPI path parameters from a path template.
// Parameters are returned in the order they appear in the path.
func pathParameters(path string) []map[string]any {
	const (
		open  = '{'
		close = '}'
	)

	seen := map[string]bool{}
	params := make([]map[string]any, 0)

	for i := 0; i < len(path); i++ {
		if path[i] != open {
			continue
		}
		end := strings.IndexByte(path[i+1:], close)
		if end < 0 {
			continue
		}
		name := path[i+1 : i+1+end]
		if name == "" || strings.ContainsAny(name, "/{}") || seen[name] {
			continue
		}
		seen[name] = true
		params = append(params, map[string]any{
			"name":     name,
			"in":       "path",
			"required": true,
			"schema": map[string]any{
				"type": "string",
			},
		})
		i += end + 1
	}

	return params
}

// operationParameters converts explicit route parameter descriptions into
// OpenAPI parameter objects.
func operationParameters(params []ParameterDescription) []map[string]any {
	if len(params) == 0 {
		return nil
	}

	out := make([]map[string]any, 0, len(params))
	for _, param := range params {
		if param.Name == "" || param.In == "" {
			continue
		}

		entry := map[string]any{
			"name":     param.Name,
			"in":       param.In,
			"required": param.Required || param.In == "path",
		}
		if param.Description != "" {
			entry["description"] = param.Description
		}
		if param.Deprecated {
			entry["deprecated"] = true
		}
		if len(param.Schema) > 0 {
			entry["schema"] = param.Schema
		} else if param.In == "path" || param.In == "query" || param.In == "header" || param.In == "cookie" {
			entry["schema"] = map[string]any{"type": "string"}
		}
		if param.Example != nil {
			entry["example"] = param.Example
		}

		out = append(out, entry)
	}

	return out
}

// mergeOperationParameters combines generated and explicit parameter
// definitions, letting explicit entries override auto-generated path params.
func mergeOperationParameters(existing any, explicit []map[string]any) []map[string]any {
	merged := make([]map[string]any, 0, len(explicit))
	index := map[string]int{}

	add := func(param map[string]any) {
		name, _ := param["name"].(string)
		in, _ := param["in"].(string)
		if name == "" || in == "" {
			return
		}
		key := in + ":" + name
		if pos, ok := index[key]; ok {
			merged[pos] = param
			return
		}
		index[key] = len(merged)
		merged = append(merged, param)
	}

	if params, ok := existing.([]map[string]any); ok {
		for _, param := range params {
			add(param)
		}
	}

	for _, param := range explicit {
		add(param)
	}

	if len(merged) == 0 {
		return nil
	}

	return merged
}

// resolvedOperationTags returns the explicit route tags when provided, or a
// stable fallback derived from the group's name when the route omits tags.
func resolvedOperationTags(g RouteGroup, rd RouteDescription) []string {
	if len(rd.Tags) > 0 {
		return cleanTags(rd.Tags)
	}

	if name := strings.TrimSpace(g.Name()); name != "" {
		return []string{name}
	}

	return nil
}

// cleanTags trims whitespace and removes empty or duplicate tags while
// preserving the first occurrence of each name.
func cleanTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	cleaned := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		cleaned = append(cleaned, tag)
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
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
