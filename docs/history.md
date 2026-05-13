<!-- SPDX-License-Identifier: EUPL-1.2 -->

# go-api — Project History and Known Limitations

Module: `dappco.re/go/api`

---

## Origins

`go-api` was created as the dedicated HTTP framework for the Lethean Go ecosystem. The motivation
was to give every Go package in the stack a consistent way to expose REST endpoints without each
package taking its own opinion on routing, middleware, response formatting, or OpenAPI generation.
It was scaffolded independently from the start — it was never extracted from a monolith — and has
no legacy forge-path dependencies. It now acts as the gateway module: go-api imports and
wires provider packages, while providers keep their dependency direction independent of go-api.

---

## Development Phases

### Phase 1 — Core Engine (36 tests)

Commits `889391a` through `22f8a69`

The initial phase established the foundational abstractions that all subsequent work builds on.

**Scaffold** (`889391a`):
Module path `dappco.re/go/api` created. `go.mod` initialised with Gin as the only
direct dependency.

**Response envelope** (`7835837`):
`Response[T]`, `Error`, and `Meta` types defined. `OK()`, `Fail()`, and `Paginated()` helpers
added. The generic envelope was established from the start rather than retrofitted.

**RouteGroup interface** (`6f5fb69`):
`RouteGroup` (Name, BasePath, RegisterRoutes) and `StreamGroup` (Channels) interfaces defined
in `group.go`. The interface-driven extension model was the core design decision of Phase 1.

**Engine** (`db75c88`):
`Engine` struct added with `New()`, `Register()`, `Handler()`, `Serve()`, and graceful shutdown.
Default listen address `:8080`. Built-in `GET /health` endpoint. Panic recovery via `gin.Recovery()`
always applied first.

**Bearer auth, request ID, CORS** (`d21734d`):
First three middleware options: `WithBearerAuth()`, `WithRequestID()`, `WithCORS()`.
The functional `Option` type and the `With*()` pattern were established here.

**Swagger UI** (`22f8a69`):
`WithSwagger()` added. The initial implementation served a static Swagger UI backed by a placeholder
spec; this was later replaced in Phase 3.

**WebSocket** (`22f8a69`):
`WithWSHandler()` added, mounting any `http.Handler` at `GET /ws`. `Engine.Channels()` added to
aggregate channel names from `StreamGroup` implementations.

By the end of Phase 1, the module had 36 tests covering the engine lifecycle, health endpoint,
response helpers, bearer auth, request ID propagation, CORS, and WebSocket mounting.

---

### Phase 2 — 21 Middleware Options (143 tests)

Commits `d760e77` through `8ba1716`

Phase 2 expanded the middleware library in four waves, reaching 21 `With*()` options total.

**Wave 1 — Security and Identity** (`d760e77` through `8f3e496`):

The Authentik integration was the most significant addition of this wave.

- `WithAuthentik()` — permissive forward-auth middleware. Reads `X-authentik-*` headers when
  `TrustedProxy: true`, validates JWTs via OIDC discovery when `Issuer` and `ClientID` are set.
  Fail-open: unauthenticated requests are never rejected by this middleware alone.
- `RequireAuth()`, `RequireGroup()` — explicit guards for protected routes, returning 401 and 403
  respectively via the standard `Fail()` envelope.
- `GetUser()` — context accessor for the current `AuthentikUser`.
- `AuthentikUser` — carries Username, Email, Name, UID, Groups, Entitlements, and JWT. `HasGroup()`
  convenience method added.
- Live integration tests added in `authentik_integration_test.go`, guarded by environment variables.
- `WithSecure()` — HSTS, X-Frame-Options DENY, X-Content-Type-Options nosniff, strict referrer
  policy. SSL redirect deliberately omitted to work correctly behind a TLS-terminating proxy.

**Wave 2 — Compression and Logging** (`6521b90` through `6bb7195`):

- `WithTimeout(d)` — per-request deadline via gin-contrib/timeout. Returns 504 with the standard
  Fail() envelope on expiry.
- `WithGzip(level...)` — gzip response compression; defaults to `gzip.DefaultCompression`.
- `WithBrotli(level...)` — Brotli compression via `andybalholm/brotli`. Custom `brotliHandler`
  wrapping `brotli.HTTPCompressor`.
- `WithSlog(logger)` — structured request logging via gin-contrib/slog. Falls back to
  `slog.Default()` when nil is passed.
- `WithStatic(prefix, root)` — static file serving via gin-contrib/static; directory listing
  disabled.

**Wave 3 — Auth, Caching, Streaming** (`0ab962a` through `7b3f99e`):

- `WithCache(ttl)` — in-memory GET response cache. Custom `cacheWriter` intercepts the response
  body without affecting the downstream handler. `X-Cache: HIT` on served cache entries.
- `WithSessions(name, secret)` — cookie-backed server sessions via gin-contrib/sessions.
- `WithAuthz(enforcer)` — Casbin policy-based authorisation via gin-contrib/authz. Subject from
  HTTP Basic Auth.
- `WithHTTPSign(secrets, opts...)` — HTTP Signatures verification via gin-contrib/httpsign.
- `WithSSE(broker)` — Server-Sent Events at `GET /events`. `SSEBroker` added with `Publish()`,
  channel filtering, 64-event per-client buffer, and `Drain()` for graceful shutdown.

**Wave 4 — Infrastructure and Protocol** (`a612d85` through `8ba1716`):

- `WithLocation()` — reverse proxy header detection via gin-contrib/location/v2.
- `WithI18n(cfg...)` — Accept-Language parsing and BCP 47 locale matching via
  `golang.org/x/text/language`. `GetLocale()` and `GetMessage()` context accessors added.
- `WithGraphQL(schema, opts...)` — gqlgen `ExecutableSchema` mounting. `WithPlayground()` and
  `WithGraphQLPath()` sub-options. Playground at `{path}/playground`.
- `WithPprof()` — Go runtime profiling at `/debug/pprof/`.
- `WithExpvar()` — expvar runtime metrics at `/debug/vars`.
- `WithTracing(name, opts...)` — OpenTelemetry distributed tracing via otelgin. `NewTracerProvider()`
  convenience helper added. W3C `traceparent` propagation.

At the end of Phase 2, the module had 143 tests.

---

### Phase 3 — OpenAPI, ToolBridge, SDK Codegen (176 tests)

Commits `465bd60` through `1910aec`

Phase 3 upgraded the Swagger integration from a placeholder to a full runtime OpenAPI 3.1 pipeline
and added two new subsystems: `ToolBridge` and `SDKGenerator`.

**DescribableGroup interface** (`465bd60`):
`DescribableGroup` added to `group.go`, extending `RouteGroup` with `Describe() []RouteDescription`.
`RouteDescription` carries HTTP method, path, summary, description, tags, and JSON Schema maps
for request body and response data. This was the contract that `SpecBuilder` and `ToolBridge`
would both consume.

**ToolBridge** (`2b63c7b`):
`ToolBridge` added to `bridge.go`. Converts `ToolDescriptor` values into `POST /{tool_name}`
Gin routes and implements `DescribableGroup` so those routes appear in the OpenAPI spec. Designed
to bridge the MCP tool model (as used by go-ai) into the REST world. `Tools()` accessor added
for external enumeration.

**SpecBuilder** (`3e96f9b`):
`SpecBuilder` added to `openapi.go`. Generates a complete OpenAPI 3.1 JSON document from registered
`RouteGroup` and `DescribableGroup` values. Includes the built-in `GET /health` endpoint, shared
`Error` and `Meta` component schemas, and the `Response[T]` envelope schema wrapping every response
body. Tags are derived from all group names, not just describable ones.

**Spec export** (`e94283b`):
`ExportSpec()` and `ExportSpecToFile()` added to `export.go`. Supports `"json"` and `"yaml"`
output formats. YAML output is produced by unmarshalling the JSON then re-encoding with
`gopkg.in/yaml.v3` at 2-space indentation. Parent directories created automatically by
`ExportSpecToFile()`.

**Swagger refactor** (`303779f`):
`registerSwagger()` in `swagger.go` rewritten to use `SpecBuilder` rather than the previous
placeholder. A `swaggerSpec` wrapper satisfies the `swag.Spec` interface and builds the spec
lazily on first access via `sync.Once`. A `swaggerSeq` atomic counter assigns unique instance
names so multiple `Engine` instances in the same test binary do not collide in the global
`swag` registry.

**SDK codegen** (`a09a4e9`, `1910aec`):
`SDKGenerator` added to `codegen.go`. Wraps `openapi-generator-cli` to generate client SDKs
for 11 target languages. `SupportedLanguages()` returns the list in sorted order (the sort was
added in `1910aec` to ensure deterministic output in tests and documentation).

At the end of Phase 3, the module has 176 tests.

---

## Known Limitations

### 1. Cache remains in-memory

`WithCache(ttl, maxEntries, maxBytes)` can now bound the cache by entry count and approximate
payload size, but it still stores responses in memory. Workloads with very large cached bodies
or a long-lived process will still consume RAM, so a disk-backed cache would be the next step if
that becomes a concern.

### 2. SDK codegen requires an external binary

`SDKGenerator.Generate()` shells out to `openapi-generator-cli`. This requires a JVM and the
openapi-generator JAR to be installed on the host. `Available()` checks whether the CLI is on
`PATH` but there is no programmatic fallback. Packaging `openapi-generator-cli` via a Docker
wrapper or replacing it with a pure-Go generator would remove this external dependency.

### 3. OpenAPI spec generation is build-time only

`SpecBuilder.Build()` generates the spec from `Describe()` return values, which are static at
the time of construction. Dynamic route generation (for example, routes registered after
`New()` returns) is not reflected in the spec. This matches the current design — all groups
must be registered before `Serve()` is called — but it would conflict with any future dynamic
route registration model.

### 4. i18n message map is a lightweight bridge only

`WithI18n()` accepts a `Messages map[string]map[string]string` for simple key-value lookups.
It does not support pluralisation, gender inflection, argument interpolation, or any of the
grammar features provided by `go-i18n`. Applications requiring production-grade localisation
should use `go-i18n` directly and use `GetLocale()` to pass the detected locale to it.

### 5. Authentik JWT validation performs OIDC discovery on first request

`getOIDCProvider()` performs an OIDC discovery request on first use and caches the resulting
`*oidc.Provider` by issuer URL. This is lazy — the first request to a non-public path will
incur a network round-trip to the issuer. A warm-up call during application startup would
eliminate this latency from the first real request.

### 6. ToolBridge has no input validation

`ToolBridge.Add()` accepts a `ToolDescriptor` with `InputSchema` and `OutputSchema` maps, but
these are used only for OpenAPI documentation. The registered `gin.HandlerFunc` is responsible
for its own input validation. There is no automatic binding or validation of incoming request
bodies against the declared JSON Schema.

### 7. SSEBroker.Drain() does not wait for clients to disconnect

`Drain()` closes all client event channels to signal disconnection but returns immediately
without waiting for client goroutines to exit. In a graceful shutdown sequence, there is a
brief window where client HTTP connections remain open. The engine's 10-second shutdown
deadline covers this window in practice, but there is no explicit coordination.
