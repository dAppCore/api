// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	api "dappco.re/go/api"
)

func TestEntitlementBridge_Good_CallbackChecksWorkspaceEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/42/entitlements/check/premium.feature" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer user-token" {
			t.Fatalf("expected forwarded Authorization header, got %q", got)
		}
		if got := r.Header.Get("Cookie"); got != "" {
			t.Fatalf("expected cookie to be omitted when Authorization is set, got %q", got)
		}
		if got := r.Header.Get("X-Workspace-Id"); got != "42" {
			t.Fatalf("expected workspace header 42, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"workspace_id":42,"feature":"premium.feature","entitlement":{"allowed":true}}`))
	}))
	defer srv.Close()

	bridge := api.NewEntitlementBridge(api.EntitlementBridgeConfig{BaseURL: srv.URL})
	headers := http.Header{}
	headers.Set("Authorization", "Bearer user-token")
	headers.Set("Cookie", "session=abc")

	callback := bridge.Callback(context.Background(), "42", headers)
	if !callback("premium.feature") {
		t.Fatal("expected entitlement callback to allow feature")
	}
}

func TestEntitlementBridge_Good_CallbackForRequestUsesCurrentWorkspaceEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/entitlements/check/mcp" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Cookie"); got != "session=abc" {
			t.Fatalf("expected forwarded cookie, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entitlement":{"can":true}}`))
	}))
	defer srv.Close()

	bridge := api.NewEntitlementBridge(api.EntitlementBridgeConfig{BaseURL: srv.URL})
	req := httptest.NewRequest(http.MethodGet, "/render", nil)
	req.Header.Set("Cookie", "session=abc")

	if !bridge.CallbackForRequest(req, "")("mcp") {
		t.Fatal("expected current-workspace entitlement callback to allow feature")
	}
}

func TestEntitlementBridge_Bad_FailsClosedOnServiceError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "broken", http.StatusInternalServerError)
	}))
	defer srv.Close()

	bridge := api.NewEntitlementBridge(api.EntitlementBridgeConfig{BaseURL: srv.URL})
	allowed, err := bridge.Check(context.Background(), "42", "premium.feature", nil)
	if err == nil {
		t.Fatal("expected service error")
	}
	if allowed {
		t.Fatal("expected direct check to fail closed")
	}
	if bridge.Callback(context.Background(), "42", nil)("premium.feature") {
		t.Fatal("expected callback to fail closed")
	}
}
