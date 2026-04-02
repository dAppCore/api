// SPDX-License-Identifier: EUPL-1.2

package api

import "time"

// CacheConfig captures the configured response cache settings for an Engine.
//
// It is intentionally small and serialisable so callers can inspect the active
// cache policy without needing to rebuild middleware state.
//
// Example:
//
//	cfg := api.CacheConfig{Enabled: true, TTL: 5 * time.Minute}
type CacheConfig struct {
	Enabled    bool
	TTL        time.Duration
	MaxEntries int
	MaxBytes   int
}

// CacheConfig returns the currently configured response cache settings for the engine.
//
// The result snapshots the Engine state at call time.
//
// Example:
//
//	cfg := engine.CacheConfig()
func (e *Engine) CacheConfig() CacheConfig {
	if e == nil {
		return CacheConfig{}
	}

	cfg := CacheConfig{
		TTL:        e.cacheTTL,
		MaxEntries: e.cacheMaxEntries,
		MaxBytes:   e.cacheMaxBytes,
	}
	if e.cacheTTL > 0 {
		cfg.Enabled = true
	}
	return cfg
}
