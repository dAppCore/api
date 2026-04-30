---
title: Development Guide
description: How to build, test, and contribute to the go-api REST framework -- prerequisites, test patterns, extension guides, and coding standards.
---

<!-- SPDX-License-Identifier: EUPL-1.2 -->

# Development Guide

This guide covers everything needed to build, test, extend, and contribute to go-api.

**Module path:** `dappco.re/go/api`
**Licence:** EUPL-1.2
**Language:** Go 1.26

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Building](#building)
3. [Testing](#testing)
4. [Test Patterns](#test-patterns)
5. [Adding a New With*() Option](#adding-a-new-with-option)
6. [Adding a RouteGroup](#adding-a-routegroup)
7. [Adding a DescribableGroup](#adding-a-describablegroup)
8. [Coding Standards](#coding-standards)
9. [Commit Guidelines](#commit-guidelines)

---

## Prerequisites

### Go toolchain

Go 1.26 or later is required. Verify the installed version:

```bash
go version
```

### Minimal dependencies

go-api depends on the provider modules that the gateway wires (`process`, `scm`, `miner`,
`proxy`, and `ws`) plus the core helper modules it uses directly. The `cmd/api/` CLI lives in
its own nested module and imports `dappco.re/go/cli`. There are no `replace` directives in the
root module.

If working within the Go workspace at `~/Code/go.work`, the workspace `use` directive handles
local module resolution automatically.

---

## Building

go-api is a library module with no `main` package. Build all packages to verify that everything
compiles cleanly:

```bash
go build ./...
```

Vet for suspicious constructs:

```bash
go vet ./...
```

Neither command produces a binary. If you need a runnable server for manual testing, create a
temporary `main.go` that imports go-api and calls `engine.Serve()`.

---

## Testing

### Run all tests

```bash
go test ./...
```

### Run a single test by name

```bash
go test -run TestName ./...
```

The `-run` flag accepts a regular expression:

```bash
go test -run TestToolBridge ./...
go test -run TestSpecBuilder_Good ./...
go test -run "Test.*_Bad" ./...
```

### Verbose output

```bash
go test -v ./...
```

### Race detector

Always run with `-race` before opening a pull request. The middleware layer uses concurrency
(SSE broker, cache store, Brotli writer pool), and the race detector catches data races
reliably:

```bash
go test -race ./...
```

Note: The repository includes `race_test.go` and `norace_test.go` build-tag files that control
which tests run under the race detector.

### Live Authentik integration tests

`authentik_integration_test.go` contains tests that require a live Authentik instance. These
are skipped automatically unless the `AUTHENTIK_ISSUER` and `AUTHENTIK_CLIENT_ID` environment
variables are set. They do not run in standard CI.

To run them locally:

```bash
AUTHENTIK_ISSUER=https://auth.example.com/application/o/my-app/ \
AUTHENTIK_CLIENT_ID=my-client-id \
go test -run TestAuthentik_Integration ./...
```

---

## Test Patterns

### Naming convention

All test functions follow the `_Good`, `_Bad`, `_Ugly` suffix pattern:

| Suffix  | Purpose |
|---------|---------|
| `_Good` | Happy path -- the input is valid and the operation succeeds |
| `_Bad`  | Expected error conditions -- invalid input, missing config, wrong state |
| `_Ugly` | Panics and extreme edge cases -- nil receivers, resource exhaustion, concurrent access |

Examples from the codebase:

```
TestNew_Good
TestNew_Good_WithAddr
TestHandler_Bad_NotFound
TestSDKGenerator_Bad_UnsupportedLanguage
TestSpecBuilder_Good_SingleDescribableGroup
```

### Engine test helpers

Tests that need a running HTTP server use `httptest.NewRecorder()` or `httptest.NewServer()`.
Build the engine handler directly rather than calling `Serve()`:

```go
gin.SetMode(gin.TestMode)
engine, _ := api.New(api.WithBearerAuth("tok"))
engine.Register(&myGroup{})
handler := engine.Handler()

req := httptest.NewRequest(http.MethodGet, "/health", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

if rec.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", rec.Code)
}
```

### SSE tests

SSE handler tests open a real HTTP connection with `httptest.NewServer()` and read the
`text/event-stream` response body line by line. Publish from a goroutine and use a deadline
to avoid hanging indefinitely:

```go
srv := httptest.NewServer(engine.Handler())
defer srv.Close()

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/events", nil)
resp, _ := http.DefaultClient.Do(req)
// read lines from resp.Body
```

### Cache tests

Cache tests must use `httptest.NewServer()` rather than a recorder because the cache middleware
needs a proper response cycle to capture and replay responses. Verify the `X-Cache: HIT` header
on the second request to the same URL.

### Authentik tests

Authentik middleware tests use raw `httptest.NewRequest()` with `X-authentik-*` headers set
directly. No live Authentik instance is required for unit tests -- the permissive middleware is
exercised entirely through header injection and context assertions.

---

## Adding a New With*() Option

`With*()` options are the primary extension point. All options follow an identical pattern.

### Step 1: Add the option function

Open `options.go` and add a new exported function that returns an `Option`:

```go
// WithRateLimit adds request rate limiting middleware.
// Requests exceeding limit per second per IP are rejected with 429 Too Many Requests.
func WithRateLimit(limit int) Option {
    return func(e *Engine) {
        e.middlewares = append(e.middlewares, rateLimitMiddleware(limit))
    }
}
```

If the option stores state on `Engine` (like `swaggerEnabled` or `sseBroker`), add the
corresponding field to the `Engine` struct in `api.go` and reference it in `build()`.

### Step 2: Implement the middleware

If the option wraps a `gin-contrib` package, follow the existing pattern in `options.go`
(inline). For options with non-trivial logic, create a dedicated file (e.g. `ratelimit.go`).
Every new source file must begin with the EUPL-1.2 SPDX identifier:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import "github.com/gin-gonic/gin"

func rateLimitMiddleware(limit int) gin.HandlerFunc {
    // implementation
}
```

### Step 3: Add the dependency to go.mod

If the option relies on a new external package:

```bash
go get github.com/example/ratelimit
go mod tidy
```

### Step 4: Write tests

Create a test file (e.g. `ratelimit_test.go`) following the `_Good`/`_Bad`/`_Ugly` naming
convention. Test with `httptest` rather than calling `Serve()`.

### Step 5: Update the build path

If the option adds a new built-in HTTP endpoint (like WebSocket at `/ws` or SSE at `/events`),
add it to the `build()` method in `api.go` after the GraphQL block but before Swagger.

---

## Adding a RouteGroup

`RouteGroup` is the standard way for subsystem packages to contribute REST endpoints.

### Minimum implementation

```go
// SPDX-License-Identifier: EUPL-1.2

package mypackage

import (
    "net/http"

    api "dappco.re/go/api"
    "github.com/gin-gonic/gin"
)

type Routes struct {
    service *Service
}

func NewRoutes(svc *Service) *Routes {
    return &Routes{service: svc}
}

func (r *Routes) Name() string     { return "mypackage" }
func (r *Routes) BasePath() string { return "/v1/mypackage" }

func (r *Routes) RegisterRoutes(rg *gin.RouterGroup) {
    rg.GET("/items", r.listItems)
    rg.POST("/items", r.createItem)
}

func (r *Routes) listItems(c *gin.Context) {
    items, err := r.service.List(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, api.Fail("internal", err.Error()))
        return
    }
    c.JSON(http.StatusOK, api.OK(items))
}
```

Register with the engine:

```go
engine.Register(mypackage.NewRoutes(svc))
```

### Adding StreamGroup

If the group publishes WebSocket channels, implement `StreamGroup` as well:

```go
func (r *Routes) Channels() []string {
    return []string{"mypackage.items.updated"}
}
```

---

## Adding a DescribableGroup

`DescribableGroup` extends `RouteGroup` with OpenAPI metadata. Implementing it ensures the
group's endpoints appear in the generated spec and Swagger UI.

Add a `Describe()` method that returns a slice of `RouteDescription`:

```go
func (r *Routes) Describe() []api.RouteDescription {
    return []api.RouteDescription{
        {
            Method:  "GET",
            Path:    "/items",
            Summary: "List items",
            Tags:    []string{"mypackage"},
            Response: map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "object"},
            },
        },
        {
            Method:  "POST",
            Path:    "/items",
            Summary: "Create an item",
            Tags:    []string{"mypackage"},
            RequestBody: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "name": map[string]any{"type": "string"},
                },
                "required": []string{"name"},
            },
            Response: map[string]any{"type": "object"},
        },
    }
}
```

Paths in `RouteDescription` are relative to `BasePath()`. The `SpecBuilder` concatenates them
when building the full OpenAPI path.

---

## Coding Standards

### Language

Use **UK English** in all comments, documentation, log messages, and user-facing strings:
`colour`, `organisation`, `centre`, `initialise`, `licence` (noun), `license` (verb),
`unauthorised`, `authorisation`.

### Licence header

Every new Go source file must carry the EUPL-1.2 SPDX identifier as the first line:

```go
// SPDX-License-Identifier: EUPL-1.2

package api
```

### Error handling

- Always return errors rather than panicking.
- Wrap errors with context: `fmt.Errorf("component.Operation: what failed: %w", err)`.
- Do not discard errors with `_` unless the operation is genuinely fire-and-forget and the
  reason is documented with a comment.
- Log errors at the point of handling, not at the point of wrapping.

### Response envelope

All HTTP handlers must use the `api.OK()`, `api.Fail()`, `api.FailWithDetails()`, or
`api.Paginated()` helpers rather than constructing `Response[T]` directly. This ensures the
envelope structure is consistent across all route groups.

### Test naming

- Function names: `Test{Type}_{Suffix}_{Description}` where `{Suffix}` is `Good`, `Bad`,
  or `Ugly`.
- Helper constructors: `newTest{Type}(t *testing.T, ...) *Type`.
- Always call `t.Helper()` at the top of every test helper function.

### Formatting

The codebase uses `gofmt` defaults. Run before committing:

```bash
gofmt -l -w .
```

### Middleware conventions

- Every `With*()` function must append to `e.middlewares`, not modify Gin routes directly.
  Routes are only registered in `build()`.
- Options that require `Engine` struct fields (like `swaggerEnabled` or `sseBroker`) must be
  readable by `build()`, not set inside a closure without a backing field.
- Middleware that exposes sensitive data (`WithPprof`, `WithExpvar`) must carry a `// WARNING:`
  comment in the godoc directing users away from production exposure without authentication.

---

## Commit Guidelines

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

Body explaining what changed and why.

Co-Authored-By: Virgil <virgil@lethean.io>
```

Types in use across the repository: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`.

Example:

```
feat(api): add WithRateLimit per-IP rate limiting middleware

Adds configurable per-IP rate limiting using a token-bucket algorithm.
Requests exceeding the limit per second are rejected with 429 Too Many
Requests and a standard Fail() error envelope.

Co-Authored-By: Virgil <virgil@lethean.io>
```
