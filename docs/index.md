---
title: go-api
description: Gin-based REST framework with OpenAPI generation, middleware composition, and SDK codegen for the Lethean Go ecosystem.
---

<!-- SPDX-License-Identifier: EUPL-1.2 -->

# go-api

**Module path:** `forge.lthn.ai/core/go-api`
**Language:** Go 1.26
**Licence:** EUPL-1.2

go-api is a REST framework built on top of [Gin](https://github.com/gin-gonic/gin). It provides
an `Engine` that subsystems plug into via the `RouteGroup` interface. Each ecosystem package
(go-ai, go-ml, go-rag, and others) registers its own route group, and go-api handles the HTTP
plumbing: middleware composition, response envelopes, WebSocket and SSE integration, GraphQL
hosting, Authentik identity, OpenAPI 3.1 specification generation, and client SDK codegen.

go-api is a library. It has no `main` package and produces no binary on its own. Callers
construct an `Engine`, register route groups, and call `Serve()`.

---

## Quick Start

```go
package main

import (
    "context"
    "os/signal"
    "syscall"

    api "forge.lthn.ai/core/go-api"
)

func main() {
    engine, _ := api.New(
        api.WithAddr(":8080"),
        api.WithBearerAuth("my-secret-token"),
        api.WithCORS("*"),
        api.WithRequestID(),
        api.WithSecure(),
        api.WithSlog(nil),
        api.WithSwagger("My API", "A service description", "1.0.0"),
    )

    engine.Register(myRoutes) // any RouteGroup implementation

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    _ = engine.Serve(ctx) // blocks until context is cancelled, then shuts down gracefully
}
```

The default listen address is `:8080`. A built-in `GET /health` endpoint is always present.
Every feature beyond panic recovery requires an explicit `With*()` option.

---

## Implementing a RouteGroup

Any type that satisfies the `RouteGroup` interface can register endpoints:

```go
type Routes struct{ service *mypackage.Service }

func (r *Routes) Name() string     { return "mypackage" }
func (r *Routes) BasePath() string { return "/v1/mypackage" }

func (r *Routes) RegisterRoutes(rg *gin.RouterGroup) {
    rg.GET("/items", r.listItems)
    rg.POST("/items", r.createItem)
}

func (r *Routes) listItems(c *gin.Context) {
    items, _ := r.service.List(c.Request.Context())
    c.JSON(200, api.OK(items))
}
```

Register with the engine:

```go
engine.Register(&Routes{service: svc})
```

---

## Package Layout

| File | Purpose |
|------|---------|
| `api.go` | `Engine` struct, `New()`, `build()`, `Serve()`, `Handler()`, `Channels()` |
| `options.go` | All `With*()` option functions (28 options) |
| `group.go` | `RouteGroup`, `StreamGroup`, `DescribableGroup` interfaces; `RouteDescription` |
| `response.go` | `Response[T]`, `Error`, `Meta`, `OK()`, `Fail()`, `FailWithDetails()`, `Paginated()` |
| `middleware.go` | `bearerAuthMiddleware()`, `requestIDMiddleware()` |
| `authentik.go` | `AuthentikUser`, `AuthentikConfig`, `GetUser()`, `RequireAuth()`, `RequireGroup()` |
| `websocket.go` | `wrapWSHandler()` helper |
| `sse.go` | `SSEBroker`, `NewSSEBroker()`, `Publish()`, `Handler()`, `Drain()`, `ClientCount()` |
| `cache.go` | `cacheStore`, `cacheEntry`, `cacheWriter`, `cacheMiddleware()` |
| `brotli.go` | `brotliHandler`, `newBrotliHandler()`; compression level constants |
| `graphql.go` | `graphqlConfig`, `GraphQLOption`, `WithPlayground()`, `WithGraphQLPath()`, `mountGraphQL()` |
| `i18n.go` | `I18nConfig`, `WithI18n()`, `i18nMiddleware()`, `GetLocale()`, `GetMessage()` |
| `tracing.go` | `WithTracing()`, `NewTracerProvider()` |
| `swagger.go` | `swaggerSpec`, `registerSwagger()`; sequence counter for multi-instance safety |
| `openapi.go` | `SpecBuilder`, `Build()`, `buildPaths()`, `buildTags()`, `envelopeSchema()` |
| `export.go` | `ExportSpec()`, `ExportSpecToFile()` |
| `bridge.go` | `ToolDescriptor`, `ToolBridge`, `NewToolBridge()`, `Add()`, `Describe()`, `Tools()` |
| `codegen.go` | `SDKGenerator`, `Generate()`, `Available()`, `SupportedLanguages()` |
| `cmd/api/` | CLI subcommands: `core api spec` and `core api sdk` |

---

## Dependencies

### Direct

| Module | Role |
|--------|------|
| `github.com/gin-gonic/gin` | HTTP router and middleware engine |
| `github.com/gin-contrib/cors` | CORS policy middleware |
| `github.com/gin-contrib/secure` | Security headers (HSTS, X-Frame-Options, nosniff) |
| `github.com/gin-contrib/gzip` | Gzip response compression |
| `github.com/gin-contrib/slog` | Structured request logging via `log/slog` |
| `github.com/gin-contrib/timeout` | Per-request deadline enforcement |
| `github.com/gin-contrib/static` | Static file serving |
| `github.com/gin-contrib/sessions` | Cookie-backed server sessions |
| `github.com/gin-contrib/authz` | Casbin policy-based authorisation |
| `github.com/gin-contrib/httpsign` | HTTP Signatures verification |
| `github.com/gin-contrib/location/v2` | Reverse proxy header detection |
| `github.com/gin-contrib/pprof` | Go profiling endpoints |
| `github.com/gin-contrib/expvar` | Runtime metrics endpoint |
| `github.com/casbin/casbin/v2` | Policy-based access control engine |
| `github.com/coreos/go-oidc/v3` | OIDC provider discovery and JWT validation |
| `github.com/andybalholm/brotli` | Brotli compression |
| `github.com/gorilla/websocket` | WebSocket upgrade support |
| `github.com/swaggo/gin-swagger` | Swagger UI handler |
| `github.com/swaggo/files` | Swagger UI static assets |
| `github.com/swaggo/swag` | Swagger spec registry |
| `github.com/99designs/gqlgen` | GraphQL schema execution (gqlgen) |
| `go.opentelemetry.io/otel` | OpenTelemetry tracing SDK |
| `go.opentelemetry.io/contrib/.../otelgin` | OpenTelemetry Gin instrumentation |
| `golang.org/x/text` | BCP 47 language tag matching |
| `gopkg.in/yaml.v3` | YAML export of OpenAPI specs |
| `forge.lthn.ai/core/cli` | CLI command registration (for `cmd/api/` subcommands) |

### Ecosystem position

go-api sits at the base of the Lethean HTTP stack. It has no imports from other Lethean
ecosystem modules (beyond `core/cli` for the CLI subcommands). Other packages import go-api
to expose their functionality as REST endpoints:

```
Application main / Core CLI
    |
    v
  go-api Engine                 <-- this module
    |         |         |
    |         |         +-- OpenAPI spec --> SDKGenerator --> openapi-generator-cli
    |         +-- ToolBridge --> go-ai / go-ml / go-rag route groups
    +-- RouteGroups ----------> any package implementing RouteGroup
```

---

## Further Reading

- [Architecture](architecture.md) -- internals, key types, data flow, middleware stack
- [Development](development.md) -- building, testing, contributing, coding standards
