# Upstream Router (`WithUpstreamRouter`) — Design

- **Date:** 2026-06-06
- **Status:** Design — approved, pending implementation plan
- **Module:** `dappco.re/go/api` (`core/api/go`)
- **Author:** Snider + Cladius (brainstorming)
- **Related:** `RFC.md` §11 (chat completions), `RFC.providers.md` (gateway), `transformer*.go` (translation), `ssrf_guard.go` (outbound policy)

---

## 1. Context & Problem

`core/api` has a list-of-endpoints problem: consumers hold a set of upstream model
endpoints (local Ollama, LAN GPU boxes, hosted inference) and have **no first-class
way to load-balance or route across them by a selector key** (typically the `model`
name, but any value).

What already exists and is reused, not rebuilt:

- **Translation layer** — `TransformerIn[I,O]` / `TransformerOut[I,O]`, chainable
  pipelines, `FieldRenamer`, schema validation (`transformer.go`,
  `transformer_in.go`, `transformer_out.go`).
- **Single-target outbound** — `OpenAPIClient` (one base URL), `SSEClient`,
  `WebSocketClient`, all funnelled through the SSRF-guarded `doHTTPClientRequest`
  (`transport_client.go`).
- **Selector pattern, wrong target** — `ModelResolver` maps `name → backend` but
  resolves to **local in-process `inference.TextModel`**, not remote HTTP, and is
  loopback-only (`chat_completions.go`).
- **Rate limiting** — `go-ratelimit` (separate module) and `WithRateLimit`.

`go-proxy` is **not** reusable here — it is a stratum mining proxy
(workers/miners/shares), not an HTTP reverse proxy.

The missing piece is a **selector-keyed reverse proxy over a pool of HTTP upstreams**,
composing with the existing translators so any consuming package gets transparent
routing: accept a foreign request shape → route by key → translate → dispatch →
translate the response back.

## 2. Goals / Non-Goals

**Goals**
- An `api.Option` (`WithUpstreamRouter`) that mounts a router on the Engine and
  inherits its auth/CORS/rate-limit/tracing middleware — drop-in for any consumer.
- Route by a pluggable selector key; default reads the JSON `model` field.
- Load-balance within a per-key pool (weighted round-robin) with passive failover.
- Runtime-mutable pool table (hot reconfigure without restart).
- A decision hook to inspect the payload and override/reject routing.
- Stream SSE / `stream:true` responses through untouched; buffer + translate
  non-streaming responses.

**Non-Goals (v1)**
- Active health-check goroutines (failover is passive/inline).
- Sticky/consistent-hash routing (noted future extension).
- Direct upstream selection from the hook bypassing the registry (key-only in v1).
- Per-chunk transformation of live streams (transformers apply to buffered responses
  only).
- Mid-stream failover (impossible once response bytes are flowing; documented).

## 3. Settled Decisions

| Fork | Decision |
|------|----------|
| Selector source | Pluggable `Selector func`; **default reads JSON body `model`** |
| Streaming | **Hybrid** — stream-through for `text/event-stream`, buffer otherwise |
| LB strategy | **Weighted round-robin + passive failover** (cooldown on failure) |
| Routing seam | **Decision hook + runtime-mutable pool registry** |
| Proxy core | stdlib `net/http/httputil.ReverseProxy` + custom `RoundTripper` that owns selection/failover |
| SSRF | **Block-by-default at registration** — reject loopback/private/link-local/reserved IP literals + metadata hosts via `ssrf_guard.go` primitives; opt-in `AllowPrivateUpstreams(cidrs...)` registry option widens acceptance for local Ollama / LAN. No request-time guard (validation is one-shot at registration). |

## 4. Public Surface

```go
// WithUpstreamRouter mounts a selector-keyed reverse proxy on the Engine.
// Mirrors WithChatCompletions: the option sets a field; the Engine mounts at build.
//
//   reg := api.NewUpstreamRegistry()
//   _ = reg.Set("lemma", api.Upstream{URL: "http://10.0.0.5:8000", Weight: 2},
//                        api.Upstream{URL: "http://10.0.0.6:8000", Weight: 1})
//   _ = reg.SetDefault(api.Upstream{URL: "http://127.0.0.1:11434"}) // local Ollama fallback
//   engine, _ := api.New(api.WithUpstreamRouter(reg))
func WithUpstreamRouter(reg *UpstreamRegistry, opts ...UpstreamRouterOption) Option

// Upstream is one backend endpoint in a pool.
type Upstream struct {
    URL     string            // http(s) base URL; validated once at registration
    Weight  int               // weighted RR weight; <=0 treated as 1
    Headers map[string]string // static headers injected on dispatch (e.g. upstream API key)
}

// UpstreamRegistry is the runtime-mutable, thread-safe pool table (key -> pool).
// Copy-on-write: writes swap an immutable snapshot under a write mutex; reads are
// lock-free via atomic load.
type UpstreamRegistry struct { /* atomic.Pointer[registrySnapshot] + write mutex */ }

func NewUpstreamRegistry(opts ...RegistryOption) *UpstreamRegistry
func (r *UpstreamRegistry) Set(key string, ups ...Upstream) error // replace pool; validates URL + IP policy
func (r *UpstreamRegistry) Add(key string, up Upstream) error     // append one; validates URL + IP policy
func (r *UpstreamRegistry) Remove(key string)                     // drop a pool
func (r *UpstreamRegistry) SetDefault(ups ...Upstream) error      // fallback for unmatched keys
func (r *UpstreamRegistry) Keys() []string                        // introspection (sorted)

// RegistryOption configures registration-time validation policy.
type RegistryOption func(*UpstreamRegistry)

// AllowPrivateUpstreams permits the given private/loopback/reserved CIDRs to pass
// registration validation (default-deny otherwise). Metadata hosts stay hard-blocked.
//
//   reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8", "10.0.0.0/8"))
func AllowPrivateUpstreams(cidrs ...string) RegistryOption

// Selector resolves the routing key from the request. body may be nil if unread.
type Selector func(c *gin.Context, body []byte) (key string, err error)

// RouteFunc inspects the payload and may override the key or reject the request.
// Returning the same key is a no-op; a non-nil error aborts (default 400).
type RouteFunc func(c *gin.Context, key string, body []byte) (newKey string, err error)

// Router options.
func WithSelector(fn Selector) UpstreamRouterOption            // default: JSON body "model"
func WithRouteHook(fn RouteFunc) UpstreamRouterOption          // the "add logic later" seam
func WithRouterPaths(paths ...string) UpstreamRouterOption     // default ["/v1/chat/completions"]
func WithUpstreamTransformerIn(t ...any) UpstreamRouterOption  // reuses compileTransformerPipeline
func WithUpstreamTransformerOut(t ...any) UpstreamRouterOption // buffered (non-stream) responses only
func WithFailover(maxAttempts int, cooldown time.Duration) UpstreamRouterOption // default: len(pool) (each tried once), 10s
func WithFailoverStatuses(statuses ...int) UpstreamRouterOption // default: >=500; 429 opt-in
func WithUpstreamTransport(rt http.RoundTripper) UpstreamRouterOption // custom TLS/timeouts base
```

**Contract rules**
- The **registry is the single source of truth** for endpoints; the hook returns a
  *key*, the registry resolves it → all LB stays in one place.
- `Set`/`Add`/`SetDefault` **return `error`** — validation happens here, once, never
  per request: URL shape (http(s) scheme, host present, port in range) **and** IP
  policy. Loopback/private/link-local/reserved IP literals and metadata hosts are
  **rejected by default**; `AllowPrivateUpstreams(cidrs...)` widens acceptance.
  Non-metadata hostnames are accepted as trusted config without registration-time DNS.
- Transformers reuse `compileTransformerPipeline`/`runTransformerPipeline`, so
  `FieldRenamer` and any `TransformerIn[I,O]`/`TransformerOut[I,O]` work unchanged —
  but on the router they operate on the **raw upstream JSON body**, *not* the
  `{success,data}` OK-envelope (upstream responses are foreign; no unwrap).
- **Same-path forwarding**: each path in `WithRouterPaths` forwards its own
  path + query to the chosen upstream base URL. One registry keyed by `model` serves
  all OpenAI-shaped paths (`/v1/chat/completions`, `/v1/embeddings`, …).
- The router mounts on the Engine **root router**, so global engine middleware
  (auth/CORS/rate-limit/tracing) wraps it. It is **not** a `RouteGroup`, so the
  group-transformer middleware does not apply — the router's own transformers do.

## 5. Components

Each unit has one purpose and is testable in isolation.

| Unit | File | Responsibility | Depends on | gin/HTTP? |
|------|------|----------------|------------|-----------|
| `UpstreamRegistry` | `upstream_registry.go` | Copy-on-write pool table; URL validation on write | `sync/atomic`, `net/url` | no |
| `upstreamBalancer` | `upstream_balancer.go` | Weighted-RR pick over a pool; shared per-key cursors + per-upstream cooldown; `markFailed`; injectable `now()` | registry types | no |
| `upstreamTransport` | `upstream_transport.go` | `http.RoundTripper`: pick → rewrite host → inject headers → base.RoundTrip → failover retry | balancer, base `RoundTripper` | http only |
| `upstreamRouterHandler` | `upstream_router.go` | gin handler orchestration + config + default `model` selector; owns one `*httputil.ReverseProxy` | all above + transformer machinery | yes |
| Engine wiring | `options.go`, `api.go` | `WithUpstreamRouter` sets `e.upstreamRouter`; build mounts each path | — | — |

**State ownership:** cooldown timestamps and RR cursors are **shared, not
per-request** (a dead upstream must stay cooling for all callers). They live in the
balancer behind its own mutex — cursors keyed by selector key, cooldown keyed by
upstream URL. The per-request pool is stashed on `req.Context()` so a single
`ReverseProxy`/transport instance serves every request (no per-request proxy alloc).

## 6. Data Flow (one request)

```
hits mounted path  (engine auth/CORS/ratelimit/tracing already ran)
  1. read body once  — MaxBytesReader(maxToolRequestBodyBytes) -> 413 on overflow
  2. Selector(c, body)        -> key        (default: JSON "model"; empty -> 400)
  3. RouteHook(c, key, body)  -> finalKey   (inspect/override/reject -> 400/403)
  4. TransformerIn pipeline   -> rewrite outbound body + ContentLength (400 on err)
  5. registry snapshot -> pool[finalKey] else default  (none -> 404 no_upstream_for_key)
  6. bind {finalKey, pool} to ctx -> ReverseProxy.ServeHTTP

  upstreamTransport.RoundTrip  (loop <= maxAttempts)
      balancer.pick(finalKey, pool) -> up        (all cooling -> stop)
      clone req; set URL.Scheme/Host=up; inject up.Headers
      base.RoundTrip
        err or status in failoverStatuses -> balancer.markFailed(up, cooldown); retry next
        else                              -> return resp

  response:
      text/event-stream    -> FlushInterval:-1 streams through; ModifyResponse passes untouched
      else + TransformerOut -> ModifyResponse buffers, transforms raw body, drops Content-Length

  ErrorHandler (all upstreams failed/cooling) -> 503 upstream_unavailable + Retry-After
  tracing span attrs: key, upstream.url, retry.count, stream(bool), status
```

**Inherent limit:** failover is **pre-response only**. Once a 2xx returns and the
proxy starts copying (especially a live stream), upstreams cannot be switched — a
mid-stream upstream death surfaces to the client. True of every streaming proxy.

## 7. Error Taxonomy

Our errors use the framework `Fail`/`FailWithDetails` envelope; backend errors pass
through verbatim. Dividing line is client-error vs infra-error.

| Condition | Status | Code | Body |
|-----------|--------|------|------|
| Body exceeds `maxToolRequestBodyBytes` | 413 | `request_too_large` | `Fail` |
| Selector can't resolve key (no `model`) | 400 | `invalid_request` | `Fail` |
| Route hook rejects | hook's (default 400) | `routing_rejected` | `Fail` |
| `TransformerIn` fails | 400 | `invalid_request_body` | `Fail` |
| No pool for key **and** no default | 404 | `no_upstream_for_key` | `Fail` |
| Upstream 4xx (non-failover, incl. 429 unless opted-in) | passthrough | upstream's | upstream body verbatim |
| Upstream transport-error / status in failover set | → failover (retry next) | — | — |
| All upstreams failed/cooling | 503 | `upstream_unavailable` | `Fail` + `Retry-After` |
| `TransformerOut` fails | 502 | `invalid_upstream_response` | `Fail` |
| Bad URL at `Set/Add/SetDefault` | — | Go `error` at **config time** | never hits request path |

- **Failover set is configurable** (`WithFailoverStatuses`); default = transport errors
  + status ≥ 500, with 429 opt-in. A non-429 4xx is a deterministic client error →
  passed straight through, no retry.
- **Upstream URLs never leak to the client.** The 503 body is generic; selected
  upstream, error, and attempt count go to **logs (warn) + trace attributes** only.

## 8. Security Notes

- **SSRF posture — block-by-default + explicit opt-in** (aligned with
  `pkg/provider/proxy.go`, not bypassed). At registration, `Set`/`Add`/`SetDefault`
  reject loopback/private/link-local/reserved IP literals and metadata hosts using the
  root `ssrf_guard.go` primitives (`blockedIPReason`). Local Ollama / LAN boxes are
  enabled by an explicit `AllowPrivateUpstreams(cidrs...)` registry option (code-level
  intent — no env reliance). Non-metadata hostnames are accepted without
  registration-time DNS (trusted config). There is **no request-time guard** —
  validation is one-shot at registration, so the hot path stays allocation-free. The
  dispatch `RoundTrip` carries a **scoped `#nosec` with justification** (upstreams are
  registration-validated operator config), mirroring `transport_client.go:493`.
- **No URL leakage** to clients (see §7).
- **Bounded request bodies** via `MaxBytesReader(maxToolRequestBodyBytes)`, reusing
  the transformer constant.
- Header injection is per-upstream static config (e.g. upstream API keys) — never
  derived from the incoming request, so a client cannot inject upstream auth.

## 9. Testing Strategy

Convention: `_Good` / `_Bad` / `_Ugly` suffixes, example tests, `-race`, `GOWORK=off`.

**Per-unit (pure, fast)**
- `UpstreamRegistry` — Good: http/https + loopback/private accepted; Bad: `ftp://`,
  missing host, bad port, `javascript:` rejected at write; Ugly: concurrent
  `Set`+snapshot under `-race`, snapshot-before-write provably unaffected (COW).
- `upstreamBalancer` — weighted spread within tolerance over N picks; cooled upstream
  skipped until **fake clock** passes cooldown; all-cooling → `pick` returns `!ok`;
  `weight<=0`→1; concurrent `pick`/`markFailed` under `-race`.
- `upstreamTransport` — **fake base RoundTripper**: success returns resp;
  transport-error → `markFailed` + retry-next → success; status-in-set fails over,
  4xx passes through; all-fail returns last err; asserts header injection + correct
  scheme/host rewrite with path preserved.

**Integration (`httptest` upstreams)**
- Weighted spread roughly matches weights over many requests.
- Failover: A always 503, B 200 → client gets 200, A cooling.
- Streaming: SSE upstream with flushes → client receives chunks incrementally, body
  byte-identical, `TransformerOut` not applied.
- Non-stream + `FieldRenamer` out → fields renamed, `Content-Length` corrected;
  `FieldRenamer` in → upstream sees renamed body.
- Selector default routes by `model`; missing `model` → 400. Hook overrides key →
  different pool; hook reject → 403.
- **SSRF posture**: `127.0.0.1` upstream **rejected at config time by default**;
  accepted after `AllowPrivateUpstreams("127.0.0.0/8")`; non-metadata hostname accepted;
  metadata host `169.254.169.254` rejected even with a broad allow-list; `ftp://` and
  missing-host rejected. Integration: an allowed `127.0.0.1` httptest upstream serves
  end-to-end (proves no request-time guard blocks it).
- All-down → 503 + `Retry-After`; assert upstream URL absent from client body.
- Multiple mounted paths each forward their own path.
- Composition: `WithBearerAuth` in front → 401 without token.

**Gates:** `GOWORK=off go test -race ./...` green; gosec clean (scoped `#nosec`).

## 10. File Layout

```
go/upstream_registry.go         + _test.go + _example_test.go
go/upstream_balancer.go         + _internal_test.go
go/upstream_transport.go        + _internal_test.go
go/upstream_router.go           + _test.go + _example_test.go   (handler, config, default selector)
go/options.go                   (+ WithUpstreamRouter, UpstreamRouterOption helpers, e.upstreamRouter field)
go/api.go                       (+ build-time mount of each path)
go/string_constants.go          (+ error codes)
```

## 11. Future Extensions (out of v1 scope)

- Sticky / consistent-hash routing as a selectable strategy.
- Active health checks with a background prober (passive failover stays the default).
- Direct upstream selection from the hook (bypass registry) for advanced cases.
- Per-chunk streaming transformers (translate a foreign SSE format → OpenAI SSE).
- Path rewrite (strip/replace prefix) per upstream.
- Per-pool rate limits via `go-ratelimit` integration.

## 12. Open Implementation Notes

- Confirm `maxToolRequestBodyBytes`, `Fail`, `FailWithDetails`,
  `compileTransformerPipeline`, `runTransformerPipeline` signatures at implementation
  time and reuse verbatim (no forks).
- `ReverseProxy.Rewrite` (Go 1.20+) preferred over the deprecated `Director`; set only
  path/query preservation there — the host is set inside `upstreamTransport.RoundTrip`
  per attempt.
- `ModifyResponse` must distinguish streaming by response `Content-Type`
  (`text/event-stream`) — not by request flags — so an upstream that streams
  unexpectedly is still passed through.
- Decide the failover-status default constant set in `string_constants.go`.
- `Upstream.URL` may include a base path (e.g. `http://host/inference`); the incoming
  request path is appended to it. Document this in the `Upstream.URL` godoc so the
  forwarding rule is unambiguous.
