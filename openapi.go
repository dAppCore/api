// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"iter"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// SpecBuilder constructs an OpenAPI 3.1 specification from registered RouteGroups.
// Title, Description, Version, and optional contact/licence/terms metadata populate the
// OpenAPI info block. Top-level external documentation metadata is also supported.
//
// Example:
//
//	builder := &api.SpecBuilder{Title: "Service", Version: "1.0.0"}
//	spec, err := builder.Build(engine.Groups())
type SpecBuilder struct {
	Title                   string
	Description             string
	Version                 string
	GraphQLPath             string
	SSEPath                 string
	TermsOfService          string
	ContactName             string
	ContactURL              string
	ContactEmail            string
	Servers                 []string
	LicenseName             string
	LicenseURL              string
	ExternalDocsDescription string
	ExternalDocsURL         string
}

type preparedRouteGroup struct {
	group RouteGroup
	descs []RouteDescription
}

// Build generates the complete OpenAPI 3.1 JSON spec.
// Groups implementing DescribableGroup contribute endpoint documentation.
// Other groups are listed as tags only.
//
// Example:
//
//	data, err := (&api.SpecBuilder{Title: "Service", Version: "1.0.0"}).Build(engine.Groups())
func (sb *SpecBuilder) Build(groups []RouteGroup) ([]byte, error) {
	prepared := prepareRouteGroups(groups)

	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       sb.Title,
			"description": sb.Description,
			"version":     sb.Version,
		},
		"paths": sb.buildPaths(prepared),
		"tags":  sb.buildTags(prepared),
		"security": []any{
			map[string]any{
				"bearerAuth": []any{},
			},
		},
	}

	if sb.LicenseName != "" {
		license := map[string]any{
			"name": sb.LicenseName,
		}
		if sb.LicenseURL != "" {
			license["url"] = sb.LicenseURL
		}
		spec["info"].(map[string]any)["license"] = license
	}

	if sb.TermsOfService != "" {
		spec["info"].(map[string]any)["termsOfService"] = sb.TermsOfService
	}

	if sb.ContactName != "" || sb.ContactURL != "" || sb.ContactEmail != "" {
		contact := map[string]any{}
		if sb.ContactName != "" {
			contact["name"] = sb.ContactName
		}
		if sb.ContactURL != "" {
			contact["url"] = sb.ContactURL
		}
		if sb.ContactEmail != "" {
			contact["email"] = sb.ContactEmail
		}
		spec["info"].(map[string]any)["contact"] = contact
	}

	if servers := normaliseServers(sb.Servers); len(servers) > 0 {
		out := make([]map[string]any, 0, len(servers))
		for _, server := range servers {
			out = append(out, map[string]any{"url": server})
		}
		spec["servers"] = out
	}

	if sb.ExternalDocsURL != "" {
		externalDocs := map[string]any{
			"url": sb.ExternalDocsURL,
		}
		if sb.ExternalDocsDescription != "" {
			externalDocs["description"] = sb.ExternalDocsDescription
		}
		spec["externalDocs"] = externalDocs
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
		"headers": deprecationHeaderComponents(),
	}

	return json.MarshalIndent(spec, "", "  ")
}

// BuildIter generates the complete OpenAPI 3.1 JSON spec from a route-group
// iterator. The iterator is snapshotted before building so the result stays
// stable even if the source changes during rendering.
func (sb *SpecBuilder) BuildIter(groups iter.Seq[RouteGroup]) ([]byte, error) {
	return sb.Build(collectRouteGroups(groups))
}

// buildPaths generates the paths object from all DescribableGroups.
func (sb *SpecBuilder) buildPaths(groups []preparedRouteGroup) map[string]any {
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

	if graphqlPath := strings.TrimSpace(sb.GraphQLPath); graphqlPath != "" {
		graphqlPath = normaliseOpenAPIPath(graphqlPath)
		paths[graphqlPath] = graphqlPathItem(graphqlPath, operationIDs)
	}

	if ssePath := strings.TrimSpace(sb.SSEPath); ssePath != "" {
		ssePath = normaliseOpenAPIPath(ssePath)
		paths[ssePath] = ssePathItem(ssePath, operationIDs)
	}

	for _, g := range groups {
		for _, rd := range g.descs {
			fullPath := joinOpenAPIPath(g.group.BasePath(), rd.Path)
			method := strings.ToLower(rd.Method)
			deprecated := rd.Deprecated || strings.TrimSpace(rd.SunsetDate) != "" || strings.TrimSpace(rd.Replacement) != ""
			deprecationHeaders := deprecationResponseHeaders(deprecated, rd.SunsetDate, rd.Replacement)

			operation := map[string]any{
				"summary":     rd.Summary,
				"description": rd.Description,
				"operationId": operationID(method, fullPath, operationIDs),
				"responses":   operationResponses(method, rd.StatusCode, rd.Response, rd.ResponseExample, rd.ResponseHeaders, rd.Security, deprecationHeaders),
			}
			if deprecated {
				operation["deprecated"] = true
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
			if tags := resolvedOperationTags(g.group, rd); len(tags) > 0 {
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
			// An example-only request body still produces a documented payload so
			// callers can see the expected shape even when a schema is omitted.
			if method != "get" && (rd.RequestBody != nil || rd.RequestExample != nil) {
				requestSchema := rd.RequestBody
				if requestSchema == nil {
					requestSchema = map[string]any{}
				}
				requestMediaType := map[string]any{
					"schema": requestSchema,
				}
				if rd.RequestExample != nil {
					requestMediaType["example"] = rd.RequestExample
				}

				operation["requestBody"] = map[string]any{
					"required": true,
					"content": map[string]any{
						"application/json": requestMediaType,
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
func operationResponses(method string, statusCode int, dataSchema map[string]any, example any, responseHeaders map[string]string, security []map[string][]string, deprecationHeaders map[string]any) map[string]any {
	documentedHeaders := documentedResponseHeaders(responseHeaders)
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders(), deprecationHeaders, documentedHeaders)
	if method == "get" {
		successHeaders = mergeHeaders(successHeaders, cacheSuccessHeaders())
	}

	isPublic := security != nil && len(security) == 0
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders(), deprecationHeaders, documentedHeaders)

	code := successStatusCode(statusCode)
	successResponse := map[string]any{
		"description": successResponseDescription(code),
		"headers":     successHeaders,
	}
	if !isNoContentStatus(code) {
		content := map[string]any{
			"schema": envelopeSchema(dataSchema),
		}
		if example != nil {
			// Example payloads are optional, but when a route provides one we
			// expose it alongside the schema so generated docs stay useful.
			content["example"] = example
		}

		successResponse["content"] = map[string]any{
			"application/json": content,
		}
	}

	responses := map[string]any{
		strconv.Itoa(code): successResponse,
		"400": map[string]any{
			"description": "Bad request",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": errorHeaders,
		},
		"429": map[string]any{
			"description": "Too many requests",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitHeaders(), deprecationHeaders, documentedHeaders),
		},
		"504": map[string]any{
			"description": "Gateway timeout",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": errorHeaders,
		},
		"500": map[string]any{
			"description": "Internal server error",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": errorHeaders,
		},
	}

	if !isPublic {
		responses["401"] = map[string]any{
			"description": "Unauthorised",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": errorHeaders,
		}
		responses["403"] = map[string]any{
			"description": "Forbidden",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": errorHeaders,
		}
	}

	return responses
}

func successStatusCode(statusCode int) int {
	if statusCode < 200 || statusCode > 299 {
		return http.StatusOK
	}
	if statusCode == 0 {
		return http.StatusOK
	}
	return statusCode
}

func isNoContentStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusNoContent, http.StatusResetContent:
		return true
	default:
		return false
	}
}

func successResponseDescription(statusCode int) string {
	switch statusCode {
	case http.StatusCreated:
		return "Created"
	case http.StatusAccepted:
		return "Accepted"
	case http.StatusNoContent:
		return "No content"
	case http.StatusResetContent:
		return "Reset content"
	default:
		return "Successful response"
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

// deprecationResponseHeaders documents the standard deprecation headers for
// deprecated or sunsetted operations.
func deprecationResponseHeaders(deprecated bool, sunsetDate, replacement string) map[string]any {
	sunsetDate = strings.TrimSpace(sunsetDate)
	replacement = strings.TrimSpace(replacement)

	if !deprecated && sunsetDate == "" && replacement == "" {
		return nil
	}

	headers := map[string]any{
		"Deprecation": map[string]any{
			"$ref": "#/components/headers/deprecation",
		},
		"X-API-Warn": map[string]any{
			"$ref": "#/components/headers/xapiwarn",
		},
	}

	if sunsetDate != "" {
		headers["Sunset"] = map[string]any{
			"$ref": "#/components/headers/sunset",
		}
	}

	if replacement != "" {
		headers["Link"] = map[string]any{
			"$ref": "#/components/headers/link",
		}
	}

	return headers
}

// deprecationHeaderComponents returns reusable OpenAPI header components for
// the standard deprecation and sunset middleware headers.
func deprecationHeaderComponents() map[string]any {
	return map[string]any{
		"deprecation": map[string]any{
			"description": "Indicates that the endpoint is deprecated.",
			"schema": map[string]any{
				"type": "string",
				"enum": []string{"true"},
			},
		},
		"sunset": map[string]any{
			"description": "The date and time after which the endpoint will no longer be supported.",
			"schema": map[string]any{
				"type":   "string",
				"format": "date-time",
			},
		},
		"link": map[string]any{
			"description": "Reference to the successor endpoint, when one is provided.",
			"schema": map[string]any{
				"type": "string",
			},
		},
		"xapiwarn": map[string]any{
			"description": "Human-readable deprecation warning for clients.",
			"schema": map[string]any{
				"type": "string",
			},
		},
	}
}

// buildTags generates the tags array from all RouteGroups.
func (sb *SpecBuilder) buildTags(groups []preparedRouteGroup) []map[string]any {
	tags := []map[string]any{
		{"name": "system", "description": "System endpoints"},
	}
	seen := map[string]bool{"system": true}

	if graphqlPath := strings.TrimSpace(sb.GraphQLPath); graphqlPath != "" && !seen["graphql"] {
		tags = append(tags, map[string]any{
			"name":        "graphql",
			"description": "GraphQL endpoints",
		})
		seen["graphql"] = true
	}

	if ssePath := strings.TrimSpace(sb.SSEPath); ssePath != "" && !seen["events"] {
		tags = append(tags, map[string]any{
			"name":        "events",
			"description": "Server-Sent Events endpoints",
		})
		seen["events"] = true
	}

	for _, g := range groups {
		name := strings.TrimSpace(g.group.Name())
		if name != "" && !seen[name] {
			tags = append(tags, map[string]any{
				"name":        name,
				"description": name + " endpoints",
			})
			seen[name] = true
		}

		for _, rd := range g.descs {
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

	sortTags(tags)

	return tags
}

// sortTags keeps system first and orders the remaining tags alphabetically so
// generated specs stay stable across registration order changes.
func sortTags(tags []map[string]any) {
	if len(tags) < 2 {
		return
	}

	sort.SliceStable(tags, func(i, j int) bool {
		left, _ := tags[i]["name"].(string)
		right, _ := tags[j]["name"].(string)

		switch {
		case left == "system":
			return true
		case right == "system":
			return false
		default:
			return left < right
		}
	})
}

func graphqlPathItem(path string, operationIDs map[string]int) map[string]any {
	return map[string]any{
		"post": map[string]any{
			"summary":     "GraphQL query",
			"description": "Executes GraphQL queries and mutations",
			"tags":        []string{"graphql"},
			"operationId": operationID("post", path, operationIDs),
			"security": []any{
				map[string]any{
					"bearerAuth": []any{},
				},
			},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": graphqlRequestSchema(),
					},
				},
			},
			"responses": graphqlResponses(),
		},
	}
}

func ssePathItem(path string, operationIDs map[string]int) map[string]any {
	return map[string]any{
		"get": map[string]any{
			"summary":     "Server-Sent Events stream",
			"description": "Streams published events as text/event-stream",
			"tags":        []string{"events"},
			"operationId": operationID("get", path, operationIDs),
			"security": []any{
				map[string]any{
					"bearerAuth": []any{},
				},
			},
			"parameters": []map[string]any{
				{
					"name":        "channel",
					"in":          "query",
					"required":    false,
					"description": "Restrict the stream to a specific channel",
					"schema": map[string]any{
						"type": "string",
					},
				},
			},
			"responses": sseResponses(),
		},
	}
}

func graphqlRequestSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type": "string",
			},
			"variables": map[string]any{
				"type":                 "object",
				"additionalProperties": true,
			},
			"operationName": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"query"},
	}
}

func graphqlResponses() map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders(), cacheSuccessHeaders())
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())

	return map[string]any{
		"200": map[string]any{
			"description": "GraphQL response",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": successHeaders,
		},
		"400": map[string]any{
			"description": "Bad request",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": errorHeaders,
		},
		"401": map[string]any{
			"description": "Unauthorised",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": errorHeaders,
		},
		"403": map[string]any{
			"description": "Forbidden",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": errorHeaders,
		},
		"429": map[string]any{
			"description": "Too many requests",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitHeaders()),
		},
		"500": map[string]any{
			"description": "Internal server error",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": errorHeaders,
		},
		"504": map[string]any{
			"description": "Gateway timeout",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": errorHeaders,
		},
	}
}

func sseResponses() map[string]any {
	successHeaders := mergeHeaders(
		standardResponseHeaders(),
		rateLimitSuccessHeaders(),
		sseResponseHeaders(),
	)
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())

	return map[string]any{
		"200": map[string]any{
			"description": "Event stream",
			"content": map[string]any{
				"text/event-stream": map[string]any{
					"schema": map[string]any{
						"type": "string",
					},
				},
			},
			"headers": successHeaders,
		},
		"401": map[string]any{
			"description": "Unauthorised",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": errorHeaders,
		},
		"403": map[string]any{
			"description": "Forbidden",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": errorHeaders,
		},
		"429": map[string]any{
			"description": "Too many requests",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitHeaders()),
		},
		"500": map[string]any{
			"description": "Internal server error",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": errorHeaders,
		},
		"504": map[string]any{
			"description": "Gateway timeout",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
			"headers": errorHeaders,
		},
	}
}

// prepareRouteGroups snapshots route descriptions once per group so iterator-
// backed implementations can be consumed safely by both tag and path builders.
func prepareRouteGroups(groups []RouteGroup) []preparedRouteGroup {
	if len(groups) == 0 {
		return nil
	}

	out := make([]preparedRouteGroup, 0, len(groups))
	for _, g := range groups {
		if g == nil {
			continue
		}
		if isHiddenRouteGroup(g) {
			continue
		}
		out = append(out, preparedRouteGroup{
			group: g,
			descs: collectRouteDescriptions(g),
		})
	}

	return out
}

func collectRouteGroups(groups iter.Seq[RouteGroup]) []RouteGroup {
	if groups == nil {
		return nil
	}

	out := make([]RouteGroup, 0)
	for group := range groups {
		out = append(out, group)
	}

	return out
}

func collectRouteDescriptions(g RouteGroup) []RouteDescription {
	descIter := routeDescriptions(g)
	if descIter == nil {
		return nil
	}

	descs := make([]RouteDescription, 0)
	for rd := range descIter {
		if rd.Hidden {
			continue
		}
		descs = append(descs, rd)
	}

	return descs
}

func isHiddenRouteGroup(g RouteGroup) bool {
	type hiddenRouteGroup interface {
		Hidden() bool
	}

	hg, ok := g.(hiddenRouteGroup)
	return ok && hg.Hidden()
}

// routeDescriptions returns OpenAPI route descriptions for a group.
// Iterator-backed implementations are preferred when available so builders
// can avoid slice allocation.
func routeDescriptions(g RouteGroup) iter.Seq[RouteDescription] {
	if dg, ok := g.(DescribableGroupIter); ok {
		if descIter := dg.DescribeIter(); descIter != nil {
			return descIter
		}
	}
	if dg, ok := g.(DescribableGroup); ok {
		descs := dg.Describe()
		return func(yield func(RouteDescription) bool) {
			for _, rd := range descs {
				if !yield(rd) {
					return
				}
			}
		}
	}
	return nil
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
	if tags := cleanTags(rd.Tags); len(tags) > 0 {
		return tags
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

// sseResponseHeaders documents the response headers emitted by the SSE stream.
func sseResponseHeaders() map[string]any {
	return map[string]any{
		"Cache-Control": map[string]any{
			"description": "Prevents intermediaries from caching the event stream",
			"schema": map[string]any{
				"type": "string",
			},
		},
		"Connection": map[string]any{
			"description": "Keeps the HTTP connection open for streaming",
			"schema": map[string]any{
				"type": "string",
			},
		},
		"X-Accel-Buffering": map[string]any{
			"description": "Disables buffering in compatible reverse proxies",
			"schema": map[string]any{
				"type": "string",
			},
		},
	}
}

// documentedResponseHeaders converts route-specific response header metadata
// into OpenAPI header objects.
func documentedResponseHeaders(headers map[string]string) map[string]any {
	if len(headers) == 0 {
		return nil
	}

	out := make(map[string]any, len(headers))
	for name, description := range headers {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		out[name] = map[string]any{
			"description": description,
			"schema": map[string]any{
				"type": "string",
			},
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
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
