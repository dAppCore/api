// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"compress/gzip"
	"log/slog"
	"net/http"
	"slices"
	"time"

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
type Option func(*Engine)

// WithAddr sets the listen address for the server.
func WithAddr(addr string) Option {
	return func(e *Engine) {
		e.addr = addr
	}
}

// WithBearerAuth adds bearer token authentication middleware.
// Requests to /health and paths starting with /swagger are exempt.
func WithBearerAuth(token string) Option {
	return func(e *Engine) {
		skip := []string{"/health", "/swagger"}
		e.middlewares = append(e.middlewares, bearerAuthMiddleware(token, skip))
	}
}

// WithRequestID adds middleware that assigns an X-Request-ID to every response.
// Client-provided IDs are preserved; otherwise a random hex ID is generated.
func WithRequestID() Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, requestIDMiddleware())
	}
}

// WithResponseMeta attaches request metadata to JSON envelope responses.
// It preserves any existing pagination metadata and merges in request_id
// and duration when available from the request context. Combine it with
// WithRequestID() to populate both fields automatically.
func WithResponseMeta() Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, responseMetaMiddleware())
	}
}

// WithCORS configures Cross-Origin Resource Sharing via gin-contrib/cors.
// Pass "*" to allow all origins, or supply specific origin URLs.
// Standard methods (GET, POST, PUT, PATCH, DELETE, OPTIONS) and common
// headers (Authorization, Content-Type, X-Request-ID) are permitted.
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
func WithMiddleware(mw ...gin.HandlerFunc) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, mw...)
	}
}

// WithStatic serves static files from the given root directory at urlPrefix.
// Directory listing is disabled; only individual files are served.
// Internally this uses gin-contrib/static as Gin middleware.
func WithStatic(urlPrefix, root string) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, static.Serve(urlPrefix, static.LocalFile(root, false)))
	}
}

// WithWSHandler registers a WebSocket handler at GET /ws.
// Typically this wraps a go-ws Hub.Handler().
func WithWSHandler(h http.Handler) Option {
	return func(e *Engine) {
		e.wsHandler = h
	}
}

// WithAuthentik adds Authentik forward-auth middleware that extracts user
// identity from X-authentik-* headers set by a trusted reverse proxy.
// The middleware is permissive: unauthenticated requests are allowed through.
func WithAuthentik(cfg AuthentikConfig) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, authentikMiddleware(cfg))
	}
}

// WithSwagger enables the Swagger UI at /swagger/.
// The title, description, and version populate the OpenAPI info block.
func WithSwagger(title, description, version string) Option {
	return func(e *Engine) {
		e.swaggerTitle = title
		e.swaggerDesc = description
		e.swaggerVersion = version
		e.swaggerEnabled = true
	}
}

// WithSwaggerContact adds contact metadata to the generated Swagger spec.
// Empty fields are ignored. Multiple calls replace the previous contact data.
func WithSwaggerContact(name, url, email string) Option {
	return func(e *Engine) {
		e.swaggerContactName = name
		e.swaggerContactURL = url
		e.swaggerContactEmail = email
	}
}

// WithSwaggerServers adds OpenAPI server metadata to the generated Swagger spec.
// Empty strings are ignored. Multiple calls append and normalise the combined
// server list so callers can compose metadata across options.
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
		e.swaggerLicenseName = name
		e.swaggerLicenseURL = url
	}
}

// WithPprof enables Go runtime profiling endpoints at /debug/pprof/.
// The standard pprof handlers (index, cmdline, profile, symbol, trace,
// allocs, block, goroutine, heap, mutex, threadcreate) are registered
// via gin-contrib/pprof.
//
// WARNING: pprof exposes sensitive runtime data and should only be
// enabled in development or behind authentication in production.
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

// WithCache adds in-memory response caching middleware for GET requests.
// Successful (2xx) GET responses are cached for the given TTL and served
// with an X-Cache: HIT header on subsequent requests. Non-GET methods
// and error responses pass through uncached.
//
// An optional maxEntries limit enables LRU eviction when the cache reaches
// capacity. A value <= 0 keeps the cache unbounded for backward compatibility.
func WithCache(ttl time.Duration, maxEntries ...int) Option {
	return func(e *Engine) {
		limit := 0
		if len(maxEntries) > 0 {
			limit = maxEntries[0]
		}
		store := newCacheStore(limit)
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
func WithRateLimit(limit int) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, rateLimitMiddleware(limit))
	}
}

// WithSessions adds server-side session management middleware via
// gin-contrib/sessions using a cookie-based store. The name parameter
// sets the session cookie name (e.g. "session") and secret is the key
// used for cookie signing and encryption.
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
func WithHTTPSign(secrets httpsign.Secrets, opts ...httpsign.Option) Option {
	return func(e *Engine) {
		auth := httpsign.NewAuthenticator(secrets, opts...)
		e.middlewares = append(e.middlewares, auth.Authenticated())
	}
}

// WithSSE registers a Server-Sent Events broker at GET /events.
// Clients connect to the endpoint and receive a streaming text/event-stream
// response. The broker manages client connections and broadcasts events
// published via its Publish method.
func WithSSE(broker *SSEBroker) Option {
	return func(e *Engine) {
		e.sseBroker = broker
	}
}

// WithLocation adds reverse proxy header detection middleware via
// gin-contrib/location. It inspects X-Forwarded-Proto and X-Forwarded-Host
// headers to determine the original scheme and host when the server runs
// behind a TLS-terminating reverse proxy such as Traefik.
//
// After this middleware runs, handlers can call location.Get(c) to retrieve
// a *url.URL with the detected scheme, host, and base path.
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
