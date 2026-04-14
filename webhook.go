// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"slices"
	"strconv"
	"time"

	core "dappco.re/go/core"
)

// Canonical webhook event identifiers from RFC §6. These constants mirror the
// PHP-side event catalogue so Go senders and receivers reference the same
// namespaced strings.
//
//	evt := api.WebhookEventLinkClicked // "link.clicked"
const (
	// WebhookEventWorkspaceCreated fires when a new workspace is provisioned.
	WebhookEventWorkspaceCreated = "workspace.created"
	// WebhookEventWorkspaceDeleted fires when a workspace is permanently removed.
	WebhookEventWorkspaceDeleted = "workspace.deleted"
	// WebhookEventSubscriptionChanged fires when a subscription plan changes.
	WebhookEventSubscriptionChanged = "subscription.changed"
	// WebhookEventSubscriptionCancelled fires when a subscription is cancelled.
	WebhookEventSubscriptionCancelled = "subscription.cancelled"
	// WebhookEventBiolinkCreated fires when a new biolink page is created.
	WebhookEventBiolinkCreated = "biolink.created"
	// WebhookEventLinkClicked fires when a tracked short link is clicked.
	// High-volume event — recipients should opt in explicitly.
	WebhookEventLinkClicked = "link.clicked"
	// WebhookEventTicketCreated fires when a support ticket is opened.
	WebhookEventTicketCreated = "ticket.created"
	// WebhookEventTicketReplied fires when a support ticket receives a reply.
	WebhookEventTicketReplied = "ticket.replied"
)

// WebhookEvents returns the canonical list of webhook event identifiers
// documented in RFC §6. The order is stable: catalogue groups share a prefix
// (workspace → subscription → biolink → link → ticket).
//
//	for _, evt := range api.WebhookEvents() {
//	    registry.Enable(evt)
//	}
func WebhookEvents() []string {
	return []string{
		WebhookEventWorkspaceCreated,
		WebhookEventWorkspaceDeleted,
		WebhookEventSubscriptionChanged,
		WebhookEventSubscriptionCancelled,
		WebhookEventBiolinkCreated,
		WebhookEventLinkClicked,
		WebhookEventTicketCreated,
		WebhookEventTicketReplied,
	}
}

// IsKnownWebhookEvent reports whether the given event name is one of the
// canonical identifiers documented in RFC §6.
//
//	if !api.IsKnownWebhookEvent(evt) {
//	    return errors.New("unknown webhook event")
//	}
func IsKnownWebhookEvent(name string) bool {
	return slices.Contains(WebhookEvents(), core.Trim(name))
}

// WebhookSigner produces and verifies HMAC-SHA256 signatures over webhook
// payloads. Spec §6: signed payloads (HMAC-SHA256) include a timestamp and
// a signature header so receivers can validate authenticity, integrity, and
// reject replays.
//
// The signature format mirrors the PHP-side WebhookSignature service:
//
//	signature = HMAC-SHA256(timestamp + "." + payload, secret)
//
//	signer := api.NewWebhookSigner("supersecret")
//	headers := signer.Headers([]byte(`{"event":"workspace.created"}`))
//	// headers["X-Webhook-Signature"] is the hex digest
//	// headers["X-Webhook-Timestamp"] is the Unix timestamp string
type WebhookSigner struct {
	secret    []byte
	tolerance time.Duration
}

const (
	// WebhookSignatureHeader is the response/request header that carries the
	// HMAC-SHA256 hex digest of a signed webhook payload.
	WebhookSignatureHeader = "X-Webhook-Signature"

	// WebhookTimestampHeader is the response/request header that carries the
	// Unix timestamp used to derive the signature.
	WebhookTimestampHeader = "X-Webhook-Timestamp"

	// DefaultWebhookTolerance is the maximum age of a webhook timestamp the
	// signer will accept. Five minutes mirrors the PHP-side default and
	// allows for reasonable clock skew between sender and receiver.
	DefaultWebhookTolerance = 5 * time.Minute
)

// NewWebhookSigner constructs a signer that uses the given shared secret with
// the default timestamp tolerance (five minutes).
//
//	signer := api.NewWebhookSigner("shared-secret")
func NewWebhookSigner(secret string) *WebhookSigner {
	return &WebhookSigner{
		secret:    []byte(secret),
		tolerance: DefaultWebhookTolerance,
	}
}

// NewWebhookSignerWithTolerance constructs a signer with a custom timestamp
// tolerance. A tolerance of zero or negative falls back to
// DefaultWebhookTolerance to avoid silently disabling replay protection.
//
//	signer := api.NewWebhookSignerWithTolerance("secret", 30*time.Second)
func NewWebhookSignerWithTolerance(secret string, tolerance time.Duration) *WebhookSigner {
	if tolerance <= 0 {
		tolerance = DefaultWebhookTolerance
	}
	return &WebhookSigner{
		secret:    []byte(secret),
		tolerance: tolerance,
	}
}

// GenerateWebhookSecret returns a hex-encoded 32-byte random string suitable
// for use as a webhook signing secret. Output length is 64 characters.
//
//	secret, err := api.GenerateWebhookSecret()
//	// secret = "9f1a..." (64 hex chars)
func GenerateWebhookSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", core.E("WebhookSigner.GenerateSecret", "read random bytes", err)
	}
	return hex.EncodeToString(buf), nil
}

// Tolerance returns the configured maximum age tolerance for verification.
//
//	d := signer.Tolerance()
func (s *WebhookSigner) Tolerance() time.Duration {
	if s == nil || s.tolerance <= 0 {
		return DefaultWebhookTolerance
	}
	return s.tolerance
}

// Sign returns the hex-encoded HMAC-SHA256 of "timestamp.payload" using the
// signer's secret. Callers may pass any non-nil payload bytes (typically a
// JSON-encoded event body).
//
//	digest := signer.Sign(payload, time.Now().Unix())
func (s *WebhookSigner) Sign(payload []byte, timestamp int64) string {
	if s == nil {
		return ""
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// SignNow signs the payload using the current Unix timestamp and returns
// both the signature and the timestamp used.
//
//	sig, ts := signer.SignNow(payload)
func (s *WebhookSigner) SignNow(payload []byte) (string, int64) {
	now := time.Now().Unix()
	return s.Sign(payload, now), now
}

// Headers returns the HTTP headers a sender should attach to a webhook
// request: X-Webhook-Signature and X-Webhook-Timestamp. The current Unix
// timestamp is used.
//
//	for k, v := range signer.Headers(payload) {
//	    req.Header.Set(k, v)
//	}
func (s *WebhookSigner) Headers(payload []byte) map[string]string {
	sig, ts := s.SignNow(payload)
	return map[string]string{
		WebhookSignatureHeader: sig,
		WebhookTimestampHeader: strconv.FormatInt(ts, 10),
	}
}

// Verify reports whether the given signature matches the payload and the
// timestamp is within the signer's tolerance window. Comparison uses
// hmac.Equal to avoid timing attacks.
//
//	if signer.Verify(payload, sig, ts) {
//	    // accept event
//	}
func (s *WebhookSigner) Verify(payload []byte, signature string, timestamp int64) bool {
	if s == nil {
		return false
	}
	if !s.IsTimestampValid(timestamp) {
		return false
	}
	expected := s.Sign(payload, timestamp)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// VerifySignatureOnly compares the signature without checking timestamp
// freshness. Use it when timestamp validation is performed elsewhere or in
// tests where a stable timestamp is required.
//
//	ok := signer.VerifySignatureOnly(payload, sig, fixedTimestamp)
func (s *WebhookSigner) VerifySignatureOnly(payload []byte, signature string, timestamp int64) bool {
	if s == nil {
		return false
	}
	expected := s.Sign(payload, timestamp)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// IsTimestampValid reports whether the given Unix timestamp is within the
// signer's configured tolerance window relative to the current time.
//
//	if !signer.IsTimestampValid(ts) {
//	    return errors.New("webhook timestamp expired")
//	}
func (s *WebhookSigner) IsTimestampValid(timestamp int64) bool {
	tol := s.Tolerance()
	now := time.Now().Unix()
	delta := now - timestamp
	if delta < 0 {
		delta = -delta
	}
	return time.Duration(delta)*time.Second <= tol
}

// VerifyRequest extracts the signature and timestamp headers from a request
// and verifies the given payload against them. Returns false when the
// headers are missing, malformed, or the signature does not match.
//
//	if !signer.VerifyRequest(r, payload) {
//	    http.Error(w, "invalid signature", http.StatusUnauthorized)
//	    return
//	}
func (s *WebhookSigner) VerifyRequest(r *http.Request, payload []byte) bool {
	if r == nil {
		return false
	}
	sig := r.Header.Get(WebhookSignatureHeader)
	rawTS := r.Header.Get(WebhookTimestampHeader)
	if sig == "" || rawTS == "" {
		return false
	}
	ts, err := strconv.ParseInt(rawTS, 10, 64)
	if err != nil {
		return false
	}
	return s.Verify(payload, sig, ts)
}
