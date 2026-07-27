// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"io"
	"net/http"
	"net/url" // Note: AX-6 — url.URL fields are structural for per-attempt upstream rewriting.

	core "dappco.re/go"
)

// upstreamTransport is the http.RoundTripper that owns weighted selection and
// passive failover. The per-request pool and key are read from the request
// context (bound by the router handler). On a transport error or a failover
// status it marks the upstream cooling and retries the next, up to maxAttempts.
//
// SECURITY: this transport intentionally dispatches to operator-configured
// upstreams without re-applying the request-time SSRF guard. Upstream URLs are
// validated once at registration (UpstreamRegistry.validate, default-deny with
// AllowPrivateUpstreams opt-in), so loopback/private model endpoints are
// permitted by design. See spec §8.
type upstreamTransport struct {
	base        http.RoundTripper
	balancer    *upstreamBalancer
	maxAttempts int
	failover    map[int]bool
}

func (t *upstreamTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	pool, ok := poolFromContext(req.Context())
	if !ok || len(pool) == 0 {
		return nil, &routerError{status: http.StatusServiceUnavailable, code: errCodeUpstreamUnavailable, message: "no upstream pool bound to request"}
	}
	key, _ := keyFromContext(req.Context())

	attempts := t.maxAttempts
	if attempts <= 0 || attempts > len(pool) {
		attempts = len(pool)
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		up, ok := t.balancer.pick(key, pool)
		if !ok {
			break
		}
		target, err := url.Parse(up.URL)
		if err != nil {
			t.balancer.markFailed(up.URL)
			lastErr = err
			continue
		}

		out := req.Clone(req.Context())
		if out.GetBody != nil {
			body, berr := out.GetBody()
			if berr != nil {
				// A failed body replay would dispatch a consumed/empty body to the
				// upstream; bail to the exhausted-path 503 (which logs the cause).
				lastErr = berr
				break
			}
			out.Body = body
		}
		applyUpstream(out, target)
		for k, v := range up.Headers {
			out.Header.Set(k, v)
		}

		//#nosec G107 -- upstream is operator-configured and validated at registration
		// (UpstreamRegistry default-deny + AllowPrivateUpstreams opt-in); the request-time
		// SSRF guard is deliberately not re-applied here. See spec §8 / Mantis upstream-router.
		resp, err := t.base.RoundTrip(out)
		if err != nil {
			t.balancer.markFailed(up.URL)
			lastErr = err
			continue
		}
		if t.failover[resp.StatusCode] {
			t.balancer.markFailed(up.URL)
			drainAndClose(resp.Body)
			lastErr = core.E("upstream", core.Sprintf("upstream %s returned %d", up.URL, resp.StatusCode), nil)
			continue
		}
		return resp, nil
	}

	if lastErr != nil {
		// Detail goes to the error (logged by ErrorHandler); the client sees a
		// generic envelope so upstream URLs never leak.
		return nil, &routerError{status: http.StatusServiceUnavailable, code: errCodeUpstreamUnavailable, message: "no healthy upstream available", cause: lastErr}
	}
	return nil, &routerError{status: http.StatusServiceUnavailable, code: errCodeUpstreamUnavailable, message: "all upstreams cooling"}
}

// applyUpstream rewrites the outbound request to target the chosen upstream.
// A base path on the upstream URL is prefixed to the incoming request path.
func applyUpstream(out *http.Request, target *url.URL) {
	out.URL.Scheme = target.Scheme
	out.URL.Host = target.Host
	out.Host = target.Host
	if base := trimTrailingSlashes(target.Path); base != "" {
		out.URL.Path = base + out.URL.Path
		if out.URL.RawPath != "" {
			out.URL.RawPath = base + out.URL.RawPath
		}
	}
}

func drainAndClose(body io.ReadCloser) {
	if body != nil {
		_, _ = io.CopyN(io.Discard, body, 4<<10) // bounded drain so the conn is reusable; cap guards a hostile error body
		_ = body.Close()
	}
}
