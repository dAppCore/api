// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	core "dappco.re/go"
)

type fakeRoundTripper struct {
	fn func(*http.Request) (*http.Response, error)
}

func (f fakeRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f.fn(r) }

func newResp(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("ok")),
		Header:     http.Header{},
	}
}

func requestWithPool(pool []Upstream, key string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader("{}"))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("{}")), nil }
	ctx := context.WithValue(req.Context(), poolCtxKey, pool)
	ctx = context.WithValue(ctx, keyCtxKey, key)
	return req.WithContext(ctx)
}

func TestUpstreamTransport_FailoverThenSuccess_Good(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	var hits []string
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		hits = append(hits, r.URL.Host)
		if r.URL.Host == "a" {
			return newResp(http.StatusBadGateway), nil
		}
		return newResp(http.StatusOK), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 2, failover: defaultFailoverStatuses()}
	pool := []Upstream{{URL: "http://a", Weight: 1}, {URL: "http://b", Weight: 1}}

	resp, err := tr.RoundTrip(requestWithPool(pool, "k"))
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(hits) != 2 {
		t.Fatalf("attempts = %v, want 2 (a then b)", hits)
	}
}

func TestUpstreamTransport_HeaderInjection_Good(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	var gotAuth string
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		return newResp(http.StatusOK), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 1, failover: defaultFailoverStatuses()}
	pool := []Upstream{{URL: "http://a", Headers: map[string]string{"Authorization": "Bearer up-key"}}}

	if _, err := tr.RoundTrip(requestWithPool(pool, "k")); err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if gotAuth != "Bearer up-key" {
		t.Fatalf("injected auth = %q, want Bearer up-key", gotAuth)
	}
}

func TestUpstreamTransport_AllFail_Bad(t *testing.T) {
	bal := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	base := fakeRoundTripper{fn: func(r *http.Request) (*http.Response, error) {
		return newResp(http.StatusServiceUnavailable), nil
	}}
	tr := &upstreamTransport{base: base, balancer: bal, maxAttempts: 2, failover: defaultFailoverStatuses()}
	pool := []Upstream{{URL: "http://a"}, {URL: "http://b"}}

	_, err := tr.RoundTrip(requestWithPool(pool, "k"))
	var re *routerError
	if !core.As(err, &re) || re.status != http.StatusServiceUnavailable {
		t.Fatalf("err = %v, want *routerError status 503", err)
	}
}
