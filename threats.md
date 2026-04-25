## SSRF audit transport_client.go:371 doHTTPClientRequest

Status: Fixed

Commit anchor: 295c0ff

### Scope

Finding source: gosec G704 taint-analysis SSRF at `transport_client.go` `client.Do(req)`.

Audited choke point: `doHTTPClientRequest(client, req)`, currently reached by:

| Call site | Request URL source | User-controlled? | Notes |
| --- | --- | --- | --- |
| `transport_client.go:229` `SSEClient.Connect` | `NewSSEClient(rawURL)` -> `c.URL` -> `http.NewRequestWithContext(ctx, GET, rawURL, nil)` | YES when a route binds request payload/config directly into `rawURL`; otherwise caller-trusted | Constructor accepts any caller string. No host allowlist is applied before construction. |
| `client.go:342` `OpenAPIClient.Call` | `c.buildURL(op, params)` from `WithBaseURL(baseURL)` or first absolute `servers[].url` loaded by `WithSpecReader`/`WithSpec` | YES if base URL or supplied OpenAPI spec is attacker-controlled; NO for operator-owned specs/config | Path/query/header/body params are encoded and do not set the URL host; host comes from base URL or spec server metadata. |

No other package call sites of `doHTTPClientRequest` were found.

### Adjacent URL Acceptors

`Webhook` destinations do not flow into `doHTTPClientRequest`; they use `ValidateWebhookURL`, which rejects non-HTTP(S), credentialed URLs, lookup failures, and private/loopback/link-local/reserved targets.

`SDKGenerator`/codegen does not perform outbound HTTP through this path. Its `SpecPath` is a filesystem path passed to `openapi-generator-cli`; generator names are allowlisted separately.

No `TransformerIn`/`TransformerOut` Go types or direct transformer URL plumbing were found in this package.

### Allowlist Presence

No explicit business host allowlist such as `config.AllowedHosts` was found before `doHTTPClientRequest`.

The active protection is the centralized `validateOutboundURL` mechanism at the `doHTTPClientRequest` choke point. It uses a deny-by-default scheme allowlist (`http`, `https`) and rejects blocked hosts/IP classes before dispatch.

### Local and Metadata Blocking

Present for direct requests:

- Literal metadata hosts including `169.254.169.254`, `metadata.google.internal`, `metadata.googleapis.com`, `metadata.azure.com`, `fd00:ec2::254`, and `100.100.100.200`.
- Literal IPs in loopback, private RFC1918/RFC4193, link-local, unspecified, and multicast ranges.
- Hostnames resolving to blocked IPs at request time, covering DNS-rebinding-style private resolution.

Required ranges are covered:

- `169.254.0.0/16`: blocked by `net.IP.IsLinkLocalUnicast`.
- `127.0.0.0/8`: blocked by `net.IP.IsLoopback`.
- `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`: blocked by `net.IP.IsPrivate`.
- `fc00::/7`: blocked by `net.IP.IsPrivate`.
- `::1/128`: blocked by `net.IP.IsLoopback`.
- `fe80::/10`: blocked by `net.IP.IsLinkLocalUnicast`.

### Finding

Direct local/metadata SSRF was already blocked at the initial request URL. A redirect-based bypass was reachable: `net/http` follows 3xx responses inside `Client.Do`, so a public first hop could redirect to metadata/local addresses without re-entering `doHTTPClientRequest`.

Fix applied in `transport_client.go`: `doHTTPClientRequest` now executes through a shallow copy of the caller's `http.Client` with a redirect guard. The guard preserves caller-supplied `CheckRedirect` behavior, preserves Go's default 10-redirect limit when no custom policy is set, and validates each redirect target with `validateOutboundURL` before the redirect is followed.

Fix coverage added in `transport_client_test.go`: a public initial URL returning `Location: http://169.254.169.254/...` is blocked with `errOutboundURLBlocked`, and the follow-up request is not issued.

### Severity Verdict

Before fix: High for attacker-controlled upstream URLs because metadata/local SSRF was reachable through redirects even though direct metadata/local URLs were blocked.

After fix: Low for local/metadata SSRF in this choke point. Direct and redirect targets are validated against the centralized block policy. Residual note: arbitrary public-host egress is still allowed by design because there is no configured business-host allowlist; callers that bind attacker input into upstream URL fields must provide trusted host policy at the application/config layer if public egress itself is out of scope.

---

## G204 codegen.go:97 audit (Cerberus #322)

- Sink: `SDKGenerator.Generate` builds `args := g.buildArgs(...)` and runs `exec.CommandContext(ctx, "openapi-generator-cli", args...)`. The command name is a string literal; the variable at the sink is the argument vector.
- Trust chain: the only production caller found is `cmd/api/cmd_sdk.go:sdkAction`. CLI options populate `--lang`, `--output`, `--spec`, and `--package`; when `--spec` is omitted, the spec path is a local temporary file generated from registered route metadata.
- Validation: `language` is trimmed and mapped through the closed `supportedLanguages` allowlist; `PackageName` is constrained by `packageNameRe`; `Available()` resolves the literal `openapi-generator-cli` with `exec.LookPath`.
- API reachability: repo grep found no `TransformerIn`, request body, query parameter, or HTTP route path reaching `SDKGenerator.Generate`; only CLI code, tests, and docs reference it.
- Severity verdict: OPERATOR-ONLY / low. Existing `#nosec G204` in `codegen.go` is justified for the current trust chain. Reassess if a future API endpoint binds request fields to `SDKGenerator`.
