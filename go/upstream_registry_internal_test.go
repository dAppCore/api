// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"sync"
	"testing"
)

// TestUpstreamRegistry_ResolveDefaultFallback_Good covers the unexported resolve
// path: an unknown key must fall back to the default pool when one is set.
func TestUpstreamRegistry_ResolveDefaultFallback_Good(t *testing.T) {
	reg := NewUpstreamRegistry()
	if err := reg.Set("known", Upstream{URL: "https://known.example.com"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := reg.SetDefault(Upstream{URL: "https://fallback.example.com"}); err != nil {
		t.Fatalf("SetDefault: %v", err)
	}

	pool, ok := reg.resolve("unknown")
	if !ok {
		t.Fatal("resolve(unknown) = !ok, want default-pool fallback")
	}
	if len(pool) != 1 || pool[0].URL != "https://fallback.example.com" {
		t.Fatalf("resolve(unknown) = %v, want the default pool", pool)
	}

	// An explicitly-registered key still resolves to its own pool, not default.
	pool, ok = reg.resolve("known")
	if !ok || len(pool) != 1 || pool[0].URL != "https://known.example.com" {
		t.Fatalf("resolve(known) = (%v,%v), want the known pool", pool, ok)
	}
}

// TestUpstreamRegistry_ResolveNoDefault_Bad covers resolve with no matching key
// and no default pool: it must report !ok.
func TestUpstreamRegistry_ResolveNoDefault_Bad(t *testing.T) {
	reg := NewUpstreamRegistry()
	if _, ok := reg.resolve("missing"); ok {
		t.Fatal("resolve(missing) = ok with no default pool, want !ok")
	}
}

// TestUpstreamRegistry_RemoveThenResolve_Good proves Remove drops a key so that
// resolve no longer returns its pool and falls through to the default.
func TestUpstreamRegistry_RemoveThenResolve_Good(t *testing.T) {
	reg := NewUpstreamRegistry()
	if err := reg.Set("k", Upstream{URL: "https://a.example.com"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := reg.SetDefault(Upstream{URL: "https://fallback.example.com"}); err != nil {
		t.Fatalf("SetDefault: %v", err)
	}

	reg.Remove("k")

	pool, ok := reg.resolve("k")
	if !ok || len(pool) != 1 || pool[0].URL != "https://fallback.example.com" {
		t.Fatalf("resolve(k) after Remove = (%v,%v), want the default pool", pool, ok)
	}
}

// TestUpstreamRegistry_Ugly_HeadersDeepCopy proves the registry deep-copies an
// upstream's Headers map so the stored snapshot does not alias the caller's map.
// One goroutine mutates the caller's original map; another iterates the stored
// (cloned) map fetched via resolve. With a shallow copy both touch the same map
// and the race detector fires (and "concurrent map writes" panics); with the
// deep copy the stored map is independent, so neither happens.
func TestUpstreamRegistry_Ugly_HeadersDeepCopy(t *testing.T) {
	reg := NewUpstreamRegistry()
	headers := map[string]string{"Authorization": "Bearer up-key"}
	if err := reg.Set("k", Upstream{URL: "https://a.example.com", Headers: headers}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	stored, ok := reg.resolve("k")
	if !ok || len(stored) != 1 || stored[0].Headers == nil {
		t.Fatalf("resolve(k) = (%v,%v), want one upstream with Headers", stored, ok)
	}
	storedHeaders := stored[0].Headers

	var wg sync.WaitGroup
	wg.Add(2)
	// Single caller-side mutator: serialised writes to the original map, so any
	// race surfaced is strictly caller-map vs stored-map aliasing, not the test
	// racing itself.
	go func() {
		defer wg.Done()
		for range 1000 {
			headers["Authorization"] = "Bearer rotated"
		}
	}()
	// Reader iterating the stored map — must be a distinct map after the deep copy.
	go func() {
		defer wg.Done()
		for range 1000 {
			for range storedHeaders {
			}
		}
	}()
	wg.Wait()

	if got := storedHeaders["Authorization"]; got != "Bearer up-key" {
		t.Fatalf("stored Headers mutated by caller = %q, want unchanged Bearer up-key", got)
	}
}
