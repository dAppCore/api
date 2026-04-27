// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"iter"
	"net/http"
	"slices"
	"time"
	"unicode" // Note: AX-6 — Unicode-aware operationId normalization has no core primitive.

	core "dappco.re/go/core"
)

// SpecBuilder constructs an OpenAPI 3.1 specification from registered RouteGroups.
// Title, Summary, Description, Version, and optional contact/licence/terms metadata populate the
// OpenAPI info block. Top-level external documentation metadata is also supported, along with
// additive extension fields that describe runtime transport, cache, i18n, and Authentik settings.
//
// Example:
//
//	builder := &api.SpecBuilder{Title: "Service", Version: "1.0.0"}
//	spec, err := builder.Build(engine.Groups())
type SpecBuilder struct {
	Title                   string
	Summary                 string
	Description             string
	Version                 string
	SwaggerEnabled          bool
	SwaggerPath             string
	GraphQLEnabled          bool
	GraphQLPath             string
	GraphQLPlayground       bool
	GraphQLPlaygroundPath   string
	WSPath                  string
	WSEnabled               bool
	SSEPath                 string
	SSEEnabled              bool
	TermsOfService          string
	ContactName             string
	ContactURL              string
	ContactEmail            string
	Servers                 []string
	LicenseName             string
	LicenseURL              string
	SecuritySchemes         map[string]any
	ExternalDocsDescription string
	ExternalDocsURL         string
	PprofEnabled            bool
	ExpvarEnabled           bool
	ChatCompletionsEnabled  bool
	ChatCompletionsPath     string
	OpenAPISpecEnabled      bool
	OpenAPISpecPath         string
	CacheEnabled            bool
	CacheTTL                string
	CacheMaxEntries         int
	CacheMaxBytes           int
	I18nDefaultLocale       string
	I18nSupportedLocales    []string
	AuthentikIssuer         string
	AuthentikClientID       string
	AuthentikTrustedProxy   bool
	AuthentikPublicPaths    []string
}

type preparedRouteGroup struct {
	name     string
	basePath string
	descs    []RouteDescription
}

const openAPIDialect = "https://spec.openapis.org/oas/3.1/dialect/base"

// Build generates the complete OpenAPI 3.1 JSON spec.
// Groups implementing DescribableGroup contribute endpoint documentation.
// Other groups are listed as tags only.
//
// Example:
//
//	data, err := (&api.SpecBuilder{Title: "Service", Version: "1.0.0"}).Build(engine.Groups())
func (sb *SpecBuilder) Build(groups []RouteGroup) ([]byte, error) {
	if sb == nil {
		sb = &SpecBuilder{}
	}
	sb = sb.snapshot()

	prepared := prepareRouteGroups(groups)

	info := map[string]any{
		"title":       sb.Title,
		"description": sb.Description,
		"version":     sb.Version,
	}
	if sb.Summary != "" {
		info["summary"] = sb.Summary
	}

	spec := map[string]any{
		"openapi":           "3.1.0",
		"jsonSchemaDialect": openAPIDialect,
		"info":              info,
		"paths":             sb.buildPaths(prepared),
		"tags":              sb.buildTags(prepared),
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
	if swaggerPath := sb.effectiveSwaggerPath(); swaggerPath != "" {
		spec["x-swagger-ui-path"] = normaliseSwaggerPath(swaggerPath)
	}
	if sb.SwaggerEnabled {
		spec["x-swagger-enabled"] = true
	}
	if sb.GraphQLEnabled {
		spec["x-graphql-enabled"] = true
	}
	if graphqlPath := sb.effectiveGraphQLPath(); graphqlPath != "" {
		spec["x-graphql-path"] = normaliseOpenAPIPath(graphqlPath)
		if sb.GraphQLPlayground {
			spec["x-graphql-playground"] = true
		}
	}
	if sb.GraphQLPlayground {
		if playgroundPath := sb.effectiveGraphQLPlaygroundPath(); playgroundPath != "" {
			spec["x-graphql-playground-path"] = normaliseOpenAPIPath(playgroundPath)
		}
	}
	if wsPath := sb.effectiveWSPath(); wsPath != "" {
		spec["x-ws-path"] = normaliseOpenAPIPath(wsPath)
	}
	if sb.WSEnabled {
		spec["x-ws-enabled"] = true
	}
	if ssePath := sb.effectiveSSEPath(); ssePath != "" {
		spec["x-sse-path"] = normaliseOpenAPIPath(ssePath)
	}
	if sb.SSEEnabled {
		spec["x-sse-enabled"] = true
	}
	if sb.PprofEnabled {
		spec["x-pprof-enabled"] = true
	}
	if sb.ExpvarEnabled {
		spec["x-expvar-enabled"] = true
	}
	if sb.ChatCompletionsEnabled {
		spec["x-chat-completions-enabled"] = true
	}
	if path := sb.effectiveChatCompletionsPath(); path != "" {
		spec["x-chat-completions-path"] = normaliseOpenAPIPath(path)
	}
	if sb.OpenAPISpecEnabled {
		spec["x-openapi-spec-enabled"] = true
	}
	if path := sb.effectiveOpenAPISpecPath(); path != "" {
		spec["x-openapi-spec-path"] = normaliseOpenAPIPath(path)
	}
	if sb.CacheEnabled {
		spec["x-cache-enabled"] = true
	}
	if ttl := sb.effectiveCacheTTL(); ttl != "" {
		spec["x-cache-ttl"] = ttl
	}
	if sb.CacheMaxEntries > 0 {
		spec["x-cache-max-entries"] = sb.CacheMaxEntries
	}
	if sb.CacheMaxBytes > 0 {
		spec["x-cache-max-bytes"] = sb.CacheMaxBytes
	}
	if locale := core.Trim(sb.I18nDefaultLocale); locale != "" {
		spec["x-i18n-default-locale"] = locale
	}
	if len(sb.I18nSupportedLocales) > 0 {
		spec["x-i18n-supported-locales"] = slices.Clone(sb.I18nSupportedLocales)
	}
	if issuer := core.Trim(sb.AuthentikIssuer); issuer != "" {
		spec["x-authentik-issuer"] = issuer
	}
	if clientID := core.Trim(sb.AuthentikClientID); clientID != "" {
		spec["x-authentik-client-id"] = clientID
	}
	if sb.AuthentikTrustedProxy {
		spec["x-authentik-trusted-proxy"] = true
	}
	if paths := sb.effectiveAuthentikPublicPaths(); len(paths) > 0 {
		spec["x-authentik-public-paths"] = paths
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
		"securitySchemes": securitySchemeComponents(sb.SecuritySchemes),
		"headers":         deprecationHeaderComponents(),
		"responses":       responseComponents(),
	}

	return marshalCoreJSONIndent(spec, "", "  ")
}

// BuildIter generates the complete OpenAPI 3.1 JSON spec from a route-group
// iterator. The iterator is snapshotted before building so the result stays
// stable even if the source changes during rendering.
//
// Example:
//
//	data, err := (&api.SpecBuilder{Title: "Service"}).BuildIter(api.RegisteredSpecGroupsIter())
func (sb *SpecBuilder) BuildIter(groups iter.Seq[RouteGroup]) ([]byte, error) {
	if sb == nil {
		sb = &SpecBuilder{}
	}

	return sb.Build(collectRouteGroups(groups))
}

// buildPaths generates the paths object from all DescribableGroups.
func (sb *SpecBuilder) buildPaths(groups []preparedRouteGroup) map[string]any {
	operationIDs := map[string]int{}
	publicPaths := sb.effectiveAuthentikPublicPaths()
	paths := map[string]any{
		// Built-in health endpoint.
		"/health": map[string]any{
			"get": map[string]any{
				"summary":     "Health check",
				"description": "Returns server health status",
				"tags":        []string{"system"},
				"operationId": operationID("get", "/health", operationIDs),
				"responses":   healthResponses(sb.CacheEnabled),
			},
		},
	}

	graphqlPath := sb.effectiveGraphQLPath()
	if graphqlPath != "" {
		graphqlPath = normaliseOpenAPIPath(graphqlPath)
		item := graphqlPathItem(graphqlPath, operationIDs, sb.CacheEnabled)
		if isPublicPathForList(graphqlPath, publicPaths) {
			makePathItemPublic(item)
		}
		paths[graphqlPath] = item
		if sb.GraphQLPlayground {
			playgroundPath := sb.effectiveGraphQLPlaygroundPath()
			if playgroundPath == "" {
				playgroundPath = graphqlPath + "/playground"
			}
			playgroundPath = normaliseOpenAPIPath(playgroundPath)
			item := graphqlPlaygroundPathItem(playgroundPath, operationIDs)
			if isPublicPathForList(playgroundPath, publicPaths) {
				makePathItemPublic(item)
			}
			paths[playgroundPath] = item
		}
	}

	if wsPath := sb.effectiveWSPath(); wsPath != "" {
		wsPath = normaliseOpenAPIPath(wsPath)
		item := wsPathItem(wsPath, operationIDs)
		if isPublicPathForList(wsPath, publicPaths) {
			makePathItemPublic(item)
		}
		paths[wsPath] = item
	}

	if ssePath := sb.effectiveSSEPath(); ssePath != "" {
		ssePath = normaliseOpenAPIPath(ssePath)
		item := ssePathItem(ssePath, operationIDs)
		if isPublicPathForList(ssePath, publicPaths) {
			makePathItemPublic(item)
		}
		paths[ssePath] = item
		if ssePath == defaultSSEPath {
			legacyPath := legacySSEPath
			legacyItem := ssePathItem(legacyPath, operationIDs)
			if isPublicPathForList(legacyPath, publicPaths) {
				makePathItemPublic(legacyItem)
			}
			paths[legacyPath] = legacyItem
		}
	}

	if sb.PprofEnabled {
		item := pprofPathItem(operationIDs)
		if isPublicPathForList("/debug/pprof", publicPaths) {
			makePathItemPublic(item)
		}
		paths["/debug/pprof"] = item
	}

	if sb.ExpvarEnabled {
		item := expvarPathItem(operationIDs)
		if isPublicPathForList("/debug/vars", publicPaths) {
			makePathItemPublic(item)
		}
		paths["/debug/vars"] = item
	}

	if specPath := sb.effectiveOpenAPISpecPath(); specPath != "" {
		specPath = normaliseOpenAPIPath(specPath)
		item := openAPISpecPathItem(specPath, operationIDs)
		if isPublicPathForList(specPath, publicPaths) {
			makePathItemPublic(item)
		}
		paths[specPath] = item
	}

	if chatPath := sb.effectiveChatCompletionsPath(); chatPath != "" {
		chatPath = normaliseOpenAPIPath(chatPath)
		item := chatCompletionsPathItem(chatPath, operationIDs)
		if isPublicPathForList(chatPath, publicPaths) {
			makePathItemPublic(item)
		}
		paths[chatPath] = item
	}

	for _, g := range groups {
		for _, rd := range g.descs {
			fullPath := joinOpenAPIPath(g.basePath, rd.Path)
			method := core.Lower(rd.Method)
			replacement := core.Trim(rd.ReplacementURL)
			if replacement == "" {
				replacement = core.Trim(rd.Replacement)
			}
			deprecated := rd.Deprecated || core.Trim(rd.SunsetDate) != "" || replacement != "" || core.Trim(rd.NoticeURL) != ""
			deprecationHeaders := deprecationResponseHeaders(deprecated, rd.SunsetDate, replacement)
			if deprecated && core.Trim(rd.NoticeURL) != "" && deprecationHeaders != nil {
				deprecationHeaders["API-Deprecation-Notice-URL"] = map[string]any{
					"$ref": "#/components/headers/apiDeprecationNoticeURL",
				}
			}
			isPublic := isPublicPathForList(fullPath, publicPaths)
			security := rd.Security
			if isPublic {
				security = []map[string][]string{}
			}

			operation := map[string]any{
				"summary":     rd.Summary,
				"description": rd.Description,
				"operationId": resolvedOperationID(rd, method, fullPath, operationIDs),
				"responses":   operationResponses(method, rd.StatusCode, rd.Response, rd.ResponseExample, rd.ResponseHeaders, security, deprecated, rd.SunsetDate, replacement, deprecationHeaders, sb.CacheEnabled, rd.CacheControl),
			}
			if deprecated {
				operation["deprecated"] = true
			}
			if isPublic {
				operation["security"] = []any{}
			} else if security != nil {
				operation["security"] = security
			} else {
				operation["security"] = []any{
					map[string]any{
						"bearerAuth": []any{},
					},
				}
			}
			if tags := resolvedOperationTags(g.name, rd); len(tags) > 0 {
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
			if renderHints := resolvedRenderHints(rd); len(renderHints) > 0 {
				operation["x-render-hints"] = renderHints
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
	basePath = core.Trim(basePath)
	routePath = core.Trim(routePath)

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

	return trimTrailingSlashes(basePath) + "/" + core.TrimPrefix(routePath, "/")
}

// normaliseOpenAPIPath trims whitespace and collapses trailing separators
// while preserving the root path and converting Gin-style path parameters.
func normaliseOpenAPIPath(path string) string {
	path = core.Trim(path)
	if path == "" {
		return "/"
	}

	segments := core.Split(path, "/")
	cleaned := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = core.Trim(segment)
		if segment == "" {
			continue
		}
		switch {
		case core.HasPrefix(segment, ":") && len(segment) > 1:
			segment = "{" + segment[1:] + "}"
		case core.HasPrefix(segment, "*") && len(segment) > 1:
			segment = "{" + segment[1:] + "}"
		}
		cleaned = append(cleaned, segment)
	}

	if len(cleaned) == 0 {
		return "/"
	}

	return "/" + core.Join("/", cleaned...)
}

// operationResponses builds the standard response set for a documented API
// operation. The framework always exposes the common envelope responses, plus
// middleware-driven 429 and 504 errors.
func operationResponses(method string, statusCode int, dataSchema map[string]any, example any, responseHeaders map[string]string, security []map[string][]string, deprecated bool, sunsetDate, replacement string, deprecationHeaders map[string]any, cacheEnabled bool, cacheControl string) map[string]any {
	documentedHeaders := documentedResponseHeaders(responseHeaders)
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders(), deprecationHeaders, documentedHeaders)
	if method == "get" && cacheEnabled {
		successHeaders = mergeHeaders(successHeaders, cacheSuccessHeaders())
	}
	if cacheControl = core.Trim(cacheControl); cacheControl != "" {
		successHeaders = mergeHeaders(successHeaders, cacheControlHeaders(cacheControl))
	}

	isPublic := security != nil && len(security) == 0
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders(), deprecationHeaders, documentedHeaders)

	code := successStatusCode(statusCode)
	if dataSchema == nil && example != nil {
		dataSchema = map[string]any{}
	}
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
		core.Itoa(code): successResponse,
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

	if deprecated && (core.Trim(sunsetDate) != "" || core.Trim(replacement) != "") {
		responses["410"] = map[string]any{
			"description": "Gone",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": errorHeaders,
		}
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

func cacheControlHeaders(cacheControl string) map[string]any {
	return map[string]any{
		"Cache-Control": map[string]any{
			"description": "Caching policy hint for successful responses",
			"schema": map[string]any{
				"type":    "string",
				"example": cacheControl,
			},
		},
	}
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
func healthResponses(cacheEnabled bool) map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	if cacheEnabled {
		successHeaders = mergeHeaders(successHeaders, cacheSuccessHeaders())
	}

	return map[string]any{
		"200": map[string]any{
			"description": "Server is healthy",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(map[string]any{"type": "string"}),
				},
			},
			"headers": successHeaders,
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
// deprecated or sunsetted operations. The header set mirrors what
// ApiSunset/ApiSunsetWith emit at runtime, including the spec §8 custom
// headers (API-Suggested-Replacement, API-Deprecation-Notice-URL).
func deprecationResponseHeaders(deprecated bool, sunsetDate, replacement string) map[string]any {
	sunsetDate = core.Trim(sunsetDate)
	replacement = core.Trim(replacement)

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
		headers["API-Suggested-Replacement"] = map[string]any{
			"$ref": "#/components/headers/apiSuggestedReplacement",
		}
	}

	return headers
}

// deprecationHeaderComponents returns reusable OpenAPI header components for
// the standard deprecation and sunset middleware headers. Includes both the
// IETF-standard headers (Deprecation, Sunset, Link) and the custom spec §8
// headers used to communicate replacement endpoints and migration guides.
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
		"apiSuggestedReplacement": map[string]any{
			"description": "Suggested replacement endpoint for clients to migrate to.",
			"schema": map[string]any{
				"type": "string",
			},
		},
		"apiDeprecationNoticeURL": map[string]any{
			"description": "URL pointing to a detailed deprecation notice.",
			"schema": map[string]any{
				"type":   "string",
				"format": "uri",
			},
		},
	}
}

// responseComponents returns reusable OpenAPI response objects for the
// common error cases exposed by the framework. The path operations still
// inline their concrete headers so existing callers keep the same output,
// but these components make the response catalogue available for reuse.
func responseComponents() map[string]any {
	return map[string]any{
		"BadRequest": map[string]any{
			"description": "Bad request",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": standardResponseHeaders(),
		},
		"Unauthorized": map[string]any{
			"description": "Unauthorised",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": standardResponseHeaders(),
		},
		"Forbidden": map[string]any{
			"description": "Forbidden",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": standardResponseHeaders(),
		},
		"RateLimitExceeded": map[string]any{
			"description": "Too many requests",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), rateLimitHeaders()),
		},
		"GatewayTimeout": map[string]any{
			"description": "Gateway timeout",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": standardResponseHeaders(),
		},
		"InternalServerError": map[string]any{
			"description": "Internal server error",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": standardResponseHeaders(),
		},
		"Gone": map[string]any{
			"description": "Gone",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": envelopeSchema(nil),
				},
			},
			"headers": mergeHeaders(standardResponseHeaders(), deprecationResponseHeaders(true, "", "")),
		},
	}
}

// securitySchemeComponents builds the OpenAPI security scheme registry.
// bearerAuth stays available by default, while callers can add or override
// additional scheme definitions for custom security requirements.
func securitySchemeComponents(overrides map[string]any) map[string]any {
	schemes := map[string]any{
		"bearerAuth": map[string]any{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
		},
	}

	for name, scheme := range overrides {
		name = core.Trim(name)
		if name == "" || scheme == nil {
			continue
		}
		schemes[name] = cloneOpenAPIValue(scheme)
	}

	return schemes
}

// buildTags generates the tags array from all RouteGroups.
func (sb *SpecBuilder) buildTags(groups []preparedRouteGroup) []map[string]any {
	tags := []map[string]any{
		{"name": "system", "description": "System endpoints"},
	}
	seen := map[string]bool{"system": true}

	if graphqlPath := sb.effectiveGraphQLPath(); graphqlPath != "" && !seen["graphql"] {
		tags = append(tags, map[string]any{
			"name":        "graphql",
			"description": "GraphQL endpoints",
		})
		seen["graphql"] = true
	}

	if ssePath := sb.effectiveSSEPath(); ssePath != "" && !seen["events"] {
		tags = append(tags, map[string]any{
			"name":        "events",
			"description": "Server-Sent Events endpoints",
		})
		seen["events"] = true
	}

	if (sb.PprofEnabled || sb.ExpvarEnabled) && !seen["debug"] {
		tags = append(tags, map[string]any{
			"name":        "debug",
			"description": "Runtime debug endpoints",
		})
		seen["debug"] = true
	}

	if sb.effectiveChatCompletionsPath() != "" && !seen["inference"] {
		tags = append(tags, map[string]any{
			"name":        "inference",
			"description": "Local inference endpoints (OpenAI-compatible)",
		})
		seen["inference"] = true
	}

	for _, g := range groups {
		name := core.Trim(g.name)
		if name != "" && !seen[name] {
			tags = append(tags, map[string]any{
				"name":        name,
				"description": name + " endpoints",
			})
			seen[name] = true
		}

		for _, rd := range g.descs {
			for _, tag := range rd.Tags {
				tag = core.Trim(tag)
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

	slices.SortStableFunc(tags, func(a, b map[string]any) int {
		left, _ := a["name"].(string)
		right, _ := b["name"].(string)

		switch {
		case left == "system":
			return -1
		case right == "system":
			return 1
		case left < right:
			return -1
		case left > right:
			return 1
		default:
			return 0
		}
	})
}

func graphqlPathItem(path string, operationIDs map[string]int, cacheEnabled bool) map[string]any {
	return map[string]any{
		"get": map[string]any{
			"summary":     "GraphQL query",
			"description": "Executes GraphQL queries over GET using query parameters",
			"tags":        []string{"graphql"},
			"operationId": operationID("get", path, operationIDs),
			"security": []any{
				map[string]any{
					"bearerAuth": []any{},
				},
			},
			"parameters": graphqlQueryParameters(),
			"responses":  graphqlResponses(cacheEnabled),
		},
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
			"responses": graphqlResponses(cacheEnabled),
		},
	}
}

func graphqlPlaygroundPathItem(path string, operationIDs map[string]int) map[string]any {
	return map[string]any{
		"get": map[string]any{
			"summary":     "GraphQL playground",
			"description": "Interactive GraphQL IDE for the configured schema",
			"tags":        []string{"graphql"},
			"operationId": operationID("get", path, operationIDs),
			"security": []any{
				map[string]any{
					"bearerAuth": []any{},
				},
			},
			"responses": graphqlPlaygroundResponses(),
		},
	}
}

func wsPathItem(path string, operationIDs map[string]int) map[string]any {
	return map[string]any{
		"get": map[string]any{
			"summary":     "WebSocket connection",
			"description": "Upgrades the connection to a WebSocket stream",
			"tags":        []string{"system"},
			"operationId": operationID("get", path, operationIDs),
			"security": []any{
				map[string]any{
					"bearerAuth": []any{},
				},
			},
			"responses": wsResponses(),
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

func wsResponses() map[string]any {
	successHeaders := mergeHeaders(
		standardResponseHeaders(),
		rateLimitSuccessHeaders(),
		wsUpgradeHeaders(),
	)
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())

	return map[string]any{
		"101": map[string]any{
			"description": "Switching protocols",
			"headers":     successHeaders,
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

func wsUpgradeHeaders() map[string]any {
	return map[string]any{
		"Upgrade": map[string]any{
			"description": "Indicates that the connection has switched to WebSocket",
			"schema": map[string]any{
				"type": "string",
			},
		},
		"Connection": map[string]any{
			"description": "Keeps the upgraded connection open",
			"schema": map[string]any{
				"type": "string",
			},
		},
		"Sec-WebSocket-Accept": map[string]any{
			"description": "Validates the WebSocket handshake",
			"schema": map[string]any{
				"type": "string",
			},
		},
	}
}

func pprofPathItem(operationIDs map[string]int) map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())

	return map[string]any{
		"get": map[string]any{
			"summary":     "pprof index",
			"description": "Lists the available Go runtime profiles",
			"tags":        []string{"debug"},
			"operationId": operationID("get", "/debug/pprof", operationIDs),
			"security": []any{
				map[string]any{
					"bearerAuth": []any{},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "pprof index",
					"content": map[string]any{
						"text/html": map[string]any{
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
			},
		},
	}
}

func expvarPathItem(operationIDs map[string]int) map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())

	return map[string]any{
		"get": map[string]any{
			"summary":     "Runtime metrics",
			"description": "Returns expvar metrics as JSON",
			"tags":        []string{"debug"},
			"operationId": operationID("get", "/debug/vars", operationIDs),
			"security": []any{
				map[string]any{
					"bearerAuth": []any{},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Runtime metrics",
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
			},
		},
	}
}

// openAPISpecPathItem returns the OpenAPI path item describing the standalone
// JSON document endpoint (RFC.endpoints.md — "GET /v1/openapi.json"). The
// endpoint is flagged public so SDK generators can fetch the description
// without credentials when Authentik is configured.
//
//	paths["/v1/openapi.json"] = openAPISpecPathItem("/v1/openapi.json", ids)
func openAPISpecPathItem(path string, operationIDs map[string]int) map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())

	return map[string]any{
		"get": map[string]any{
			"summary":     "OpenAPI specification",
			"description": "Returns the generated OpenAPI 3.1 JSON document for this API.",
			"tags":        []string{"system"},
			"operationId": operationID("get", path, operationIDs),
			"security":    []any{},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "OpenAPI 3.1 JSON document",
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
				"500": map[string]any{
					"description": "Failed to render specification",
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
			},
		},
	}
}

// chatCompletionsPathItem returns the OpenAPI path item describing the
// OpenAI-compatible chat completions endpoint (RFC §11). The path documents
// the streaming and non-streaming response shapes, the Gemma 4 calibrated
// sampling defaults, and the OpenAI-compatible error envelope so SDK
// generators can bind to the same surface as the hand-written client.
//
//	paths["/v1/chat/completions"] = chatCompletionsPathItem("/v1/chat/completions", ids)
func chatCompletionsPathItem(path string, operationIDs map[string]int) map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())

	return map[string]any{
		"post": map[string]any{
			"summary":     "Chat completions",
			"description": "OpenAI-compatible chat completion endpoint. Defaults to temperature=1.0, top_p=0.95, top_k=64, max_tokens=2048 (Gemma 4 calibrated). Set stream=true to receive Server-Sent Events matching OpenAI's streaming format.",
			"tags":        []string{"inference"},
			"operationId": operationID("post", path, operationIDs),
			"security":    []any{},
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": chatCompletionsRequestSchema(),
					},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Chat completion response",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": chatCompletionsResponseSchema(),
						},
						"text/event-stream": map[string]any{
							"schema": chatCompletionsStreamSchema(),
						},
					},
					"headers": successHeaders,
				},
				"400": map[string]any{
					"description": "Invalid request",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": chatCompletionsErrorSchema(),
						},
					},
					"headers": errorHeaders,
				},
				"404": map[string]any{
					"description": "Model not found",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": chatCompletionsErrorSchema(),
						},
					},
					"headers": errorHeaders,
				},
				"503": map[string]any{
					"description": "Model loading or unavailable",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": chatCompletionsErrorSchema(),
						},
					},
					"headers": errorHeaders,
				},
				"500": map[string]any{
					"description": "Inference error",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": chatCompletionsErrorSchema(),
						},
					},
					"headers": errorHeaders,
				},
			},
		},
	}
}

// chatCompletionsRequestSchema is the OpenAPI schema for
// ChatCompletionRequest. Gemma 4 calibrated defaults (temperature=1.0,
// top_p=0.95, top_k=64, max_tokens=2048) are documented in the example.
//
//	schema := chatCompletionsRequestSchema()
func chatCompletionsRequestSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"model": map[string]any{
				"type":        "string",
				"description": "Model name (lemer, lemma, lemmy, lemrd, or any identifier resolvable via ~/.core/models.yaml)",
			},
			"messages": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"role":    map[string]any{"type": "string", "enum": []string{"system", "user", "assistant"}},
						"content": map[string]any{"type": "string"},
					},
					"required": []string{"role", "content"},
				},
			},
			"temperature": map[string]any{"type": "number", "description": "Sampling temperature (default 1.0 for Gemma 4)"},
			"top_p":       map[string]any{"type": "number", "description": "Nucleus sampling (default 0.95)"},
			"top_k":       map[string]any{"type": "integer", "description": "Top-K sampling (default 64)"},
			"max_tokens":  map[string]any{"type": "integer", "description": "Output token cap (default 2048)"},
			"stream":      map[string]any{"type": "boolean", "description": "Enable SSE streaming"},
			"stop":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"user":        map[string]any{"type": "string", "description": "Opaque end-user identifier"},
		},
		"required": []string{"model", "messages"},
	}
}

// chatCompletionsResponseSchema is the OpenAPI schema for a non-streaming
// ChatCompletionResponse. See RFC §11.3.
//
//	schema := chatCompletionsResponseSchema()
func chatCompletionsResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":      map[string]any{"type": "string"},
			"object":  map[string]any{"type": "string"},
			"created": map[string]any{"type": "integer"},
			"model":   map[string]any{"type": "string"},
			"choices": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"index": map[string]any{"type": "integer"},
						"message": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"role":    map[string]any{"type": "string"},
								"content": map[string]any{"type": "string"},
							},
						},
						"finish_reason": map[string]any{"type": "string", "enum": []string{"stop", "length", "error"}},
					},
				},
			},
			"usage": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"prompt_tokens":     map[string]any{"type": "integer"},
					"completion_tokens": map[string]any{"type": "integer"},
					"total_tokens":      map[string]any{"type": "integer"},
				},
			},
			"thought": map[string]any{"type": "string", "description": "Thinking channel content when the model emits <|channel>thought tokens"},
		},
	}
}

// chatCompletionsStreamSchema documents the text/event-stream chunk shape for
// Server-Sent Events responses. See RFC §11.4.
//
//	schema := chatCompletionsStreamSchema()
func chatCompletionsStreamSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "data: <json> events terminated by data: [DONE] per OpenAI's SSE format",
	}
}

// chatCompletionsErrorSchema is the OpenAI-compatible error envelope emitted
// by the chat completions endpoint. See RFC §11.7.
//
//	schema := chatCompletionsErrorSchema()
func chatCompletionsErrorSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"error": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message": map[string]any{"type": "string"},
					"type":    map[string]any{"type": "string"},
					"param":   map[string]any{"type": "string"},
					"code":    map[string]any{"type": "string"},
				},
				"required": []string{"message", "type", "code"},
			},
		},
		"required": []string{"error"},
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

func graphqlQueryParameters() []map[string]any {
	return []map[string]any{
		{
			"name":        "query",
			"in":          "query",
			"required":    true,
			"description": "GraphQL query or mutation document",
			"schema": map[string]any{
				"type": "string",
			},
		},
		{
			"name":        "variables",
			"in":          "query",
			"required":    false,
			"description": "JSON-encoded GraphQL variables",
			"schema": map[string]any{
				"type": "string",
			},
		},
		{
			"name":        "operationName",
			"in":          "query",
			"required":    false,
			"description": "Operation name to execute",
			"schema": map[string]any{
				"type": "string",
			},
		},
	}
}

func graphqlResponses(cacheEnabled bool) map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	if cacheEnabled {
		successHeaders = mergeHeaders(successHeaders, cacheSuccessHeaders())
	}
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

func graphqlPlaygroundResponses() map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())

	return map[string]any{
		"200": map[string]any{
			"description": "GraphQL playground",
			"content": map[string]any{
				"text/html": map[string]any{
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
		if g == nil || isNilRouteGroup(g) {
			continue
		}
		if isHiddenRouteGroup(g) {
			continue
		}
		out = append(out, preparedRouteGroup{
			name:     g.Name(),
			basePath: g.BasePath(),
			descs:    collectRouteDescriptions(g),
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
		resolved := resolveRouteDescription(rd)
		if resolved.Hidden {
			continue
		}
		descs = append(descs, resolved)
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

func resolveRouteDescription(rd RouteDescription) RouteDescription {
	resolved := cloneRouteDescription(rd)
	handler := routeDescribable(resolved.Handler)
	if handler == nil {
		return resolved
	}

	handlerDesc := cloneRouteDescription(handler.Describe())

	if core.Trim(resolved.Method) == "" {
		resolved.Method = handlerDesc.Method
	}
	if core.Trim(resolved.Path) == "" {
		resolved.Path = handlerDesc.Path
	}
	if core.Trim(resolved.Summary) == "" {
		resolved.Summary = firstNonEmpty(handler.Summary(), handlerDesc.Summary)
	}
	if core.Trim(resolved.Description) == "" {
		resolved.Description = firstNonEmpty(handler.Description(), handlerDesc.Description)
	}
	if tags := cleanTags(resolved.Tags); len(tags) > 0 {
		resolved.Tags = tags
	} else if tags := cleanTags(handler.Tags()); len(tags) > 0 {
		resolved.Tags = tags
	} else {
		resolved.Tags = cleanTags(handlerDesc.Tags)
	}
	if core.Trim(resolved.CacheControl) == "" {
		resolved.CacheControl = handlerDesc.CacheControl
	}
	if !resolved.Hidden && handlerDesc.Hidden {
		resolved.Hidden = true
	}
	if !resolved.Deprecated && handlerDesc.Deprecated {
		resolved.Deprecated = true
	}
	if core.Trim(resolved.SunsetDate) == "" {
		resolved.SunsetDate = handlerDesc.SunsetDate
	}
	if core.Trim(resolved.ReplacementURL) == "" {
		resolved.ReplacementURL = handlerDesc.ReplacementURL
	}
	if core.Trim(resolved.Replacement) == "" {
		resolved.Replacement = handlerDesc.Replacement
	}
	if core.Trim(resolved.NoticeURL) == "" {
		resolved.NoticeURL = handlerDesc.NoticeURL
	}
	if resolved.StatusCode == 0 {
		resolved.StatusCode = handlerDesc.StatusCode
	}
	if resolved.Security == nil && handlerDesc.Security != nil {
		resolved.Security = cloneSecurityRequirements(handlerDesc.Security)
	}
	if resolved.Parameters == nil && handlerDesc.Parameters != nil {
		resolved.Parameters = cloneParameterDescriptions(handlerDesc.Parameters)
	}
	if resolved.RequestBody == nil && handlerDesc.RequestBody != nil {
		resolved.RequestBody = cloneOpenAPIObject(handlerDesc.RequestBody)
	}
	if resolved.RequestExample == nil && handlerDesc.RequestExample != nil {
		resolved.RequestExample = cloneOpenAPIValue(handlerDesc.RequestExample)
	}
	if resolved.Response == nil && handlerDesc.Response != nil {
		resolved.Response = cloneOpenAPIObject(handlerDesc.Response)
	}
	if resolved.ResponseExample == nil && handlerDesc.ResponseExample != nil {
		resolved.ResponseExample = cloneOpenAPIValue(handlerDesc.ResponseExample)
	}
	if resolved.ResponseHeaders == nil && handlerDesc.ResponseHeaders != nil {
		resolved.ResponseHeaders = cloneStringMap(handlerDesc.ResponseHeaders)
	}

	return resolved
}

func routeDescribable(handler any) Describable {
	if isNilValue(handler) {
		return nil
	}

	d, ok := handler.(Describable)
	if !ok || isNilValue(d) {
		return nil
	}

	return d
}

func routeRenderable(handler any) Renderable {
	if isNilValue(handler) {
		return nil
	}

	r, ok := handler.(Renderable)
	if !ok || isNilValue(r) {
		return nil
	}

	return r
}

func resolvedOperationID(rd RouteDescription, method, path string, operationIDs map[string]int) string {
	if handler := routeDescribable(rd.Handler); handler != nil {
		if operationID := registerOperationID(handler.OperationID(), operationIDs); operationID != "" {
			return operationID
		}
	}

	return operationID(method, path, operationIDs)
}

func registerOperationID(id string, operationIDs map[string]int) string {
	id = core.Trim(id)
	if id == "" {
		return ""
	}
	if operationIDs == nil {
		return id
	}

	count := operationIDs[id]
	operationIDs[id] = count + 1
	if count == 0 {
		return id
	}

	return id + "_" + core.Itoa(count+1)
}

func resolvedRenderHints(rd RouteDescription) map[string]any {
	handler := routeRenderable(rd.Handler)
	if handler == nil {
		return nil
	}

	return renderHintsExtension(handler.Render())
}

func renderHintsExtension(hints RenderHints) map[string]any {
	extension := map[string]any{}

	if kind := core.Trim(hints.Kind); kind != "" {
		extension["kind"] = kind
	}
	if fields := cloneFieldHints(hints.Fields); len(fields) > 0 {
		extension["fields"] = fields
	}
	if actions := cloneActionHints(hints.Actions); len(actions) > 0 {
		extension["actions"] = actions
	}
	if len(extension) == 0 {
		return nil
	}

	return extension
}

func cloneFieldHints(fields []FieldHint) []FieldHint {
	if len(fields) == 0 {
		return nil
	}

	out := make([]FieldHint, len(fields))
	for i, field := range fields {
		out[i] = field
		out[i].Validation = cloneOpenAPIObject(field.Validation)
	}

	return out
}

func cloneActionHints(actions []ActionHint) []ActionHint {
	if len(actions) == 0 {
		return nil
	}

	return slices.Clone(actions)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = core.Trim(value); value != "" {
			return value
		}
	}

	return ""
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
		end := indexByteFrom(path, close, i+1)
		if end < 0 {
			continue
		}
		name := path[i+1 : end]
		if name == "" || containsPathParameterDelimiter(name) || seen[name] {
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
		i = end
	}

	return params
}

func indexByteFrom(value string, target byte, start int) int {
	for i := start; i < len(value); i++ {
		if value[i] == target {
			return i
		}
	}
	return -1
}

func containsPathParameterDelimiter(name string) bool {
	return core.Contains(name, "/") ||
		core.Contains(name, "{") ||
		core.Contains(name, "}")
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
func resolvedOperationTags(groupName string, rd RouteDescription) []string {
	if tags := cleanTags(rd.Tags); len(tags) > 0 {
		return tags
	}

	if name := core.Trim(groupName); name != "" {
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
		tag = core.Trim(tag)
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

// effectiveGraphQLPath returns the configured GraphQL path or the default
// GraphQL path when GraphQL is enabled without an explicit path. An explicit
// path also surfaces on its own so spec generation reflects configuration
// authored ahead of runtime activation. Returns an empty string only when no
// configuration is present.
//
//	sb.effectiveGraphQLPath()  // "/graphql" when enabled or configured
func (sb *SpecBuilder) effectiveGraphQLPath() string {
	graphqlPath := core.Trim(sb.GraphQLPath)
	if graphqlPath != "" {
		return graphqlPath
	}
	if sb.GraphQLEnabled || sb.GraphQLPlayground {
		return defaultGraphQLPath
	}
	return ""
}

// effectiveGraphQLPlaygroundPath returns the configured playground path when
// GraphQL playground is enabled.
func (sb *SpecBuilder) effectiveGraphQLPlaygroundPath() string {
	if !sb.GraphQLPlayground {
		return ""
	}

	path := core.Trim(sb.GraphQLPlaygroundPath)
	if path != "" {
		return path
	}

	base := sb.effectiveGraphQLPath()
	if base == "" {
		base = defaultGraphQLPath
	}

	return base + "/playground"
}

// effectiveSwaggerPath returns the configured Swagger UI path or the default
// path when Swagger is enabled without an explicit override. An explicit path
// also surfaces on its own so spec generation reflects configuration authored
// ahead of runtime activation. Returns an empty string only when no
// configuration is present.
//
//	sb.effectiveSwaggerPath()  // "/swagger" when enabled or configured
func (sb *SpecBuilder) effectiveSwaggerPath() string {
	swaggerPath := core.Trim(sb.SwaggerPath)
	if swaggerPath != "" {
		return swaggerPath
	}
	if sb.SwaggerEnabled {
		return defaultSwaggerPath
	}
	return ""
}

// effectiveWSPath returns the configured WebSocket path or the default path
// when WebSockets are enabled without an explicit override. An explicit path
// also surfaces on its own so spec generation reflects configuration authored
// ahead of runtime activation. Returns an empty string only when no
// configuration is present.
//
//	sb.effectiveWSPath()  // "/ws" when enabled or configured
func (sb *SpecBuilder) effectiveWSPath() string {
	wsPath := core.Trim(sb.WSPath)
	if wsPath != "" {
		return wsPath
	}
	if sb.WSEnabled {
		return defaultWSPath
	}
	return ""
}

// effectiveSSEPath returns the configured SSE path or the default path when
// SSE is enabled without an explicit override. An explicit path also surfaces
// on its own so spec generation reflects configuration authored ahead of
// runtime activation. Returns an empty string only when no configuration is
// present.
//
//	sb.effectiveSSEPath()  // "/events" when enabled or configured
func (sb *SpecBuilder) effectiveSSEPath() string {
	ssePath := core.Trim(sb.SSEPath)
	if ssePath != "" {
		return ssePath
	}
	if sb.SSEEnabled {
		return defaultSSEPath
	}
	return ""
}

// effectiveChatCompletionsPath returns the configured chat completions path or
// the RFC §11.1 default when chat completions is enabled without an explicit
// override. An explicit path also surfaces on its own so spec generation
// reflects configuration authored ahead of runtime activation.
//
//	sb.effectiveChatCompletionsPath()  // "/v1/chat/completions" when enabled
func (sb *SpecBuilder) effectiveChatCompletionsPath() string {
	path := core.Trim(sb.ChatCompletionsPath)
	if path != "" {
		return path
	}
	if sb.ChatCompletionsEnabled {
		return defaultChatCompletionsPath
	}
	return ""
}

// effectiveOpenAPISpecPath returns the configured standalone OpenAPI JSON
// endpoint path or the RFC default "/v1/openapi.json" when enabled without an
// explicit override. An explicit path also surfaces on its own so the spec
// reflects configuration authored ahead of runtime activation.
//
//	sb.effectiveOpenAPISpecPath()  // "/v1/openapi.json" when enabled
func (sb *SpecBuilder) effectiveOpenAPISpecPath() string {
	path := core.Trim(sb.OpenAPISpecPath)
	if path != "" {
		return path
	}
	if sb.OpenAPISpecEnabled {
		return defaultOpenAPISpecPath
	}
	return ""
}

// effectiveCacheTTL returns a normalised cache TTL when it parses to a
// positive duration.
func (sb *SpecBuilder) effectiveCacheTTL() string {
	ttl := core.Trim(sb.CacheTTL)
	if ttl == "" {
		return ""
	}

	d, err := time.ParseDuration(ttl)
	if err != nil || d <= 0 {
		return ""
	}

	return ttl
}

// effectiveAuthentikPublicPaths returns the public paths that Authentik skips
// in practice, including the always-public health endpoint and both the
// default and any configured Swagger UI paths. The runtime middleware also
// always skips "/swagger" unconditionally (see authentikMiddleware), so the
// spec mirrors that behaviour even when a custom swagger path is mounted.
//
//	paths := sb.effectiveAuthentikPublicPaths()  // [/health /swagger ...]
func (sb *SpecBuilder) effectiveAuthentikPublicPaths() []string {
	if !sb.hasAuthentikMetadata() {
		return nil
	}

	paths := []string{"/health", defaultSwaggerPath}
	if swaggerPath := sb.effectiveSwaggerPath(); swaggerPath != "" {
		paths = append(paths, swaggerPath)
	}
	paths = append(paths, sb.AuthentikPublicPaths...)
	return normalisePublicPaths(paths)
}

// snapshot returns a trimmed copy of the builder so Build operates on stable
// input even when callers reuse or mutate their original configuration.
func (sb *SpecBuilder) snapshot() *SpecBuilder {
	if sb == nil {
		return &SpecBuilder{}
	}

	out := *sb
	out.Title = core.Trim(out.Title)
	out.Summary = core.Trim(out.Summary)
	out.Description = core.Trim(out.Description)
	out.Version = core.Trim(out.Version)
	out.SwaggerPath = core.Trim(out.SwaggerPath)
	out.GraphQLPath = core.Trim(out.GraphQLPath)
	out.GraphQLPlaygroundPath = core.Trim(out.GraphQLPlaygroundPath)
	out.WSPath = core.Trim(out.WSPath)
	out.SSEPath = core.Trim(out.SSEPath)
	out.TermsOfService = core.Trim(out.TermsOfService)
	out.ContactName = core.Trim(out.ContactName)
	out.ContactURL = core.Trim(out.ContactURL)
	out.ContactEmail = core.Trim(out.ContactEmail)
	out.LicenseName = core.Trim(out.LicenseName)
	out.LicenseURL = core.Trim(out.LicenseURL)
	out.ExternalDocsDescription = core.Trim(out.ExternalDocsDescription)
	out.ExternalDocsURL = core.Trim(out.ExternalDocsURL)
	out.CacheTTL = core.Trim(out.CacheTTL)
	out.I18nDefaultLocale = core.Trim(out.I18nDefaultLocale)
	out.Servers = slices.Clone(sb.Servers)
	out.I18nSupportedLocales = slices.Clone(sb.I18nSupportedLocales)
	out.AuthentikPublicPaths = normalisePublicPaths(sb.AuthentikPublicPaths)
	out.SecuritySchemes = cloneSecuritySchemes(sb.SecuritySchemes)

	return &out
}

// isPublicOperationPath reports whether an OpenAPI path should be documented
// as public because Authentik bypasses it in the running engine.
func (sb *SpecBuilder) isPublicOperationPath(path string) bool {
	return isPublicPathForList(path, sb.effectiveAuthentikPublicPaths())
}

// hasAuthentikMetadata reports whether the spec carries any Authentik-related
// configuration worth surfacing.
func (sb *SpecBuilder) hasAuthentikMetadata() bool {
	if sb == nil {
		return false
	}

	return core.Trim(sb.AuthentikIssuer) != "" ||
		core.Trim(sb.AuthentikClientID) != "" ||
		sb.AuthentikTrustedProxy ||
		len(sb.AuthentikPublicPaths) > 0
}

// makePathItemPublic strips auth-specific responses and marks every operation
// within the path item as public.
func makePathItemPublic(pathItem map[string]any) {
	for _, rawOperation := range pathItem {
		operation, ok := rawOperation.(map[string]any)
		if !ok {
			continue
		}

		operation["security"] = []any{}
		responses, ok := operation["responses"].(map[string]any)
		if !ok {
			continue
		}
		delete(responses, "401")
		delete(responses, "403")
	}
}

// isPublicPathForList reports whether path should be documented as public
// when compared against a precomputed list of public paths.
func isPublicPathForList(path string, publicPaths []string) bool {
	for _, publicPath := range publicPaths {
		if isPublicPath(path, publicPath) {
			return true
		}
	}
	return false
}

// documentedResponseHeaders converts route-specific response header metadata
// into OpenAPI header objects.
func documentedResponseHeaders(headers map[string]string) map[string]any {
	if len(headers) == 0 {
		return nil
	}

	out := make(map[string]any, len(headers))
	for name, description := range headers {
		name = core.Trim(name)
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
	b := core.NewBuilder()
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

	out := trimOperationIDUnderscores(b.String())
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
	return out + "_" + core.Itoa(count+1)
}

func trimOperationIDUnderscores(value string) string {
	for core.HasPrefix(value, "_") {
		value = core.TrimPrefix(value, "_")
	}
	for core.HasSuffix(value, "_") {
		value = core.TrimSuffix(value, "_")
	}
	return value
}
