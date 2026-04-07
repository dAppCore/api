// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

type rateLimitTestGroup struct{}

func (r *rateLimitTestGroup) Name() string     { return "rate-limit" }
func (r *rateLimitTestGroup) BasePath() string { return "/rate" }
func (r *rateLimitTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("pong"))
	})
}

func TestWithRateLimit_Good_AllowsBurstThenRejects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRateLimit(2))
	e.Register(&rateLimitTestGroup{})

	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req1.RemoteAddr = "203.0.113.10:1234"
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", w1.Code)
	}
	if got := w1.Header().Get("X-RateLimit-Limit"); got != "2" {
		t.Fatalf("expected X-RateLimit-Limit=2, got %q", got)
	}
	if got := w1.Header().Get("X-RateLimit-Remaining"); got != "1" {
		t.Fatalf("expected X-RateLimit-Remaining=1, got %q", got)
	}
	if got := w1.Header().Get("X-RateLimit-Reset"); got == "" {
		t.Fatal("expected X-RateLimit-Reset on successful response")
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req2.RemoteAddr = "203.0.113.10:1234"
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected second request to succeed, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req3.RemoteAddr = "203.0.113.10:1234"
	h.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected third request to be rate limited, got %d", w3.Code)
	}

	if got := w3.Header().Get("Retry-After"); got == "" {
		t.Fatal("expected Retry-After header on 429 response")
	}
	if got := w3.Header().Get("X-RateLimit-Limit"); got != "2" {
		t.Fatalf("expected X-RateLimit-Limit=2 on 429, got %q", got)
	}
	if got := w3.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Fatalf("expected X-RateLimit-Remaining=0 on 429, got %q", got)
	}
	if got := w3.Header().Get("X-RateLimit-Reset"); got == "" {
		t.Fatal("expected X-RateLimit-Reset on 429 response")
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w3.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected Success=false for rate limited response")
	}
	if resp.Error == nil || resp.Error.Code != "rate_limit_exceeded" {
		t.Fatalf("expected rate_limit_exceeded error, got %+v", resp.Error)
	}
}

func TestWithRateLimit_Good_IsolatesPerIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRateLimit(1))
	e.Register(&rateLimitTestGroup{})

	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req1.RemoteAddr = "203.0.113.10:1234"
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first IP to succeed, got %d", w1.Code)
	}
	if got := w1.Header().Get("X-RateLimit-Limit"); got != "1" {
		t.Fatalf("expected X-RateLimit-Limit=1, got %q", got)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req2.RemoteAddr = "203.0.113.11:1234"
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected second IP to have its own bucket, got %d", w2.Code)
	}
}

func TestWithRateLimit_Good_IsolatesPerAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRateLimit(1))
	e.Register(&rateLimitTestGroup{})

	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req1.RemoteAddr = "203.0.113.20:1234"
	req1.Header.Set("X-API-Key", "key-a")
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first API key request to succeed, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req2.RemoteAddr = "203.0.113.20:1234"
	req2.Header.Set("X-API-Key", "key-b")
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected second API key to have its own bucket, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req3.RemoteAddr = "203.0.113.20:1234"
	req3.Header.Set("X-API-Key", "key-a")
	h.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected repeated API key to be rate limited, got %d", w3.Code)
	}
}

func TestWithRateLimit_Good_UsesBearerTokenWhenPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRateLimit(1))
	e.Register(&rateLimitTestGroup{})

	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req1.RemoteAddr = "203.0.113.30:1234"
	req1.Header.Set("Authorization", "Bearer token-a")
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first bearer token request to succeed, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req2.RemoteAddr = "203.0.113.30:1234"
	req2.Header.Set("Authorization", "Bearer token-b")
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected second bearer token to have its own bucket, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req3.RemoteAddr = "203.0.113.30:1234"
	req3.Header.Set("Authorization", "Bearer token-a")
	h.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected repeated bearer token to be rate limited, got %d", w3.Code)
	}
}

func TestWithRateLimit_Good_RefillsOverTime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRateLimit(1))
	e.Register(&rateLimitTestGroup{})

	h := e.Handler()

	req, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
	req.RemoteAddr = "203.0.113.12:1234"

	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, req.Clone(req.Context()))
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2 := req.Clone(req.Context())
	req2.RemoteAddr = req.RemoteAddr
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be rate limited, got %d", w2.Code)
	}

	time.Sleep(1100 * time.Millisecond)

	w3 := httptest.NewRecorder()
	req3 := req.Clone(req.Context())
	req3.RemoteAddr = req.RemoteAddr
	h.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected bucket to refill after waiting, got %d", w3.Code)
	}
}

func TestWithRateLimit_Ugly_NonPositiveLimitDisablesMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithRateLimit(0))
	e.Register(&rateLimitTestGroup{})

	h := e.Handler()

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/rate/ping", nil)
		req.RemoteAddr = "203.0.113.13:1234"
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected request %d to succeed with disabled limiter, got %d", i+1, w.Code)
		}
	}
}
