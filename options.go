// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"compress/gzip"
	"log/slog"
	"net/http"
	"slices"
	"time"

	core "dappco.re/go/core"

	"github.com/99designs/gqlgen/graphql"
	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/authz"
	"github.com/gin-contrib/cors"
	gingzip "github.com/gin-contrib/gzip"
	"github.com/gin-contrib/httpsign"
	"github.com/gin-contrib/location/v2"
	"github.com/gin-contrib/secure"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	ginslog "github.com/gin-contrib/slog"
	"github.com/gin-contrib/static"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

// Option configures an Engine during construction.
//
// Example:
//
//	engine, _ := api.New(api.WithAddr(":8080"))
type Option func(*Engine)

// WithAddr sets the listen address for the server.
//
// Example:
//
//	api.New(api.WithAddr(":8443"))
func WithAddr(addr string) Option {
	return func(e *Engine) {
		e.addr = addr
	}
}

// WithHTTP3 enables HTTP/3 advertisement and configures the QUIC listen
// address used by ServeH3. Pass an empty address to reuse the main HTTP
// address at serve time.
//
// HTTP/3 requires TLS. ServeH3 returns ErrHTTP3TLSRequired when called without
// a TLS configuration.
//
// Example:
//
//	api.New(api.WithHTTP3(":8443"))
func WithHTTP3(addr string) Option {
	return func(e *Engine) {
		e.http3Enabled = true
		e.http3Addr = core.Trim(addr)
	}
}

// WithBearerAuth adds bearer token authentication middleware.
// Requests to /health and the Swagger UI path are exempt.
//
// Example:
//
//	api.New(api.WithBearerAuth("secret"))
func WithBearerAuth(token string) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, bearerAuthMiddleware(token, func() []string {
			skip := []string{"/health"}
			if swaggerPath := resolveSwaggerPath(e.swaggerPath); swaggerPath != "" {
				skip = append(skip, swaggerPath)
			}
			if openAPISpecPath := resolveOpenAPISpecPath(e.openAPISpecPath); openAPISpecPath != "" {
				skip = append(skip, openAPISpecPath)
			}
			return skip
		}))
	}
}

// WithRequestID adds middleware that assigns an X-Request-ID to every response.
// Client-provided IDs are preserved; otherwise a random hex ID is generated.
//
// Example:
//
//	api.New(api.WithRequestID())
func WithRequestID() Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, requestIDMiddleware())
	}
}

// WithResponseMeta attaches request metadata to JSON envelope responses.
// It preserves any existing pagination metadata and merges in request_id
// and duration when available from the request context. Combine it with
// WithRequestID() to populate both fields automatically.
//
// Example:
//
//	api.New(api.WithRequestID(), api.WithResponseMeta())
func WithResponseMeta() Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, responseMetaMiddleware())
	}
}

// WithCORS configures Cross-Origin Resource Sharing via gin-contrib/cors.
// Pass "*" to allow all origins, or supply specific origin URLs.
// Standard methods (GET, POST, PUT, PATCH, DELETE, OPTIONS) and common
// headers (Authorization, Content-Type, X-Request-ID) are permitted.
//
// Example:
//
//	api.New(api.WithCORS("*"))
func WithCORS(allowOrigins ...string) Option {
	return func(e *Engine) {
		cfg := cors.Config{
			AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Authorization", "Content-Type", "X-Request-ID"},
			MaxAge:       12 * time.Hour,
		}

		if slices.Contains(allowOrigins, "*") {
			cfg.AllowAllOrigins = true
		}
		if !cfg.AllowAllOrigins {
			cfg.AllowOrigins = allowOrigins
		}

		e.middlewares = append(e.middlewares, cors.New(cfg))
	}
}

// WithMiddleware appends arbitrary Gin middleware to the engine.
//
// Example:
//
//	api.New(api.WithMiddleware(loggingMiddleware))
func WithMiddleware(mw ...gin.HandlerFunc) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, mw...)
	}
}

// WithStatic serves static files from the given root directory at urlPrefix.
// Directory listing is disabled; only individual files are served.
// Internally this uses gin-contrib/static as Gin middleware.
//
// Example:
//
//	api.New(api.WithStatic("/assets", "./public"))
func WithStatic(urlPrefix, root string) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, static.Serve(urlPrefix, static.LocalFile(root, false)))
	}
}

// WithWSHandler registers a WebSocket handler at GET /ws.
// Use WithWSPath to customise the route before mounting the handler.
// Typically this wraps a go-ws Hub.Handler().
//
// Example:
//
//	api.New(api.WithWSHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
func WithWSHandler(h http.Handler) Option {
	return func(e *Engine) {
		e.wsHandler = h
	}
}

// WithWebSocket registers a Gin-native WebSocket handler at GET /ws.
//
// This is the gin-handler form of WithWSHandler. The handler receives the
// request via *gin.Context and is responsible for performing the upgrade
// (typically with gorilla/websocket) and managing the message loop.
// Use WithWSPath to customise the route before mounting the handler.
//
// Example:
//
//	api.New(api.WithWebSocket(func(c *gin.Context) {
//	    // upgrade and handle messages
//	}))
func WithWebSocket(h gin.HandlerFunc) Option {
	return func(e *Engine) {
		if h == nil {
			return
		}
		e.wsGinHandler = h
	}
}

// WithWSPath sets a custom URL path for the WebSocket endpoint.
// The default path is "/ws".
//
// Example:
//
//	api.New(api.WithWSPath("/socket"))
func WithWSPath(path string) Option {
	return func(e *Engine) {
		e.wsPath = normaliseWSPath(path)
	}
}

// WithAuthentik adds Authentik forward-auth middleware that extracts user
// identity from X-authentik-* headers set by a trusted reverse proxy.
// The middleware is permissive: unauthenticated requests are allowed through.
//
// Example:
//
//	api.New(api.WithAuthentik(api.AuthentikConfig{TrustedProxy: true}))
func WithAuthentik(cfg AuthentikConfig) Option {
	return func(e *Engine) {
		snapshot := cloneAuthentikConfig(cfg)
		e.authentikConfig = snapshot
		e.middlewares = append(e.middlewares, authentikMiddleware(snapshot, func() []string {
			return []string{
				resolveSwaggerPath(e.swaggerPath),
				resolveOpenAPISpecPath(e.openAPISpecPath),
			}
		}))
	}
}

// WithSunset adds deprecation headers to every response.
// The middleware appends Deprecation, optional Sunset, optional Link, and
// X-API-Warn headers without clobbering any existing header values. Use it to
// deprecate an entire route group or API version.
//
// Example:
//
//	api.New(api.WithSunset("2026-12-31", "https://api.example.com/v2"))
func WithSunset(sunsetDate, replacement string) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, ApiSunset(sunsetDate, replacement))
	}
}

// WithSwagger enables the Swagger UI at /swagger/ by default.
// The title, description, and version populate the OpenAPI info block.
// Use WithSwaggerSummary() to set the optional info.summary field.
//
// Example:
//
//	api.New(api.WithSwagger("Service", "Public API", "1.0.0"))
func WithSwagger(title, description, version string) Option {
	return func(e *Engine) {
		e.swaggerTitle = core.Trim(title)
		e.swaggerDesc = core.Trim(description)
		e.swaggerVersion = core.Trim(version)
		e.swaggerEnabled = true
	}
}

// WithSwaggerSummary adds the OpenAPI info.summary field to generated specs.
//
// Example:
//
//	api.WithSwaggerSummary("Service overview")
func WithSwaggerSummary(summary string) Option {
	return func(e *Engine) {
		if summary = core.Trim(summary); summary != "" {
			e.swaggerSummary = summary
		}
	}
}

// WithSwaggerPath sets a custom URL path for the Swagger UI.
// The default path is "/swagger".
//
// Example:
//
//	api.New(api.WithSwaggerPath("/docs"))
func WithSwaggerPath(path string) Option {
	return func(e *Engine) {
		e.swaggerPath = normaliseSwaggerPath(path)
	}
}

// WithSwaggerTermsOfService adds the terms of service URL to the generated Swagger spec.
// Empty strings are ignored.
//
// Example:
//
//	api.WithSwaggerTermsOfService("https://example.com/terms")
func WithSwaggerTermsOfService(url string) Option {
	return func(e *Engine) {
		if url = core.Trim(url); url != "" {
			e.swaggerTermsOfService = url
		}
	}
}

// WithSwaggerContact adds contact metadata to the generated Swagger spec.
// Empty fields are ignored. Multiple calls replace the previous contact data.
//
// Example:
//
//	api.WithSwaggerContact("API Support", "https://example.com/support", "support@example.com")
func WithSwaggerContact(name, url, email string) Option {
	return func(e *Engine) {
		if name = core.Trim(name); name != "" {
			e.swaggerContactName = name
		}
		if url = core.Trim(url); url != "" {
			e.swaggerContactURL = url
		}
		if email = core.Trim(email); email != "" {
			e.swaggerContactEmail = email
		}
	}
}

// WithSwaggerServers adds OpenAPI server metadata to the generated Swagger spec.
// Empty strings are ignored. Multiple calls append and normalise the combined
// server list so callers can compose metadata across options.
//
// Example:
//
//	api.WithSwaggerServers("https://api.example.com", "https://docs.example.com")
func WithSwaggerServers(servers ...string) Option {
	return func(e *Engine) {
		e.swaggerServers = normaliseServers(append(e.swaggerServers, servers...))
	}
}

// WithSwaggerLicense adds licence metadata to the generated Swagger spec.
// Pass both a name and URL to populate the OpenAPI info block consistently.
//
// Example:
//
//	api.WithSwaggerLicense("EUPL-1.2", "https://eupl.eu/1.2/en/")
func WithSwaggerLicense(name, url string) Option {
	return func(e *Engine) {
		if name = core.Trim(name); name != "" {
			e.swaggerLicenseName = name
		}
		if url = core.Trim(url); url != "" {
			e.swaggerLicenseURL = url
		}
	}
}

// WithSwaggerSecuritySchemes merges custom OpenAPI security schemes into the
// generated Swagger spec. Existing schemes are preserved unless the new map
// defines the same key, in which case the later definition wins.
//
// Example:
//
//	api.WithSwaggerSecuritySchemes(map[string]any{
//		"apiKeyAuth": map[string]any{
//			"type": "apiKey",
//			"in":   "header",
//			"name": "X-API-Key",
//		},
//	})
func WithSwaggerSecuritySchemes(schemes map[string]any) Option {
	return func(e *Engine) {
		if len(schemes) == 0 {
			return
		}
		if e.swaggerSecuritySchemes == nil {
			e.swaggerSecuritySchemes = make(map[string]any, len(schemes))
		}
		for name, scheme := range schemes {
			name = core.Trim(name)
			if name == "" || scheme == nil {
				continue
			}
			e.swaggerSecuritySchemes[name] = cloneOpenAPIValue(scheme)
		}
	}
}

// WithSwaggerExternalDocs adds top-level external documentation metadata to
// the generated Swagger spec.
// Empty URLs are ignored; the description is optional.
//
// Example:
//
//	api.WithSwaggerExternalDocs("Developer guide", "https://example.com/docs")
func WithSwaggerExternalDocs(description, url string) Option {
	return func(e *Engine) {
		if description = core.Trim(description); description != "" {
			e.swaggerExternalDocsDescription = description
		}
		if url = core.Trim(url); url != "" {
			e.swaggerExternalDocsURL = url
		}
	}
}

// WithPprof enables Go runtime profiling endpoints at /debug/pprof/.
// The standard pprof handlers (index, cmdline, profile, symbol, trace,
// allocs, block, goroutine, heap, mutex, threadcreate) are registered
// via gin-contrib/pprof.
//
// WARNING: pprof exposes sensitive runtime data and should only be
// enabled in development or behind authentication in production.
//
// Example:
//
//	api.New(api.WithPprof())
func WithPprof() Option {
	return func(e *Engine) {
		e.pprofEnabled = true
	}
}

// WithExpvar enables the Go runtime metrics endpoint at /debug/vars.
// The endpoint serves JSON containing memstats, cmdline, and any
// custom expvar variables registered by the application. Powered by
// gin-contrib/expvar wrapping Go's standard expvar.Handler().
//
// WARNING: expvar exposes runtime internals (memory allocation,
// goroutine counts, command-line arguments) and should only be
// enabled in development or behind authentication in production.
//
// Example:
//
//	api.New(api.WithExpvar())
func WithExpvar() Option {
	return func(e *Engine) {
		e.expvarEnabled = true
	}
}

// WithSecure adds security headers middleware via gin-contrib/secure.
// Default policy sets HSTS (1 year, includeSubDomains), X-Frame-Options DENY,
// X-Content-Type-Options nosniff, and Referrer-Policy strict-origin-when-cross-origin.
// SSL redirect is not enabled so the middleware works behind a reverse proxy
// that terminates TLS.
//
// Example:
//
//	api.New(api.WithSecure())
func WithSecure() Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, secure.New(secure.Config{
			STSSeconds:           31536000,
			STSIncludeSubdomains: true,
			FrameDeny:            true,
			ContentTypeNosniff:   true,
			ReferrerPolicy:       "strict-origin-when-cross-origin",
			IsDevelopment:        false,
		}))
	}
}

// WithGzip adds gzip response compression middleware via gin-contrib/gzip.
// An optional compression level may be supplied (e.g. gzip.BestSpeed,
// gzip.BestCompression). If omitted, gzip.DefaultCompression is used.
//
// Example:
//
//	api.New(api.WithGzip())
func WithGzip(level ...int) Option {
	return func(e *Engine) {
		l := gzip.DefaultCompression
		if len(level) > 0 {
			l = level[0]
		}
		e.middlewares = append(e.middlewares, gingzip.Gzip(l))
	}
}

// WithBrotli adds Brotli response compression middleware using andybalholm/brotli.
// An optional compression level may be supplied (e.g. BrotliBestSpeed,
// BrotliBestCompression). If omitted, BrotliDefaultCompression is used.
//
// Example:
//
//	api.New(api.WithBrotli())
func WithBrotli(level ...int) Option {
	return func(e *Engine) {
		l := BrotliDefaultCompression
		if len(level) > 0 {
			l = level[0]
		}
		e.middlewares = append(e.middlewares, newBrotliHandler(l).Handle)
	}
}

// WithSlog adds structured request logging middleware via gin-contrib/slog.
// Each request is logged with method, path, status code, latency, and client IP.
// If logger is nil, slog.Default() is used.
//
// Example:
//
//	api.New(api.WithSlog(nil))
func WithSlog(logger *slog.Logger) Option {
	return func(e *Engine) {
		if logger == nil {
			logger = slog.Default()
		}
		e.middlewares = append(e.middlewares, ginslog.SetLogger(
			ginslog.WithLogger(func(_ *gin.Context, l *slog.Logger) *slog.Logger {
				return logger
			}),
		))
	}
}

// WithTimeout adds per-request timeout middleware via gin-contrib/timeout.
// If a handler exceeds the given duration, the request is aborted with a
// 504 Gateway Timeout carrying the standard error envelope:
//
//	{"success":false,"error":{"code":"timeout","message":"Request timed out"}}
//
// A zero or negative duration effectively disables the timeout (the handler
// runs without a deadline) — this is safe and will not panic.
//
// Example:
//
//	api.New(api.WithTimeout(5 * time.Second))
func WithTimeout(d time.Duration) Option {
	return func(e *Engine) {
		if d <= 0 {
			return
		}
		e.middlewares = append(e.middlewares, timeout.New(
			timeout.WithTimeout(d),
			timeout.WithResponse(timeoutResponse),
		))
	}
}

// timeoutResponse writes a 504 Gateway Timeout with the standard error envelope.
func timeoutResponse(c *gin.Context) {
	c.JSON(http.StatusGatewayTimeout, Fail("timeout", "Request timed out"))
}

// cacheDefaultMaxEntries is the entry cap applied by WithCache when the caller
// does not supply explicit limits. Prevents unbounded growth when WithCache is
// called with only a TTL argument.
const cacheDefaultMaxEntries = 1_000

// WithCache adds in-memory response caching middleware for GET requests.
// Successful (2xx) GET responses are cached for the given TTL and served
// with an X-Cache: HIT header on subsequent requests. Non-GET methods
// and error responses pass through uncached.
//
// Optional integer limits enable LRU eviction:
//   - maxEntries limits the number of cached responses
//   - maxBytes limits the approximate total cached payload size
//
// At least one limit must be positive; when called with only a TTL the entry
// cap defaults to cacheDefaultMaxEntries (1 000) to prevent unbounded growth.
// A non-positive TTL disables the middleware entirely.
//
// Example:
//
//	engine, _ := api.New(api.WithCache(5*time.Minute, 100, 10<<20))
func WithCache(ttl time.Duration, maxEntries ...int) Option {
	entryLimit := cacheDefaultMaxEntries
	byteLimit := 0
	if len(maxEntries) > 0 {
		entryLimit = maxEntries[0]
	}
	if len(maxEntries) > 1 {
		byteLimit = maxEntries[1]
	}
	return WithCacheLimits(ttl, entryLimit, byteLimit)
}

// WithCacheLimits adds in-memory response caching middleware for GET requests
// with explicit entry and payload-size bounds.
//
// This is the clearer form of WithCache when call sites want to make the
// eviction policy self-documenting.
//
// Example:
//
//	engine, _ := api.New(api.WithCacheLimits(5*time.Minute, 100, 10<<20))
func WithCacheLimits(ttl time.Duration, maxEntries, maxBytes int) Option {
	return func(e *Engine) {
		if ttl <= 0 {
			return
		}
		// newCacheStore returns nil when both limits are non-positive (unbounded),
		// which is a footgun; skip middleware registration in that case.
		store := newCacheStore(maxEntries, maxBytes)
		if store == nil {
			return
		}
		e.cacheTTL = ttl
		e.cacheMaxEntries = maxEntries
		e.cacheMaxBytes = maxBytes
		e.middlewares = append(e.middlewares, cacheMiddleware(store, ttl))
	}
}

// WithRateLimit adds token-bucket rate limiting middleware.
// Requests are bucketed by API key or bearer token when present, and
// otherwise by client IP. Passing requests are annotated with
// X-RateLimit-Limit, X-RateLimit-Remaining, and X-RateLimit-Reset headers.
// Requests exceeding the configured limit are rejected with 429 Too Many
// Requests, Retry-After, and the standard Fail() error envelope.
// A zero or negative limit disables rate limiting.
//
// Example:
//
//	engine, _ := api.New(api.WithRateLimit(100))
func WithRateLimit(limit int) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, rateLimitMiddleware(limit))
	}
}

// WithSessions adds server-side session management middleware via
// gin-contrib/sessions using a cookie-based store. The name parameter
// sets the session cookie name (e.g. "session") and secret is the key
// used for cookie signing and encryption.
//
// Example:
//
//	api.New(api.WithSessions("session", []byte("secret")))
func WithSessions(name string, secret []byte) Option {
	return func(e *Engine) {
		store := cookie.NewStore(secret)
		e.middlewares = append(e.middlewares, sessions.Sessions(name, store))
	}
}

// WithAuthz adds Casbin policy-based authorisation middleware via
// gin-contrib/authz. The caller provides a pre-configured Casbin enforcer
// holding the desired model and policy rules. The middleware extracts the
// subject from HTTP Basic Authentication, evaluates it against the request
// method and path, and returns 403 Forbidden when the policy denies access.
//
// Example:
//
//	api.New(api.WithAuthz(enforcer))
func WithAuthz(enforcer *casbin.Enforcer) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, authz.NewAuthorizer(enforcer))
	}
}

// WithHTTPSign adds HTTP signature verification middleware via
// gin-contrib/httpsign. Incoming requests must carry a valid cryptographic
// signature in the Authorization or Signature header as defined by the HTTP
// Signatures specification (draft-cavage-http-signatures).
//
// The caller provides a key store mapping key IDs to secrets (each pairing a
// shared key with a signing algorithm). Optional httpsign.Option values may
// configure required headers or custom validators; sensible defaults apply
// when omitted (date, digest, and request-target headers are required; date
// and digest validators are enabled).
//
// Requests with a missing, malformed, or invalid signature are rejected with
// 401 Unauthorised or 400 Bad Request.
//
// Example:
//
//	api.New(api.WithHTTPSign(secrets))
func WithHTTPSign(secrets httpsign.Secrets, opts ...httpsign.Option) Option {
	return func(e *Engine) {
		auth := httpsign.NewAuthenticator(secrets, opts...)
		e.middlewares = append(e.middlewares, auth.Authenticated())
	}
}

// WithSSE registers a Server-Sent Events broker at the configured path.
// By default the endpoint is mounted at GET /events; use WithSSEPath to
// customise the route. Clients receive a streaming text/event-stream
// response and the broker manages client connections and broadcasts events
// published via its Publish method.
//
// Example:
//
//	broker := api.NewSSEBroker()
//	engine, _ := api.New(api.WithSSE(broker))
func WithSSE(broker *SSEBroker) Option {
	return func(e *Engine) {
		e.sseBroker = broker
	}
}

// WithSSEPath sets a custom URL path for the SSE endpoint.
// The default path is "/events".
//
// Example:
//
//	api.New(api.WithSSEPath("/stream"))
func WithSSEPath(path string) Option {
	return func(e *Engine) {
		e.ssePath = normaliseSSEPath(path)
	}
}

// WithLocation adds reverse proxy header detection middleware via
// gin-contrib/location. It inspects X-Forwarded-Proto and X-Forwarded-Host
// headers to determine the original scheme and host when the server runs
// behind a TLS-terminating reverse proxy such as Traefik.
//
// After this middleware runs, handlers can call location.Get(c) to retrieve
// a *url.URL with the detected scheme, host, and base path.
//
// Example:
//
//	engine, _ := api.New(api.WithLocation())
func WithLocation() Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, location.Default())
	}
}

// WithGraphQL mounts a GraphQL endpoint serving the given gqlgen ExecutableSchema.
// By default the endpoint is mounted at "/graphql". Use GraphQLOption helpers to
// enable the playground UI or customise the path:
//
//	api.New(
//	    api.WithGraphQL(schema, api.WithPlayground(), api.WithGraphQLPath("/gql")),
//	)
//
// Example:
//
//	engine, _ := api.New(api.WithGraphQL(schema))
func WithGraphQL(schema graphql.ExecutableSchema, opts ...GraphQLOption) Option {
	return func(e *Engine) {
		cfg := &graphqlConfig{
			schema: schema,
			path:   defaultGraphQLPath,
		}
		for _, opt := range opts {
			opt(cfg)
		}
		e.graphql = cfg
	}
}

// WithChatCompletions mounts an OpenAI-compatible POST /v1/chat/completions
// endpoint backed by the given ModelResolver. The resolver maps model names to
// loaded inference.TextModel instances (see chat_completions.go).
//
// Use WithChatCompletionsPath to override the default "/v1/chat/completions"
// mount point. The endpoint streams Server-Sent Events when the request body
// sets "stream": true, and otherwise returns a single JSON response that
// mirrors OpenAI's chat completion payload.
//
// Example:
//
//	resolver := api.NewModelResolver()
//	engine, _ := api.New(api.WithChatCompletions(resolver))
func WithChatCompletions(resolver *ModelResolver) Option {
	return func(e *Engine) {
		e.chatCompletionsResolver = resolver
	}
}

// WithChatCompletionsPath sets a custom URL path for the chat completions
// endpoint. The default path is "/v1/chat/completions".
//
// Example:
//
//	api.New(api.WithChatCompletionsPath("/api/v1/chat/completions"))
func WithChatCompletionsPath(path string) Option {
	return func(e *Engine) {
		e.chatCompletionsPath = normaliseChatCompletionsPath(path)
	}
}

// WithSDKGen mounts POST /v1/sdk/generate. The endpoint exposes the RFC SDK
// generation contract and currently returns 501 until an artifact backend is
// configured around SDKGenerator.
func WithSDKGen() Option {
	return func(e *Engine) {
		e.sdkGenEnabled = true
	}
}

// WithOpenAPISpec mounts a standalone JSON document endpoint at
// "/v1/openapi.json" (RFC.endpoints.md — "GET /v1/openapi.json"). The generated
// spec mirrors the document surfaced by the Swagger UI but is served
// application/json directly so SDK generators and ToolBridge consumers can
// fetch it without loading the UI bundle.
//
// Example:
//
//	engine, _ := api.New(api.WithOpenAPISpec())
func WithOpenAPISpec() Option {
	return func(e *Engine) {
		e.openAPISpecEnabled = true
	}
}

// WithOpenAPISpecPath sets a custom URL path for the standalone OpenAPI JSON
// endpoint. An empty string falls back to the RFC default "/v1/openapi.json".
// The override also enables the endpoint so callers can configure the URL
// without an additional WithOpenAPISpec() call.
//
// Example:
//
//	api.New(api.WithOpenAPISpecPath("/api/v1/openapi.json"))
func WithOpenAPISpecPath(path string) Option {
	return func(e *Engine) {
		path = core.Trim(path)
		if path == "" {
			e.openAPISpecPath = defaultOpenAPISpecPath
			e.openAPISpecEnabled = true
			return
		}
		if !core.HasPrefix(path, "/") {
			path = "/" + path
		}
		e.openAPISpecPath = path
		e.openAPISpecEnabled = true
	}
}
