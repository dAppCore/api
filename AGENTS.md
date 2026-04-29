<!-- SPDX-License-Identifier: EUPL-1.2 -->

# Agent Notes

This repository is the Go Core API framework. Treat it as infrastructure: keep
changes narrow, preserve the public `RouteGroup` and `Engine` contracts, and
verify the whole module before handing work back.

## Code Map

- `api.go` owns `Engine`, route group registration, handler construction, and
  graceful serving.
- `options.go` contains public `With*` options. New middleware should enter
  through this file unless it is strictly internal.
- `response.go`, `middleware.go`, `authentik.go`, `sse.go`, `websocket.go`,
  `graphql.go`, `swagger.go`, `openapi.go`, `export.go`, and `codegen.go`
  implement the user-facing HTTP features.
- `bridge.go` maps `ToolDescriptor` values to REST endpoints and OpenAPI
  descriptions.
- `cmd/api` wires Core CLI actions for spec export and SDK generation.
- `cmd/gateway` builds the provider gateway binary.
- `pkg/provider` contains provider discovery, registry, and reverse proxy
  support.
- `pkg/stream` contains declarative stream route groups.

## Compliance Rules

Follow the v0.9.0 Core compliance shape. Use `dappco.re/go` wrappers for JSON,
errors, formatting, strings, buffers, filesystem, process, and environment
helpers whenever a wrapper exists. Do not add files named `ax7*.go`, versioned
test files, or monolithic compliance dumps.

For every production source file with public symbols, keep tests and examples
beside that file. Test names use `Test<File>_<Symbol>_Good`,
`Test<File>_<Symbol>_Bad`, and `Test<File>_<Symbol>_Ugly`. Examples use
`Example<Symbol>` or a valid lowercase suffix variant and print through Core
`Println`.

## Before Stopping

Use the exact repository gate, with `GOWORK=off` for Go commands:

```bash
GOWORK=off go mod tidy
GOWORK=off go vet ./...
GOWORK=off go test -count=1 ./...
gofmt -l .
bash /Users/snider/Code/core/go/tests/cli/v090-upgrade/audit.sh .
```

If the sandbox cannot write the default Go build cache, set `GOCACHE` to a
temporary directory while running the same commands.
