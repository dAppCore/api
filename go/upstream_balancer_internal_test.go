// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"sync"
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
	clock := func() time.Time { return now } // now is mutated below; clock() reads it live by closure
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

// TestUpstreamBalancer_ConcurrentPickMark_Ugly hammers the shared mutex from
// many goroutines (Task 3 drives this balancer concurrently). It asserts
// nothing beyond running clean under -race: no data race, no panic.
func TestUpstreamBalancer_ConcurrentPickMark_Ugly(t *testing.T) {
	b := newUpstreamBalancer(time.Minute, time.Now)
	pool := []Upstream{{URL: "a", Weight: 2}, {URL: "b", Weight: 1}}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); _, _ = b.pick("k", pool) }()
		go func() { defer wg.Done(); b.markFailed("a") }()
	}
	wg.Wait()
}
