// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestModelResolver_Knows_Good(t *testing.T) {
	r := NewModelResolver()
	// Seed the loaded-by-name cache directly (internal test) to simulate a known model.
	r.loadedByName["lemer"] = nil
	if !r.Knows("lemer") {
		t.Fatal("Knows(lemer) = false, want true (cache hit)")
	}
}

func TestModelResolver_Knows_CaseInsensitive_Good(t *testing.T) {
	r := NewModelResolver()
	// Cache stores the lowercased name; Knows must mirror ResolveModel's
	// normalisation so a mixed-case request still hits the known model.
	r.loadedByName["gpt-4"] = nil
	if !r.Knows("GPT-4") {
		t.Fatal("Knows(GPT-4) = false, want true (case-insensitive cache hit)")
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
