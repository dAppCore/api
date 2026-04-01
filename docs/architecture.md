---
title: Architecture
description: Internals of the go-api REST framework -- Engine, RouteGroup, middleware composition, response envelope, authentication, real-time transports, OpenAPI generation, and SDK codegen.
---

<!-- SPDX-License-Identifier: EUPL-1.2 -->

# Architecture

This document explains how go-api works internally. It covers every major subsystem, the key
types, and the data flow from incoming HTTP request to outgoing JSON response.

---

## 1. Engine

### 1.1 The Engine struct

`Engine` is the central container. It holds the listen address, the ordered list of registered
route groups, the middleware chain, and all optional integrations:

```go
type Engine struct {
    addr           string
    groups         []RouteGroup
    middlewares    []gin.HandlerFunc
    wsHandler      http.Handler
    sseBroker      *SSEBroker
    swaggerEnabled bool
    swaggerTitle   string
    swaggerDesc    string
    swaggerVersion string
    pprofEnabled   bool
    expvarEnabled  bool
    graphql        *graphqlConfig
}
```

All fields are private. Configuration happens exclusively through `Option` functions passed
to `New()`.

### 1.2 Construction

`New()` applies functional options and returns a configured engine. The default listen address
is `:8080`. No middleware is added automatically beyond Gin's built-in panic recovery; every
feature requires an explicit `With*()` option:

```go
engine, err := api.New(
    api.WithAddr(":9000"),
    api.WithBearerAuth("secret"),
    api.WithCORS("*"),
    api.WithRequestID(),
    api.WithSlog(nil),
    api.WithSwagger("My API", "Description", "1.0.0"),
)
```

After construction, call `engine.Register(group)` to add route groups, then either
`engine.Serve(ctx)` to start an HTTP server or `engine.Handler()` to obtain an `http.Handler`
for use with `httptest` or an external server.

### 1.3 Build sequence

`Engine.build()` is called internally by `Handler()` and `Serve()`. It assembles a fresh
`*gin.Engine` each time. The order is fixed:

1. `gin.Recovery()` -- panic recovery (always first).
2. User middleware in registration order -- all `With*()` options that append to `e.middlewares`.
3. Built-in `GET /health` endpoint -- always present, returns `{"success":true,"data":"healthy"}`.
4. Route groups -- each mounted at its `BasePath()`.
5. WebSocket handler at `GET /ws` -- when `WithWSHandler()` was called.
6. SSE broker at `GET /events` -- when `WithSSE()` was called.
7. GraphQL endpoint -- when `WithGraphQL()` was called.
8. Swagger UI at `GET /swagger/*any` -- when `WithSwagger()` was called.
9. pprof endpoints at `GET /debug/pprof/*` -- when `WithPprof()` was called.
10. expvar endpoint at `GET /debug/vars` -- when `WithExpvar()` was called.

### 1.4 Graceful shutdown

`Serve()` starts an `http.Server` in a goroutine and blocks on `ctx.Done()`. When the context
is cancelled, a 10-second shutdown deadline is applied. In-flight requests complete or time out
before the server exits. Any listen error that occurred before shutdown is returned to the
caller.

### 1.5 Iterators

`Engine` provides iterator methods following Go 1.23+ conventions:

- `GroupsIter()` returns `iter.Seq[RouteGroup]` over all registered groups.
- `ChannelsIter()` returns `iter.Seq[string]` over WebSocket channel names from groups that
  implement `StreamGroup`.

---

## 2. RouteGroup, StreamGroup, and DescribableGroup

Three interfaces form the extension point model:

```go
// RouteGroup is the minimum interface. All subsystems implement this.
type RouteGroup interface {
    Name() string
    BasePath() string
    RegisterRoutes(rg *gin.RouterGroup)
}

// StreamGroup is optionally implemented by groups that publish WebSocket channels.
type StreamGroup interface {
    Channels() []string
}

// DescribableGroup extends RouteGroup with OpenAPI metadata.
// Groups implementing this have their endpoints included in the generated spec.
type DescribableGroup interface {
    RouteGroup
    Describe() []RouteDescription
}
```

`RouteDescription` carries the HTTP method, path (relative to `BasePath()`), summary,
description, tags, and JSON Schema maps for the request body and response data:

```go
type RouteDescription struct {
    Method      string
    Path        string
    Summary     string
    Description string
    Tags        []string
    RequestBody map[string]any
    Response    map[string]any
}
```

`Engine.Channels()` iterates all registered groups and collects channel names from those that
implement `StreamGroup`. This list is used when initialising a WebSocket hub.

---

## 3. Middleware Stack

All middleware options append to `Engine.middlewares` in the order they are passed to `New()`.
They execute after `gin.Recovery()` but before any route handler. The `Option` type is simply
`func(*Engine)`.

### Complete option reference

| Option | Purpose | Key detail |
|--------|---------|-----------|
| `WithAddr(addr)` | Listen address | Default `:8080` |
| `WithBearerAuth(token)` | Static bearer token authentication | Skips `/health` and `/swagger` |
| `WithRequestID()` | `X-Request-ID` propagation | Preserves client-supplied IDs; generates 16-byte hex otherwise |
| `WithResponseMeta()` | Request metadata in JSON envelopes | Merges `request_id` and `duration` into standard responses |
| `WithCORS(origins...)` | CORS policy | `"*"` enables `AllowAllOrigins`; 12-hour `MaxAge` |
| `WithRateLimit(limit)` | Per-IP token-bucket rate limiting | `429 Too Many Requests`; `Retry-After` on rejection; zero or negative disables |
| `WithMiddleware(mw...)` | Arbitrary Gin middleware | Escape hatch for custom middleware |
| `WithStatic(prefix, root)` | Static file serving | Directory listing disabled |
| `WithWSHandler(h)` | WebSocket at `/ws` | Wraps any `http.Handler` |
| `WithAuthentik(cfg)` | Authentik forward-auth + OIDC JWT | Permissive; populates context, never rejects |
| `WithSwagger(title, desc, ver)` | Swagger UI at `/swagger/` | Runtime spec via `SpecBuilder` |
| `WithPprof()` | Go profiling at `/debug/pprof/` | WARNING: do not expose in production without authentication |
| `WithExpvar()` | Runtime metrics at `/debug/vars` | WARNING: do not expose in production without authentication |
| `WithSecure()` | Security headers | HSTS 1 year, X-Frame-Options DENY, nosniff, strict referrer |
| `WithGzip(level...)` | Gzip response compression | Default compression if level omitted |
| `WithBrotli(level...)` | Brotli response compression | Writer pool for efficiency; default compression if level omitted |
| `WithSlog(logger)` | Structured request logging | Falls back to `slog.Default()` if nil |
| `WithTimeout(d)` | Per-request deadline | 504 with standard error envelope on timeout |
| `WithCache(ttl)` | In-memory GET response caching | `X-Cache: HIT` header on cache hits; 2xx only |
| `WithSessions(name, secret)` | Cookie-backed server sessions | gin-contrib/sessions with cookie store |
| `WithAuthz(enforcer)` | Casbin policy-based authorisation | Subject from HTTP Basic Auth; 403 on deny |
| `WithHTTPSign(secrets, opts...)` | HTTP Signatures verification | draft-cavage-http-signatures; 401/400 on failure |
| `WithSSE(broker)` | Server-Sent Events at `/events` | `?channel=` query parameter filtering |
| `WithLocation()` | Reverse proxy header detection | X-Forwarded-Proto / X-Forwarded-Host |
| `WithI18n(cfg...)` | Accept-Language locale detection | BCP 47 matching via `golang.org/x/text/language` |
| `WithTracing(name, opts...)` | OpenTelemetry distributed tracing | otelgin + W3C `traceparent` header propagation |
| `WithGraphQL(schema, opts...)` | GraphQL endpoint | gqlgen `ExecutableSchema`; optional playground UI |

### Bearer authentication flow

`bearerAuthMiddleware` validates the `Authorization: Bearer <token>` header. Requests to paths
in the skip list (`/health`, `/swagger`) pass through without authentication. Missing or
invalid tokens produce a `401 Unauthorised` response using the standard error envelope.

### Request ID flow

`requestIDMiddleware` checks for an incoming `X-Request-ID` header. If present, the value is
preserved. Otherwise, a cryptographically random 16-byte hex string is generated. The ID is
stored in the Gin context under the key `"request_id"` and set as an `X-Request-ID` response
header.

---

## 4. Response Envelope

All API responses use a single generic envelope:

```go
type Response[T any] struct {
    Success bool   `json:"success"`
    Data    T      `json:"data,omitempty"`
    Error   *Error `json:"error,omitempty"`
    Meta    *Meta  `json:"meta,omitempty"`
}
```

Supporting types:

```go
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

type Meta struct {
    RequestID string `json:"request_id,omitempty"`
    Duration  string `json:"duration,omitempty"`
    Page      int    `json:"page,omitempty"`
    PerPage   int    `json:"per_page,omitempty"`
    Total     int    `json:"total,omitempty"`
}
```

### Constructor helpers

| Helper | Produces |
|--------|----------|
| `OK(data)` | `{"success":true,"data":...}` |
| `Fail(code, message)` | `{"success":false,"error":{"code":"...","message":"..."}}` |
| `FailWithDetails(code, message, details)` | Same as `Fail` with an additional `details` field |
| `Paginated(data, page, perPage, total)` | `{"success":true,"data":...,"meta":{"page":...,"per_page":...,"total":...}}` |

All handlers should use these helpers rather than constructing `Response[T]` manually. This
guarantees a consistent envelope across every route group.

---

## 5. Authentik Integration

The `WithAuthentik()` option installs a permissive identity middleware that runs on every
non-public request. It has two extraction paths:

### Path 1 -- Forward-auth headers (TrustedProxy: true)

When a reverse proxy (e.g. Traefik) is configured with Authentik forward-auth, it injects
headers: `X-authentik-username`, `X-authentik-email`, `X-authentik-name`, `X-authentik-uid`,
`X-authentik-groups` (pipe-separated), `X-authentik-entitlements` (pipe-separated), and
`X-authentik-jwt`. The middleware reads these and populates an `AuthentikUser` in the Gin
context.

### Path 2 -- OIDC JWT validation

For direct API clients that present a `Bearer` token, the middleware validates the JWT against
the configured OIDC issuer and client ID. Providers are cached by issuer URL to avoid repeated
discovery requests.

### Fail-open behaviour

In both paths, if extraction fails the request continues unauthenticated. The middleware never
rejects requests. Handlers check identity with:

```go
user := api.GetUser(c) // returns nil when unauthenticated
```

### Route guards

For protected routes, apply guards as Gin middleware on individual routes:

```go
rg.GET("/private", api.RequireAuth(), handler)          // 401 if no user
rg.GET("/admin",   api.RequireGroup("admins"), handler) // 403 if wrong group
```

`RequireAuth()` returns 401 when `GetUser(c)` is nil. `RequireGroup(group)` returns 401 when
no user is present, or 403 when the user lacks the specified group membership.

### AuthentikUser type

```go
type AuthentikUser struct {
    Username     string   `json:"username"`
    Email        string   `json:"email"`
    Name         string   `json:"name"`
    UID          string   `json:"uid"`
    Groups       []string `json:"groups,omitempty"`
    Entitlements []string `json:"entitlements,omitempty"`
    JWT          string   `json:"-"`
}
```

The `HasGroup(group string) bool` method provides a convenience check for group membership.

### Configuration

```go
type AuthentikConfig struct {
    Issuer       string   // OIDC issuer URL
    ClientID     string   // OAuth2 client identifier
    TrustedProxy bool     // Whether to read X-authentik-* headers
    PublicPaths  []string // Additional paths exempt from header extraction
}
```

`/health` and `/swagger` are always public. Additional paths may be specified via
`PublicPaths`.

---

## 6. WebSocket and Server-Sent Events

### WebSocket

`WithWSHandler(h)` mounts any `http.Handler` at `GET /ws`. The handler is responsible for
upgrading the connection. The intended pairing is a WebSocket hub (e.g. from go-ws):

```go
hub := ws.NewHub()
go hub.Run(ctx)
engine, _ := api.New(api.WithWSHandler(hub.Handler()))
```

Groups implementing `StreamGroup` declare channel names, which `Engine.Channels()` aggregates
into a single slice.

### Server-Sent Events

`SSEBroker` manages persistent SSE connections at `GET /events`. Clients optionally subscribe
to a named channel via the `?channel=<name>` query parameter. Clients without a channel
parameter receive events on all channels.

```go
broker := api.NewSSEBroker()
engine, _ := api.New(api.WithSSE(broker))

// Publish from anywhere:
broker.Publish("deployments", "deploy.started", payload)
```

Key implementation details:

- Each client has a 64-event buffered channel. Overflow events are dropped without blocking
  the publisher.
- `SSEBroker.ClientCount()` returns the number of currently connected clients.
- `SSEBroker.Drain()` signals all clients to disconnect, useful during graceful shutdown.
- The response is streamed with `Content-Type: text/event-stream`, `Cache-Control: no-cache`,
  `Connection: keep-alive`, and `X-Accel-Buffering: no` headers.
- Data payloads are JSON-encoded before being written as SSE `data:` fields.

---

## 7. GraphQL

`WithGraphQL()` mounts a gqlgen `ExecutableSchema` at `/graphql` (or a custom path via
`WithGraphQLPath()`). An optional `WithPlayground()` adds the interactive GraphQL Playground
at `{path}/playground`.

```go
engine, _ := api.New(
    api.WithGraphQL(schema,
        api.WithPlayground(),
        api.WithGraphQLPath("/gql"),
    ),
)
```

The endpoint accepts all HTTP methods (POST for queries and mutations, GET for playground
redirects and introspection). The GraphQL handler is created via gqlgen's
`handler.NewDefaultServer()`.

---

## 8. Response Caching

`WithCache(ttl)` installs a URL-keyed in-memory response cache scoped to GET requests:

- Only successful 2xx responses are cached.
- Non-GET methods pass through uncached.
- Cached responses are served with an `X-Cache: HIT` header.
- Expired entries are evicted lazily on the next access for the same key.
- The cache is not shared across `Engine` instances.
- There is no size limit on the cache.

The implementation uses a `cacheWriter` that wraps `gin.ResponseWriter` to intercept and
capture the response body and status code for storage.

---

## 9. Brotli Compression

`WithBrotli(level...)` adds Brotli response compression. The middleware checks the
`Accept-Encoding` header for `br` support before compressing.

Key implementation details:

- A `sync.Pool` of `brotli.Writer` instances is used to avoid allocation per request.
- Error responses (4xx and above) bypass compression and are sent uncompressed.
- Three compression level constants are exported: `BrotliBestSpeed`, `BrotliBestCompression`,
  and `BrotliDefaultCompression`.

---

## 10. Internationalisation

`WithI18n(cfg)` parses the `Accept-Language` header on every request and stores the resolved
BCP 47 locale tag in the Gin context. The `golang.org/x/text/language` matcher handles
quality-weighted negotiation and script/region subtag matching.

```go
engine, _ := api.New(
    api.WithI18n(api.I18nConfig{
        DefaultLocale: "en",
        Supported:     []string{"en", "fr", "de"},
        Messages: map[string]map[string]string{
            "en": {"greeting": "Hello"},
            "fr": {"greeting": "Bonjour"},
        },
    }),
)
```

Handlers retrieve the locale and optional localised messages:

```go
locale := api.GetLocale(c)               // e.g. "en", "fr"
msg, ok := api.GetMessage(c, "greeting") // from configured Messages map
```

The built-in message map is a lightweight bridge. The `go-i18n` grammar engine is the intended
replacement for production-grade localisation.

---

## 11. OpenTelemetry Tracing

`WithTracing(serviceName)` adds otelgin middleware that creates a span for each request, tagged
with HTTP method, route template, and response status code. Trace context is propagated via the
W3C `traceparent` header.

`NewTracerProvider(exporter)` is a convenience helper for tests and simple deployments that
constructs a synchronous `TracerProvider` and installs it globally:

```go
tp := api.NewTracerProvider(exporter)
defer tp.Shutdown(ctx)

engine, _ := api.New(api.WithTracing("my-service"))
```

Production deployments should build a batching provider with appropriate resource attributes
and span processors.

---

## 12. OpenAPI Specification Generation

### SpecBuilder

`SpecBuilder` generates an OpenAPI 3.1 JSON document from registered route groups:

```go
builder := &api.SpecBuilder{
    Title:       "My API",
    Description: "Service description",
    Version:     "1.0.0",
}
data, err := builder.Build(engine.Groups())
```

The built document includes:

- The `GET /health` endpoint under the `system` tag.
- One path entry per `RouteDescription` returned by `DescribableGroup.Describe()`.
- `#/components/schemas/Error` and `#/components/schemas/Meta` shared schemas.
- All response bodies wrapped in the `Response[T]` envelope schema.
- Tags derived from every registered group's `Name()`.

Groups that implement `RouteGroup` but not `DescribableGroup` contribute a tag but no paths.

### Export

Two convenience functions write the spec to an `io.Writer` or directly to a file:

```go
// Write JSON or YAML to a writer:
api.ExportSpec(os.Stdout, "yaml", builder, engine.Groups())

// Write to a file (parent directories created automatically):
api.ExportSpecToFile("./api/openapi.yaml", "yaml", builder, engine.Groups())
```

### Swagger UI

When `WithSwagger()` is active, the spec is built lazily on first access by a `swaggerSpec`
wrapper that satisfies the `swag.Spec` interface. It is registered in the global `swag` registry
with a unique sequence-based instance name (via `atomic.Uint64`), so multiple `Engine` instances
in the same process do not collide.

---

## 13. ToolBridge

`ToolBridge` converts tool descriptors into REST endpoints and OpenAPI paths. It implements both
`RouteGroup` and `DescribableGroup`. This is the primary mechanism for exposing MCP tool
descriptors as a REST API.

```go
bridge := api.NewToolBridge("/v1/tools")
bridge.Add(api.ToolDescriptor{
    Name:        "file_read",
    Description: "Read a file",
    Group:       "files",
    InputSchema: map[string]any{"type": "object", "properties": ...},
}, fileReadHandler)

engine.Register(bridge)
```

Each registered tool becomes a `POST /v1/tools/{tool_name}` endpoint. The bridge provides:

- `Tools()` / `ToolsIter()` -- enumerate registered tool descriptors.
- `Describe()` / `DescribeIter()` -- generate `RouteDescription` entries for OpenAPI.

`ToolDescriptor` carries:

```go
type ToolDescriptor struct {
    Name         string         // Tool name (becomes POST path segment)
    Description  string         // Human-readable description
    Group        string         // OpenAPI tag group
    InputSchema  map[string]any // JSON Schema for request body
    OutputSchema map[string]any // JSON Schema for response data (optional)
}
```

---

## 14. SDK Codegen

`SDKGenerator` wraps `openapi-generator-cli` to generate client SDKs from an exported OpenAPI
spec:

```go
gen := &api.SDKGenerator{
    SpecPath:    "./api/openapi.yaml",
    OutputDir:   "./sdk",
    PackageName: "myapi",
}
if gen.Available() {
    _ = gen.Generate(ctx, "typescript-fetch")
    _ = gen.Generate(ctx, "python")
}
```

Supported target languages (11 total): `csharp`, `go`, `java`, `kotlin`, `php`, `python`,
`ruby`, `rust`, `swift`, `typescript-axios`, `typescript-fetch`.

- `SupportedLanguages()` returns the full list in sorted order.
- `SupportedLanguagesIter()` returns an `iter.Seq[string]` over the same list.
- `SDKGenerator.Available()` checks whether `openapi-generator-cli` is on `PATH`.

---

## 15. CLI Subcommands

The `cmd/api/` package registers two CLI subcommands under the `core api` namespace:

### `core api spec`

Generates an OpenAPI 3.1 specification from registered route groups.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | (stdout) | Write spec to file |
| `--format` | `-f` | `json` | Output format: `json` or `yaml` |
| `--title` | `-t` | `Lethean Core API` | API title |
| `--version` | `-V` | `1.0.0` | API version |

### `core api sdk`

Generates client SDKs from an OpenAPI spec using `openapi-generator-cli`.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--lang` | `-l` | (required) | Target language(s), comma-separated |
| `--output` | `-o` | `./sdk` | Output directory |
| `--spec` | `-s` | (auto-generated) | Path to existing OpenAPI spec |
| `--package` | `-p` | `lethean` | Package name for generated SDK |

---

## 16. Data Flow Summary

```
HTTP Request
    |
    v
gin.Recovery()            -- panic recovery
    |
    v
User middleware chain     -- WithBearerAuth, WithCORS, WithRequestID, WithAuthentik, etc.
    |                        (in registration order)
    v
Route matching            -- /health (built-in) or BasePath() + route from RouteGroup
    |
    v
Handler function          -- uses api.OK(), api.Fail(), api.Paginated()
    |
    v
Response[T] envelope      -- {"success": bool, "data": T, "error": Error, "meta": Meta}
    |
    v
HTTP Response
```

Real-time transports (WebSocket at `/ws`, SSE at `/events`) and development endpoints
(Swagger at `/swagger/`, pprof at `/debug/pprof/`, expvar at `/debug/vars`) are mounted
alongside the route groups during the build phase.
