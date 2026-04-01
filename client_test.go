// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	api "dappco.re/go/core/api"
)

func TestOpenAPIClient_Good_CallOperationByID(t *testing.T) {
	errCh := make(chan error, 2)
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errCh <- fmt.Errorf("expected GET, got %s", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if got := r.URL.Query().Get("name"); got != "Ada" {
			errCh <- fmt.Errorf("expected query name=Ada, got %q", got)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"message":"hello"}}`))
	})
	mux.HandleFunc("/users/123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errCh <- fmt.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if got := r.URL.Query().Get("verbose"); got != "true" {
			errCh <- fmt.Errorf("expected query verbose=true, got %q", got)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"123","name":"Ada"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	specPath := writeTempSpec(t, `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /hello:
    get:
      operationId: get_hello
  /users/{id}:
    post:
      operationId: update_user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
`)

	client := api.NewOpenAPIClient(
		api.WithSpec(specPath),
		api.WithBaseURL(srv.URL),
	)

	result, err := client.Call("get_hello", map[string]any{
		"name": "Ada",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}

	hello, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if hello["message"] != "hello" {
		t.Fatalf("expected message=hello, got %#v", hello["message"])
	}

	result, err = client.Call("update_user", map[string]any{
		"path": map[string]any{
			"id": "123",
		},
		"query": map[string]any{
			"verbose": true,
		},
		"body": map[string]any{
			"name": "Ada",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}

	updated, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if updated["id"] != "123" {
		t.Fatalf("expected id=123, got %#v", updated["id"])
	}
	if updated["name"] != "Ada" {
		t.Fatalf("expected name=Ada, got %#v", updated["name"])
	}
}

func TestOpenAPIClient_Good_CallHeadOperationWithRequestBody(t *testing.T) {
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/head", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			errCh <- fmt.Errorf("expected HEAD, got %s", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if got := r.URL.RawQuery; got != "" {
			errCh <- fmt.Errorf("expected no query string, got %q", got)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			errCh <- fmt.Errorf("read body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if string(body) != `{"name":"Ada"}` {
			errCh <- fmt.Errorf("expected JSON body {\"name\":\"Ada\"}, got %q", string(body))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	specPath := writeTempSpec(t, `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /head:
    head:
      operationId: head_check
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
`)

	client := api.NewOpenAPIClient(
		api.WithSpec(specPath),
		api.WithBaseURL(srv.URL),
	)

	result, err := client.Call("head_check", map[string]any{
		"name": "Ada",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
	if result != nil {
		t.Fatalf("expected nil result for empty HEAD response body, got %T", result)
	}
}

func TestOpenAPIClient_Good_CallOperationWithRepeatedQueryValues(t *testing.T) {
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errCh <- fmt.Errorf("expected GET, got %s", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if got := r.URL.Query()["tag"]; len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
			errCh <- fmt.Errorf("expected repeated tag values [alpha beta], got %v", got)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			errCh <- fmt.Errorf("expected page=2, got %q", got)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"ok":true}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	specPath := writeTempSpec(t, `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /search:
    get:
      operationId: search_items
`)

	client := api.NewOpenAPIClient(
		api.WithSpec(specPath),
		api.WithBaseURL(srv.URL),
	)

	result, err := client.Call("search_items", map[string]any{
		"tag":  []string{"alpha", "beta"},
		"page": 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}

	decoded, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if okValue, ok := decoded["ok"].(bool); !ok || !okValue {
		t.Fatalf("expected ok=true, got %#v", decoded["ok"])
	}
}

func TestOpenAPIClient_Good_UsesTopLevelQueryParametersOnPost(t *testing.T) {
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errCh <- fmt.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if got := r.URL.Query().Get("verbose"); got != "true" {
			errCh <- fmt.Errorf("expected query verbose=true, got %q", got)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			errCh <- fmt.Errorf("read body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if string(body) != `{"name":"Ada"}` {
			errCh <- fmt.Errorf("expected JSON body {\"name\":\"Ada\"}, got %q", string(body))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"ok":true}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	specPath := writeTempSpec(t, `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /submit:
    post:
      operationId: submit_item
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
      parameters:
        - name: verbose
          in: query
`)

	client := api.NewOpenAPIClient(
		api.WithSpec(specPath),
		api.WithBaseURL(srv.URL),
	)

	result, err := client.Call("submit_item", map[string]any{
		"verbose": true,
		"name":    "Ada",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}

	decoded, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if okValue, ok := decoded["ok"].(bool); !ok || !okValue {
		t.Fatalf("expected ok=true, got %#v", decoded["ok"])
	}
}

func TestOpenAPIClient_Good_UsesHeaderAndCookieParameters(t *testing.T) {
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/inspect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errCh <- fmt.Errorf("expected GET, got %s", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if got := r.Header.Get("X-Trace-ID"); got != "trace-123" {
			errCh <- fmt.Errorf("expected X-Trace-ID=trace-123, got %q", got)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if got := r.Header.Get("X-Custom-Header"); got != "custom-value" {
			errCh <- fmt.Errorf("expected X-Custom-Header=custom-value, got %q", got)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		session, err := r.Cookie("session_id")
		if err != nil {
			errCh <- fmt.Errorf("expected session_id cookie: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if session.Value != "cookie-123" {
			errCh <- fmt.Errorf("expected session_id=cookie-123, got %q", session.Value)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pref, err := r.Cookie("pref")
		if err != nil {
			errCh <- fmt.Errorf("expected pref cookie: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if pref.Value != "dark" {
			errCh <- fmt.Errorf("expected pref=dark, got %q", pref.Value)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"ok":true}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	specPath := writeTempSpec(t, `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /inspect:
    get:
      operationId: inspect_request
      parameters:
        - name: X-Trace-ID
          in: header
        - name: session_id
          in: cookie
`)

	client := api.NewOpenAPIClient(
		api.WithSpec(specPath),
		api.WithBaseURL(srv.URL),
	)

	result, err := client.Call("inspect_request", map[string]any{
		"X-Trace-ID": "trace-123",
		"session_id": "cookie-123",
		"header": map[string]any{
			"X-Custom-Header": "custom-value",
		},
		"cookie": map[string]any{
			"pref": "dark",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}

	decoded, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if okValue, ok := decoded["ok"].(bool); !ok || !okValue {
		t.Fatalf("expected ok=true, got %#v", decoded["ok"])
	}
}

func TestOpenAPIClient_Good_UsesFirstAbsoluteServer(t *testing.T) {
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errCh <- fmt.Errorf("expected GET, got %s", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"message":"hello"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	specPath := writeTempSpec(t, `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: " `+srv.URL+` "
  - url: /
  - url: " `+srv.URL+` "
paths:
  /hello:
    get:
      operationId: get_hello
`)

	client := api.NewOpenAPIClient(
		api.WithSpec(specPath),
	)

	result, err := client.Call("get_hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}

	hello, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if hello["message"] != "hello" {
		t.Fatalf("expected message=hello, got %#v", hello["message"])
	}
}

func TestOpenAPIClient_Bad_ValidatesRequestBodyAgainstSchema(t *testing.T) {
	called := make(chan struct{}, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		called <- struct{}{}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"123"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	specPath := writeTempSpec(t, `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      operationId: create_user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                  data:
                    type: object
                    properties:
                      id:
                        type: string
`)

	client := api.NewOpenAPIClient(
		api.WithSpec(specPath),
		api.WithBaseURL(srv.URL),
	)

	if _, err := client.Call("create_user", map[string]any{
		"body": map[string]any{},
	}); err == nil {
		t.Fatal("expected request body validation error, got nil")
	}

	select {
	case <-called:
		t.Fatal("expected request validation to fail before the HTTP call")
	default:
	}
}

func TestOpenAPIClient_Bad_ValidatesResponseAgainstSchema(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":123}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	specPath := writeTempSpec(t, `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: list_users
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                required: [success, data]
                properties:
                  success:
                    type: boolean
                  data:
                    type: object
                    required: [id]
                    properties:
                      id:
                        type: string
`)

	client := api.NewOpenAPIClient(
		api.WithSpec(specPath),
		api.WithBaseURL(srv.URL),
	)

	if _, err := client.Call("list_users", nil); err == nil {
		t.Fatal("expected response validation error, got nil")
	}
}

func TestOpenAPIClient_Bad_MissingOperation(t *testing.T) {
	specPath := writeTempSpec(t, `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths: {}
`)

	client := api.NewOpenAPIClient(
		api.WithSpec(specPath),
		api.WithBaseURL("http://example.invalid"),
	)

	if _, err := client.Call("missing", nil); err == nil {
		t.Fatal("expected error for missing operation, got nil")
	}
}

func writeTempSpec(t *testing.T, contents string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "openapi.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	return path
}
