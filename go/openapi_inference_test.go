// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"testing"

	api "dappco.re/go/api"
)

// specObject builds the engine's OpenAPI spec and returns the whole document.
func specObject(t *testing.T, e *api.Engine) map[string]any {
	t.Helper()
	data, err := e.OpenAPISpecBuilder().Build(nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}
	return spec
}

// specPaths builds the engine's OpenAPI spec and returns its "paths" object.
func specPaths(t *testing.T, e *api.Engine) map[string]any {
	t.Helper()
	paths, ok := specObject(t, e)["paths"].(map[string]any)
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
	spec := specObject(t, e)

	// The typed chat path must surface for a remote-only backend.
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatalf("spec has no paths object")
	}
	if !hasTag(postTags(paths, "/v1/chat/completions"), "inference") {
		t.Fatalf("remote-only chat endpoint missing/untagged in spec; paths present: %v", keysOf(paths))
	}

	// The top-level capability flag must report chat as enabled. This is the
	// load-bearing assertion for ChatCompletionsEnabled honouring e.chatRemote:
	// it fails without the transport.go change and passes with it.
	if enabled, _ := spec["x-chat-completions-enabled"].(bool); !enabled {
		t.Fatalf("x-chat-completions-enabled missing/false for a remote-only chat engine")
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
