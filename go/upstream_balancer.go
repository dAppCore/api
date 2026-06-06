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
