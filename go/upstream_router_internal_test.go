// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// serveInternal builds a test engine with the upstream router mounted and
// returns a live server. It mirrors the external serve helper but lives in
// package api so tests can override unexported package vars.
func serveInternal(t *testing.T, reg *UpstreamRegistry, opts ...UpstreamRouterOption) *httptest.Server {
	t.Helper()
	e, err := New(WithUpstreamRouter(reg, opts...))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return httptest.NewServer(e.Handler())
}

// TestUpstreamRouter_OversizeBuffered_Bad asserts the non-stream + out-transformer
// path rejects an upstream body larger than maxUpstreamResponseBytes with a 502
// invalid_upstream_response, rather than buffering it unbounded. The limit is
// lowered for the test (save/restore) so no real 10 MiB body is needed.
func TestUpstreamRouter_OversizeBuffered_Bad(t *testing.T) {
	prev := maxUpstreamResponseBytes
	maxUpstreamResponseBytes = 16
	defer func() { maxUpstreamResponseBytes = prev }()

	oversize := strings.Repeat("x", 64) // > 16-byte cap
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":"`+oversize+`"}`)
	}))
	defer up.Close()

	reg := NewUpstreamRegistry(AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.Set("m", Upstream{URL: up.URL}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	// An out-transformer is required to reach the buffered branch at all.
	srv := serveInternal(t, reg, WithUpstreamTransformerOut(RenameFields(map[string]string{"data": "payload"})))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"model":"m"}`))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502 (oversize buffered response)", resp.StatusCode)
	}
	got, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(got), errCodeInvalidUpstreamResp) {
		t.Fatalf("body = %s, want %s envelope", got, errCodeInvalidUpstreamResp)
	}
}

// TestUpstreamRouter_SmallBufferedTransforms_Good is the regression guard: a body
// at or under the (lowered) cap still transforms cleanly through the same path.
func TestUpstreamRouter_SmallBufferedTransforms_Good(t *testing.T) {
	prev := maxUpstreamResponseBytes
	maxUpstreamResponseBytes = 4096
	defer func() { maxUpstreamResponseBytes = prev }()

	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"internal_id":42}`)
	}))
	defer up.Close()

	reg := NewUpstreamRegistry(AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.Set("m", Upstream{URL: up.URL}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	srv := serveInternal(t, reg, WithUpstreamTransformerOut(RenameFields(map[string]string{"internal_id": "id"})))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"model":"m"}`))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	got, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(got), `"id":42`) {
		t.Fatalf("body = %s, want renamed internal_id->id", got)
	}
}
