// SPDX-License-Identifier: EUPL-1.2

package api

import (
	core "dappco.re/go"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestWebhook_GenerateWebhookSecret_Bad_ReturnsErrorWhenReaderFails(t *testing.T) {
	original := randomRead
	randomRead = func([]byte) (int, error) {
		return 0, io.ErrUnexpectedEOF
	}
	defer func() { randomRead = original }()

	secret, err := GenerateWebhookSecret()
	if err == nil {
		t.Fatal("expected GenerateWebhookSecret to fail when the random reader errors")
	}
	if secret != "" {
		t.Fatalf("expected empty secret on error, got %q", secret)
	}
}

func TestWebhook_GenerateWebhookSecret_Ugly_ReturnsErrorAfterPartialRead(t *testing.T) {
	original := randomRead
	randomRead = func(p []byte) (int, error) {
		if len(p) == 0 {
			return 0, io.ErrUnexpectedEOF
		}
		n := len(p) / 2
		if n == 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			p[i] = byte(i)
		}
		return n, io.ErrUnexpectedEOF
	}
	defer func() { randomRead = original }()

	secret, err := GenerateWebhookSecret()
	if err == nil {
		t.Fatal("expected GenerateWebhookSecret to fail when the random reader is truncated")
	}
	if secret != "" {
		t.Fatalf("expected empty secret on partial-read error, got %q", secret)
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

	r := httptest.NewRequest(http.MethodPost, "/incoming", core.NewReader(string(payload)))
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
	r := httptest.NewRequest(http.MethodPost, "/incoming", core.NewReader("body"))
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

// TestWebhook_WebhookEvents_Good_ListsCanonicalIdentifiers verifies the
// exported list of canonical event names documented in RFC §6.
func TestWebhook_WebhookEvents_Good_ListsCanonicalIdentifiers(t *testing.T) {
	got := WebhookEvents()
	want := []string{
		"workspace.created",
		"workspace.deleted",
		"subscription.changed",
		"subscription.cancelled",
		"biolink.created",
		"link.clicked",
		"ticket.created",
		"ticket.replied",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d events, got %d: %v", len(want), len(got), got)
	}
	for i, evt := range want {
		if got[i] != evt {
			t.Fatalf("index %d: expected %q, got %q", i, evt, got[i])
		}
	}
}

// TestWebhook_WebhookEvents_Good_ReturnsFreshSlice ensures the returned slice
// is safe to mutate — callers never corrupt the canonical list.
func TestWebhook_WebhookEvents_Good_ReturnsFreshSlice(t *testing.T) {
	first := WebhookEvents()
	first[0] = "mutated"
	second := WebhookEvents()
	if second[0] != "workspace.created" {
		t.Fatalf("expected canonical list to be immutable, got %q", second[0])
	}
}

// TestWebhook_IsKnownWebhookEvent_Good_RecognisesCanonical confirms canonical
// identifiers pass the known-event predicate.
func TestWebhook_IsKnownWebhookEvent_Good_RecognisesCanonical(t *testing.T) {
	for _, evt := range WebhookEvents() {
		if !IsKnownWebhookEvent(evt) {
			t.Fatalf("expected %q to be recognised", evt)
		}
	}
}

// TestWebhook_IsKnownWebhookEvent_Good_TrimsWhitespace ensures whitespace
// around user-supplied event names does not defeat the lookup.
func TestWebhook_IsKnownWebhookEvent_Good_TrimsWhitespace(t *testing.T) {
	if !IsKnownWebhookEvent("  workspace.created  ") {
		t.Fatal("expected whitespace-padded identifier to be recognised")
	}
}

// TestWebhook_IsKnownWebhookEvent_Bad_RejectsUnknown guards against silent
// acceptance of typos or out-of-catalogue event identifiers.
func TestWebhook_IsKnownWebhookEvent_Bad_RejectsUnknown(t *testing.T) {
	if IsKnownWebhookEvent("") {
		t.Fatal("expected empty string to be rejected")
	}
	if IsKnownWebhookEvent("workspace.created.extra") {
		t.Fatal("expected suffixed event to be rejected")
	}
	if IsKnownWebhookEvent("Workspace.Created") {
		t.Fatal("expected differently-cased identifier to be rejected")
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

// TestWebhook_IsTimestampValid_Good_HandlesFutureTimestamps exercises the
// symmetric window used for replay protection.
func TestWebhook_IsTimestampValid_Good_HandlesFutureTimestamps(t *testing.T) {
	s := NewWebhookSignerWithTolerance("x", time.Minute)
	future := time.Now().Add(30 * time.Second).Unix()
	if !s.IsTimestampValid(future) {
		t.Fatal("expected future timestamp within tolerance to be valid")
	}
}

// TestWebhook_Tolerance_Ugly_NilReceiverFallsBackToDefault covers the
// defensive nil-receiver branch used by verification helpers.
func TestWebhook_Tolerance_Ugly_NilReceiverFallsBackToDefault(t *testing.T) {
	var s *WebhookSigner
	if got := s.Tolerance(); got != DefaultWebhookTolerance {
		t.Fatalf("expected default tolerance for nil receiver, got %s", got)
	}
}

// TestWebhook_Verify_Bad_NilReceiverReturnsFalse ensures callers can safely
// invoke Verify on a nil signer without panicking.
func TestWebhook_Verify_Bad_NilReceiverReturnsFalse(t *testing.T) {
	var s *WebhookSigner
	if s.Verify([]byte("body"), "sig", time.Now().Unix()) {
		t.Fatal("expected nil receiver verification to fail")
	}
}

// TestWebhook_VerifySignatureOnly_Bad_NilReceiverReturnsFalse mirrors Verify
// for the signature-only variant.
func TestWebhook_VerifySignatureOnly_Bad_NilReceiverReturnsFalse(t *testing.T) {
	var s *WebhookSigner
	if s.VerifySignatureOnly([]byte("body"), "sig", time.Now().Unix()) {
		t.Fatal("expected nil receiver verification to fail")
	}
}

// TestWebhook_IsTimestampValid_Ugly_NilReceiverFallsBackToDefault documents
// that a nil receiver still applies the default tolerance window.
func TestWebhook_IsTimestampValid_Ugly_NilReceiverFallsBackToDefault(t *testing.T) {
	var s *WebhookSigner
	if !s.IsTimestampValid(time.Now().Unix()) {
		t.Fatal("expected nil receiver timestamp check to use default tolerance")
	}
}

func TestWebhook_ValidateWebhookURL_Good_AllowsPublicHTTPIP(t *testing.T) {
	if err := ValidateWebhookURL("https://8.8.8.8/inbox"); err != nil {
		t.Fatalf("expected public IP URL to be accepted, got %v", err)
	}
}

func TestWebhook_ValidateWebhookURL_Good_AllowsResolvablePublicHostname(t *testing.T) {
	original := lookupIP
	lookupIP = func(host string) ([]net.IP, error) {
		if host != "hooks.example.test" {
			return nil, net.UnknownNetworkError("unexpected host")
		}
		return []net.IP{
			net.ParseIP("1.1.1.1"),
			net.ParseIP("2606:4700:4700::1111"),
		}, nil
	}
	defer func() { lookupIP = original }()

	if err := ValidateWebhookURL("https://hooks.example.test/inbox"); err != nil {
		t.Fatalf("expected hostname resolving to public IPs to be accepted, got %v", err)
	}
}

func TestWebhook_DialPath_Bad_RevalidatesHostnameAtRequestTime(t *testing.T) {
	originalLookupIP := lookupIP
	originalResolveHost := resolveHost
	t.Cleanup(func() {
		lookupIP = originalLookupIP
		resolveHost = originalResolveHost
	})

	raw := "https://hooks.example.test/inbox"
	lookupIP = func(host string) ([]net.IP, error) {
		if host != "hooks.example.test" {
			return nil, net.UnknownNetworkError("unexpected host")
		}
		return []net.IP{net.ParseIP("93.184.216.34")}, nil
	}
	if err := ValidateWebhookURL(raw); err != nil {
		t.Fatalf("expected construction-time validation to accept public hostname, got %v", err)
	}

	resolveHost = func(host string) ([]net.IP, error) {
		if host != "hooks.example.test" {
			return nil, net.UnknownNetworkError("unexpected host")
		}
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}

	var attempts int
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			return nil, core.NewError("request should have been blocked before transport")
		}),
	}
	req, err := http.NewRequest(http.MethodPost, raw, core.NewReader("{}"))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	resp, err := doHTTPClientRequest(client, req)
	if err == nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		t.Fatal("expected dial-time SSRF guard to block rebound loopback resolution")
	}
	if !core.Is(err, errOutboundURLBlocked) {
		t.Fatalf("expected errOutboundURLBlocked, got %v", err)
	}
	if attempts != 0 {
		t.Fatalf("expected request to be blocked before transport, got %d attempts", attempts)
	}
}

func TestWebhook_DialPath_Good_DialsPublicHostname(t *testing.T) {
	originalLookupIP := lookupIP
	originalResolveHost := resolveHost
	t.Cleanup(func() {
		lookupIP = originalLookupIP
		resolveHost = originalResolveHost
	})

	lookupIP = func(host string) ([]net.IP, error) {
		if host != "hooks.example.test" {
			return nil, net.UnknownNetworkError("unexpected host")
		}
		return []net.IP{net.ParseIP("93.184.216.34")}, nil
	}
	resolveHost = func(host string) ([]net.IP, error) {
		if host != "hooks.example.test" {
			return nil, net.UnknownNetworkError("unexpected host")
		}
		return []net.IP{net.ParseIP("93.184.216.34")}, nil
	}

	raw := "http://hooks.example.test/inbox"
	if err := ValidateWebhookURL(raw); err != nil {
		t.Fatalf("expected construction-time validation to accept public hostname, got %v", err)
	}

	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, "accepted")
	}))
	defer srv.Close()

	client := &http.Client{Transport: localServerTransport(t, srv)}
	req, err := http.NewRequest(http.MethodPost, raw, core.NewReader("{}"))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	resp, err := doHTTPClientRequest(client, req)
	if err != nil {
		t.Fatalf("expected public webhook hostname to be requested, got %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 response, got %d", resp.StatusCode)
	}
	if sawPath != "/inbox" {
		t.Fatalf("expected webhook path to reach server, got %q", sawPath)
	}
}

func TestWebhook_ValidateWebhookURL_Bad_RejectsHostnameResolvingToPrivateIP(t *testing.T) {
	original := lookupIP
	lookupIP = func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.0.0.1")}, nil
	}
	defer func() { lookupIP = original }()

	if err := ValidateWebhookURL("https://hooks.example.test/inbox"); err == nil {
		t.Fatal("expected hostname resolving to a private IP to be rejected")
	}
}

func TestWebhook_ValidateWebhookURL_Bad_RejectsDirectPrivateHTTPSIP(t *testing.T) {
	if err := ValidateWebhookURL("https://10.0.0.1/inbox"); err == nil {
		t.Fatal("expected direct private IP webhook URL to be rejected")
	}
}

func TestWebhook_ValidateWebhookURL_Ugly_RejectsLookupFailuresAndEmptyResults(t *testing.T) {
	original := lookupIP
	defer func() { lookupIP = original }()

	lookupIP = func(string) ([]net.IP, error) {
		return nil, net.UnknownNetworkError("lookup failed")
	}
	if err := ValidateWebhookURL("https://hooks.example.test/inbox"); err == nil {
		t.Fatal("expected lookup failure to be rejected")
	}

	lookupIP = func(string) ([]net.IP, error) {
		return []net.IP{}, nil
	}
	if err := ValidateWebhookURL("https://hooks.example.test/inbox"); err == nil {
		t.Fatal("expected empty lookup results to be rejected")
	}
}

func TestWebhook_ValidateWebhookURL_Bad_RejectsBlockedDestinations(t *testing.T) {
	cases := []string{
		"http://127.0.0.1/inbox",
		"http://10.0.0.1/inbox",
		"http://169.254.1.1/inbox",
		"http://203.0.113.10/inbox",
		"http://0.0.0.0/inbox",
		"http://255.255.255.255/inbox",
		"http://224.0.0.1/inbox",
		"http://[::1]/inbox",
		"http://[fc00::1]/inbox",
		"http://[::]/inbox",
		"http://[ff01::1]/inbox",
		"http://[ff00::1]/inbox",
		"https://localhost/inbox",
	}

	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			if err := ValidateWebhookURL(raw); err == nil {
				t.Fatalf("expected blocked destination %q to be rejected", raw)
			}
		})
	}
}

func TestWebhook_ValidateWebhookURL_Ugly_RejectsMalformedAndCredentialedURLs(t *testing.T) {
	cases := []string{
		"ftp://8.8.8.8/inbox",
		"https://user:pass@8.8.8.8/inbox",
		"not-a-url",
		"http:///missing-host",
		"https://:443/inbox",
		"https://[::1",
		"https://%zz",
	}

	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			if err := ValidateWebhookURL(raw); err == nil {
				t.Fatalf("expected malformed URL %q to be rejected", raw)
			}
		})
	}
}

func TestWebhook_isBlockedWebhookIP_Good_AllowsPublicIP(t *testing.T) {
	if isBlockedWebhookIP(net.ParseIP("8.8.8.8")) {
		t.Fatal("expected public IPv4 address to be allowed")
	}
	if isBlockedWebhookIP(net.ParseIP("2001:4860:4860::8888")) {
		t.Fatal("expected public IPv6 address to be allowed")
	}
}

func TestWebhook_isBlockedWebhookIP_Bad_RejectsReservedAndPrivateIPs(t *testing.T) {
	cases := []string{
		"127.0.0.1",
		"10.0.0.1",
		"169.254.1.1",
		"203.0.113.10",
		"0.0.0.0",
		"255.255.255.255",
		"224.0.0.1",
		"::1",
		"fc00::1",
		"::",
		"ff01::1",
		"ff00::1",
	}

	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			if !isBlockedWebhookIP(net.ParseIP(raw)) {
				t.Fatalf("expected %q to be blocked", raw)
			}
		})
	}
}

func TestWebhook_isBlockedWebhookIP_Ugly_TreatsNilAsBlocked(t *testing.T) {
	if !isBlockedWebhookIP(nil) {
		t.Fatal("expected nil IP to be blocked")
	}
}
