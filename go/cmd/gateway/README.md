# Core Gateway

`cmd/gateway` is the unified Core API mount-point for provider packages.

## Build

```bash
go build -o core-gateway ./cmd/gateway/
```

## Configuration

`CORE_GATEWAY_BIND` sets the listen address. The default is `0.0.0.0:8080`.

`CORE_GATEWAY_ENABLE` optionally limits mounted providers. It is a comma-separated list of provider names:

```bash
CORE_GATEWAY_ENABLE=scm,process ./core-gateway
```

When unset, all gateway providers are attempted. If one provider panics or returns nil during construction, the gateway logs the failure and continues mounting the remaining providers.

## Mounted Providers

| Provider | Prefix | Source |
| --- | --- | --- |
| `brain` | `/api/brain` | `core/agent/pkg/brain` |
| `brain-mcp` | `/api/mcp/brain` | `core/mcp/pkg/mcp/brain` |
| `scm` | `/scm` | `go-scm/pkg/api` |
| `process` | `/api/process` | `go-process/pkg/api` |
| `build` | `/api/v1/build` | `go-build/pkg/api` |
| `miner` | root (`/miners`, `/profiles`, `/history`) | `go-miner/pkg/api` |
| `proxy` | `/1` | `go-proxy/api` |

## Operator Notes

To disable a provider, set `CORE_GATEWAY_ENABLE` to only the providers you want to mount:

```bash
CORE_GATEWAY_ENABLE=brain,process,build ./core-gateway
```

To add a provider, add one entry to `gatewayProviderSpecs()` in `main.go`. Keep the constructor inside that entry so the dependency is easy to remove for subset builds.
