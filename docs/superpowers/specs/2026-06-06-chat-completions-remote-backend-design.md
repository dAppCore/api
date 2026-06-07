# Chat Completions — Remote Backend + Format Adapters — Design

- **Date:** 2026-06-06
- **Status:** Design — approved, pending implementation plan
- **Module:** `dappco.re/go/api` (`core/api/go`)
- **Author:** Snider + Cladius (brainstorming)
- **Builds on:** `WithUpstreamRouter` (`docs/superpowers/specs/2026-06-06-upstream-router-design.md`) — reuses `UpstreamRegistry`, `upstreamBalancer`, `upstreamTransport` unchanged.
- **Related:** `RFC.md` §11 (chat completions), `RFC.providers.md` Open Question 4 (go-ai backend) + §7.3 (PHP-direct anti-pattern).

---

## 1. Context & Problem

`RFC.md` §11 specifies an OpenAI-compatible `POST /v1/chat/completions`. Today `WithChatCompletions(resolver *ModelResolver)` resolves a model name to a **local, in-process** `inference.TextModel` (`chat_completions.go:714`) and is **loopback-only** (`:693`). There is no way for that endpoint to serve a model hosted on a **remote** OpenAI-compatible server (Ollama, LiteLLM, vLLM) or a non-OpenAI server (Ollama-native, Anthropic).

`RFC.providers.md` Open Question 4 — *"does go-ai proxy to Ollama / LiteLLM, run in-process, or hybrid?"* — is flagged there as the highest-leverage architectural decision. This feature answers it: **hybrid**. Local models are served in-process; everything else is routed to a remote pool via the already-built upstream router, with per-model format adapters for non-OpenAI backends.

This also enables the fix for the `RFC.providers.md` §7.3 anti-pattern (PHP calling external model services directly): one stable Go endpoint fronts heterogeneous backends.

## 2. Goals / Non-Goals

**Goals**
- One `/v1/chat/completions` endpoint that serves **local in-process** models and **remote** models, decided per request by model name.
- Reuse the upstream router's `UpstreamRegistry` + weighted-RR + passive-failover transport for the remote path.
- **Passthrough by default** for OpenAI-compatible upstreams (verbatim request + response, preserving fields our struct doesn't model).
- **Per-model format adapters** for non-OpenAI upstreams: request mapping, non-streaming response mapping, and **per-chunk streaming transcoding**. Built-ins: Ollama-native, Anthropic.
- Opt-in to expose the endpoint off-loopback, gated by a configured bearer.

**Non-Goals (v1)**
- A generic/pluggable streaming-transcoder framework beyond the two built-in adapters (consumers can implement `ChatFormatAdapter` themselves, but only Ollama + Anthropic ship).
- Tool/function-calling translation across formats (passthrough preserves OpenAI `tools`; adapter tool-mapping is a future extension).
- Embeddings/scoring endpoints (separate go-ai provider work, `RFC.providers.md` §4.1).
- Changing the local inference path (`serveStreaming`/`serveNonStreaming`) — reused unchanged.

## 3. Settled Decisions

| Fork | Decision |
|------|----------|
| Dispatch precedence | **Local-first** (`resolver.Knows(model)`) → else **remote** (per-model pool or `SetDefault`) → else 404 |
| Bind posture | **Configurable opt-in** (`WithChatCompletionsAllowRemoteClients`), allowed off-loopback only when a bearer is configured |
| Translation | **Per-pool adapters**: passthrough default; Ollama + Anthropic built-ins with request + non-stream + **streaming** transcoding |
| Proxy core | **Reuse `upstreamBalancer`+`upstreamTransport` directly** (not `httputil.ReverseProxy`) — ReverseProxy can't rewrite request bodies per-format or stream-transcode |
| Scope | One spec; internal unit boundaries kept crisp (dispatcher / passthrough / adapter iface / Ollama / Anthropic / bind) |

## 4. Public Surface

```go
// WithChatCompletionsRemote attaches a remote backend to /v1/chat/completions.
// Use WITH WithChatCompletions for hybrid (local-first); ALONE for remote-only.
//
//   reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("10.0.0.0/8"))
//   _ = reg.Set("claude-3-opus", api.Upstream{URL: "https://anthropic-gw.lthn.sh"})
//   _ = reg.Set("llama3:70b", api.Upstream{URL: "http://gpu1:11434"}, api.Upstream{URL: "http://gpu2:11434"})
//   _ = reg.SetDefault(api.Upstream{URL: "https://llm.lthn.sh"})   // OpenAI-compatible — passthrough
//   engine, _ := api.New(
//       api.WithChatCompletions(localResolver),                     // local-first (optional)
//       api.WithChatCompletionsRemote(reg,
//           api.WithChatModelAdapter("llama3:70b",    api.OllamaAdapter()),
//           api.WithChatModelAdapter("claude-3-opus", api.AnthropicAdapter()),
//       ),
//   )
func WithChatCompletionsRemote(reg *UpstreamRegistry, opts ...ChatRemoteOption) Option

type ChatRemoteOption func(*chatRemoteConfig)
func WithChatModelAdapter(model string, a ChatFormatAdapter) ChatRemoteOption // non-OpenAI models only
func WithChatRemoteFailover(maxAttempts int, cooldown time.Duration) ChatRemoteOption
func WithChatRemoteTransport(rt http.RoundTripper) ChatRemoteOption

// WithChatCompletionsAllowRemoteClients permits non-loopback clients, but only
// when a bearer is configured (WithBearerAuth) — mirrors the engine's
// ErrPublicBindNoBearer invariant. Without it, the endpoint stays loopback-only.
func WithChatCompletionsAllowRemoteClients() Option

// ChatFormatAdapter maps between the OpenAI chat shape and a non-OpenAI upstream.
// Passthrough (OpenAI-compatible) upstreams need NO adapter — that is the default.
type ChatFormatAdapter interface {
    Name() string                                              // "ollama", "anthropic"
    UpstreamPath() string                                      // "/api/chat", "/v1/messages"
    // BuildRequest maps the OpenAI request into the upstream body + protocol headers
    // (Content-Type, anthropic-version). Operator secrets (x-api-key) live in Upstream.Headers.
    BuildRequest(req ChatCompletionRequest) (body []byte, headers map[string]string, err error)
    // DecodeResponse maps a complete (non-streaming) upstream body into the OpenAI response.
    DecodeResponse(model string, upstream []byte) (ChatCompletionResponse, error)
    // Transcoder converts the upstream stream into OpenAI chunk SSE. nil = non-stream only.
    Transcoder() ChatStreamTranscoder
}

// ChatStreamTranscoder converts an upstream response stream into OpenAI
// chat.completion.chunk SSE events written to w (flushing as it goes); it emits
// the terminating "data: [DONE]". Returns on upstream EOF or ctx cancellation.
type ChatStreamTranscoder interface {
    Transcode(w io.Writer, flush func(), upstream io.Reader, meta ChatStreamMeta) error
}
type ChatStreamMeta struct {
    ID      string
    Model   string
    Created int64
}

// Built-in adapters (only the non-OpenAI formats need one).
func OllamaAdapter() ChatFormatAdapter      // OpenAI ⇄ Ollama-native /api/chat (NDJSON stream)
func AnthropicAdapter() ChatFormatAdapter   // OpenAI ⇄ Anthropic /v1/messages (event-stream)
```

**Contract rules**
- **Passthrough is the default; adapters are per-model exceptions.** A model with no `WithChatModelAdapter` is forwarded verbatim (raw request bytes up, raw response bytes down), preserving fields the `ChatCompletionRequest`/`Response` structs don't model (`tools`, `response_format`, `logprobs`, …).
- **Composable**: `WithChatCompletions` (local) and `WithChatCompletionsRemote` (remote) each set an Engine field; `build()` mounts one handler holding `resolver` (optional) + `remote` (optional). Local-only, remote-only, and hybrid all fall out.
- **`WithChatModelAdapter` keys by model** (the registry key); the adapter owns the upstream path + both-direction mapping + protocol headers.
- Remote failover/transport config reuses the router's machinery; defaults: `maxAttempts = len(pool)`, `cooldown = 10s`, base transport = cloned `http.DefaultTransport`.

## 5. Dispatch Flow

```
ServeHTTP(c):
 1. not-configured: resolver==nil && remote==nil → 503 service_unavailable
 2. bind guard: loopback → OK; non-loopback → OK only if allowRemote && bearerConfigured, else 403
 3. decode body → req; KEEP raw bytes; invalid → 400 invalid_request_error (param body)
 4. validate(req) (existing); invalid → mapped 400
 5. LOCAL:
      - PURE-LOCAL (remote == nil): model := resolver.ResolveModel(req.Model) directly
            → existing serveStreaming / serveNonStreaming. This is the CURRENT behaviour,
            unchanged — no Knows() gate, so no risk of a loadable model 404ing.
      - HYBRID (remote != nil): if resolver != nil && resolver.Knows(req.Model):
            model := resolver.ResolveModel(req.Model)  // load; err → mapResolverError
            → existing serveStreaming / serveNonStreaming; return
            else fall through to remote (avoids loading a remote-only model locally).
 6. REMOTE (remote != nil): pool, ok := remote.reg.resolve(req.Model); !ok → 404 model_not_found
        adapter := remote.adapters[req.Model]       // nil ⇒ passthrough
        dispatchRemote(c, req, raw, pool, adapter); return
 7. else → 404 model_not_found

Note: `resolve` returns the default pool (if `SetDefault` was called) for ANY unmatched
model, so with a default pool set the 404 in step 6 fires only when no default exists —
unknown models are proxied to the default upstream (which returns its own model_not_found).
`Knows()` MUST mirror `ResolveModel`'s resolution sources exactly (cache ∪ models.yaml ∪
discovery) so a Knows()-false model is genuinely one ResolveModel couldn't serve.

dispatchRemote(c, req, raw, pool, adapter):
  a. build upstream request:
       passthrough (adapter==nil): path "/v1/chat/completions", body = raw
       adapter:                    path = adapter.UpstreamPath(); body, hdrs = adapter.BuildRequest(req)
       outReq := POST(path, body); set GetBody (replay); apply hdrs
  b. bind {pool, key} on ctx; resp, err := transport.RoundTrip(outReq)   // weighted pick + failover (reused)
       err (*routerError) → OpenAI error shape (503 upstream_unavailable + Retry-After / 502)
  c. deliver:
       upstream non-2xx: passthrough → copy status+body verbatim; adapter → wrap into OpenAI error shape
       req.Stream:       passthrough → SSE headers; flushing io.Copy(resp.Body → c.Writer)
                         adapter     → tr := adapter.Transcoder(); tr==nil → 400 (param stream);
                                       else SSE headers; tr.Transcode(c.Writer, flush, resp.Body, meta)
       non-stream:       passthrough → copy resp body verbatim (200, application/json)
                         adapter     → out := adapter.DecodeResponse(model, body); err → 502; c.JSON(200, out)
```

### 5.1 `ModelResolver.Knows(name) bool` (new)

`ResolveModel` *loads* the model (`inference.LoadModel`), so it cannot be used as a cheap local-vs-remote test — it would load a remote-only model locally. The spec adds a **no-load existence check**:

```go
// Knows reports whether the resolver can serve name without loading it: a hit in
// the loaded-model cache, the models.yaml mapping, or the (cached) discovery set.
func (r *ModelResolver) Knows(name string) bool
```

Uses internals it already has (`loadedByName`, `modelsYAMLMapping`, `resolveDiscoveredPath`). Discovery results are cached so `Knows` stays cheap on the hot path.

### 5.2 Delivery writer

Streaming and buffered responses are written through gin's `c.Writer` (it implements `http.Flusher`) — never the unwrapped raw writer. This is the lesson carried from the upstream router: keeps gin's `Written()` tracking correct and avoids the superfluous-`WriteHeader` warning. The transcoder's `flush` callback is `c.Writer.Flush`.

## 6. Format Adapters

### 6.1 OllamaAdapter — OpenAI ⇄ Ollama-native `/api/chat`

| Direction | Mapping |
|---|---|
| Request | `{model, messages:[{role,content}], stream, options:{temperature, top_p, top_k, num_predict←max_tokens, stop←stop}}`; headers `Content-Type: application/json`. NOTE: Ollama reads `stop` **inside `options`**, not at the top level. |
| Non-stream resp | Ollama `{message:{role,content}, done, done_reason, prompt_eval_count, eval_count}` → content=`message.content`; `usage{prompt_tokens←prompt_eval_count, completion_tokens←eval_count}`; finish_reason=`length` if `done_reason=="length"` else `stop` |
| Stream (NDJSON) | each line `{message:{content:<delta>}, done:false}` → OpenAI chunk `delta.content`; first chunk adds `delta.role:"assistant"`; final line `{done:true, done_reason, eval_count}` → final chunk `finish_reason`, then `data: [DONE]`. Flush per line. |

### 6.2 AnthropicAdapter — OpenAI ⇄ Anthropic `/v1/messages`

| Direction | Mapping |
|---|---|
| Request | OpenAI `role:"system"` messages → top-level `system`; rest → `messages:[{role,content}]`; `max_tokens` (mandatory — default if absent), `temperature, top_p, top_k, stop_sequences←stop, stream`; headers `anthropic-version: 2023-06-01`, `Content-Type: application/json` |
| Non-stream resp | `{content:[{type:"text",text}], stop_reason, usage:{input_tokens,output_tokens}}` → content=concat text blocks; `usage{prompt_tokens←input_tokens, completion_tokens←output_tokens}`; finish_reason=map(`end_turn`→stop, `max_tokens`→length, `stop_sequence`→stop) |
| Stream (event-stream) | parse named SSE events: `message_start` (seed id/usage), `content_block_delta`+`text_delta` → OpenAI `delta.content` (first adds `delta.role:"assistant"`), `message_delta` (capture `stop_reason`), `message_stop` → final chunk `finish_reason`, then `data: [DONE]`. Flush per delta. |

Each adapter is an isolated unit (own file + tests). The Anthropic streaming transcoder is the fiddliest piece and gets the most adversarial coverage (fixture-driven).

## 7. Bind Opt-in + Error Taxonomy

**Bind.** The handler is constructed with `allowRemote` + `bearerConfigured` from the engine. Per-request guard: loopback always OK; non-loopback OK only if `allowRemote && bearerConfigured`, else 403. Mirrors `ErrPublicBindNoBearer` at the request layer. Documented caveat: `WithBearerAuth` is permissive, so operators must pair this with an auth-guarded route (`RequireAuth`) for true enforcement; the handler gate is the structural "don't expose local inference off-box without a configured bearer" guard.

**Errors** — OpenAI shape (`{"error":{message,type,param,code}}`) via the existing `writeChatCompletionError`; upstream URLs never leak (details → logs).

| Condition | HTTP | code |
|---|---|---|
| Not configured | 503 | service_unavailable |
| Non-loopback w/o allowRemote+bearer | 403 | — |
| Body decode / validation fail | 400 | (existing) |
| Local load error | mapped | `mapResolverError` (`model_not_found`/`model_loading`/`inference_error`) |
| Known neither locally nor remotely | 404 | `model_not_found` |
| `adapter.BuildRequest` fail | 500 | `inference_error` |
| All upstreams failed/cooling | 503 | `upstream_unavailable` + `Retry-After` |
| `adapter.DecodeResponse` fail | 502 | `invalid_upstream_response` |
| Stream requested, adapter non-streaming | 400 | — (param `stream`) |
| Upstream non-2xx, passthrough | verbatim | upstream's OpenAI-ish error copied through |
| Upstream non-2xx, adapter | mapped | upstream status/body wrapped into OpenAI error shape |

## 8. Testing Strategy

Reuses the router's tested `balancer`/`transport` (no re-test). Convention: `_Good/_Bad/_Ugly`, example test, `-race`, `GOWORK=off`.

**Per-unit**
- `ModelResolver.Knows()` — `_Good`: cache / `models.yaml` / discovered hits → true **without loading** (sentinel resolver asserts no load); `_Bad`: unknown → false.
- Dispatcher (fake resolver + httptest remote): `Knows`-true → local; registered remote → proxied; default-pool → proxied; unknown → 404 `model_not_found`.
- OllamaAdapter — table-driven `BuildRequest` / `DecodeResponse`; `Transcoder` fed a captured NDJSON fixture → OpenAI chunks, role-on-first, finish_reason, `data: [DONE]`.
- AnthropicAdapter — `BuildRequest` (system extraction, mandatory `max_tokens` default, sampling, `anthropic-version`), `DecodeResponse` (text-block concat, `stop_reason` map, usage), `Transcoder` fed a captured event-stream fixture → OpenAI SSE + `[DONE]`. Most adversarial coverage.

**Integration (`httptest` upstreams)**
- Hybrid: local-first model in-process + remote model proxied on one endpoint.
- Passthrough remote: request forwarded verbatim incl. an unmodelled field (`tools`) — fidelity; response verbatim; SSE passthrough.
- Ollama e2e: upstream speaking `/api/chat` (non-stream + NDJSON stream) → client gets OpenAI shape.
- Anthropic e2e: upstream speaking `/v1/messages` (non-stream + event-stream) → client gets OpenAI shape; `anthropic-version` sent.
- Failover (reuses transport): dead+live upstreams → fails over.
- Bind: non-loopback → 403 by default; with `WithChatCompletionsAllowRemoteClients`+`WithBearerAuth` → allowed; opt-in **without** bearer → still 403.
- Errors: unknown → 404 `model_not_found`; stream+non-streaming-adapter → 400; all-down → 503 `upstream_unavailable`+`Retry-After`+no-URL-leak; upstream 4xx passthrough verbatim.

**Gates:** `GOWORK=off go test ./ -race`; vet; gofmt; gosec.

## 9. File Layout

```
go/chat_remote.go             chatRemoteConfig, WithChatCompletionsRemote + opts, dispatchRemote, bind opt-in (+ _test, _example_test)
go/chat_adapter.go            ChatFormatAdapter / ChatStreamTranscoder / ChatStreamMeta
go/chat_adapter_ollama.go     OllamaAdapter (+ _test, testdata NDJSON fixture)
go/chat_adapter_anthropic.go  AnthropicAdapter (+ _test, testdata event-stream fixture)
go/chat_completions.go (mod)  handler holds resolver?+remote?+allowRemote+bearerConfigured; bind guard; local-first dispatch; ModelResolver.Knows
go/options.go (mod)           WithChatCompletionsRemote, WithChatModelAdapter, WithChatRemoteFailover, WithChatRemoteTransport, WithChatCompletionsAllowRemoteClients
go/api.go (mod)               Engine fields (chatRemote *chatRemoteConfig, chatAllowRemote bool); build wiring
```

## 10. Future Extensions (out of v1)

- Generic/pluggable streaming-transcoder registry beyond the two built-ins.
- Tool/function-calling translation across non-OpenAI formats.
- Additional adapters (Gemini, Cohere, …) implementing `ChatFormatAdapter`.
- Per-model rate limiting (ties to `RFC.md` §5 + go-ratelimit; shared with the router's deferred per-pool limits).
- Surfacing the remote/adapter routes in the generated OpenAPI spec (the broader describability gap).

## 11. Open Implementation Notes

- Confirm `ModelResolver` internals (`loadedByName`, `modelsYAMLMapping`, `resolveDiscoveredPath`) at implementation time and build `Knows` to reuse them with no load.
- Confirm `isLoopbackRequest`, `writeChatCompletionError`, `mapResolverError`, `ChatCompletionRequest/Response/Chunk`, `newChatCompletionID`, `NewThinkingExtractor` signatures (all in `chat_completions.go`) and reuse verbatim.
- Reuse `upstreamTransport` via its context-bound pool/key contract (`poolCtxKey`/`keyCtxKey`); construct the balancer+transport in `chatRemoteConfig.finalise()` mirroring the router's `buildProxy`.
- Capture small representative Ollama NDJSON and Anthropic event-stream samples as `testdata/` fixtures (or inline consts) for the transcoder tests.
- `BuildRequest` returning headers is a refinement of the interface beyond the router's transformer shape — keep operator secrets in `Upstream.Headers`, adapter contributes only protocol headers.
