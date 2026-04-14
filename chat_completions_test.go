// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// TestChatCompletions_WithChatCompletions_Good verifies that WithChatCompletions
// mounts the endpoint and unknown model names produce a 404 body conforming to
// RFC §11.7.
func TestChatCompletions_WithChatCompletions_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := api.NewModelResolver()
	engine, err := api.New(api.WithChatCompletions(resolver))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{
		"model": "missing-model",
		"messages": [{"role":"user","content":"hi"}]
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	engine.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d (body=%s)", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Param   string `json:"param"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if payload.Error.Code != "model_not_found" {
		t.Fatalf("expected code=model_not_found, got %q", payload.Error.Code)
	}
	if payload.Error.Type != "model_not_found" {
		t.Fatalf("expected type=model_not_found, got %q", payload.Error.Type)
	}
	if payload.Error.Param != "model" {
		t.Fatalf("expected param=model, got %q", payload.Error.Param)
	}
}

// TestChatCompletions_WithChatCompletionsPath_Good verifies the custom mount path override.
func TestChatCompletions_WithChatCompletionsPath_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := api.NewModelResolver()
	engine, err := api.New(
		api.WithChatCompletions(resolver),
		api.WithChatCompletionsPath("/chat"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{
		"model": "missing-model",
		"messages": [{"role":"user","content":"hi"}]
	}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	engine.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

// TestChatCompletions_ValidateRequest_Bad verifies that missing messages produces a 400.
func TestChatCompletions_ValidateRequest_Bad(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := api.NewModelResolver()
	engine, _ := api.New(api.WithChatCompletions(resolver))

	cases := []struct {
		name string
		body string
		code string
	}{
		{
			name: "missing-messages",
			body: `{"model":"lemer"}`,
			code: "invalid_request_error",
		},
		{
			name: "missing-model",
			body: `{"messages":[{"role":"user","content":"hi"}]}`,
			code: "invalid_request_error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			engine.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d (body=%s)", rec.Code, rec.Body.String())
			}

			var payload struct {
				Error struct {
					Type string `json:"type"`
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("invalid JSON body: %v", err)
			}
			if payload.Error.Type != tc.code {
				t.Fatalf("expected type=%q, got %q", tc.code, payload.Error.Type)
			}
		})
	}
}

// TestChatCompletions_NoResolver_Ugly verifies graceful handling when an engine
// is constructed WITHOUT a resolver — no route is mounted.
func TestChatCompletions_NoResolver_Ugly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, _ := api.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	engine.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when no resolver is configured, got %d", rec.Code)
	}
}
