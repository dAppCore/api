# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Core API is the REST framework for the Lethean ecosystem, providing both a **Go HTTP engine** (Gin-based, with OpenAPI generation, WebSocket/SSE, ToolBridge) and a **PHP Laravel package** (rate limiting, webhooks, API key management, OpenAPI documentation). Both halves serve the same purpose in their respective stacks.

Module: `dappco.re/go/api` | Package: `dappco.re/php/service` | Licence: EUPL-1.2

## Repo Layout

```
core/api/
├── go.work                            ← workspace root (one level above the module)
├── external/<dappco.re-go-dep>/       ← git submodules tracking dev branches on github
├── go/                                ← Go module root (module dappco.re/go/api)
├── php/                               ← PHP package
├── docs/                              ← engine docs (symlinked from go/)
├── sdk-config/                        ← cross-language SDK gen configs
└── scripts/                           ← cross-language build helpers
```

Cross-language symmetry target: `dappco.re/<lang>/api/<feature>` ↔ `core/api/<lang>/<feature>` (Go today, PHP today, TS+Py later).

## Go Resolution Modes

Two ways the same `go/go.mod` resolves dappco.re/go/* deps:

| Mode | When | What runs |
|------|------|-----------|
| **Workspace ON** (default for devs) | `go build` / `go test` from any subdir of `core/api/` | Walks up to `go.work`, uses local `external/<x>` checkouts at submodule pin (typically dev tip). Fast iteration; finds upstream bugs early. |
| **`GOWORK=off`** | Woodpecker CI | Pure go.mod, fetches the pinned tag from the proxy. Reproducible builds. No replace directives — 0% policy intact. |

### Working with submodules

```bash
git clone --recursive https://github.com/dappcore/api.git    # full dev workspace
git submodule update --init --recursive                      # if cloned without --recursive

# Bump a single dep to its current dev tip
git submodule update --remote external/go-process

# See latest tag in a dep
( cd external/go-process && git tag --sort=-v:refname | head )

# When ready to bump the api go.mod to that dep's new tag
( cd go && go get dappco.re/go/process@v0.11.0 && go mod tidy )
```

### Workspace mode caveat

Workspace mode validates each `external/<x>/go.sum` against current proxy bits. If an upstream repo has a stale or missing go.sum entry, the build errors with `verifying ...: checksum mismatch`. Two ways to handle:

1. **Fix upstream**: `cd external/<x> && go mod tidy`, commit + push to that repo's dev, then `git submodule update --remote external/<x>` here.
2. **CI mode locally**: `GOWORK=off go test ./go/...` — uses the api repo's pinned hashes.

This is the "find bugs as they roll out" payoff: workspace mode surfaces stale-sum issues across the whole tree.

## Build and Test Commands

### Go

```bash
core build                          # Build binary (if cmd/ has main)
go build ./...                      # Build library

core go test                        # Run all Go tests
core go test --run TestName         # Run a single test
core go cov                         # Coverage report
core go cov --open                  # Open HTML coverage in browser
core go qa                          # Format + vet + lint + test
core go qa full                     # Also race detector, vuln scan, security audit
core go fmt                         # gofmt
core go lint                        # golangci-lint
core go vet                         # go vet
```

### PHP (from repo root)

```bash
composer test                                    # Run all PHP tests (Pest)
composer test -- --filter=ApiKey                 # Single test
composer lint                                    # Laravel Pint (PSR-12)
./vendor/bin/pint --dirty                        # Format changed files
```

Tests live in `php/src/Api/Tests/Feature/` (in-source) and `php/tests/` (standalone).

## Architecture

### Go Engine (root-level .go files)

`Engine` is the central type, configured via functional `Option` functions passed to `New()`:

```go
engine, _ := api.New(api.WithAddr(":8080"), api.WithCORS("*"), api.WithSwagger(...))
engine.Register(myRouteGroup)
engine.Serve(ctx)
```

**Extension interfaces** (`group.go`):
- `RouteGroup` — minimum: `Name()`, `BasePath()`, `RegisterRoutes(*gin.RouterGroup)`
- `StreamGroup` — optional: `Channels() []string` for WebSocket
- `DescribableGroup` — extends RouteGroup with `Describe() []RouteDescription` for OpenAPI

**ToolBridge** (`bridge.go`): Converts `ToolDescriptor` structs into `POST /{tool_name}` REST endpoints with auto-generated OpenAPI paths.

**Authentication** (`authentik.go`): Authentik OIDC integration + static bearer token. Permissive middleware with `RequireAuth()` / `RequireGroup()` guards.

**OpenAPI** (`openapi.go`, `export.go`, `codegen.go`): `SpecBuilder.Build()` generates OpenAPI 3.1 JSON. `SDKGenerator` wraps openapi-generator-cli for 11 languages.

**CLI** (`cmd/api/`): Registers `core api spec` and `core api sdk` commands.

### PHP Package (`php/`)

Three namespace roots:

| Namespace | Path | Role |
|-----------|------|------|
| `Core\Front\Api` | `php/src/Front/Api/` | API frontage — middleware, versioning, auto-discovered provider |
| `Core\Api` | `php/src/Api/` | Backend — auth, scopes, models, webhooks, OpenAPI docs |
| `Core\Website\Api` | `php/src/Website/Api/` | Documentation UI — controllers, Blade views, web routes |

Boot chain: `Front\Api\Boot` (auto-discovered) fires `ApiRoutesRegistering` -> `Api\Boot` registers middleware and routes.

Key services: `WebhookService`, `RateLimitService`, `IpRestrictionService`, `OpenApiBuilder`, `ApiKeyService`.

## Conventions

- **UK English** in all user-facing strings and docs (colour, organisation, unauthorised)
- **SPDX headers** in Go files: `// SPDX-License-Identifier: EUPL-1.2`
- **`declare(strict_types=1);`** in every PHP file
- **Full type hints** on all PHP parameters and return types
- **Pest syntax** for PHP tests (not PHPUnit)
- **Flux Pro** components in Livewire views; **Font Awesome** icons
- **Conventional commits**: `type(scope): description`
- **Co-Author**: `Co-Authored-By: Virgil <virgil@lethean.io>`
- Go test names use `_Good` / `_Bad` / `_Ugly` suffixes

## Key Dependencies

| Go module | Role |
|-----------|------|
| `dappco.re/go/cli` | CLI command registration for the nested `cmd/api` module |
| `github.com/gin-gonic/gin` | HTTP router |
| `github.com/casbin/casbin/v2` | Authorisation policies |
| `github.com/coreos/go-oidc/v3` | OIDC / Authentik |
| `go.opentelemetry.io/otel` | OpenTelemetry tracing |

PHP: `lthn/php` (Core framework), Laravel 12, `symfony/yaml`.

Go workspace: this module is part of `~/Code/go.work`. Requires Go 1.26+, PHP 8.2+.
