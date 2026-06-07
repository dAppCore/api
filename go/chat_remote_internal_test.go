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
	const offLoopback = "203.0.113.7:5555"
	body := `{"model":"m","messages":[{"role":"user","content":"x"}]}`
	// serve dispatches a non-loopback request through the handler and returns the
	// recorded status code. bearer, when non-empty, is sent as a Bearer header.
	serve := func(h *chatCompletionsHandler, bearer string) int {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
		req.RemoteAddr = offLoopback
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		c.Request = req
		h.ServeHTTP(c)
		return w.Code
	}

	// (a) non-loopback, no opt-in → 403 regardless of any bearer.
	hNoOptIn := newChatCompletionsHandler(nil, &chatRemoteConfig{}, false, bearerValidator("secret"))
	if code := serve(hNoOptIn, "secret"); code != http.StatusForbidden {
		t.Fatalf("(a) non-loopback w/o opt-in: code = %d, want 403", code)
	}

	// (b) non-loopback, opt-in + validator that accepts the matching bearer → NOT 403.
	hAccept := newChatCompletionsHandler(nil, &chatRemoteConfig{reg: NewUpstreamRegistry()}, true, bearerValidator("secret"))
	if code := serve(hAccept, "secret"); code == http.StatusForbidden {
		t.Fatalf("(b) non-loopback opt-in + valid bearer: code = 403, want allowed")
	}

	// (c) non-loopback, opt-in + validator but request has NO/wrong bearer → 403.
	if code := serve(hAccept, ""); code != http.StatusForbidden {
		t.Fatalf("(c) non-loopback opt-in + missing bearer: code = %d, want 403", code)
	}
	if code := serve(hAccept, "wrong"); code != http.StatusForbidden {
		t.Fatalf("(c) non-loopback opt-in + wrong bearer: code = %d, want 403", code)
	}

	// (d) non-loopback, opt-in but validateBearer == nil (no bearer configured) → 403.
	hNoBearer := newChatCompletionsHandler(nil, &chatRemoteConfig{}, true, bearerValidator(""))
	if code := serve(hNoBearer, "secret"); code != http.StatusForbidden {
		t.Fatalf("(d) non-loopback opt-in WITHOUT configured bearer: code = %d, want 403", code)
	}
}
