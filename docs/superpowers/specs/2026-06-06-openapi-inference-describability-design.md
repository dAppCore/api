# OpenAPI Describability for the Inference Surface — Design

- **Date:** 2026-06-06
- **Status:** Design — approved, pending implementation plan
- **Module:** `dappco.re/go/api` (`core/api/go`)
- **Author:** Snider + Cladius (brainstorming)
- **Builds on:** `WithUpstreamRouter` + `WithChatCompletionsRemote` (the two prior specs in this directory).
- **Related:** `RFC.md` §7 (SDK generation), `RFC.documentation.md` (OpenAPI/SDK tooling — the framework's headline value-prop).

---

## 1. Context & Problem

The framework auto-generates `/v1/openapi.json` from `DescribableGroup.Describe()` plus a few special-cased path items (`chatCompletionsPathItem`, `openAPISpecPathItem`) gated by `runtime.Transport.*` flags (`openapi.go` `Build()`). SDK generation (`RFC.md` §7, `RFC.documentation.md`) consumes that spec.

Two parts of the inference surface we just built are **invisible to the spec/SDKs**:

1. **Remote / hybrid chat-completions.** The rich `chatCompletionsPathItem` (request/response/SSE/error schemas, tag `inference`) already exists, but `transport.go:53` sets `ChatCompletionsEnabled: e.chatCompletionsResolver != nil` — only the **local** resolver flips it. A `WithChatCompletionsRemote`-only (or hybrid) engine serves `/v1/chat/completions` but it never appears in the spec.
2. **`WithUpstreamRouter` mounted paths.** The router mounts via `r.Any` at the engine root — not a `DescribableGroup`, not special-cased — so its paths are absent entirely.

Shipping a live inference surface that SDK consumers can't see is incoherent with the framework's purpose.

## 2. Goals / Non-Goals

**Goals**
- The chat-completions path item appears in the spec whenever a local resolver **or** a remote backend is configured (local / remote / hybrid).
- Every `WithUpstreamRouter` mounted path appears in the spec as a minimal, honest `POST` proxy item.
- De-dupe: a real item (chat, openapi-spec, swagger, or a `DescribableGroup` path) always wins over the minimal proxy item at the same path.
- Follow the existing special-cased-path mechanism — no new abstraction.

**Non-Goals**
- Inferring real request/response schemas for generic router paths (the router proxies arbitrary shapes — the minimal item is deliberately loose).
- Documenting all HTTP methods the router's `r.Any` accepts (POST only — see §3).
- Surfacing runtime routing data (model→pool table, adapters) in the static spec.
- Changing how `DescribableGroup` or `chatCompletionsPathItem` themselves work.

## 3. What Appears in the Spec

### 3.1 Chat-completions (local / remote / hybrid)
The existing `chatCompletionsPathItem` (full OpenAI request/response/SSE/error schemas, tag `inference`) is emitted whenever chat is configured by either path. No schema change — only the enabling condition widens. The remote backend is OpenAI-shaped (passthrough or adapted), so the existing schema remains accurate.

### 3.2 Upstream router paths (minimal proxy item)
Each `WithUpstreamRouter` mounted path (from `WithRouterPaths`, default `["/v1/chat/completions"]`) gets a minimal but honest `POST` item:

- **Method:** `POST` only. The router uses `r.Any`, but documenting all seven methods with freeform bodies is misleading noise; `POST` matches the inference convention and the default path.
- **Tag:** `proxy` (distinct from the real `inference` chat item, so consumers can tell a generic proxy path from the typed chat endpoint).
- **Request body:** generic `object` (`additionalProperties: true`), `required: true`, with the description: *"Selector-routed proxy. The request body must carry the selector field (default `model`); the concrete request/response schema depends on the target upstream/model."*
- **Responses:** `200` with `application/json` (generic `object`) **and** `text/event-stream` (the router streams); `404` (`no_upstream_for_key`); `503` (`upstream_unavailable`, with a `Retry-After` response header) — matching the router's real envelopes.
- **Security:** same `isPublicPathForList` treatment as the other path items (no forced-public; honours configured public paths).

### 3.3 De-dup rule
The router-path loop runs **after** the chat/openapi-spec special items and the `DescribableGroup` loop. For each router path, normalise it and skip if the `paths` map already has that key. So:
- Router mounted at `/v1/chat/completions` while chat is enabled → only the `inference` chat item (real schema) appears, never a duplicate `proxy` item.
- A router path colliding with the openapi-spec/swagger/group path → skipped.

## 4. Wiring (4 files)

1. **`transport.go`** — `TransportConfig`:
   - `ChatCompletionsEnabled: e.chatCompletionsResolver != nil || e.chatRemote != nil`.
   - New field `UpstreamRouterPaths []string`; in `TransportConfig()`, set from `e.upstreamRouter.paths` when `e.upstreamRouter != nil`, else nil.
2. **`runtime_config.go`** — no change (`Transport: e.TransportConfig()` already carries the new field).
3. **`spec_builder_helper.go`** — `builder.UpstreamRouterPaths = runtime.Transport.UpstreamRouterPaths` (beside the existing `ChatCompletionsEnabled`/`Path` assignments).
4. **`openapi.go`**:
   - `SpecBuilder` struct gains `UpstreamRouterPaths []string`.
   - New `upstreamRouterPathItem(path string, operationIDs map[string]int) map[string]any` — the §3.2 item.
   - `Build()`: after the chat/openapi-spec items and the group loop, iterate `sb.UpstreamRouterPaths`; normalise; `if _, exists := paths[norm]; exists { continue }`; else add `upstreamRouterPathItem`, applying the `isPublicPathForList` security treatment.
   - Optional `x-upstream-router-paths` extension (informational, symmetric with `x-chat-completions-*`).

The data already exists statically at spec-build time: `e.upstreamRouter.paths` (set by `WithRouterPaths`) and `e.chatRemote` (set by `WithChatCompletionsRemote`). No runtime/dynamic lookup.

## 5. Testing

Internal spec-builder tests (mirror `openapi_test.go`'s build/parse pattern — construct the `SpecBuilder` from the engine's runtime config, or fetch `/v1/openapi.json`):

- **Chat in spec — remote-only:** `api.New(WithChatCompletionsRemote(reg))` → `/v1/chat/completions` POST present with the `inference` request/response/SSE schema. `_Good` (the core gap).
- **Chat in spec — hybrid + local:** both still present (local regression guard). `_Good`
- **Chat absent** when neither local nor remote configured. `_Good`
- **Router paths in spec:** `WithUpstreamRouter(reg, WithRouterPaths("/v1/embeddings", "/v1/score"))` → both appear as `POST`, tag `proxy`, generic schema, `404` + `503` responses. `_Good`
- **De-dup (key case):** router mounted at `/v1/chat/completions` with chat enabled → exactly one item at that path, and it's the `inference` chat item (assert tag `inference` / the chat request schema, NOT `proxy`). `_Ugly`
- **De-dup vs spec/swagger path:** a router path colliding with the openapi-spec or swagger path is skipped (real item retained). `_Good`
- **OpenAPI 3.1 validity:** the produced spec still parses/validates (reuse the existing spec-validation test harness).

Gates: `_Good/_Bad/_Ugly`, `GOWORK=off go test ./ -race`, `go vet ./`, `gofmt`.

## 6. File Layout

```
go/transport.go             (mod) ChatCompletionsEnabled |= chatRemote; + UpstreamRouterPaths field + population
go/openapi.go               (mod) SpecBuilder.UpstreamRouterPaths; upstreamRouterPathItem(); Build() router loop + dedup; optional x-extension
go/spec_builder_helper.go   (mod) builder.UpstreamRouterPaths = runtime.Transport.UpstreamRouterPaths
go/openapi_inference_test.go (new, or extend openapi_test.go) describability tests
```

## 7. Future Extensions (out of v1)

- Real per-path schemas for the generic router via consumer-supplied `RouteDescription`s (the considered-but-deferred option (b) from brainstorming).
- Per-model documentation (enumerate registry keys) — runtime data, deliberately excluded from the static contract.
- Surfacing the MCP HTTP bridge + other un-described engine routes (broader describability sweep).

## 8. Open Implementation Notes

- Confirm `e.upstreamRouter` exposes `.paths` and `e.chatRemote` is the field name set by `WithChatCompletionsRemote` (both from the prior specs) at implementation time.
- Confirm `chatCompletionsPathItem`, `isPublicPathForList`, `normaliseOpenAPIPath`, `operationID`, the `paths` map population order, and the `mimeJSON` constant — reuse verbatim.
- Place the router-path loop after the `DescribableGroup` loop so the dedup covers group-contributed paths too.
- The minimal item's schema is generic on BOTH request and response: `{"type":"object","additionalProperties":true}` for the JSON request and JSON response. For the `text/event-stream` response use a generic schema too (`{"type":"string"}` or a free-form object) — do NOT reuse `chatCompletionsStreamSchema()`, which would falsely imply OpenAI chunk shape on a generic proxy whose stream format depends on the upstream.
