<!-- SPDX-License-Identifier: EUPL-1.2 -->

# dappco.re/go/api

> Gin-based HTTP framework + multi-language REST gateway for the Core ecosystem.

[![CI](https://github.com/dappcore/api/actions/workflows/ci.yml/badge.svg?branch=dev)](https://github.com/dappcore/api/actions/workflows/ci.yml)
[![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=dappcore_api&metric=alert_status)](https://sonarcloud.io/dashboard?id=dappcore_api)
[![Coverage](https://codecov.io/gh/dappcore/api/branch/dev/graph/badge.svg)](https://codecov.io/gh/dappcore/api)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=dappcore_api&metric=security_rating)](https://sonarcloud.io/dashboard?id=dappcore_api)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=dappcore_api&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=dappcore_api)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=dappcore_api&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=dappcore_api)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=dappcore_api&metric=code_smells)](https://sonarcloud.io/dashboard?id=dappcore_api)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=dappcore_api&metric=ncloc)](https://sonarcloud.io/dashboard?id=dappcore_api)
[![Go Reference](https://pkg.go.dev/badge/dappco.re/go/api.svg)](https://pkg.go.dev/dappco.re/go/api)
[![License: EUPL-1.2](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](https://eupl.eu/1.2/en/)

## Overview

`dappco.re/go/api` is the Gin-based HTTP framework used by the Core Go
ecosystem. It provides a small `Engine` type, option-driven middleware
configuration, route group mounting, response envelopes, OpenAPI 3.1
generation, SDK export/codegen helpers, SSE and WebSocket wiring, GraphQL
hosting, Authentik identity middleware, and the `core api` CLI commands.

The package is a library first. Applications construct an engine, register
one or more `RouteGroup` implementations, then either call `Serve(ctx)` or use
`Handler()` with their own server.

```go
engine, err := api.New(
    api.WithAddr(":8080"),
    api.WithRequestID(),
    api.WithResponseMeta(),
    api.WithSwagger("Core API", "Core service endpoints", "1.0.0"),
)
if err != nil {
    return err
}
engine.Register(myRoutes)
return engine.Serve(ctx)
```

## Repository Layout

```
api/
├── go/                       Go module — module path: dappco.re/go/api
│   ├── api.go, options.go    HTTP engine surface
│   ├── cmd/api/              core api spec + sdk CLI subcommands
│   ├── cmd/gateway/          runnable gateway, mounts Core providers
│   ├── pkg/provider/         provider discovery + proxy
│   └── pkg/stream/           SSE + WebSocket route group
├── php/                      Laravel Core API package (REST middleware,
│                             webhooks, OpenAPI, rate limiting)
├── docs/                     Engine docs
├── sdk-config/               Multi-language SDK generator config
├── go.work + external/       Dev workspace mode (see CLAUDE.md)
└── .woodpecker.yml + .github/workflows/  CI (internal + public)
```

Cross-language symmetry target: `dappco.re/<lang>/api/<feature>` ↔
`api/<lang>/<feature>` (Go today, PHP today, TS + Py later).

## Local Verification

Run the repository with the workspace disabled when checking this module
in isolation:

```bash
cd go
GOWORK=off go mod tidy
GOWORK=off go vet ./...
GOWORK=off go test -count=1 ./...
gofmt -l .
bash /Users/snider/Code/core/go/tests/cli/v090-upgrade/audit.sh .
```

The audit is part of the development contract. Public symbols need
sibling triplet tests and examples, Core wrappers are used instead of
banned standard library imports, and generated AX7 dump files are not
accepted.

## CI

- **Internal** (homelab, full sonar.lthn.sh detail): Woodpecker pipeline
  defined in `.woodpecker.yml` — runs lint, test with race + coverage,
  and pushes results to `sonar.lthn.sh`.
- **Public** (mirror on github.com, badge surface): GitHub Actions
  workflow at `.github/workflows/ci.yml` — runs the same shape and
  pushes coverage to Codecov + analysis to SonarCloud.

## Branch Model

- `dev` — active development. All Cladius / codex lane work lands here
  first.
- `main` — squash-stable. Promotion happens via the squash-and-push gate
  on the public mirror only.

## Licence

EUPL-1.2 — see [LICENCE](LICENCE).

## Authorship

Maintained by Cladius (Snider's in-house Opus persona) via the
`agent/cladius` workspace at `forge.lthn.sh/agent/cladius`. Most
substantive commits land via the codex lane pattern documented in
`factory/`.
