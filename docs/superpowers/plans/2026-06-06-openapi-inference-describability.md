# OpenAPI Describability for the Inference Surface — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Surface the remote/hybrid chat-completions endpoint and the `WithUpstreamRouter` mounted paths in the generated OpenAPI spec (and therefore SDK gen), which today omits both.

**Architecture:** Extend the existing special-cased-path mechanism in the spec builder — no new abstraction. Widen `ChatCompletionsEnabled` to fire for a remote backend too, and add an `UpstreamRouterPaths` field that flows engine → `TransportConfig` → `SpecBuilder`, where `Build()` emits a minimal honest `POST` proxy item per path, deduped so real items always win.

**Tech Stack:** Go 1.26, the existing `openapi.go` SpecBuilder + `transport.go` + `spec_builder_helper.go`. Spec: `docs/superpowers/specs/2026-06-06-openapi-inference-describability-design.md`.

**Conventions:** SPDX header on new files. UK English. `_Good/_Bad/_Ugly` test suffixes. Run from `core/api/go` with `GOWORK=off go test ./ ...`. Commit `Co-Authored-By: Virgil <virgil@lethean.io>`.

**Reused symbols (already in package `api` — do NOT redefine):** `SpecBuilder`, `(*Engine).OpenAPISpecBuilder()`, `(*SpecBuilder).Build([]RouteGroup)`, `chatCompletionsPathItem`, `openAPISpecPathItem`, `normaliseOpenAPIPath`, `isPublicPathForList`, `makePathItemPublic`, `operationID`, `mergeHeaders`, `standardResponseHeaders`, `rateLimitSuccessHeaders`, `mimeJSON`. Engine fields `e.chatCompletionsResolver`, `e.chatRemote` (set by `WithChatCompletionsRemote`), `e.upstreamRouter` (set by `WithUpstreamRouter`; has a `paths []string` field). `TransportConfig` + `(*Engine).TransportConfig()` in `transport.go`. Test pattern: `e.OpenAPISpecBuilder().Build(nil)` → JSON bytes (see `spec_builder_helper_test.go`).

---

## File Structure

| File | Change |
|------|--------|
| `go/transport.go` | `ChatCompletionsEnabled` fires for `e.chatRemote` too; new `UpstreamRouterPaths []string` field + population |
| `go/spec_builder_helper.go` | `builder.UpstreamRouterPaths = runtime.Transport.UpstreamRouterPaths` |
| `go/openapi.go` | `SpecBuilder.UpstreamRouterPaths`; `upstreamRouterPathItem()`; `Build()` router-path loop with dedup |
| `go/openapi_inference_test.go` | new — describability tests (`package api_test`) |

---

## Task 1: Chat-completions describability (local / remote / hybrid)

**Files:**
- Modify: `go/transport.go:53`
- Test: `go/openapi_inference_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `go/openapi_inference_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"testing"

	api "dappco.re/go/api"
)

// specPaths builds the engine's OpenAPI spec and returns its "paths" object.
func specPaths(t *testing.T, e *api.Engine) map[string]any {
	t.Helper()
	data, err := e.OpenAPISpecBuilder().Build(nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatalf("spec has no paths object")
	}
	return paths
}

// postTags returns the tags of the POST operation at path, or nil.
func postTags(paths map[string]any, path string) []string {
	item, ok := paths[path].(map[string]any)
	if !ok {
		return nil
	}
	post, ok := item["post"].(map[string]any)
	if !ok {
		return nil
	}
	raw, _ := post["tags"].([]any)
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		if s, ok := t.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func TestOpenAPISpec_ChatCompletions_RemoteOnly_Good(t *testing.T) {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.SetDefault(api.Upstream{URL: "http://127.0.0.1:11434"}); err != nil {
		t.Fatal(err)
	}
	e, err := api.New(api.WithChatCompletionsRemote(reg))
	if err != nil {
		t.Fatal(err)
	}
	paths := specPaths(t, e)
	if !hasTag(postTags(paths, "/v1/chat/completions"), "inference") {
		t.Fatalf("remote-only chat endpoint missing/untagged in spec; paths present: %v", keysOf(paths))
	}
}

func TestOpenAPISpec_ChatCompletions_Absent_Good(t *testing.T) {
	e, err := api.New() // neither local nor remote chat configured
	if err != nil {
		t.Fatal(err)
	}
	paths := specPaths(t, e)
	if _, exists := paths["/v1/chat/completions"]; exists {
		t.Fatalf("chat endpoint present in spec with no chat configured")
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestOpenAPISpec_ChatCompletions`
Expected: `TestOpenAPISpec_ChatCompletions_RemoteOnly_Good` FAILS (chat path absent — `ChatCompletionsEnabled` is false for remote-only); `_Absent_Good` passes.

- [ ] **Step 3: Widen the enabling condition**

In `go/transport.go`, in `TransportConfig()`, change line 53 from:

```go
		ChatCompletionsEnabled: e.chatCompletionsResolver != nil,
```
to:
```go
		ChatCompletionsEnabled: e.chatCompletionsResolver != nil || e.chatRemote != nil,
```

The `ChatCompletionsPath` resolution (line 73-75) already fires for `core.Trim(e.chatCompletionsPath) != ""`, and `New()` sets the default path when a resolver OR remote backend is configured, so the path is already correct — only the enabled flag needed widening.

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestOpenAPISpec_ChatCompletions -race`
Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/transport.go go/openapi_inference_test.go
git commit -m "$(printf 'feat(api): OpenAPI spec includes chat-completions for remote/hybrid backends\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 2: Upstream router path items

**Files:**
- Modify: `go/transport.go` (struct + population), `go/spec_builder_helper.go`, `go/openapi.go`
- Test: `go/openapi_inference_test.go` (extend)

- [ ] **Step 1: Write the failing tests**

Append to `go/openapi_inference_test.go`:

```go
func TestOpenAPISpec_RouterPaths_Good(t *testing.T) {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.SetDefault(api.Upstream{URL: "http://127.0.0.1:11434"}); err != nil {
		t.Fatal(err)
	}
	e, err := api.New(api.WithUpstreamRouter(reg, api.WithRouterPaths("/v1/embeddings", "/v1/score")))
	if err != nil {
		t.Fatal(err)
	}
	paths := specPaths(t, e)
	for _, p := range []string{"/v1/embeddings", "/v1/score"} {
		if !hasTag(postTags(paths, p), "proxy") {
			t.Fatalf("router path %s missing/untagged in spec; paths: %v", p, keysOf(paths))
		}
		item := paths[p].(map[string]any)
		post := item["post"].(map[string]any)
		responses := post["responses"].(map[string]any)
		for _, code := range []string{"404", "503"} {
			if _, ok := responses[code]; !ok {
				t.Errorf("router path %s missing %s response", p, code)
			}
		}
	}
}

func TestOpenAPISpec_RouterDedupChat_Ugly(t *testing.T) {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.SetDefault(api.Upstream{URL: "http://127.0.0.1:11434"}); err != nil {
		t.Fatal(err)
	}
	// Router mounted at the default chat path AND chat enabled (remote).
	e, err := api.New(
		api.WithChatCompletionsRemote(reg),
		api.WithUpstreamRouter(reg), // default WithRouterPaths == /v1/chat/completions
	)
	if err != nil {
		t.Fatal(err)
	}
	paths := specPaths(t, e)
	tags := postTags(paths, "/v1/chat/completions")
	if !hasTag(tags, "inference") {
		t.Fatalf("chat path lost its inference item to the proxy dedup; tags=%v", tags)
	}
	if hasTag(tags, "proxy") {
		t.Fatalf("chat path was clobbered by the proxy item; tags=%v", tags)
	}
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestOpenAPISpec_Router`
Expected: `_RouterPaths_Good` FAILS (router paths absent). `_RouterDedupChat_Ugly` passes already (no proxy item exists yet, so the chat item is intact) — it locks in the dedup once Step 3 lands.

- [ ] **Step 3: Add the `UpstreamRouterPaths` field + population (`transport.go`)**

In `go/transport.go`, add to the `TransportConfig` struct (after `OpenAPISpecPath string`):

```go
	UpstreamRouterPaths []string
```

In `TransportConfig()`, after the `cfg.OpenAPISpecPath` block (around line 78), add:

```go
	if e.upstreamRouter != nil {
		cfg.UpstreamRouterPaths = append([]string(nil), e.upstreamRouter.paths...)
	}
```

- [ ] **Step 4: Pass it into the builder (`spec_builder_helper.go`)**

In `go/spec_builder_helper.go`, after `builder.OpenAPISpecPath = runtime.Transport.OpenAPISpecPath` (line 84), add:

```go
	builder.UpstreamRouterPaths = runtime.Transport.UpstreamRouterPaths
```

- [ ] **Step 5: Add the SpecBuilder field + the path item + the Build loop (`openapi.go`)**

In `go/openapi.go`, add to the `SpecBuilder` struct (after `OpenAPISpecPath string`):

```go
	UpstreamRouterPaths []string
```

Add the path-item builder (place it next to `openAPISpecPathItem`):

```go
// upstreamRouterPathItem documents a WithUpstreamRouter mounted path as a
// minimal, honest POST proxy operation. The router proxies arbitrary shapes by
// selector key, so request/response schemas are generic by design; the path is
// tagged "proxy" to distinguish it from the typed "inference" chat endpoint.
func upstreamRouterPathItem(path string, operationIDs map[string]int) map[string]any {
	successHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	errorHeaders := mergeHeaders(standardResponseHeaders(), rateLimitSuccessHeaders())
	genericObject := func() map[string]any {
		return map[string]any{"type": "object", "additionalProperties": true}
	}

	return map[string]any{
		"post": map[string]any{
			"summary":     "Upstream router (selector-routed proxy)",
			"description": "Selector-routed reverse proxy. The request body must carry the selector field (default \"model\"); the concrete request and response schemas depend on the target upstream/model. Streams Server-Sent Events when the upstream does.",
			"tags":        []string{"proxy"},
			"operationId": operationID("post", path, operationIDs),
			"requestBody": map[string]any{
				"required": true,
				"content": map[string]any{
					mimeJSON: map[string]any{"schema": genericObject()},
				},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Proxied upstream response",
					"content": map[string]any{
						mimeJSON:            map[string]any{"schema": genericObject()},
						"text/event-stream": map[string]any{"schema": map[string]any{"type": "string"}},
					},
					"headers": successHeaders,
				},
				"404": map[string]any{
					"description": "No upstream registered for the selector key",
					"content":     map[string]any{mimeJSON: map[string]any{"schema": genericObject()}},
					"headers":     errorHeaders,
				},
				"503": map[string]any{
					"description": "All upstreams unavailable",
					"content":     map[string]any{mimeJSON: map[string]any{"schema": genericObject()}},
					"headers": mergeHeaders(errorHeaders, map[string]any{
						"Retry-After": map[string]any{
							"description": "Seconds to wait before retrying.",
							"schema":      map[string]any{"type": "integer"},
						},
					}),
				},
			},
		},
	}
}
```

In `Build()`, **immediately after the `for _, g := range groups { ... }` loop closes** (so the dedup covers group-contributed paths too), add:

```go
	for _, rawPath := range sb.UpstreamRouterPaths {
		routerPath := normaliseOpenAPIPath(rawPath)
		if routerPath == "" {
			continue
		}
		if _, exists := paths[routerPath]; exists {
			continue // a real item (chat, spec, swagger, or a group) already documents this path
		}
		item := upstreamRouterPathItem(routerPath, operationIDs)
		if isPublicPathForList(routerPath, publicPaths) {
			makePathItemPublic(item)
		}
		paths[routerPath] = item
	}
```

> Note: `upstreamRouterPathItem` does NOT hard-code `"security"`. Public paths get `makePathItemPublic` applied (matching the other items); non-public paths inherit the document's global security — which is the intended "honour configured public paths, don't force-public" behaviour from spec §3.2.

- [ ] **Step 6: Run to verify it passes**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestOpenAPISpec -race`
Expected: all 4 PASS (`_RemoteOnly_Good`, `_Absent_Good`, `_RouterPaths_Good`, `_RouterDedupChat_Ugly`).

- [ ] **Step 7: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/transport.go go/spec_builder_helper.go go/openapi.go go/openapi_inference_test.go
git commit -m "$(printf 'feat(api): OpenAPI spec documents WithUpstreamRouter paths (deduped proxy items)\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 3: QA gate + final review

**Files:** none (verification only)

- [ ] **Step 1: Full QA gate**

Run:
```bash
cd /Users/snider/Code/core/api/go
gofmt -l transport.go openapi.go spec_builder_helper.go openapi_inference_test.go
GOWORK=off go vet ./
GOWORK=off go test ./ -race -count=1
GOWORK=off go build -o /dev/null ./cmd/gateway/
```
Expected: `gofmt -l` empty; vet clean; full suite PASS under `-race` (no regression to the ~1686 existing tests, esp. the existing `openapi_test.go` / `spec_builder_helper_test.go`); gateway builds.

- [ ] **Step 2: OpenAPI 3.1 validity sanity**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run 'TestSpec|TestOpenAPI|TestSwagger' -count=1`
Expected: PASS — the existing spec-shape/validity tests still hold with the new path items present.

- [ ] **Step 3: Commit any formatting fixes**

```bash
cd /Users/snider/Code/core/api
git add -A go/ && git commit -m "$(printf 'chore(api): gofmt pass for inference describability\n\nCo-Authored-By: Virgil <virgil@lethean.io>')" || echo "nothing to commit"
```

---

## Spec coverage check

| Spec section | Task |
|---|---|
| §3.1 chat-completions for local/remote/hybrid | Task 1 |
| §3.2 minimal router proxy item (POST, `proxy` tag, generic schema, 404/503+Retry-After) | Task 2 (Step 5) |
| §3.3 dedup (real items win; router-at-chat-path → inference item) | Task 2 (Build loop + `_RouterDedupChat_Ugly`) |
| §4 wiring (transport → spec_builder_helper → openapi.go) | Tasks 1, 2 |
| §5 testing matrix | Tasks 1, 2 (+ §5 OpenAPI-validity reuse in Task 3) |
| §6 file layout | all |

**Deferred per spec §7 (not in this plan):** real per-path schemas via consumer `RouteDescription`s, per-model enumeration, broader un-described-route sweep.
