// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/httpsign"
	"github.com/gin-contrib/httpsign/crypto"
	"github.com/gin-gonic/gin"

	api "forge.lthn.ai/core/api"
)

const testSecretKey = "test-secret-key-for-hmac-sha256"

// testKeyID is the key ID used in HTTP signature tests.
var testKeyID = httpsign.KeyID("test-client")

// newTestSecrets builds a Secrets map with a single HMAC-SHA256 key for testing.
func newTestSecrets() httpsign.Secrets {
	return httpsign.Secrets{
		testKeyID: &httpsign.Secret{
			Key:       testSecretKey,
			Algorithm: &crypto.HmacSha256{},
		},
	}
}

// signRequest constructs a valid HTTP Signature Authorization header for the
// given request, signing the specified headers with HMAC-SHA256 and the test
// secret key. The Date header is set to the current time if not already present.
func signRequest(req *http.Request, keyID httpsign.KeyID, secret string, headers []string) {
	// Ensure a Date header exists.
	if req.Header.Get("Date") == "" {
		req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}

	// Build the signing string in the same way the library does:
	// each header as "header: value", joined by newlines.
	var parts []string
	for _, h := range headers {
		var val string
		switch h {
		case "(request-target)":
			val = fmt.Sprintf("%s %s", strings.ToLower(req.Method), req.URL.RequestURI())
		case "host":
			val = req.Host
		default:
			val = req.Header.Get(h)
		}
		parts = append(parts, fmt.Sprintf("%s: %s", h, val))
	}
	signingString := strings.Join(parts, "\n")

	// Sign with HMAC-SHA256.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingString))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Build the Authorization header.
	authValue := fmt.Sprintf(
		"Signature keyId=\"%s\",algorithm=\"hmac-sha256\",headers=\"%s\",signature=\"%s\"",
		keyID,
		strings.Join(headers, " "),
		sig,
	)
	req.Header.Set("Authorization", authValue)
}

// ── WithHTTPSign ──────────────────────────────────────────────────────────

func TestWithHTTPSign_Good_ValidSignatureAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use only (request-target) and date as required headers, disable
	// validators to keep the test focused on signature verification.
	requiredHeaders := []string{"(request-target)", "date"}

	e, _ := api.New(api.WithHTTPSign(
		newTestSecrets(),
		httpsign.WithRequiredHeaders(requiredHeaders),
		httpsign.WithValidator(), // no validators — pure signature check
	))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	signRequest(req, testKeyID, testSecretKey, requiredHeaders)

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for validly signed request, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestWithHTTPSign_Bad_InvalidSignatureRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requiredHeaders := []string{"(request-target)", "date"}

	e, _ := api.New(api.WithHTTPSign(
		newTestSecrets(),
		httpsign.WithRequiredHeaders(requiredHeaders),
		httpsign.WithValidator(),
	))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)

	// Sign with the wrong secret so the signature is invalid.
	signRequest(req, testKeyID, "wrong-secret-key", requiredHeaders)

	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid signature, got %d", w.Code)
	}
}

func TestWithHTTPSign_Bad_MissingSignatureRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requiredHeaders := []string{"(request-target)", "date"}

	e, _ := api.New(api.WithHTTPSign(
		newTestSecrets(),
		httpsign.WithRequiredHeaders(requiredHeaders),
		httpsign.WithValidator(),
	))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()

	// Send a request with no signature at all.
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing signature, got %d", w.Code)
	}
}

func TestWithHTTPSign_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requiredHeaders := []string{"(request-target)", "date"}

	e, _ := api.New(
		api.WithRequestID(),
		api.WithHTTPSign(
			newTestSecrets(),
			httpsign.WithRequiredHeaders(requiredHeaders),
			httpsign.WithValidator(),
		),
	)
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	signRequest(req, testKeyID, testSecretKey, requiredHeaders)

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	// Verify that WithRequestID also ran.
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}

func TestWithHTTPSign_Ugly_UnknownKeyIDRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requiredHeaders := []string{"(request-target)", "date"}

	e, _ := api.New(api.WithHTTPSign(
		newTestSecrets(),
		httpsign.WithRequiredHeaders(requiredHeaders),
		httpsign.WithValidator(),
	))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)

	// Sign with an unknown key ID that does not exist in the secrets map.
	unknownKeyID := httpsign.KeyID("unknown-client")
	signRequest(req, unknownKeyID, testSecretKey, requiredHeaders)

	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unknown key ID, got %d", w.Code)
	}
}
