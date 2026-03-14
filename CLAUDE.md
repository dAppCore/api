# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Core API is the REST framework for the Lethean ecosystem, providing both a **Go HTTP engine** (Gin-based, with OpenAPI generation, WebSocket/SSE, ToolBridge) and a **PHP Laravel package** (rate limiting, webhooks, API key management, OpenAPI documentation). Both halves serve the same purpose in their respective stacks.

Module: `forge.lthn.ai/core/api` | Package: `lthn/api` | Licence: EUPL-1.2

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

Tests live in `src/php/src/Api/Tests/Feature/` (in-source) and `src/php/tests/` (standalone).

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

### PHP Package (`src/php/`)

Three namespace roots:

| Namespace | Path | Role |
|-----------|------|------|
| `Core\Front\Api` | `src/php/src/Front/Api/` | API frontage — middleware, versioning, auto-discovered provider |
| `Core\Api` | `src/php/src/Api/` | Backend — auth, scopes, models, webhooks, OpenAPI docs |
| `Core\Website\Api` | `src/php/src/Website/Api/` | Documentation UI — controllers, Blade views, web routes |

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
| `forge.lthn.ai/core/cli` | CLI command registration |
| `github.com/gin-gonic/gin` | HTTP router |
| `github.com/casbin/casbin/v2` | Authorisation policies |
| `github.com/coreos/go-oidc/v3` | OIDC / Authentik |
| `go.opentelemetry.io/otel` | OpenTelemetry tracing |

PHP: `lthn/php` (Core framework), Laravel 12, `symfony/yaml`.

Go workspace: this module is part of `~/Code/go.work`. Requires Go 1.26+, PHP 8.2+.
