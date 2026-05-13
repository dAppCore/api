// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newRateLimitTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return ctx, recorder
}

func TestRatelimit_clientRateLimitKey_Good_PrioritisesPrincipalOverOtherInputs(t *testing.T) {
	ctx, _ := newRateLimitTestContext()
	ctx.Set("principal", "workspace-1")
	ctx.Set("userID", "user-1")
	ctx.Request.Header.Set("X-API-Key", "key-a")
	ctx.Request.Header.Set("Authorization", "Bearer token-a")
	ctx.Request.RemoteAddr = "203.0.113.10:1234"

	if got := clientRateLimitKey(ctx); got != "principal:workspace-1" {
		t.Fatalf("expected principal bucket, got %q", got)
	}
}

func TestRatelimit_clientRateLimitKey_Bad_FallsBackToUserIDWhenPrincipalIsBlank(t *testing.T) {
	ctx, _ := newRateLimitTestContext()
	ctx.Set("principal", "")
	ctx.Set("userID", "user-1")
	ctx.Request.Header.Set("X-API-Key", "key-a")
	ctx.Request.Header.Set("Authorization", "Bearer token-a")
	ctx.Request.RemoteAddr = "203.0.113.10:1234"

	if got := clientRateLimitKey(ctx); got != "user:user-1" {
		t.Fatalf("expected user bucket, got %q", got)
	}
}

func TestRatelimit_clientRateLimitKey_Ugly_HashesCredentialsBeforeFallingBackToIP(t *testing.T) {
	ctx, _ := newRateLimitTestContext()
	ctx.Request.Header.Set("X-API-Key", "key-a")

	sum := sha256.Sum256([]byte("key-a"))
	want := "cred:sha256:" + hex.EncodeToString(sum[:])

	if got := clientRateLimitKey(ctx); got != want {
		t.Fatalf("expected hashed API key bucket, got %q", got)
	}

	ctx, _ = newRateLimitTestContext()
	ctx.Request.RemoteAddr = "203.0.113.10:1234"
	ctx.Request.Header.Set("Authorization", "Token abc")

	if got := clientRateLimitKey(ctx); got != "ip:203.0.113.10" {
		t.Fatalf("expected malformed bearer token to fall back to IP, got %q", got)
	}
}

func TestRatelimit_setRateLimitHeaders_Good_WritesLimitRemainingAndReset(t *testing.T) {
	ctx, _ := newRateLimitTestContext()
	resetAt := time.Now().Add(time.Minute)

	setRateLimitHeaders(ctx, 60, 42, resetAt)

	if got := ctx.Writer.Header().Get("X-RateLimit-Limit"); got != "60" {
		t.Fatalf("expected X-RateLimit-Limit=60, got %q", got)
	}
	if got := ctx.Writer.Header().Get("X-RateLimit-Remaining"); got != "42" {
		t.Fatalf("expected X-RateLimit-Remaining=42, got %q", got)
	}
	if got := ctx.Writer.Header().Get("X-RateLimit-Reset"); got == "" {
		t.Fatal("expected X-RateLimit-Reset to be set")
	}
}

func TestRatelimit_setRateLimitHeaders_Bad_ClampsNegativeRemaining(t *testing.T) {
	ctx, _ := newRateLimitTestContext()

	setRateLimitHeaders(ctx, 60, -5, time.Now().Add(time.Minute))

	if got := ctx.Writer.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Fatalf("expected negative remaining to be clamped to 0, got %q", got)
	}
}

func TestRatelimit_setRateLimitHeaders_Ugly_SkipsLimitAndResetForZeroValues(t *testing.T) {
	ctx, _ := newRateLimitTestContext()

	setRateLimitHeaders(ctx, 0, 0, time.Time{})

	if got := ctx.Writer.Header().Get("X-RateLimit-Limit"); got != "" {
		t.Fatalf("expected no limit header for zero limit, got %q", got)
	}
	if got := ctx.Writer.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Fatalf("expected remaining to be serialised as 0, got %q", got)
	}
	if got := ctx.Writer.Header().Get("X-RateLimit-Reset"); got != "" {
		t.Fatalf("expected no reset header for zero reset time, got %q", got)
	}
}

func TestRatelimit_timeUntilFull_Good_ReturnsZeroForFullBucket(t *testing.T) {
	if got := timeUntilFull(10, 10); got != 0 {
		t.Fatalf("expected full bucket to require no wait, got %s", got)
	}
}

func TestRatelimit_timeUntilFull_Bad_CalculatesCeilingForPartialBucket(t *testing.T) {
	if got := timeUntilFull(9, 10); got != 100*time.Millisecond {
		t.Fatalf("expected 100ms refill window, got %s", got)
	}
}

func TestRatelimit_timeUntilFull_Ugly_ReturnsZeroForNonPositiveLimit(t *testing.T) {
	if got := timeUntilFull(0, 0); got != 0 {
		t.Fatalf("expected zero wait for non-positive limit, got %s", got)
	}
	if got := timeUntilFull(0, -1); got != 0 {
		t.Fatalf("expected zero wait for negative limit, got %s", got)
	}
}
