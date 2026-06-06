# Upstream Router (`WithUpstreamRouter`) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a selector-keyed reverse-proxy Option (`WithUpstreamRouter`) to `dappco.re/go/api` that load-balances each request across a runtime-mutable pool of HTTP upstreams, with weighted round-robin + passive failover, hybrid streaming, a decision hook, and composition with the existing TransformerIn/Out layer.

**Architecture:** A copy-on-write `UpstreamRegistry` (key→pool) is the source of truth and validates URLs at registration (block-by-default SSRF, opt-in `AllowPrivateUpstreams`). A pure `upstreamBalancer` does smooth weighted round-robin + cooldown. An `upstreamTransport` (`http.RoundTripper`) owns per-attempt selection + failover. One `httputil.ReverseProxy` per router does streaming (`FlushInterval:-1`), buffered `TransformerOut` (`ModifyResponse`), and clean error envelopes (`ErrorHandler`). Mounted at the gin root by `Engine.build()`, so engine middleware (auth/CORS/rate-limit/tracing) wraps it.

**Tech Stack:** Go 1.26, `net/http/httputil`, `gin`, `dappco.re/go` (core), existing `transformer*.go` / `ssrf_guard.go` / `response.go` helpers. Reference implementation for proxy mechanics: `go/pkg/provider/proxy.go`. Spec: `docs/superpowers/specs/2026-06-06-upstream-router-design.md`.

**Conventions:** SPDX header `// SPDX-License-Identifier: EUPL-1.2` on every file. UK English in strings/docs. `_Good/_Bad/_Ugly` test suffixes. Run tests with `GOWORK=off go test` from `core/api/go`. Commit with `Co-Authored-By: Virgil <virgil@lethean.io>`.

---

## File Structure

| File | Responsibility |
|------|----------------|
| `go/upstream_registry.go` | `Upstream`, `UpstreamRegistry` (COW), `RegistryOption`, `AllowPrivateUpstreams`, registration-time validation |
| `go/upstream_balancer.go` | `upstreamBalancer` — smooth weighted RR + per-URL cooldown, injectable clock |
| `go/upstream_transport.go` | `upstreamTransport` — `http.RoundTripper` doing per-attempt selection + failover |
| `go/upstream_router.go` | `Selector`, `RouteFunc`, default selector, `upstreamRouterConfig`, `UpstreamRouterOption`, handler + `httputil.ReverseProxy` assembly, `routerError`, ctx keys |
| `go/options.go` (modify) | `WithUpstreamRouter` + the `UpstreamRouterOption` helpers |
| `go/api.go` (modify) | `Engine.upstreamRouter` field; mount in `build()` |
| Tests | `upstream_registry_test.go`, `upstream_balancer_internal_test.go`, `upstream_transport_internal_test.go`, `upstream_router_test.go`, `upstream_router_example_test.go` |

Error codes are defined as `const` at the top of `upstream_router.go` (not `string_constants.go`, which is for cross-file shared literals — these are router-local).

---

## Task 1: `Upstream` + `UpstreamRegistry` (COW + validation)

**Files:**
- Create: `go/upstream_registry.go`
- Test: `go/upstream_registry_test.go`

- [ ] **Step 1: Write the failing tests**

Create `go/upstream_registry_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"sync"
	"testing"

	api "dappco.re/go/api"
)

func TestUpstreamRegistry_Good(t *testing.T) {
	reg := api.NewUpstreamRegistry()
	if err := reg.Set("lemma", api.Upstream{URL: "https://a.example.com:8000", Weight: 2}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := reg.Add("lemma", api.Upstream{URL: "https://b.example.com"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := reg.SetDefault(api.Upstream{URL: "https://fallback.example.com"}); err != nil {
		t.Fatalf("SetDefault: %v", err)
	}
	keys := reg.Keys()
	if len(keys) != 1 || keys[0] != "lemma" {
		t.Fatalf("Keys = %v, want [lemma]", keys)
	}
}

func TestUpstreamRegistry_Bad(t *testing.T) {
	reg := api.NewUpstreamRegistry()
	cases := map[string]string{
		"scheme":   "ftp://a.example.com",
		"no-host":  "http://",
		"bad-port": "http://a.example.com:99999",
		"creds":    "http://user:pass@a.example.com",
		"loopback": "http://127.0.0.1:11434",
		"private":  "http://10.0.0.5:8000",
		"metadata": "http://169.254.169.254",
	}
	for name, raw := range cases {
		if err := reg.Set("k", api.Upstream{URL: raw}); err == nil {
			t.Errorf("%s: Set(%q) = nil error, want rejection", name, raw)
		}
	}
}

func TestUpstreamRegistry_AllowPrivate_Good(t *testing.T) {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.Set("local", api.Upstream{URL: "http://127.0.0.1:11434"}); err != nil {
		t.Fatalf("Set loopback with allow-list: %v", err)
	}
	// Metadata stays hard-blocked even with a broad allow-list.
	reg2 := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("0.0.0.0/0"))
	if err := reg2.Set("meta", api.Upstream{URL: "http://169.254.169.254"}); err == nil {
		t.Fatal("metadata host accepted under broad allow-list, want rejection")
	}
}

func TestUpstreamRegistry_Ugly_ConcurrentWriteSnapshot(t *testing.T) {
	reg := api.NewUpstreamRegistry()
	_ = reg.Set("k", api.Upstream{URL: "https://a.example.com"})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); _ = reg.Add("k", api.Upstream{URL: "https://b.example.com"}) }()
		go func() { defer wg.Done(); _ = reg.Keys() }()
	}
	wg.Wait()
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestUpstreamRegistry`
Expected: FAIL — `undefined: api.NewUpstreamRegistry`, `api.Upstream`, `api.AllowPrivateUpstreams`.

- [ ] **Step 3: Write the implementation**

Create `go/upstream_registry.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net"     // Note: AX-6 — net.ParseIP/ParseCIDR are structural for SSRF IP-range checks.
	"net/url" // Note: AX-6 — url.URL fields are structural for upstream URL validation.
	"sort"
	"strconv"
	"sync"
	"sync/atomic"

	core "dappco.re/go"
)

// Upstream is one backend endpoint in a routing pool.
//
// Example:
//
//	api.Upstream{URL: "http://10.0.0.5:8000", Weight: 2}
type Upstream struct {
	URL     string            // http(s) base URL; validated at registration
	Weight  int               // weighted round-robin weight; <=0 treated as 1
	Headers map[string]string // static headers injected on dispatch (e.g. upstream API key)
}

// registrySnapshot is the immutable read-side view swapped atomically on writes.
type registrySnapshot struct {
	pools   map[string][]Upstream
	deflt   []Upstream
}

// UpstreamRegistry is the runtime-mutable, thread-safe pool table consumed by
// WithUpstreamRouter. Reads are lock-free (atomic snapshot load); writes take a
// mutex, clone, mutate, and swap (copy-on-write).
//
// Example:
//
//	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
//	_ = reg.Set("lemma", api.Upstream{URL: "http://127.0.0.1:11434"})
type UpstreamRegistry struct {
	mu      sync.Mutex
	snap    atomic.Pointer[registrySnapshot]
	allow   []*net.IPNet
	cidrErr error
}

// RegistryOption configures registration-time validation policy.
type RegistryOption func(*UpstreamRegistry)

// AllowPrivateUpstreams permits the given private/loopback/reserved CIDRs to
// pass registration validation. Without it the registry denies loopback,
// private, link-local, reserved, and metadata destinations by default. Metadata
// hosts stay hard-blocked regardless of the allow-list.
//
// Example:
//
//	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8", "10.0.0.0/8"))
func AllowPrivateUpstreams(cidrs ...string) RegistryOption {
	return func(r *UpstreamRegistry) {
		for _, raw := range cidrs {
			raw = core.Trim(raw)
			if raw == "" {
				continue
			}
			_, network, err := net.ParseCIDR(raw)
			if err != nil {
				if r.cidrErr == nil {
					r.cidrErr = core.E("UpstreamRegistry", "invalid AllowPrivateUpstreams CIDR "+raw, err)
				}
				continue
			}
			r.allow = append(r.allow, network)
		}
	}
}

// NewUpstreamRegistry creates an empty registry. Apply AllowPrivateUpstreams to
// widen the default-deny validation policy.
func NewUpstreamRegistry(opts ...RegistryOption) *UpstreamRegistry {
	r := &UpstreamRegistry{}
	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}
	r.snap.Store(&registrySnapshot{pools: map[string][]Upstream{}})
	return r
}

// Set replaces the pool for key. Returns an error (without mutating) if any
// upstream URL fails validation.
func (r *UpstreamRegistry) Set(key string, ups ...Upstream) error {
	if err := r.validateAll(ups); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	next := r.clone()
	next.pools[key] = cloneUpstreams(ups)
	r.snap.Store(next)
	return nil
}

// Add appends one upstream to the pool for key.
func (r *UpstreamRegistry) Add(key string, up Upstream) error {
	if err := r.validate(up); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	next := r.clone()
	next.pools[key] = append(cloneUpstreams(next.pools[key]), up)
	r.snap.Store(next)
	return nil
}

// Remove drops the pool for key.
func (r *UpstreamRegistry) Remove(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	next := r.clone()
	delete(next.pools, key)
	r.snap.Store(next)
}

// SetDefault sets the fallback pool used when a key has no explicit pool.
func (r *UpstreamRegistry) SetDefault(ups ...Upstream) error {
	if err := r.validateAll(ups); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	next := r.clone()
	next.deflt = cloneUpstreams(ups)
	r.snap.Store(next)
	return nil
}

// Keys returns the sorted set of explicitly-registered pool keys.
func (r *UpstreamRegistry) Keys() []string {
	snap := r.snap.Load()
	keys := make([]string, 0, len(snap.pools))
	for k := range snap.pools {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// resolve returns the pool for key (or the default pool) and whether one exists.
func (r *UpstreamRegistry) resolve(key string) ([]Upstream, bool) {
	snap := r.snap.Load()
	if pool, ok := snap.pools[key]; ok && len(pool) > 0 {
		return pool, true
	}
	if len(snap.deflt) > 0 {
		return snap.deflt, true
	}
	return nil, false
}

func (r *UpstreamRegistry) clone() *registrySnapshot {
	cur := r.snap.Load()
	next := &registrySnapshot{
		pools: make(map[string][]Upstream, len(cur.pools)),
		deflt: cur.deflt,
	}
	for k, v := range cur.pools {
		next.pools[k] = v
	}
	return next
}

func (r *UpstreamRegistry) validateAll(ups []Upstream) error {
	if len(ups) == 0 {
		return core.E("UpstreamRegistry", "pool must contain at least one upstream", nil)
	}
	for _, up := range ups {
		if err := r.validate(up); err != nil {
			return err
		}
	}
	return nil
}

func (r *UpstreamRegistry) validate(up Upstream) error {
	if r.cidrErr != nil {
		return r.cidrErr
	}
	return validateUpstreamURL(up.URL, r.allow)
}

// validateUpstreamURL enforces the block-by-default registration policy, reusing
// the root SSRF primitives (allowedSchemes, metadataHosts, blockedIPReason).
// Non-metadata hostnames are accepted without registration-time DNS (trusted
// config). IP literals in a denied range are rejected unless covered by allow.
func validateUpstreamURL(rawURL string, allow []*net.IPNet) error {
	rawURL = core.Trim(rawURL)
	if rawURL == "" {
		return core.E("UpstreamRegistry", "upstream URL is required", nil)
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return core.E("UpstreamRegistry", "invalid upstream URL "+rawURL, err)
	}
	if u.User != nil {
		return core.E("UpstreamRegistry", "upstream URL must not include credentials: "+rawURL, nil)
	}
	if _, ok := allowedSchemes[core.Lower(u.Scheme)]; !ok {
		return core.E("UpstreamRegistry", "upstream URL scheme must be http or https: "+rawURL, nil)
	}
	host := u.Hostname()
	if host == "" {
		return core.E("UpstreamRegistry", "upstream URL must include a host: "+rawURL, nil)
	}
	if port := u.Port(); port != "" {
		n, perr := strconv.Atoi(port)
		if perr != nil || n < 1 || n > 65535 {
			return core.E("UpstreamRegistry", "upstream URL port is invalid: "+rawURL, perr)
		}
	}
	if _, ok := metadataHosts[core.Lower(host)]; ok {
		return core.E("UpstreamRegistry", "metadata host is not permitted: "+host, nil)
	}
	if ip := net.ParseIP(host); ip != nil {
		if reason := blockedIPReason(ip); reason != "" && !ipAllowed(ip, allow) {
			return core.E("UpstreamRegistry", reason+" not permitted (use AllowPrivateUpstreams): "+host, nil)
		}
	}
	return nil
}

func ipAllowed(ip net.IP, allow []*net.IPNet) bool {
	for _, network := range allow {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func cloneUpstreams(ups []Upstream) []Upstream {
	if len(ups) == 0 {
		return nil
	}
	out := make([]Upstream, len(ups))
	copy(out, ups)
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestUpstreamRegistry -race`
Expected: PASS (all four tests, no data race).

- [ ] **Step 5: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/upstream_registry.go go/upstream_registry_test.go
git commit -m "$(printf 'feat(api): UpstreamRegistry — COW pool table + registration SSRF policy\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 2: `upstreamBalancer` (weighted RR + cooldown)

**Files:**
- Create: `go/upstream_balancer.go`
- Test: `go/upstream_balancer_internal_test.go`

- [ ] **Step 1: Write the failing tests**

Create `go/upstream_balancer_internal_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"testing"
	"time"
)

func TestUpstreamBalancer_WeightedSpread_Good(t *testing.T) {
	b := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	pool := []Upstream{{URL: "a", Weight: 2}, {URL: "b", Weight: 1}}
	counts := map[string]int{}
	for i := 0; i < 30; i++ {
		up, ok := b.pick("k", pool)
		if !ok {
			t.Fatal("pick returned !ok with healthy pool")
		}
		counts[up.URL]++
	}
	if counts["a"] != 20 || counts["b"] != 10 {
		t.Fatalf("weighted spread = %v, want a:20 b:10", counts)
	}
}

func TestUpstreamBalancer_CooldownSkip_Good(t *testing.T) {
	now := time.Unix(1000, 0)
	clock := func() time.Time { return now }
	b := newUpstreamBalancer(10*time.Second, clock)
	pool := []Upstream{{URL: "a", Weight: 1}, {URL: "b", Weight: 1}}

	b.markFailed("a")
	for i := 0; i < 5; i++ {
		up, ok := b.pick("k", pool)
		if !ok || up.URL != "b" {
			t.Fatalf("during cooldown got (%v,%v), want b", up.URL, ok)
		}
	}
	now = now.Add(11 * time.Second) // cooldown elapsed
	seen := map[string]bool{}
	for i := 0; i < 10; i++ {
		up, _ := b.pick("k", pool)
		seen[up.URL] = true
	}
	if !seen["a"] {
		t.Fatal("a not picked after cooldown elapsed")
	}
}

func TestUpstreamBalancer_AllCooling_Bad(t *testing.T) {
	b := newUpstreamBalancer(time.Minute, func() time.Time { return time.Unix(0, 0) })
	pool := []Upstream{{URL: "a"}, {URL: "b"}}
	b.markFailed("a")
	b.markFailed("b")
	if _, ok := b.pick("k", pool); ok {
		t.Fatal("pick returned ok with all upstreams cooling")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestUpstreamBalancer`
Expected: FAIL — `undefined: newUpstreamBalancer`.

- [ ] **Step 3: Write the implementation**

Create `go/upstream_balancer.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"sync"
	"time"
)

// upstreamBalancer performs smooth weighted round-robin selection over a pool,
// skipping upstreams in a cooldown window after a failure. State (per-key
// current weights, per-URL cooldown) is shared across requests behind a mutex —
// a failed upstream cools for every caller. The now func is injectable for tests.
type upstreamBalancer struct {
	mu       sync.Mutex
	current  map[string]map[string]int // key -> url -> SWRR current weight
	cooldown map[string]time.Time      // url -> cooling-until (global across keys)
	cool     time.Duration
	now      func() time.Time
}

func newUpstreamBalancer(cool time.Duration, now func() time.Time) *upstreamBalancer {
	if now == nil {
		now = time.Now
	}
	return &upstreamBalancer{
		current:  map[string]map[string]int{},
		cooldown: map[string]time.Time{},
		cool:     cool,
		now:      now,
	}
}

// pick selects the next upstream for key via smooth weighted round-robin over the
// non-cooling members of pool. Returns false when every member is cooling.
func (b *upstreamBalancer) pick(key string, pool []Upstream) (Upstream, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	t := b.now()
	cw := b.current[key]
	if cw == nil {
		cw = map[string]int{}
		b.current[key] = cw
	}

	bestIdx, total := -1, 0
	for i := range pool {
		up := pool[i]
		if until, ok := b.cooldown[up.URL]; ok && t.Before(until) {
			continue
		}
		w := up.Weight
		if w <= 0 {
			w = 1
		}
		cw[up.URL] += w
		total += w
		if bestIdx == -1 || cw[up.URL] > cw[pool[bestIdx].URL] {
			bestIdx = i
		}
	}
	if bestIdx == -1 {
		return Upstream{}, false
	}
	cw[pool[bestIdx].URL] -= total
	return pool[bestIdx], true
}

// markFailed puts url into a cooldown window starting now.
func (b *upstreamBalancer) markFailed(url string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cooldown[url] = b.now().Add(b.cool)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestUpstreamBalancer -race`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/upstream_balancer.go go/upstream_balancer_internal_test.go
git commit -m "$(printf 'feat(api): upstreamBalancer — smooth weighted RR + cooldown\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 3: `upstreamTransport` (failover RoundTripper)

**Files:**
- Create: `go/upstream_transport.go`
- Test: `go/upstream_transport_internal_test.go`

- [ ] **Step 1: Write the failing tests**

Create `go/upstream_transport_internal_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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
```

> Note: `core.As`, `routerError`, `poolCtxKey`, `keyCtxKey`, and `defaultFailoverStatuses` are defined in Task 4's `upstream_router.go`. This test file will not compile until Task 4 lands. Implement Task 3's production file now; if running tests before Task 4, expect a compile error naming those symbols (that IS the failing state). Otherwise reorder to write Task 4's `upstream_router.go` symbol stubs first — the recommended path is to do Steps 3 of Task 3 and Task 4 together, then run both test suites.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestUpstreamTransport`
Expected: FAIL — `undefined: upstreamTransport` (and `routerError`/`poolCtxKey`/etc. until Task 4).

- [ ] **Step 3: Write the implementation**

Create `go/upstream_transport.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import (
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
			if body, berr := out.GetBody(); berr == nil {
				out.Body = body
			}
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

func drainAndClose(body interface{ Close() error }) {
	if body != nil {
		_ = body.Close()
	}
}
```

- [ ] **Step 4: Run tests to verify they pass** (after Task 4 lands the shared symbols)

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestUpstreamTransport -race`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/upstream_transport.go go/upstream_transport_internal_test.go
git commit -m "$(printf 'feat(api): upstreamTransport — selection + passive failover RoundTripper\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 4: Router config, options, default selector, engine wiring

**Files:**
- Create: `go/upstream_router.go`
- Modify: `go/options.go` (add `WithUpstreamRouter` + option helpers)
- Modify: `go/api.go` (add `upstreamRouter` field; mount in `build()`)

- [ ] **Step 1: Write the implementation file (`upstream_router.go`)**

Create `go/upstream_router.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil" // Note: AX-6 — reverse-proxy mechanics are structural; no core primitive.
	"net/url"           // Note: AX-6 — url.Parse is structural for the Rewrite placeholder target.
	"strconv"
	"time"

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
)

const (
	defaultUpstreamRouterPath = "/v1/chat/completions"
	defaultUpstreamCooldown   = 10 * time.Second

	errCodeInvalidRequest      = "invalid_request"
	errCodeInvalidRequestBody  = "invalid_request_body"
	errCodeRoutingRejected     = "routing_rejected"
	errCodeNoUpstream          = "no_upstream_for_key"
	errCodeRequestTooLarge     = "request_too_large"
	errCodeUpstreamUnavailable = "upstream_unavailable"
	errCodeInvalidUpstreamResp = "invalid_upstream_response"
)

type ctxKey int

const (
	poolCtxKey ctxKey = iota
	keyCtxKey
	ginCtxKey
)

// Selector resolves the routing key from the request. body holds the (bounded)
// request body, already read by the handler; it may be empty for bodyless requests.
type Selector func(c *gin.Context, body []byte) (key string, err error)

// RouteFunc inspects the payload after the selector and may override the key or
// reject the request. Returning the same key is a no-op; a non-nil error aborts.
type RouteFunc func(c *gin.Context, key string, body []byte) (newKey string, err error)

// UpstreamRouterOption configures a router built by WithUpstreamRouter.
type UpstreamRouterOption func(*upstreamRouterConfig)

type upstreamRouterConfig struct {
	registry    *UpstreamRegistry
	selector    Selector
	hook        RouteFunc
	paths       []string
	inRaw       []any
	outRaw      []any
	in          []compiledTransformer
	out         []compiledTransformer
	maxAttempts int
	cooldown    time.Duration
	failover    map[int]bool
	transport   http.RoundTripper
}

// routerError carries an HTTP status + envelope code from the transport or
// ModifyResponse to the ReverseProxy ErrorHandler.
type routerError struct {
	status  int
	code    string
	message string
	cause   error
}

func (e *routerError) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
	}
	return e.message
}

func (e *routerError) Unwrap() error { return e.cause }

// WithSelector overrides the routing-key selector. Default: defaultModelSelector.
func WithSelector(fn Selector) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.selector = fn }
}

// WithRouteHook installs a decision hook to inspect the payload and override/reject.
func WithRouteHook(fn RouteFunc) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.hook = fn }
}

// WithRouterPaths sets the mounted paths (default ["/v1/chat/completions"]).
// Each path forwards its own path + query to the chosen upstream.
func WithRouterPaths(paths ...string) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.paths = paths }
}

// WithUpstreamTransformerIn adds request-body transformers (reuses the existing
// TransformerIn machinery; FieldRenamer etc. work). Operates on the raw body.
func WithUpstreamTransformerIn(t ...any) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.inRaw = append(cfg.inRaw, t...) }
}

// WithUpstreamTransformerOut adds response-body transformers, applied only to
// buffered (non-streaming) responses, on the raw upstream body.
func WithUpstreamTransformerOut(t ...any) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.outRaw = append(cfg.outRaw, t...) }
}

// WithFailover sets the max upstream attempts (default len(pool), each tried once)
// and the cooldown applied to a failed upstream (default 10s).
func WithFailover(maxAttempts int, cooldown time.Duration) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) {
		cfg.maxAttempts = maxAttempts
		if cooldown > 0 {
			cfg.cooldown = cooldown
		}
	}
}

// WithFailoverStatuses overrides which response statuses trigger failover
// (default: all >= 500). Pass e.g. 429 to also fail over on rate-limit responses.
func WithFailoverStatuses(statuses ...int) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) {
		cfg.failover = map[int]bool{}
		for _, s := range statuses {
			cfg.failover[s] = true
		}
	}
}

// WithUpstreamTransport sets the base RoundTripper used for dispatch (custom TLS,
// timeouts). Default: a clone of http.DefaultTransport.
func WithUpstreamTransport(rt http.RoundTripper) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.transport = rt }
}

// defaultFailoverStatuses returns the default failover status set: all >= 500.
func defaultFailoverStatuses() map[int]bool {
	m := map[int]bool{}
	for s := 500; s <= 599; s++ {
		m[s] = true
	}
	return m
}

// defaultModelSelector reads the OpenAI-style "model" field from a JSON body.
func defaultModelSelector(_ *gin.Context, body []byte) (string, error) {
	var probe struct {
		Model string `json:"model"`
	}
	if res := core.JSONUnmarshal(body, &probe); !res.OK {
		return "", core.E("upstream.selector", "request body is not valid JSON", nil)
	}
	if core.Trim(probe.Model) == "" {
		return "", core.E("upstream.selector", "request body has no \"model\" field", nil)
	}
	return probe.Model, nil
}

func poolFromContext(ctx context.Context) ([]Upstream, bool) {
	pool, ok := ctx.Value(poolCtxKey).([]Upstream)
	return pool, ok
}

func keyFromContext(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(keyCtxKey).(string)
	return key, ok
}

// finalise resolves defaults and compiles transformer pipelines. Returns an
// error if a transformer fails to compile.
func (cfg *upstreamRouterConfig) finalise() error {
	if cfg.selector == nil {
		cfg.selector = defaultModelSelector
	}
	if len(cfg.paths) == 0 {
		cfg.paths = []string{defaultUpstreamRouterPath}
	}
	if cfg.cooldown <= 0 {
		cfg.cooldown = defaultUpstreamCooldown
	}
	if cfg.failover == nil {
		cfg.failover = defaultFailoverStatuses()
	}
	if cfg.transport == nil {
		cfg.transport = http.DefaultTransport
	}
	in, err := compileTransformerPipeline(transformerDirectionIn, cfg.inRaw)
	if err != nil {
		return err
	}
	out, err := compileTransformerPipeline(transformerDirectionOut, cfg.outRaw)
	if err != nil {
		return err
	}
	cfg.in, cfg.out = in, out
	return nil
}

// buildProxy constructs the shared ReverseProxy for the router.
func (cfg *upstreamRouterConfig) buildProxy() *httputil.ReverseProxy {
	balancer := newUpstreamBalancer(cfg.cooldown, time.Now)
	transport := &upstreamTransport{
		base:        cfg.transport,
		balancer:    balancer,
		maxAttempts: cfg.maxAttempts,
		failover:    cfg.failover,
	}
	return &httputil.ReverseProxy{
		Transport:     transport,
		FlushInterval: -1, // stream SSE / chunked responses through immediately
		Rewrite: func(pr *httputil.ProxyRequest) {
			// Placeholder target so the pipeline has a valid URL; the transport
			// overrides scheme/host/path per attempt for the selected upstream.
			if pool, ok := poolFromContext(pr.In.Context()); ok && len(pool) > 0 {
				if target, err := url.Parse(pool[0].URL); err == nil {
					pr.Out.URL.Scheme = target.Scheme
					pr.Out.URL.Host = target.Host
				}
			}
			pr.SetXForwarded()
		},
		ModifyResponse: cfg.modifyResponse,
		ErrorHandler:   cfg.errorHandler,
	}
}

func (cfg *upstreamRouterConfig) modifyResponse(resp *http.Response) error {
	if len(cfg.out) == 0 {
		return nil
	}
	if isEventStream(resp.Header.Get("Content-Type")) {
		return nil // streaming: pass through untransformed
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return &routerError{status: http.StatusBadGateway, code: errCodeInvalidUpstreamResp, message: "could not read upstream response", cause: err}
	}
	c, _ := resp.Request.Context().Value(ginCtxKey).(*gin.Context)
	transformed, err := runTransformerPipeline(c, body, cfg.out)
	if err != nil {
		return &routerError{status: http.StatusBadGateway, code: errCodeInvalidUpstreamResp, message: "response transform failed", cause: err}
	}
	resp.Body = io.NopCloser(bytes.NewReader(transformed))
	resp.ContentLength = int64(len(transformed))
	resp.Header.Set("Content-Length", strconv.Itoa(len(transformed)))
	return nil
}

func (cfg *upstreamRouterConfig) errorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	re := &routerError{status: http.StatusBadGateway, code: errCodeUpstreamUnavailable, message: "upstream request failed"}
	var got *routerError
	if core.As(err, &got) {
		re = got
	}
	slog.Warn("upstream router dispatch failed", "code", re.code, "err", err.Error())
	w.Header().Set("Content-Type", "application/json")
	if re.status == http.StatusServiceUnavailable {
		w.Header().Set("Retry-After", strconv.Itoa(int(cfg.cooldown.Seconds())))
	}
	w.WriteHeader(re.status)
	_ = json.NewEncoder(w).Encode(Fail(re.code, re.message))
}

// handler returns the gin.HandlerFunc mounted at each router path.
func (cfg *upstreamRouterConfig) handler(proxy *httputil.ReverseProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, ok := readUpstreamBody(c)
		if !ok {
			return
		}

		key, err := cfg.selector(c, body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Fail(errCodeInvalidRequest, err.Error()))
			return
		}
		if cfg.hook != nil {
			newKey, herr := cfg.hook(c, key, body)
			if herr != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, Fail(errCodeRoutingRejected, herr.Error()))
				return
			}
			if core.Trim(newKey) != "" {
				key = newKey
			}
		}

		if len(cfg.in) > 0 {
			body, err = runTransformerPipeline(c, body, cfg.in)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, Fail(errCodeInvalidRequestBody, err.Error()))
				return
			}
		}

		pool, ok := cfg.registry.resolve(key)
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, Fail(errCodeNoUpstream, "no upstream registered for key: "+key))
			return
		}

		bound := body // capture for GetBody closure
		c.Request.Body = io.NopCloser(bytes.NewReader(bound))
		c.Request.ContentLength = int64(len(bound))
		c.Request.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(bound)), nil }

		ctx := context.WithValue(c.Request.Context(), poolCtxKey, pool)
		ctx = context.WithValue(ctx, keyCtxKey, key)
		ctx = context.WithValue(ctx, ginCtxKey, c)
		c.Request = c.Request.WithContext(ctx)

		proxy.ServeHTTP(upstreamResponseWriter(c), c.Request)
	}
}

// upstreamResponseWriter unwraps gin's ResponseWriter to the underlying
// http.ResponseWriter, which httputil.ReverseProxy requires for flush/cancel.
func upstreamResponseWriter(c *gin.Context) http.ResponseWriter {
	var w http.ResponseWriter = c.Writer
	if uw, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
		w = uw.Unwrap()
	}
	return w
}

func readUpstreamBody(c *gin.Context) ([]byte, bool) {
	limited := http.MaxBytesReader(c.Writer, c.Request.Body, maxToolRequestBodyBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		if err.Error() == "http: request body too large" {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, Fail(errCodeRequestTooLarge, "Request body exceeds the maximum allowed size"))
			return nil, false
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, Fail(errCodeInvalidRequest, "Unable to read request body"))
		return nil, false
	}
	return body, true
}

func isEventStream(contentType string) bool {
	return core.HasPrefix(core.Lower(core.Trim(contentType)), "text/event-stream")
}
```

> The Rewrite target is only a placeholder to satisfy `httputil.ReverseProxy` (which requires a non-nil `Rewrite`/`Director`); `upstreamTransport.RoundTrip` overrides scheme/host/path per attempt for the actually-selected upstream, so `pool[0]` here is never the real dispatch target.

- [ ] **Step 2: Add the engine field and mount (modify `api.go`)**

In `go/api.go`, add the field to the `Engine` struct (after `noRouteHandler gin.HandlerFunc` at line ~115):

```go
	// upstreamRouter, when set via WithUpstreamRouter, mounts a selector-keyed
	// reverse proxy over a pool of HTTP upstreams at the configured paths.
	upstreamRouter *upstreamRouterConfig
```

In `go/api.go` `build()`, after the chat-completions mount block (line ~443) add:

```go
	// Mount the selector-keyed upstream router when configured.
	if e.upstreamRouter != nil {
		proxy := e.upstreamRouter.buildProxy()
		h := e.upstreamRouter.handler(proxy)
		for _, p := range e.upstreamRouter.paths {
			r.Any(p, h)
		}
	}
```

- [ ] **Step 3: Add `WithUpstreamRouter` (modify `options.go`)**

In `go/options.go`, after `WithChatCompletionsPath` (line ~849) add:

```go
// WithUpstreamRouter mounts a selector-keyed reverse proxy that load-balances
// each request across a runtime-mutable pool of HTTP upstreams (weighted
// round-robin + passive failover, hybrid streaming, decision hook, transformer
// composition). The registry is the source of truth for upstreams.
//
// Example:
//
//	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
//	_ = reg.Set("lemma", api.Upstream{URL: "http://127.0.0.1:11434"})
//	engine, _ := api.New(api.WithUpstreamRouter(reg))
func WithUpstreamRouter(reg *UpstreamRegistry, opts ...UpstreamRouterOption) Option {
	return func(e *Engine) {
		if reg == nil {
			return
		}
		cfg := &upstreamRouterConfig{registry: reg}
		for _, opt := range opts {
			if opt != nil {
				opt(cfg)
			}
		}
		if err := cfg.finalise(); err != nil {
			// Transformer compile errors mirror the panic contract used by
			// transformerRouteConfigForDescription (transformer_in.go:78).
			panic(err)
		}
		e.upstreamRouter = cfg
	}
}
```

- [ ] **Step 4: Build and run all prior unit suites together**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go build ./ && GOWORK=off go test ./ -run 'TestUpstream' -race`
Expected: PASS for `TestUpstreamRegistry*`, `TestUpstreamBalancer*`, `TestUpstreamTransport*` (Task 3's tests now compile and pass).

- [ ] **Step 5: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/upstream_router.go go/options.go go/api.go go/upstream_transport_internal_test.go
git commit -m "$(printf 'feat(api): WithUpstreamRouter — config, options, default model selector, engine mount\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 5: Integration tests (httptest end-to-end)

**Files:**
- Create: `go/upstream_router_test.go`

- [ ] **Step 1: Write the failing integration tests**

Create `go/upstream_router_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	api "dappco.re/go/api"
	"github.com/gin-gonic/gin"
)

// newEngine builds a test engine with the router mounted, returning a live server.
func serve(t *testing.T, reg *api.UpstreamRegistry, opts ...api.UpstreamRouterOption) *httptest.Server {
	t.Helper()
	e, err := api.New(api.WithUpstreamRouter(reg, opts...))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return httptest.NewServer(e.Handler())
}

func post(t *testing.T, base, path, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(base+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func TestUpstreamRouter_RoutesByModel_Good(t *testing.T) {
	upA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"upstream":"A"}`)
	}))
	defer upA.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.Set("lemma", api.Upstream{URL: upA.URL}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	srv := serve(t, reg)
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"lemma"}`)
	defer resp.Body.Close()
	got, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(got), `"upstream":"A"`) {
		t.Fatalf("body = %s, want routed to A", got)
	}
}

func TestUpstreamRouter_MissingModel_Bad(t *testing.T) {
	reg := api.NewUpstreamRegistry()
	_ = reg.SetDefault(api.Upstream{URL: "https://example.com"})
	srv := serve(t, reg)
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestUpstreamRouter_Failover_Good(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer dead.Close()
	live := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer live.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.Set("m", api.Upstream{URL: dead.URL}, api.Upstream{URL: live.URL}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	srv := serve(t, reg)
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m"}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (failed over to live)", resp.StatusCode)
	}
}

func TestUpstreamRouter_AllDown_503_Ugly(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer dead.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("m", api.Upstream{URL: dead.URL})
	srv := serve(t, reg)
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m"}`)
	defer resp.Body.Close()
	got, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Error("missing Retry-After header on 503")
	}
	if strings.Contains(string(got), dead.URL) {
		t.Error("upstream URL leaked into client response body")
	}
}

func TestUpstreamRouter_StreamingPassthrough_Good(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f, _ := w.(http.Flusher)
		for _, chunk := range []string{"data: a\n\n", "data: b\n\n", "data: [DONE]\n\n"} {
			_, _ = io.WriteString(w, chunk)
			if f != nil {
				f.Flush()
			}
		}
	}))
	defer up.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("m", api.Upstream{URL: up.URL})
	// Out transformer present to prove it is NOT applied to streams.
	srv := serve(t, reg, api.WithUpstreamTransformerOut(api.RenameFields(map[string]string{"x": "y"})))
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m"}`)
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}
	sc := bufio.NewScanner(resp.Body)
	var lines int
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "data:") {
			lines++
		}
	}
	if lines != 3 {
		t.Fatalf("got %d data lines, want 3 (stream byte-preserved)", lines)
	}
}

func TestUpstreamRouter_TransformInOut_Good(t *testing.T) {
	var gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = io.WriteString(w, `{"internal_id":42}`)
	}))
	defer up.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("m", api.Upstream{URL: up.URL})
	srv := serve(t, reg,
		api.WithUpstreamTransformerIn(api.RenameFields(map[string]string{"q": "prompt"})),
		api.WithUpstreamTransformerOut(api.RenameFields(map[string]string{"internal_id": "id"})),
	)
	defer srv.Close()

	// Selector reads "model" from the original body; the in-transform then renames
	// q->prompt before dispatch so the upstream sees the translated shape.
	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m","q":"hello"}`)
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if !strings.Contains(gotBody, `"prompt"`) {
		t.Errorf("upstream body = %s, want renamed q->prompt", gotBody)
	}
	if !strings.Contains(string(out), `"id":42`) {
		t.Errorf("client body = %s, want renamed internal_id->id", out)
	}
}

func TestUpstreamRouter_RouteHookOverride_Good(t *testing.T) {
	upB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"pool":"B"}`)
	}))
	defer upB.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.Set("b", api.Upstream{URL: upB.URL})
	srv := serve(t, reg, api.WithRouteHook(func(_ *gin.Context, _ string, _ []byte) (string, error) {
		return "b", nil
	}))
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"anything"}`)
	defer resp.Body.Close()
	got, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(got), `"pool":"B"`) {
		t.Fatalf("body = %s, want hook-overridden pool B", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail then pass**

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestUpstreamRouter -race`
Expected: after fixing the `*gin.Context` import note, PASS for all seven integration tests.

- [ ] **Step 3: SSRF-posture integration assertion**

Add to `go/upstream_router_test.go`:

```go
func TestUpstreamRouter_SSRFPosture_Bad(t *testing.T) {
	reg := api.NewUpstreamRegistry() // no allow-list
	if err := reg.Set("m", api.Upstream{URL: "http://127.0.0.1:11434"}); err == nil {
		t.Fatal("loopback accepted without AllowPrivateUpstreams, want rejection")
	}
}

func TestUpstreamRouter_Composition_Middleware_Good(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer up.Close()

	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	_ = reg.SetDefault(api.Upstream{URL: up.URL})
	// WithSunset adds a Sunset header to every response via engine middleware.
	e, _ := api.New(api.WithSunset("2026-12-31", "https://api.example.com/v2"), api.WithUpstreamRouter(reg))
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp := post(t, srv.URL, "/v1/chat/completions", `{"model":"m"}`)
	defer resp.Body.Close()
	if resp.Header.Get("Sunset") == "" {
		t.Fatal("Sunset header absent — engine middleware did not wrap the mounted router")
	}
}
```

> Uses `WithSunset` (deterministic per-response header) rather than auth to prove engine middleware wraps the mounted router — the API's bearer middleware is permissive, so a missing token does not reliably 401. Confirm the exact header name is `Sunset` (RFC 8594; see `sunset.go`).

Run: `cd /Users/snider/Code/core/api/go && GOWORK=off go test ./ -run TestUpstreamRouter -race`
Expected: PASS (all integration tests including SSRF + composition).

- [ ] **Step 4: Write the example test (godoc-facing)**

Create `go/upstream_router_example_test.go`:

```go
// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"fmt"

	api "dappco.re/go/api"
)

func ExampleWithUpstreamRouter() {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8", "10.0.0.0/8"))
	_ = reg.Set("lemma",
		api.Upstream{URL: "http://10.0.0.5:8000", Weight: 2},
		api.Upstream{URL: "http://10.0.0.6:8000", Weight: 1},
	)
	_ = reg.SetDefault(api.Upstream{URL: "http://127.0.0.1:11434"})

	engine, err := api.New(api.WithUpstreamRouter(reg))
	if err != nil {
		panic(err)
	}
	fmt.Println(engine.Addr())
	// Output: :8080
}
```

- [ ] **Step 5: Commit**

```bash
cd /Users/snider/Code/core/api
git add go/upstream_router_test.go go/upstream_router_example_test.go
git commit -m "$(printf 'test(api): upstream router integration — routing, failover, streaming, transforms, SSRF, composition\n\nCo-Authored-By: Virgil <virgil@lethean.io>')"
```

---

## Task 6: Full QA gate

**Files:** none (verification only)

- [ ] **Step 1: Format + vet + full test with race**

Run:
```bash
cd /Users/snider/Code/core/api/go
gofmt -l upstream_registry.go upstream_balancer.go upstream_transport.go upstream_router.go
GOWORK=off go vet ./
GOWORK=off go test ./ -race -count=1
```
Expected: `gofmt -l` prints nothing (all formatted); `vet` clean; all tests PASS.

- [ ] **Step 2: Lint + security audit (matches repo `core go qa full`)**

Run:
```bash
cd /Users/snider/Code/core/api/go
GOWORK=off go test ./ -run 'Example' -count=1   # godoc examples compile + match Output
golangci-lint run ./ 2>/dev/null || echo "run golangci-lint if available"
gosec -quiet ./ 2>/dev/null || echo "run gosec if available"
```
Expected: example output matches; lint clean; `gosec` reports only the annotated `#nosec G107` on `upstreamTransport.RoundTrip` (justified — registration-validated operator upstreams).

- [ ] **Step 3: Confirm the gateway binary still builds**

Run: `cd /Users/snider/Code/core/api/go && go build ./cmd/gateway/ && GOWORK=off go build ./...`
Expected: exit 0 (no regression to existing build).

- [ ] **Step 4: Commit any formatting/lint fixes**

```bash
cd /Users/snider/Code/core/api
git add -A go/
git commit -m "$(printf 'chore(api): gofmt + lint pass for upstream router\n\nCo-Authored-By: Virgil <virgil@lethean.io>')" || echo "nothing to commit"
```

---

## Spec coverage check

| Spec section | Task |
|---|---|
| §4 `WithUpstreamRouter`, `Upstream`, `UpstreamRegistry`, `Selector`, `RouteFunc`, options | Tasks 1, 4 |
| §4 `AllowPrivateUpstreams` registry option | Task 1 |
| §5 `UpstreamRegistry` / `upstreamBalancer` / `upstreamTransport` / handler units | Tasks 1, 2, 3, 4 |
| §6 data flow (body→selector→hook→transformIn→pool→proxy→transport→response) | Tasks 4, 5 |
| §7 error taxonomy (400/403/404/413/502/503 + passthrough) | Tasks 4, 5 |
| §8 SSRF block-by-default + opt-in + `#nosec` + no URL leak | Tasks 1, 3, 5 |
| §9 testing matrix (Good/Bad/Ugly, weighted spread, cooldown, streaming, transforms, composition) | Tasks 1–5 |
| §10 file layout | all |

**Deferred to future extensions (spec §11), not in this plan:** sticky/consistent-hash, active health checks, direct-upstream hook return, per-chunk stream transforms, per-pool rate limits.
