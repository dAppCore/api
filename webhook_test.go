// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestWebhook_NewWebhookSigner_Good_BuildsSignerWithDefaults verifies the
// constructor sets up a usable signer with the documented default tolerance.
func TestWebhook_NewWebhookSigner_Good_BuildsSignerWithDefaults(t *testing.T) {
	s := NewWebhookSigner("hello")
	if s == nil {
		t.Fatal("expected non-nil signer")
	}
	if s.Tolerance() != DefaultWebhookTolerance {
		t.Fatalf("expected default tolerance %s, got %s", DefaultWebhookTolerance, s.Tolerance())
	}
}

// TestWebhook_NewWebhookSignerWithTolerance_Good_OverridesTolerance ensures the
// custom-tolerance constructor is honoured for positive durations.
func TestWebhook_NewWebhookSignerWithTolerance_Good_OverridesTolerance(t *testing.T) {
	s := NewWebhookSignerWithTolerance("x", 30*time.Second)
	if s.Tolerance() != 30*time.Second {
		t.Fatalf("expected 30s tolerance, got %s", s.Tolerance())
	}
}

// TestWebhook_NewWebhookSignerWithTolerance_Ugly_FallsBackOnZero verifies a
// non-positive tolerance falls back to the documented default rather than
// silently disabling replay protection.
func TestWebhook_NewWebhookSignerWithTolerance_Ugly_FallsBackOnZero(t *testing.T) {
	s := NewWebhookSignerWithTolerance("x", 0)
	if s.Tolerance() != DefaultWebhookTolerance {
		t.Fatalf("expected default tolerance after zero override, got %s", s.Tolerance())
	}
	s = NewWebhookSignerWithTolerance("x", -5*time.Minute)
	if s.Tolerance() != DefaultWebhookTolerance {
		t.Fatalf("expected default tolerance after negative override, got %s", s.Tolerance())
	}
}

// TestWebhook_GenerateWebhookSecret_Good_Returns64HexChars ensures the helper
// returns a stable-format secret of the documented length.
func TestWebhook_GenerateWebhookSecret_Good_Returns64HexChars(t *testing.T) {
	secret, err := GenerateWebhookSecret()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secret) != 64 {
		t.Fatalf("expected 64-char secret, got %d", len(secret))
	}
	for _, r := range secret {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			t.Fatalf("expected lowercase hex characters, got %q", secret)
		}
	}
}

// TestWebhook_Sign_Good_ProducesStableHexDigest ensures the sign helper is
// deterministic for the same payload, secret, and timestamp.
func TestWebhook_Sign_Good_ProducesStableHexDigest(t *testing.T) {
	s := NewWebhookSigner("secret")
	first := s.Sign([]byte("payload"), 1234567890)
	second := s.Sign([]byte("payload"), 1234567890)
	if first != second {
		t.Fatalf("expected stable digest, got %s vs %s", first, second)
	}
	if len(first) != 64 {
		t.Fatalf("expected 64-char hex digest, got %d", len(first))
	}
}

// TestWebhook_Sign_Bad_ReturnsEmptyOnNilReceiver guards the nil-receiver
// behaviour required for safe defensive use in middleware.
func TestWebhook_Sign_Bad_ReturnsEmptyOnNilReceiver(t *testing.T) {
	var s *WebhookSigner
	if got := s.Sign([]byte("x"), 1); got != "" {
		t.Fatalf("expected empty digest from nil receiver, got %q", got)
	}
}

// TestWebhook_SignNow_Good_RoundTripsCurrentTimestamp verifies SignNow returns
// a fresh timestamp that the verifier accepts.
func TestWebhook_SignNow_Good_RoundTripsCurrentTimestamp(t *testing.T) {
	s := NewWebhookSigner("secret")
	payload := []byte(`{"event":"workspace.created"}`)
	sig, ts := s.SignNow(payload)
	if !s.Verify(payload, sig, ts) {
		t.Fatal("expected SignNow output to verify")
	}
}

// TestWebhook_Verify_Good_AcceptsMatchingSignature exercises the happy path of
// matching payload/signature/timestamp inside the tolerance window.
func TestWebhook_Verify_Good_AcceptsMatchingSignature(t *testing.T) {
	s := NewWebhookSigner("secret")
	payload := []byte("body")
	now := time.Now().Unix()
	sig := s.Sign(payload, now)
	if !s.Verify(payload, sig, now) {
		t.Fatal("expected valid signature to verify")
	}
}

// TestWebhook_Verify_Bad_RejectsTamperedPayload ensures payload mutation
// invalidates the signature even when the secret/timestamp are valid.
func TestWebhook_Verify_Bad_RejectsTamperedPayload(t *testing.T) {
	s := NewWebhookSigner("secret")
	now := time.Now().Unix()
	sig := s.Sign([]byte("body"), now)
	if s.Verify([]byte("tampered"), sig, now) {
		t.Fatal("expected verification to fail for tampered payload")
	}
}

// TestWebhook_Verify_Bad_RejectsExpiredTimestamp ensures stale timestamps fail
// even when the signature itself is valid for the older timestamp.
func TestWebhook_Verify_Bad_RejectsExpiredTimestamp(t *testing.T) {
	s := NewWebhookSignerWithTolerance("secret", time.Minute)
	old := time.Now().Add(-2 * time.Minute).Unix()
	sig := s.Sign([]byte("body"), old)
	if s.Verify([]byte("body"), sig, old) {
		t.Fatal("expected stale timestamp to be rejected")
	}
}

// TestWebhook_VerifySignatureOnly_Good_IgnoresExpiredTimestamp lets callers
// validate signature integrity even when timestamps fall outside tolerance.
func TestWebhook_VerifySignatureOnly_Good_IgnoresExpiredTimestamp(t *testing.T) {
	s := NewWebhookSignerWithTolerance("secret", time.Second)
	old := time.Now().Add(-time.Hour).Unix()
	sig := s.Sign([]byte("body"), old)
	if !s.VerifySignatureOnly([]byte("body"), sig, old) {
		t.Fatal("expected signature-only verification to pass for expired timestamp")
	}
}

// TestWebhook_Headers_Good_PopulatesSignatureAndTimestamp verifies the header
// helper returns both the signature and the timestamp that produced it.
func TestWebhook_Headers_Good_PopulatesSignatureAndTimestamp(t *testing.T) {
	s := NewWebhookSigner("secret")
	headers := s.Headers([]byte("body"))
	if headers[WebhookSignatureHeader] == "" {
		t.Fatal("expected signature header to be set")
	}
	if headers[WebhookTimestampHeader] == "" {
		t.Fatal("expected timestamp header to be set")
	}

	ts, err := strconv.ParseInt(headers[WebhookTimestampHeader], 10, 64)
	if err != nil {
		t.Fatalf("expected numeric timestamp header, got %q", headers[WebhookTimestampHeader])
	}
	if !s.Verify([]byte("body"), headers[WebhookSignatureHeader], ts) {
		t.Fatal("expected Headers() output to verify")
	}
}

// TestWebhook_VerifyRequest_Good_AcceptsValidHeaders uses the request helper
// to ensure middleware can verify webhooks straight from an http.Request.
func TestWebhook_VerifyRequest_Good_AcceptsValidHeaders(t *testing.T) {
	s := NewWebhookSigner("secret")
	payload := []byte(`{"event":"link.clicked"}`)
	headers := s.Headers(payload)

	r := httptest.NewRequest(http.MethodPost, "/incoming", strings.NewReader(string(payload)))
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	if !s.VerifyRequest(r, payload) {
		t.Fatal("expected VerifyRequest to accept valid signed request")
	}
}

// TestWebhook_VerifyRequest_Bad_RejectsMissingHeaders rejects requests with
// missing or malformed signature/timestamp headers.
func TestWebhook_VerifyRequest_Bad_RejectsMissingHeaders(t *testing.T) {
	s := NewWebhookSigner("secret")
	r := httptest.NewRequest(http.MethodPost, "/incoming", strings.NewReader("body"))
	if s.VerifyRequest(r, []byte("body")) {
		t.Fatal("expected VerifyRequest to fail with no headers")
	}

	r.Header.Set(WebhookSignatureHeader, "deadbeef")
	if s.VerifyRequest(r, []byte("body")) {
		t.Fatal("expected VerifyRequest to fail with missing timestamp header")
	}

	r.Header.Set(WebhookTimestampHeader, "not-a-number")
	if s.VerifyRequest(r, []byte("body")) {
		t.Fatal("expected VerifyRequest to fail with malformed timestamp header")
	}
}

// TestWebhook_VerifyRequest_Ugly_NilRequestReturnsFalse documents the
// defensive nil-request guard so middleware can safely call this helper.
func TestWebhook_VerifyRequest_Ugly_NilRequestReturnsFalse(t *testing.T) {
	s := NewWebhookSigner("secret")
	if s.VerifyRequest(nil, []byte("body")) {
		t.Fatal("expected VerifyRequest(nil) to return false")
	}
}

// TestWebhook_IsTimestampValid_Good_UsesConfiguredTolerance exercises the
// inclusive boundary where the timestamp falls right at the tolerance edge.
func TestWebhook_IsTimestampValid_Good_UsesConfiguredTolerance(t *testing.T) {
	s := NewWebhookSignerWithTolerance("x", time.Minute)
	now := time.Now().Unix()

	if !s.IsTimestampValid(now) {
		t.Fatal("expected current timestamp to be valid")
	}
	if !s.IsTimestampValid(now - 30) {
		t.Fatal("expected timestamp within tolerance to be valid")
	}
	if s.IsTimestampValid(now - 120) {
		t.Fatal("expected timestamp outside tolerance to be invalid")
	}
}
