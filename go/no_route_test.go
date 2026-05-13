// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestWithNoRoute_Good_FiresOnUnknownPath verifies the canonical SPA-host
// path: a handler set via WithNoRoute catches a request whose path doesn't
// match any registered route, and the response body is whatever the
// handler writes (not gin's default 404 body).
func TestWithNoRoute_Good_FiresOnUnknownPath(t *testing.T) {
	const payload = "SPA-INDEX"

	e, err := New(WithNoRoute(func(c *gin.Context) {
		c.String(http.StatusOK, payload)
	}))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent/spa/route", nil)
	e.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Body.String(); got != payload {
		t.Fatalf("body = %q, want %q", got, payload)
	}
}

// TestWithNoRoute_Bad_DoesNotShadowExplicitRoutes verifies the registration
// order — an explicit route handler must win over the NoRoute fallback even
// when both could plausibly serve the request.
func TestWithNoRoute_Bad_DoesNotShadowExplicitRoutes(t *testing.T) {
	e, err := New(WithNoRoute(func(c *gin.Context) {
		c.String(http.StatusOK, "fallback")
	}))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	// /health is always registered by build().
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	e.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Body.String(); got == "fallback" {
		t.Fatalf("NoRoute shadowed /health — body = %q", got)
	}
}

// TestWithNoRoute_Ugly_NilStaysAsDefault404 verifies the degenerate path —
// no NoRoute set means gin's default 404 surfaces unchanged. This protects
// the contract that WithNoRoute is opt-in and unset Engines stay
// compatible with the pre-WithNoRoute API.
func TestWithNoRoute_Ugly_NilStaysAsDefault404(t *testing.T) {
	e, err := New() // no WithNoRoute
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	e.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
