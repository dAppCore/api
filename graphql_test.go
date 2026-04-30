// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"context"
	core "dappco.re/go"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"

	api "dappco.re/go/api"
)

// newTestSchema creates a minimal ExecutableSchema that responds to { name }
// with {"name":"test"}. This avoids importing gqlgen's internal testserver
// while providing a realistic schema for handler tests.
func newTestSchema() graphql.ExecutableSchema {
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: `
		type Query {
			name: String!
		}
	`})

	return &graphql.ExecutableSchemaMock{
		SchemaFunc: func() *ast.Schema {
			return schema
		},
		ExecFunc: func(ctx context.Context) graphql.ResponseHandler {
			ran := false
			return func(ctx context.Context) *graphql.Response {
				if ran {
					return nil
				}
				ran = true
				return &graphql.Response{Data: []byte(`{"name":"test"}`)}
			}
		},
		ComplexityFunc: func(_ context.Context, _, _ string, childComplexity int, _ map[string]any) (int, bool) {
			return childComplexity, true
		},
	}
}

// ── GraphQL endpoint ──────────────────────────────────────────────────

func TestWithGraphQL_Good_EndpointResponds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithGraphQL(newTestSchema()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	body := `{"query":"{ name }"}`
	resp, err := http.Post(srv.URL+"/graphql", "application/json", core.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if !core.Contains(string(respBody), `"name":"test"`) {
		t.Fatalf("expected response containing name:test, got %q", string(respBody))
	}
}

func TestWithGraphQL_Good_PlaygroundServesHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithGraphQL(newTestSchema(), api.WithPlayground()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/graphql/playground")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !core.Contains(ct, "text/html") {
		t.Fatalf("expected Content-Type containing text/html, got %q", ct)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if !core.Contains(string(body), "GraphQL") {
		t.Fatalf("expected playground HTML containing 'GraphQL', got %q", string(body)[:200])
	}
}

func TestWithGraphQL_Good_NoPlaygroundByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Without WithPlayground(), /graphql/playground should return 404.
	e, err := api.New(api.WithGraphQL(newTestSchema()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/graphql/playground", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /graphql/playground without WithPlayground, got %d", w.Code)
	}
}

func TestWithGraphQL_Good_CustomPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithGraphQL(newTestSchema(), api.WithGraphQLPath("/gql"), api.WithPlayground()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	// Query endpoint should be at /gql.
	body := `{"query":"{ name }"}`
	resp, err := http.Post(srv.URL+"/gql", "application/json", core.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 at /gql, got %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if !core.Contains(string(respBody), `"name":"test"`) {
		t.Fatalf("expected response containing name:test, got %q", string(respBody))
	}

	// Playground should be at /gql/playground.
	pgResp, err := http.Get(srv.URL + "/gql/playground")
	if err != nil {
		t.Fatalf("playground request failed: %v", err)
	}
	defer pgResp.Body.Close()

	if pgResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 at /gql/playground, got %d", pgResp.StatusCode)
	}

	// The default path should not exist.
	defaultResp, err := http.Post(srv.URL+"/graphql", "application/json", core.NewReader(body))
	if err != nil {
		t.Fatalf("default path request failed: %v", err)
	}
	defer defaultResp.Body.Close()

	if defaultResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 at /graphql when custom path is /gql, got %d", defaultResp.StatusCode)
	}
}

func TestWithGraphQL_Good_NormalisesCustomPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithGraphQL(newTestSchema(), api.WithGraphQLPath(" /gql/ "), api.WithPlayground()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	body := `{"query":"{ name }"}`
	resp, err := http.Post(srv.URL+"/gql", "application/json", core.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 at normalised /gql, got %d", resp.StatusCode)
	}

	pgResp, err := http.Get(srv.URL + "/gql/playground")
	if err != nil {
		t.Fatalf("playground request failed: %v", err)
	}
	defer pgResp.Body.Close()

	if pgResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 at normalised /gql/playground, got %d", pgResp.StatusCode)
	}
}

func TestWithGraphQL_Good_DefaultPathWhenEmptyCustomPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithGraphQL(newTestSchema(), api.WithGraphQLPath(""), api.WithPlayground()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	body := `{"query":"{ name }"}`
	resp, err := http.Post(srv.URL+"/graphql", "application/json", core.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 at default /graphql, got %d", resp.StatusCode)
	}

	pgResp, err := http.Get(srv.URL + "/graphql/playground")
	if err != nil {
		t.Fatalf("playground request failed: %v", err)
	}
	defer pgResp.Body.Close()

	if pgResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 at default /graphql/playground, got %d", pgResp.StatusCode)
	}
}

func TestWithGraphQL_Ugly_RootPathFallsBackToDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(api.WithGraphQL(newTestSchema(), api.WithGraphQLPath(" / "), api.WithPlayground()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	body := `{"query":"{ name }"}`
	resp, err := http.Post(srv.URL+"/graphql", "application/json", core.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 at default /graphql after root path normalisation, got %d", resp.StatusCode)
	}

	pgResp, err := http.Get(srv.URL + "/graphql/playground")
	if err != nil {
		t.Fatalf("playground request failed: %v", err)
	}
	defer pgResp.Body.Close()

	if pgResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 at default /graphql/playground after root path normalisation, got %d", pgResp.StatusCode)
	}
}

func TestWithGraphQL_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	e, err := api.New(
		api.WithRequestID(),
		api.WithGraphQL(newTestSchema()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	body := `{"query":"{ name }"}`
	resp, err := http.Post(srv.URL+"/graphql", "application/json", core.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// RequestID middleware should have injected the header.
	reqID := resp.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Fatal("expected X-Request-ID header from RequestID middleware")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if !core.Contains(string(respBody), `"name":"test"`) {
		t.Fatalf("expected response containing name:test, got %q", string(respBody))
	}
}
