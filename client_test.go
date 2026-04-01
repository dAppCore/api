// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"fmt"
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
