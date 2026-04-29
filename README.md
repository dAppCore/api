<!-- SPDX-License-Identifier: EUPL-1.2 -->

# dappco.re/go/api

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

## Repository Shape

The root package contains the HTTP engine and most middleware. `cmd/api`
registers Core CLI subcommands for OpenAPI export and SDK generation.
`cmd/gateway` is a runnable gateway that mounts selected Core provider
packages behind one API engine. `pkg/provider` discovers and proxies provider
manifests. `pkg/stream` contains a declarative stream route group for SSE and
WebSocket handlers.

## Local Verification

Run the repository with the workspace disabled when checking this module in
isolation:

```bash
GOWORK=off go mod tidy
GOWORK=off go vet ./...
GOWORK=off go test -count=1 ./...
gofmt -l .
bash /Users/snider/Code/core/go/tests/cli/v090-upgrade/audit.sh .
```

The audit is part of the development contract. Public symbols need sibling
triplet tests and examples, Core wrappers are used instead of banned standard
library imports, and generated AX7 dump files are not accepted.
