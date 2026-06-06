# Chat Completions — Remote Backend + Format Adapters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `/v1/chat/completions` serve remote models (OpenAI-compatible passthrough + per-model Ollama/Anthropic adapters) alongside the existing local in-process path, choosing local-first by model name.

**Architecture:** The chat handler gains an optional remote backend. Local-first: if `ModelResolver.Knows(model)` → existing in-process path; else → remote via the upstream router's `upstreamBalancer`+`upstreamTransport` (weighted RR + failover, reused unchanged). Passthrough is default (verbatim bytes both ways); a per-model `ChatFormatAdapter` maps non-OpenAI formats (Ollama-native, Anthropic) including per-chunk streaming transcoders. Off-loopback access is an opt-in gated by a configured bearer.

**Tech Stack:** Go 1.26, gin, `dappco.re/go` (core), `dappco.re/go/inference`, the existing `chat_completions.go` + `upstream_*.go` (router). Spec: `docs/superpowers/specs/2026-06-06-chat-completions-remote-backend-design.md`.

**Conventions:** SPDX header on every file. UK English in strings. `_Good/_Bad/_Ugly` test suffixes. Run from `core/api/go` with `GOWORK=off go test ./ ...`. Commit `Co-Authored-By: Virgil <virgil@lethean.io>`.

**Reused symbols (already in package `api`, do NOT redefine):** `UpstreamRegistry`/`NewUpstreamRegistry`/`AllowPrivateUpstreams`/`Upstream`/`.resolve`, `upstreamBalancer`/`newUpstreamBalancer`, `upstreamTransport`, `routerError`, `poolCtxKey`/`keyCtxKey`, `defaultFailoverStatuses`, `defaultUpstreamCooldown`, `maxUpstreamResponseBytes`, `maxToolRequestBodyBytes`. Chat: `ChatCompletionRequest/Response/Chunk/ChatMessage/ChatChoice/ChatUsage/ChatChunkChoice/ChatMessageDelta`, `isLoopbackRequest`, `writeChatCompletionError(c,status,errType,param,message,code)`, `mapResolverError`, `newChatCompletionID`, `decodeJSONBody`, `validateChatRequest`, `defaultChatCompletionsPath`, `chatDefaultMaxTokens`. Engine: `e.bearerConfigured`, `e.chatCompletionsResolver`, `e.chatCompletionsPath`.

---

## File Structure

| File | Responsibility |
|------|----------------|
| `go/chat_completions.go` (modify) | Add `ModelResolver.Knows`; extend `chatCompletionsHandler` (resolver?+remote?+allowRemote+bearerConfigured), bind guard, local-first dispatch |
| `go/chat_remote.go` (create) | `chatRemoteConfig`, `dispatchRemote`, response delivery (passthrough/adapter), `*routerError`→OpenAI mapping |
| `go/chat_adapter.go` (create) | `ChatFormatAdapter`, `ChatStreamTranscoder`, `ChatStreamMeta` interfaces + small shared SSE helpers |
| `go/chat_adapter_ollama.go` (create) | `OllamaAdapter` |
| `go/chat_adapter_anthropic.go` (create) | `AnthropicAdapter` |
| `go/options.go` (modify) | `WithChatCompletionsRemote`, `WithChatModelAdapter`, `WithChatRemoteFailover`, `WithChatRemoteTransport`, `WithChatCompletionsAllowRemoteClients` |
| `go/api.go` (modify) | Engine fields `chatRemote *chatRemoteConfig`, `chatAllowRemote bool`; pass into handler in `build()` |

---

## Task 1: `ModelResolver.Knows()` — cheap local existence check

**Files:**
- Modify: `go/chat_completions.go`
- Test: `go/chat_remote_internal_test.go` (create; `package api`)

- [ ] **Step 1: Write the failing test**

Create `go/chat_remote_internal_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import "testing"

func TestModelResolver_Knows_Good(t *testing.T) {
	r := NewModelResolver()
	// Seed the loaded-by-name cache directly (internal test) to simulate a known model.
	r.loadedByName["lemer"] = nil
	if !r.Knows("lemer") {
		t.Fatal("Knows(lemer) = false, want true (cache hit)")
	}
}

func TestModelResolver_Knows_Bad(t *testing.T) {
	r := NewModelResolver()
	if r.Knows("does-not-exist") {
		t.Fatal("Knows(does-not-exist) = true, want false")
	}
	if r.Knows("") {
		t.Fatal("Knows(empty) = true, want false")
	}
	var nilR *ModelResolver
	if nilR.Knows("x") {
		t.Fatal("nil resolver Knows = true, want false")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestModelResolver_Knows`
Expected: FAIL — `r.Knows undefined`.

- [ ] **Step 3: Implement `Knows`**

In `go/chat_completions.go`, add after `ResolveModel` (around line 300):

```go
// Knows reports whether the resolver can serve name WITHOUT loading it — a hit
// in the loaded-model cache, the models.yaml mapping, or the discovery set. It
// mirrors ResolveModel's three resolution sources so a false result means
// ResolveModel could not have served the model either. Used by the chat handler
// to route local-vs-remote without triggering a model load (see chat_remote.go).
func (r *ModelResolver) Knows(name string) bool {
	if r == nil || core.Trim(name) == "" {
		return false
	}
	r.mu.RLock()
	_, cached := r.loadedByName[name]
	r.mu.RUnlock()
	if cached {
		return true
	}
	if _, ok := r.lookupModelPath(name); ok {
		return true
	}
	if _, ok := r.resolveDiscoveredPath(name); ok {
		return true
	}
	return false
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestModelResolver_Knows -race`
Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/chat_completions.go go/chat_remote_internal_test.go
git commit -m "$(printf 'feat(api): ModelResolver.Knows — no-load local existence check\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 2: Remote backend core — config, options, dispatch (passthrough), wiring

**Files:**
- Create: `go/chat_adapter.go`, `go/chat_remote.go`
- Modify: `go/options.go`, `go/api.go`, `go/chat_completions.go`
- Test: `go/chat_remote_test.go` (create; `package api_test`)

- [ ] **Step 1: Write the adapter interfaces (`chat_adapter.go`)**

Create `go/chat_adapter.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import "io" // Note: AX-6 — io.Writer/Reader are the transcoder stream boundary.

// ChatFormatAdapter maps between the OpenAI chat shape and a non-OpenAI upstream.
// OpenAI-compatible upstreams need NO adapter — passthrough is the default.
type ChatFormatAdapter interface {
	// Name identifies the adapter, e.g. "ollama", "anthropic".
	Name() string
	// UpstreamPath is the path under the upstream base URL, e.g. "/api/chat".
	UpstreamPath() string
	// BuildRequest maps the OpenAI request into the upstream body + protocol
	// headers (Content-Type, anthropic-version). Operator secrets (x-api-key)
	// belong in Upstream.Headers, not here.
	BuildRequest(req ChatCompletionRequest) (body []byte, headers map[string]string, err error)
	// DecodeResponse maps a complete (non-streaming) upstream body into the
	// OpenAI response.
	DecodeResponse(model string, upstream []byte) (ChatCompletionResponse, error)
	// Transcoder converts the upstream stream into OpenAI chunk SSE; nil means
	// the adapter supports non-streaming only.
	Transcoder() ChatStreamTranscoder
}

// ChatStreamTranscoder converts an upstream response stream into OpenAI
// chat.completion.chunk SSE events written to w (flushing via flush as it goes).
// It emits the terminating "data: [DONE]". Returns on upstream EOF or error.
type ChatStreamTranscoder interface {
	Transcode(w io.Writer, flush func(), upstream io.Reader, meta ChatStreamMeta) error
}

// ChatStreamMeta carries the OpenAI identity fields a transcoder stamps on every chunk.
type ChatStreamMeta struct {
	ID      string
	Model   string
	Created int64
}
```

- [ ] **Step 2: Write the config + options + engine wiring**

Create `go/chat_remote.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
)

// chatRemoteConfig is the remote backend attached to /v1/chat/completions via
// WithChatCompletionsRemote. It reuses the upstream router's balancer/transport.
type chatRemoteConfig struct {
	reg         *UpstreamRegistry
	adapters    map[string]ChatFormatAdapter
	maxAttempts int
	cooldown    time.Duration
	failover    map[int]bool
	transport   http.RoundTripper
	rt          *upstreamTransport // built in finalise
}

func (cfg *chatRemoteConfig) finalise() {
	if cfg.cooldown <= 0 {
		cfg.cooldown = defaultUpstreamCooldown
	}
	if cfg.failover == nil {
		cfg.failover = defaultFailoverStatuses()
	}
	if cfg.transport == nil {
		cfg.transport = http.DefaultTransport.(*http.Transport).Clone()
	}
	balancer := newUpstreamBalancer(cfg.cooldown, time.Now)
	cfg.rt = &upstreamTransport{
		base:        cfg.transport,
		balancer:    balancer,
		maxAttempts: cfg.maxAttempts,
		failover:    cfg.failover,
	}
}

// dispatchRemote proxies a chat request to the resolved remote pool, applying the
// per-model adapter (or verbatim passthrough when adapter == nil).
func (h *chatCompletionsHandler) dispatchRemote(c *gin.Context, req ChatCompletionRequest, raw []byte, pool []Upstream, adapter ChatFormatAdapter) {
	// Stream-capability check BEFORE dispatch (so we can still send an error body).
	if req.Stream && adapter != nil && adapter.Transcoder() == nil {
		writeChatCompletionError(c, http.StatusBadRequest, "invalid_request_error", "stream", "the adapter for this model does not support streaming", "")
		return
	}

	path := defaultChatCompletionsPath
	body := raw
	var hdrs map[string]string
	if adapter != nil {
		b, hh, err := adapter.BuildRequest(req)
		if err != nil {
			writeChatCompletionError(c, http.StatusInternalServerError, "inference_error", "model", err.Error(), "inference_error")
			return
		}
		path, body, hdrs = adapter.UpstreamPath(), b, hh
	}

	outReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		writeChatCompletionError(c, http.StatusInternalServerError, "inference_error", "model", err.Error(), "inference_error")
		return
	}
	bound := body
	outReq.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(bound)), nil }
	outReq.ContentLength = int64(len(bound))
	outReq.Header.Set("Content-Type", "application/json")
	for k, v := range hdrs {
		outReq.Header.Set(k, v)
	}
	ctx := context.WithValue(outReq.Context(), poolCtxKey, pool)
	ctx = context.WithValue(ctx, keyCtxKey, req.Model)
	outReq = outReq.WithContext(ctx)

	resp, err := h.remote.rt.RoundTrip(outReq)
	if err != nil {
		status, code := http.StatusServiceUnavailable, "upstream_unavailable"
		var re *routerError
		if core.As(err, &re) {
			status, code = re.status, re.code
		}
		if status == http.StatusServiceUnavailable {
			c.Header("Retry-After", strconv.Itoa(int(h.remote.cooldown.Seconds())))
		}
		writeChatCompletionError(c, status, "invalid_request_error", "model", "upstream request failed", code)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	h.deliverRemote(c, req, adapter, resp)
}

func (h *chatCompletionsHandler) deliverRemote(c *gin.Context, req ChatCompletionRequest, adapter ChatFormatAdapter, resp *http.Response) {
	// Non-2xx: passthrough copies verbatim; adapter wraps in the OpenAI error shape.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamResponseBytes))
		if adapter == nil {
			c.Header("Content-Type", "application/json")
			c.Status(resp.StatusCode)
			_, _ = c.Writer.Write(body)
			return
		}
		writeChatCompletionError(c, resp.StatusCode, "invalid_request_error", "model", "upstream error: "+string(body), "upstream_error")
		return
	}

	if req.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Status(http.StatusOK)
		flush := c.Writer.Flush
		if adapter == nil {
			copyFlushing(c.Writer, resp.Body, flush)
			return
		}
		meta := ChatStreamMeta{ID: newChatCompletionID(), Model: req.Model, Created: time.Now().Unix()}
		_ = adapter.Transcoder().Transcode(c.Writer, flush, resp.Body, meta)
		return
	}

	// Non-streaming.
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamResponseBytes))
	if adapter == nil {
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(body)
		return
	}
	out, err := adapter.DecodeResponse(req.Model, body)
	if err != nil {
		writeChatCompletionError(c, http.StatusBadGateway, "invalid_request_error", "model", "could not decode upstream response", "invalid_upstream_response")
		return
	}
	c.JSON(http.StatusOK, out)
}

// copyFlushing streams src to dst, flushing after each read so SSE chunks reach
// the client immediately.
func copyFlushing(dst io.Writer, src io.Reader, flush func()) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return
			}
			if flush != nil {
				flush()
			}
		}
		if err != nil {
			return
		}
	}
}
```

- [ ] **Step 3: Add the options (`options.go`) + engine fields (`api.go`)**

In `go/options.go`, after `WithChatCompletionsPath` (~line 849):

```go
// WithChatCompletionsRemote attaches a remote backend to /v1/chat/completions.
// Compose with WithChatCompletions for hybrid (local-first); use alone for
// remote-only. Models with no WithChatModelAdapter are forwarded verbatim
// (OpenAI passthrough); adapters map non-OpenAI upstreams (see chat_adapter.go).
//
//	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("10.0.0.0/8"))
//	_ = reg.SetDefault(api.Upstream{URL: "https://llm.lthn.sh"})
//	api.New(api.WithChatCompletions(local), api.WithChatCompletionsRemote(reg))
func WithChatCompletionsRemote(reg *UpstreamRegistry, opts ...ChatRemoteOption) Option {
	return func(e *Engine) {
		if reg == nil {
			return
		}
		cfg := &chatRemoteConfig{reg: reg, adapters: map[string]ChatFormatAdapter{}}
		for _, opt := range opts {
			if opt != nil {
				opt(cfg)
			}
		}
		cfg.finalise()
		e.chatRemote = cfg
	}
}

// ChatRemoteOption configures the chat remote backend.
type ChatRemoteOption func(*chatRemoteConfig)

// WithChatModelAdapter maps a model name to a non-OpenAI format adapter.
func WithChatModelAdapter(model string, a ChatFormatAdapter) ChatRemoteOption {
	return func(cfg *chatRemoteConfig) {
		if core.Trim(model) != "" && a != nil {
			cfg.adapters[model] = a
		}
	}
}

// WithChatRemoteFailover sets max upstream attempts + per-upstream cooldown for
// the remote backend (default: len(pool), 10s).
func WithChatRemoteFailover(maxAttempts int, cooldown time.Duration) ChatRemoteOption {
	return func(cfg *chatRemoteConfig) {
		cfg.maxAttempts = maxAttempts
		if cooldown > 0 {
			cfg.cooldown = cooldown
		}
	}
}

// WithChatRemoteTransport sets the base RoundTripper for remote dispatch.
func WithChatRemoteTransport(rt http.RoundTripper) ChatRemoteOption {
	return func(cfg *chatRemoteConfig) { cfg.transport = rt }
}

// WithChatCompletionsAllowRemoteClients permits non-loopback clients on the chat
// endpoint, but ONLY when a bearer is configured (WithBearerAuth) — mirrors the
// engine's ErrPublicBindNoBearer invariant. Without it, the endpoint stays
// loopback-only. Pair with an auth-guarded route for real enforcement.
func WithChatCompletionsAllowRemoteClients() Option {
	return func(e *Engine) { e.chatAllowRemote = true }
}
```

Confirm `options.go` already imports `time`, `net/http`, `core` (it does — used by other options). 

In `go/api.go`, add to the `Engine` struct (after `upstreamRouter *upstreamRouterConfig`):

```go
	// chatRemote, when set via WithChatCompletionsRemote, adds a remote backend
	// to the chat completions endpoint (local-first dispatch).
	chatRemote *chatRemoteConfig
	// chatAllowRemote permits non-loopback chat clients when a bearer is set.
	chatAllowRemote bool
```

- [ ] **Step 4: Wire the handler (`chat_completions.go` + `api.go` build)**

In `go/chat_completions.go`, replace the `chatCompletionsHandler` struct + constructor + `ServeHTTP` head with:

```go
type chatCompletionsHandler struct {
	resolver         *ModelResolver
	remote           *chatRemoteConfig
	allowRemote      bool
	bearerConfigured bool
}

func newChatCompletionsHandler(resolver *ModelResolver, remote *chatRemoteConfig, allowRemote, bearerConfigured bool) *chatCompletionsHandler {
	return &chatCompletionsHandler{
		resolver:         resolver,
		remote:           remote,
		allowRemote:      allowRemote,
		bearerConfigured: bearerConfigured,
	}
}

func (h *chatCompletionsHandler) ServeHTTP(c *gin.Context) {
	if h == nil || (h.resolver == nil && h.remote == nil) {
		writeChatCompletionError(c, http.StatusServiceUnavailable, "invalid_request_error", "model", "chat handler is not configured", "service_unavailable")
		return
	}

	if !isLoopbackRequest(c.Request) && !(h.allowRemote && h.bearerConfigured) {
		writeChatCompletionError(c, http.StatusForbidden, "invalid_request_error", "request", "chat completions is only available on loopback interfaces", "")
		return
	}

	raw, ok := readChatBody(c)
	if !ok {
		return
	}
	var req ChatCompletionRequest
	if err := decodeJSONBody(bytes.NewReader(raw), &req); err != nil {
		writeChatCompletionError(c, http.StatusBadRequest, "invalid_request_error", "body", "invalid request body", "")
		return
	}
	if err := validateChatRequest(&req); err != nil {
		chatErr, isChatErr := err.(*chatCompletionRequestError)
		if !isChatErr {
			writeChatCompletionError(c, http.StatusBadRequest, "invalid_request_error", "body", err.Error(), "")
			return
		}
		writeChatCompletionError(c, chatErr.Status, chatErr.Type, chatErr.Param, chatErr.Message, chatErr.Code)
		return
	}

	// PURE-LOCAL: unchanged current behaviour (no Knows gate).
	if h.remote == nil {
		h.serveLocal(c, req)
		return
	}
	// HYBRID: local-first if the resolver knows the model; else remote.
	if h.resolver != nil && h.resolver.Knows(req.Model) {
		h.serveLocal(c, req)
		return
	}
	pool, found := h.remote.reg.resolve(req.Model)
	if !found {
		writeChatCompletionError(c, http.StatusNotFound, "invalid_request_error", "model", "model not found: "+req.Model, "model_not_found")
		return
	}
	h.dispatchRemote(c, req, raw, pool, h.remote.adapters[req.Model])
}

// readChatBody reads the bounded request body once (so it can drive both the
// selector and a verbatim upstream forward).
func readChatBody(c *gin.Context) ([]byte, bool) {
	limited := http.MaxBytesReader(c.Writer, c.Request.Body, maxToolRequestBodyBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		if err.Error() == "http: request body too large" {
			writeChatCompletionError(c, http.StatusRequestEntityTooLarge, "invalid_request_error", "body", "request body too large", "")
			return nil, false
		}
		writeChatCompletionError(c, http.StatusBadRequest, "invalid_request_error", "body", "unable to read request body", "")
		return nil, false
	}
	return body, true
}
```

Then refactor the existing local logic (resolve → options → serve) from the old `ServeHTTP` body into a new method `serveLocal` (move lines that were after the decode/validate block — `resolver.ResolveModel`, `chatRequestOptions`, `normalizedStopSequences`, message conversion, stream dispatch):

```go
func (h *chatCompletionsHandler) serveLocal(c *gin.Context, req ChatCompletionRequest) {
	if h.resolver == nil {
		writeChatCompletionError(c, http.StatusNotFound, "invalid_request_error", "model", "model not found: "+req.Model, "model_not_found")
		return
	}
	model, err := h.resolver.ResolveModel(req.Model)
	if err != nil {
		status, errType, errCode, errParam := mapResolverError(err)
		writeChatCompletionError(c, status, errType, errParam, err.Error(), errCode)
		return
	}
	reqForOptions := req
	reqForOptions.Stop = nil
	options, err := chatRequestOptions(&reqForOptions)
	if err != nil {
		writeChatCompletionError(c, http.StatusBadRequest, "invalid_request_error", "stop", err.Error(), "")
		return
	}
	stopSequences, err := normalizedStopSequences(req.Stop)
	if err != nil {
		writeChatCompletionError(c, http.StatusBadRequest, "invalid_request_error", "stop", err.Error(), "")
		return
	}
	messages := make([]inference.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, inference.Message{Role: msg.Role, Content: msg.Content})
	}
	if req.Stream {
		h.serveStreaming(c, model, req, messages, stopSequences, options...)
		return
	}
	h.serveNonStreaming(c, model, req, messages, stopSequences, options...)
}
```

Add `"bytes"` and `"io"` to `chat_completions.go` imports if not present (`io` likely is not — add both).

> Confirm `decodeJSONBody(reader any, dest any)` accepts an `io.Reader` — the original `ServeHTTP` called it with `c.Request.Body` (an `io.Reader`), so `bytes.NewReader(raw)` is compatible. If it type-asserts to `io.ReadCloser` specifically, wrap with `io.NopCloser(bytes.NewReader(raw))`.

In `go/api.go` `build()`, replace the chat-completions mount block:

```go
	// Mount the OpenAI-compatible chat completion endpoint when a local resolver
	// and/or a remote backend is configured.
	if e.chatCompletionsResolver != nil || e.chatRemote != nil {
		path := e.chatCompletionsPath
		if core.Trim(path) == "" {
			path = defaultChatCompletionsPath
		}
		h := newChatCompletionsHandler(e.chatCompletionsResolver, e.chatRemote, e.chatAllowRemote, e.bearerConfigured)
		r.POST(path, h.ServeHTTP)
	}
```

And in `New()` (api.go ~138), broaden the default-path guard so remote-only also gets the default path:

```go
	if (e.chatCompletionsResolver != nil || e.chatRemote != nil) && core.Trim(e.chatCompletionsPath) == "" {
		e.chatCompletionsPath = defaultChatCompletionsPath
	}
```

- [ ] **Step 5: Write integration tests (`chat_remote_test.go`)**

Create `go/chat_remote_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	api "dappco.re/go/api"
)

// chatPost sends a chat request from a loopback client.
func chatPost(t *testing.T, base, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(base+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	return resp
}

func TestChatRemote_Passthrough_Good(t *testing.T) {
	var gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = io.WriteString(w, `{"id":"x","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}]}`)
	}))
	defer up.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.SetDefault(api.Upstream{URL: up.URL})
	e, _ := api.New(api.WithChatCompletionsRemote(reg))
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	// Send an unmodelled field (tools) to prove verbatim passthrough fidelity.
	resp := chatPost(t, srv.URL, `{"model":"gpt-x","messages":[{"role":"user","content":"hi"}],"tools":[{"type":"function"}]}`)
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if !strings.Contains(gotBody, `"tools"`) {
		t.Errorf("upstream did not receive verbatim body (tools dropped): %s", gotBody)
	}
	if !strings.Contains(string(out), `"content":"hi"`) {
		t.Errorf("client did not get upstream response: %s", out)
	}
}

func TestChatRemote_UnknownModel_Bad(t *testing.T) {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("known", api.Upstream{URL: "http://127.0.0.1:1"}) // no default
	e, _ := api.New(api.WithChatCompletionsRemote(reg))
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp := chatPost(t, srv.URL, `{"model":"nope","messages":[{"role":"user","content":"x"}]}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "model_not_found") {
		t.Errorf("want model_not_found, got %s", body)
	}
}

func TestChatRemote_Failover_Good(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(503) }))
	defer dead.Close()
	live := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)
	}))
	defer live.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("m", api.Upstream{URL: dead.URL}, api.Upstream{URL: live.URL})
	e, _ := api.New(api.WithChatCompletionsRemote(reg))
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp := chatPost(t, srv.URL, `{"model":"m","messages":[{"role":"user","content":"x"}]}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (failed over)", resp.StatusCode)
	}
}

func TestChatRemote_StreamingPassthrough_Good(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f, _ := w.(http.Flusher)
		for _, ch := range []string{"data: {\"x\":1}\n\n", "data: [DONE]\n\n"} {
			_, _ = io.WriteString(w, ch)
			if f != nil {
				f.Flush()
			}
		}
	}))
	defer up.Close()
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.SetDefault(api.Upstream{URL: up.URL})
	e, _ := api.New(api.WithChatCompletionsRemote(reg))
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp := chatPost(t, srv.URL, `{"model":"m","messages":[{"role":"user","content":"x"}],"stream":true}`)
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want SSE", ct)
	}
	out, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(out), "[DONE]") {
		t.Errorf("stream not passed through: %s", out)
	}
}

func TestChatRemote_BindOptIn_Bad(t *testing.T) {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.SetDefault(api.Upstream{URL: "http://127.0.0.1:1"})
	// No allow-remote, no bearer: non-loopback would be rejected. We assert the
	// guard logic via a loopback request still works (positive) and that the
	// option+bearer path is constructed without error.
	e, _ := api.New(
		api.WithBearerAuth("secret"),
		api.WithChatCompletionsAllowRemoteClients(),
		api.WithChatCompletionsRemote(reg),
	)
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()
	// Loopback client is always allowed regardless of opt-in.
	resp := chatPost(t, srv.URL, `{"model":"m","messages":[{"role":"user","content":"x"}]}`)
	defer resp.Body.Close()
	// httptest client is loopback → not 403. (Off-loopback 403 is covered by the
	// internal guard unit test below.)
	if resp.StatusCode == http.StatusForbidden {
		t.Fatalf("loopback client got 403, want allowed")
	}
}
```

> The off-loopback 403 path is hard to exercise via httptest (always 127.0.0.1). Add an internal guard unit test in `chat_remote_internal_test.go` that calls the guard directly:

```go
func TestChatHandler_BindGuard_Ugly(t *testing.T) {
	// non-loopback remote addr, no opt-in → must be rejected.
	h := newChatCompletionsHandler(nil, &chatRemoteConfig{}, false, false)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}]}`))
	req.RemoteAddr = "203.0.113.7:5555"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	h.ServeHTTP(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("non-loopback w/o opt-in: code = %d, want 403", w.Code)
	}
	// With opt-in + bearer configured → not 403 (proceeds to dispatch/404 etc.).
	h2 := newChatCompletionsHandler(nil, &chatRemoteConfig{reg: NewUpstreamRegistry()}, true, true)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	r2 := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}]}`))
	r2.RemoteAddr = "203.0.113.7:5555"
	c2.Request = r2
	h2.ServeHTTP(c2)
	if w2.Code == http.StatusForbidden {
		t.Fatalf("non-loopback WITH opt-in+bearer: code = 403, want allowed")
	}

	// Opt-in but NO bearer configured → still 403 (mirrors ErrPublicBindNoBearer).
	h3 := newChatCompletionsHandler(nil, &chatRemoteConfig{}, true, false)
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	r3 := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}]}`))
	r3.RemoteAddr = "203.0.113.7:5555"
	c3.Request = r3
	h3.ServeHTTP(c3)
	if w3.Code != http.StatusForbidden {
		t.Fatalf("non-loopback opt-in WITHOUT bearer: code = %d, want 403", w3.Code)
	}
}
```

Add imports `net/http`, `net/http/httptest`, `strings`, `github.com/gin-gonic/gin` to `chat_remote_internal_test.go`.

- [ ] **Step 6: Run, verify, commit**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go build ./ && GOWORK=off go test ./ -run 'TestChatRemote|TestChatHandler|TestModelResolver_Knows' -race`
Expected: PASS.

```bash
cd /Users/snider/Code/core/api
git add go/chat_adapter.go go/chat_remote.go go/chat_remote_test.go go/chat_remote_internal_test.go go/options.go go/api.go go/chat_completions.go
git commit -m "$(printf 'feat(api): chat-completions remote backend — local-first dispatch + OpenAI passthrough\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 3: OllamaAdapter

**Files:**
- Create: `go/chat_adapter_ollama.go`
- Test: `go/chat_adapter_ollama_test.go` (`package api_test`)

- [ ] **Step 1: Write the failing tests**

Create `go/chat_adapter_ollama_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	api "dappco.re/go/api"
)

func TestOllamaAdapter_BuildRequest_Good(t *testing.T) {
	a := api.OllamaAdapter()
	mt := 64
	body, hdrs, err := a.BuildRequest(api.ChatCompletionRequest{
		Model: "llama3", Messages: []api.ChatMessage{{Role: "user", Content: "hi"}}, MaxTokens: &mt, Stream: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if hdrs["Content-Type"] != "application/json" {
		t.Errorf("missing content-type header")
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	if got["model"] != "llama3" || got["stream"] != true {
		t.Errorf("bad ollama body: %s", body)
	}
	opts, _ := got["options"].(map[string]any)
	if opts["num_predict"].(float64) != 64 {
		t.Errorf("max_tokens not mapped to num_predict: %s", body)
	}
}

func TestOllamaAdapter_DecodeResponse_Good(t *testing.T) {
	a := api.OllamaAdapter()
	out, err := a.DecodeResponse("llama3", []byte(`{"message":{"role":"assistant","content":"4"},"done":true,"done_reason":"stop","prompt_eval_count":3,"eval_count":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Choices[0].Message.Content != "4" || out.Choices[0].FinishReason != "stop" {
		t.Errorf("bad decode: %+v", out)
	}
	if out.Usage.PromptTokens != 3 || out.Usage.CompletionTokens != 1 {
		t.Errorf("bad usage: %+v", out.Usage)
	}
}

func TestOllamaAdapter_Transcode_Good(t *testing.T) {
	a := api.OllamaAdapter()
	stream := strings.Join([]string{
		`{"message":{"role":"assistant","content":"He"},"done":false}`,
		`{"message":{"role":"assistant","content":"llo"},"done":false}`,
		`{"message":{"role":"assistant","content":""},"done":true,"done_reason":"stop"}`,
	}, "\n")
	var buf bytes.Buffer
	err := a.Transcoder().Transcode(&buf, func() {}, strings.NewReader(stream), api.ChatStreamMeta{ID: "id", Model: "llama3", Created: 1})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"content":"He"`) || !strings.Contains(got, `"content":"llo"`) {
		t.Errorf("missing deltas: %s", got)
	}
	if !strings.Contains(got, `"finish_reason":"stop"`) || !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing terminal/[DONE]: %s", got)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestOllamaAdapter`
Expected: FAIL — `api.OllamaAdapter undefined`.

- [ ] **Step 3: Implement `OllamaAdapter`**

Create `go/chat_adapter_ollama.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio"
	"encoding/json"
	"io"

	core "dappco.re/go"
)

type ollamaAdapter struct{}

// OllamaAdapter maps OpenAI chat completions to/from Ollama's native /api/chat
// (JSON request with an "options" block; newline-delimited JSON stream).
func OllamaAdapter() ChatFormatAdapter { return ollamaAdapter{} }

func (ollamaAdapter) Name() string         { return "ollama" }
func (ollamaAdapter) UpstreamPath() string { return "/api/chat" }

func (ollamaAdapter) BuildRequest(req ChatCompletionRequest) ([]byte, map[string]string, error) {
	msgs := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	options := map[string]any{}
	if req.Temperature != nil {
		options["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		options["top_p"] = *req.TopP
	}
	if req.TopK != nil {
		options["top_k"] = *req.TopK
	}
	if req.MaxTokens != nil {
		options["num_predict"] = *req.MaxTokens
	}
	body := map[string]any{
		"model":    req.Model,
		"messages": msgs,
		"stream":   req.Stream,
	}
	if len(options) > 0 {
		body["options"] = options
	}
	if len(req.Stop) > 0 {
		body["stop"] = []string(req.Stop)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, nil, core.E("ollama", "marshal request", err)
	}
	return raw, map[string]string{"Content-Type": "application/json"}, nil
}

type ollamaResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done            bool   `json:"done"`
	DoneReason      string `json:"done_reason"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

func ollamaFinish(doneReason string) string {
	if doneReason == "length" {
		return "length"
	}
	return "stop"
}

func (ollamaAdapter) DecodeResponse(model string, upstream []byte) (ChatCompletionResponse, error) {
	var or ollamaResponse
	if err := json.Unmarshal(upstream, &or); err != nil {
		return ChatCompletionResponse{}, core.E("ollama", "decode response", err)
	}
	return ChatCompletionResponse{
		ID:      newChatCompletionID(),
		Object:  "chat.completion",
		Model:   model,
		Choices: []ChatChoice{{Index: 0, Message: ChatMessage{Role: "assistant", Content: or.Message.Content}, FinishReason: ollamaFinish(or.DoneReason)}},
		Usage:   ChatUsage{PromptTokens: or.PromptEvalCount, CompletionTokens: or.EvalCount, TotalTokens: or.PromptEvalCount + or.EvalCount},
	}, nil
}

func (ollamaAdapter) Transcoder() ChatStreamTranscoder { return ollamaTranscoder{} }

type ollamaTranscoder struct{}

func (ollamaTranscoder) Transcode(w io.Writer, flush func(), upstream io.Reader, meta ChatStreamMeta) error {
	scanner := bufio.NewScanner(upstream)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	first := true
	for scanner.Scan() {
		line := core.Trim(scanner.Text())
		if line == "" {
			continue
		}
		var or ollamaResponse
		if err := json.Unmarshal([]byte(line), &or); err != nil {
			continue // skip malformed line
		}
		if or.Done {
			fr := ollamaFinish(or.DoneReason)
			writeChatChunk(w, flush, ChatCompletionChunk{
				ID: meta.ID, Object: "chat.completion.chunk", Created: meta.Created, Model: meta.Model,
				Choices: []ChatChunkChoice{{Index: 0, Delta: ChatMessageDelta{}, FinishReason: &fr}},
			})
			break
		}
		delta := ChatMessageDelta{Content: or.Message.Content}
		if first {
			delta.Role = "assistant"
			first = false
		}
		writeChatChunk(w, flush, ChatCompletionChunk{
			ID: meta.ID, Object: "chat.completion.chunk", Created: meta.Created, Model: meta.Model,
			Choices: []ChatChunkChoice{{Index: 0, Delta: delta, FinishReason: nil}},
		})
	}
	writeSSEDone(w, flush)
	return scanner.Err()
}
```

Add the shared SSE writers to `go/chat_adapter.go`:

```go
// writeChatChunk marshals a chunk as one SSE "data:" event and flushes.
func writeChatChunk(w io.Writer, flush func(), chunk ChatCompletionChunk) {
	data := core.JSONMarshal(chunk)
	raw, ok := data.Value.([]byte)
	if !data.OK || !ok {
		return
	}
	_, _ = io.WriteString(w, "data: ")
	_, _ = w.Write(raw)
	_, _ = io.WriteString(w, "\n\n")
	if flush != nil {
		flush()
	}
}

// writeSSEDone emits the terminating sentinel.
func writeSSEDone(w io.Writer, flush func()) {
	_, _ = io.WriteString(w, "data: [DONE]\n\n")
	if flush != nil {
		flush()
	}
}
```

Add `core "dappco.re/go"` to `chat_adapter.go` imports.

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run 'TestOllamaAdapter' -race`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/chat_adapter_ollama.go go/chat_adapter_ollama_test.go go/chat_adapter.go
git commit -m "$(printf 'feat(api): OllamaAdapter — OpenAI <-> Ollama-native /api/chat\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 4: AnthropicAdapter

**Files:**
- Create: `go/chat_adapter_anthropic.go`
- Test: `go/chat_adapter_anthropic_test.go` (`package api_test`)

- [ ] **Step 1: Write the failing tests**

Create `go/chat_adapter_anthropic_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	api "dappco.re/go/api"
)

func TestAnthropicAdapter_BuildRequest_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	body, hdrs, err := a.BuildRequest(api.ChatCompletionRequest{
		Model: "claude-3", Messages: []api.ChatMessage{{Role: "system", Content: "be terse"}, {Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if hdrs["anthropic-version"] == "" {
		t.Errorf("missing anthropic-version header")
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	if got["system"] != "be terse" {
		t.Errorf("system not extracted: %s", body)
	}
	msgs, _ := got["messages"].([]any)
	if len(msgs) != 1 { // system removed from messages
		t.Errorf("system not removed from messages: %s", body)
	}
	if _, ok := got["max_tokens"]; !ok {
		t.Errorf("max_tokens (mandatory) missing: %s", body)
	}
}

func TestAnthropicAdapter_DecodeResponse_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	out, err := a.DecodeResponse("claude-3", []byte(`{"content":[{"type":"text","text":"Hi"},{"type":"text","text":" there"}],"stop_reason":"max_tokens","usage":{"input_tokens":5,"output_tokens":2}}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Choices[0].Message.Content != "Hi there" {
		t.Errorf("text blocks not concatenated: %q", out.Choices[0].Message.Content)
	}
	if out.Choices[0].FinishReason != "length" {
		t.Errorf("max_tokens not mapped to length: %s", out.Choices[0].FinishReason)
	}
	if out.Usage.PromptTokens != 5 || out.Usage.CompletionTokens != 2 {
		t.Errorf("bad usage: %+v", out.Usage)
	}
}

func TestAnthropicAdapter_Transcode_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	// Minimal Anthropic event stream.
	stream := strings.Join([]string{
		"event: message_start",
		`data: {"type":"message_start","message":{"usage":{"input_tokens":5}}}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"He"}}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"llo"}}`,
		"",
		"event: message_delta",
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"}}`,
		"",
		"event: message_stop",
		`data: {"type":"message_stop"}`,
		"",
	}, "\n")
	var buf bytes.Buffer
	err := a.Transcoder().Transcode(&buf, func() {}, strings.NewReader(stream), api.ChatStreamMeta{ID: "id", Model: "claude-3", Created: 1})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"content":"He"`) || !strings.Contains(got, `"content":"llo"`) {
		t.Errorf("missing deltas: %s", got)
	}
	if !strings.Contains(got, `"finish_reason":"stop"`) || !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing terminal/[DONE]: %s", got)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestAnthropicAdapter`
Expected: FAIL — `api.AnthropicAdapter undefined`.

- [ ] **Step 3: Implement `AnthropicAdapter`**

Create `go/chat_adapter_anthropic.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio"
	"encoding/json"
	"io"

	core "dappco.re/go"
)

const anthropicVersion = "2023-06-01"

type anthropicAdapter struct{}

// AnthropicAdapter maps OpenAI chat completions to/from Anthropic's /v1/messages
// (top-level system field, mandatory max_tokens, content blocks, SSE event stream).
func AnthropicAdapter() ChatFormatAdapter { return anthropicAdapter{} }

func (anthropicAdapter) Name() string         { return "anthropic" }
func (anthropicAdapter) UpstreamPath() string { return "/v1/messages" }

func anthropicFinish(stopReason string) string {
	switch stopReason {
	case "max_tokens":
		return "length"
	default: // end_turn, stop_sequence, etc.
		return "stop"
	}
}

func (anthropicAdapter) BuildRequest(req ChatCompletionRequest) ([]byte, map[string]string, error) {
	var system string
	msgs := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == "system" {
			if system != "" {
				system += "\n"
			}
			system += m.Content
			continue
		}
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	maxTokens := chatDefaultMaxTokens
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}
	body := map[string]any{
		"model":      req.Model,
		"messages":   msgs,
		"max_tokens": maxTokens,
		"stream":     req.Stream,
	}
	if system != "" {
		body["system"] = system
	}
	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		body["top_p"] = *req.TopP
	}
	if req.TopK != nil {
		body["top_k"] = *req.TopK
	}
	if len(req.Stop) > 0 {
		body["stop_sequences"] = []string(req.Stop)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, nil, core.E("anthropic", "marshal request", err)
	}
	return raw, map[string]string{"Content-Type": "application/json", "anthropic-version": anthropicVersion}, nil
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (anthropicAdapter) DecodeResponse(model string, upstream []byte) (ChatCompletionResponse, error) {
	var ar anthropicResponse
	if err := json.Unmarshal(upstream, &ar); err != nil {
		return ChatCompletionResponse{}, core.E("anthropic", "decode response", err)
	}
	var content string
	for _, b := range ar.Content {
		if b.Type == "text" {
			content += b.Text
		}
	}
	return ChatCompletionResponse{
		ID:      newChatCompletionID(),
		Object:  "chat.completion",
		Model:   model,
		Choices: []ChatChoice{{Index: 0, Message: ChatMessage{Role: "assistant", Content: content}, FinishReason: anthropicFinish(ar.StopReason)}},
		Usage:   ChatUsage{PromptTokens: ar.Usage.InputTokens, CompletionTokens: ar.Usage.OutputTokens, TotalTokens: ar.Usage.InputTokens + ar.Usage.OutputTokens},
	}, nil
}

func (anthropicAdapter) Transcoder() ChatStreamTranscoder { return anthropicTranscoder{} }

type anthropicTranscoder struct{}

type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Delta struct {
		Type       string `json:"type"`
		Text       string `json:"text"`
		StopReason string `json:"stop_reason"`
	} `json:"delta"`
}

func (anthropicTranscoder) Transcode(w io.Writer, flush func(), upstream io.Reader, meta ChatStreamMeta) error {
	scanner := bufio.NewScanner(upstream)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	first := true
	stopReason := "end_turn"
	for scanner.Scan() {
		line := core.Trim(scanner.Text())
		if !core.HasPrefix(line, "data:") {
			continue // skip "event:" and blank lines; the data line carries type
		}
		payload := core.Trim(line[len("data:"):])
		if payload == "" {
			continue
		}
		var ev anthropicStreamEvent
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue
		}
		switch ev.Type {
		case "content_block_delta":
			if ev.Delta.Type != "text_delta" || ev.Delta.Text == "" {
				continue
			}
			delta := ChatMessageDelta{Content: ev.Delta.Text}
			if first {
				delta.Role = "assistant"
				first = false
			}
			writeChatChunk(w, flush, ChatCompletionChunk{
				ID: meta.ID, Object: "chat.completion.chunk", Created: meta.Created, Model: meta.Model,
				Choices: []ChatChunkChoice{{Index: 0, Delta: delta, FinishReason: nil}},
			})
		case "message_delta":
			if ev.Delta.StopReason != "" {
				stopReason = ev.Delta.StopReason
			}
		case "message_stop":
			fr := anthropicFinish(stopReason)
			writeChatChunk(w, flush, ChatCompletionChunk{
				ID: meta.ID, Object: "chat.completion.chunk", Created: meta.Created, Model: meta.Model,
				Choices: []ChatChunkChoice{{Index: 0, Delta: ChatMessageDelta{}, FinishReason: &fr}},
			})
			writeSSEDone(w, flush)
			return scanner.Err()
		}
	}
	// Stream ended without an explicit message_stop — still terminate cleanly.
	writeSSEDone(w, flush)
	return scanner.Err()
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run 'TestAnthropicAdapter' -race`
Expected: PASS (3 tests).

- [ ] **Step 5: End-to-end adapter integration test**

Add to `go/chat_remote_test.go`:

```go
func TestChatRemote_OllamaAdapter_E2E_Good(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("upstream path = %s, want /api/chat", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"message":{"role":"assistant","content":"pong"},"done":true,"done_reason":"stop","prompt_eval_count":2,"eval_count":1}`)
	}))
	defer up.Close()
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("llama3", api.Upstream{URL: up.URL})
	e, _ := api.New(api.WithChatCompletionsRemote(reg, api.WithChatModelAdapter("llama3", api.OllamaAdapter())))
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp := chatPost(t, srv.URL, `{"model":"llama3","messages":[{"role":"user","content":"ping"}]}`)
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(out), `"content":"pong"`) || !strings.Contains(string(out), `"object":"chat.completion"`) {
		t.Errorf("ollama not adapted to OpenAI shape: %s", out)
	}
}

func TestChatRemote_AnthropicAdapter_E2E_Good(t *testing.T) {
	var gotVersion string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotVersion = r.Header.Get("anthropic-version")
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"pong"}],"stop_reason":"end_turn","usage":{"input_tokens":2,"output_tokens":1}}`)
	}))
	defer up.Close()
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("claude-3", api.Upstream{URL: up.URL})
	e, _ := api.New(api.WithChatCompletionsRemote(reg, api.WithChatModelAdapter("claude-3", api.AnthropicAdapter())))
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp := chatPost(t, srv.URL, `{"model":"claude-3","messages":[{"role":"user","content":"ping"}]}`)
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if gotVersion != "2023-06-01" {
		t.Errorf("anthropic-version header not sent: %q", gotVersion)
	}
	if !strings.Contains(string(out), `"content":"pong"`) {
		t.Errorf("anthropic not adapted: %s", out)
	}
}
```

- [ ] **Step 6: Run + commit**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run 'TestAnthropicAdapter|TestChatRemote' -race`
Expected: PASS.

```bash
cd /Users/snider/Code/core/api
git add go/chat_adapter_anthropic.go go/chat_adapter_anthropic_test.go go/chat_remote_test.go
git commit -m "$(printf 'feat(api): AnthropicAdapter — OpenAI <-> Anthropic /v1/messages + e2e adapter tests\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 5: Example test + QA gate + final review

**Files:**
- Create: `go/chat_remote_example_test.go`

- [ ] **Step 1: Example test**

Create `go/chat_remote_example_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"fmt"

	api "dappco.re/go/api"
)

func ExampleWithChatCompletionsRemote() {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("10.0.0.0/8"))
	_ = reg.Set("llama3:70b", api.Upstream{URL: "http://10.0.0.5:11434"})
	_ = reg.SetDefault(api.Upstream{URL: "https://llm.lthn.sh"}) // OpenAI-compatible — passthrough

	engine, err := api.New(
		api.WithChatCompletionsRemote(reg,
			api.WithChatModelAdapter("llama3:70b", api.OllamaAdapter()),
		),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(engine.Addr())
	// Output: :8080
}
```

- [ ] **Step 2: Full QA gate**

Run:
```bash
cd /Users/snider/Code/core/api/go
gofmt -l chat_remote.go chat_adapter.go chat_adapter_ollama.go chat_adapter_anthropic.go chat_completions.go chat_remote_test.go chat_remote_internal_test.go chat_adapter_ollama_test.go chat_adapter_anthropic_test.go chat_remote_example_test.go
GOWORK=off go vet ./
GOWORK=off go test ./ -race -count=1
GOWORK=off go build -o /dev/null ./cmd/gateway/
```
Expected: `gofmt -l` empty; vet clean; full suite PASS under `-race`; gateway builds.

- [ ] **Step 3: gosec**

Run: `cd /Users/snider/Code/core/api/go && gosec -quiet ./ 2>/dev/null | tail -5 || echo "gosec unavailable"`
Expected: no new findings in the chat_* files (no `#nosec` needed — the SSRF-bypass annotation lives in `upstream_transport.go`, reused unchanged).

- [ ] **Step 4: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/chat_remote_example_test.go
git commit -m "$(printf 'test(api): ExampleWithChatCompletionsRemote + QA gate\n\nCo-Authored-By: Virgil <virgil@lethean.io>')" || echo "nothing to commit"
```

---

## Spec coverage check

| Spec section | Task |
|---|---|
| §4 `WithChatCompletionsRemote`, `WithChatModelAdapter`, failover/transport opts, `WithChatCompletionsAllowRemoteClients` | Task 2 |
| §4 `ChatFormatAdapter`/`ChatStreamTranscoder`/`ChatStreamMeta` | Task 2 |
| §5 dispatch flow (local-first, pure-local unchanged, remote, 404) | Task 2 |
| §5.1 `ModelResolver.Knows` | Task 1 |
| §5.2 deliver via gin `c.Writer` | Task 2 |
| §6.1 OllamaAdapter (request/non-stream/stream) | Task 3 |
| §6.2 AnthropicAdapter (request/non-stream/stream) | Task 4 |
| §7 bind opt-in + error taxonomy | Tasks 2 (bind, errors), 3/4 (adapter errors) |
| §8 testing matrix | Tasks 1–5 |
| §9 file layout | all |

**Deferred per spec §10 (not in this plan):** generic transcoder registry, tool-calling translation, more adapters, per-model rate limiting, OpenAPI describability.
