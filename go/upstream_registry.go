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
	pools map[string][]Upstream
	deflt []Upstream
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
	if len(ups) == 0 {
		return core.E("UpstreamRegistry", "SetDefault requires at least one upstream", nil)
	}
	if err := r.validateEach(ups); err != nil {
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
// The returned slice is the snapshot's own backing slice — callers must treat it
// as read-only and never mutate its elements or append to it in place.
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
		deflt: cloneUpstreams(cur.deflt),
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
	return r.validateEach(ups)
}

// validateEach validates every upstream in ups without the non-empty check, so
// callers can supply their own empty-pool message.
func (r *UpstreamRegistry) validateEach(ups []Upstream) error {
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

// cloneUpstreams returns a deep copy of ups. The Headers map on each upstream is
// copied into a fresh map so a caller mutating their original map after a write
// — or the transport iterating it concurrently — cannot race the stored snapshot.
func cloneUpstreams(ups []Upstream) []Upstream {
	if len(ups) == 0 {
		return nil
	}
	out := make([]Upstream, len(ups))
	copy(out, ups)
	for i := range out {
		if len(out[i].Headers) == 0 {
			out[i].Headers = nil
			continue
		}
		headers := make(map[string]string, len(out[i].Headers))
		for k, v := range out[i].Headers {
			headers[k] = v
		}
		out[i].Headers = headers
	}
	return out
}
