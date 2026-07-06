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
		// Not public and no special-cased auth: the proxy POST is a network
		// gateway under engine auth, so SDK gen must emit an authenticated
		// client. Assert the operation carries a non-empty security
		// requirement (bearerAuth, mirroring the GraphQL/group-loop items).
		security, ok := post["security"].([]any)
		if !ok || len(security) == 0 {
			t.Errorf("router path %s proxy POST missing/empty security; got %v", p, post["security"])
		}
	}
}

func TestOpenAPISpec_RouterDedupSpecPath_Good(t *testing.T) {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.SetDefault(api.Upstream{URL: "http://127.0.0.1:11434"}); err != nil {
		t.Fatal(err)
	}
	// Router mounted at the OpenAPI spec path: the real spec GET item must win,
	// and the proxy POST must be skipped by the dedup.
	e, err := api.New(
		api.WithOpenAPISpecPath("/v1/openapi.json"),
		api.WithUpstreamRouter(reg, api.WithRouterPaths("/v1/openapi.json")),
	)
	if err != nil {
		t.Fatal(err)
	}
	paths := specPaths(t, e)
	item, ok := paths["/v1/openapi.json"].(map[string]any)
	if !ok {
		t.Fatalf("spec path missing from paths; paths: %v", keysOf(paths))
	}
	if _, ok := item["get"].(map[string]any); !ok {
		t.Errorf("spec path lost its real GET item to the proxy dedup; item: %v", keysOf(item))
	}
	if _, ok := item["post"]; ok {
		t.Errorf("spec path was clobbered by the proxy POST item; item: %v", keysOf(item))
	}
}
