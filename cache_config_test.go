// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"testing"
	"time"

	api "dappco.re/go/api"
)

// TestCacheConfig_Good_SnapshotsConfiguredEngine verifies that CacheConfig
// reflects the cache limits supplied during engine construction.
func TestCacheConfig_Good_SnapshotsConfiguredEngine(t *testing.T) {
	e, err := api.New(api.WithCacheLimits(5*time.Minute, 10, 1024))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.CacheConfig()

	if !cfg.Enabled {
		t.Fatal("expected cache config to be enabled")
	}
	if cfg.TTL != 5*time.Minute {
		t.Fatalf("expected TTL %v, got %v", 5*time.Minute, cfg.TTL)
	}
	if cfg.MaxEntries != 10 {
		t.Fatalf("expected MaxEntries 10, got %d", cfg.MaxEntries)
	}
	if cfg.MaxBytes != 1024 {
		t.Fatalf("expected MaxBytes 1024, got %d", cfg.MaxBytes)
	}
}

// TestCacheConfig_Bad_NilEngineReturnsZeroValue verifies the nil-receiver
// guard returns an empty snapshot instead of panicking.
func TestCacheConfig_Bad_NilEngineReturnsZeroValue(t *testing.T) {
	var e *api.Engine

	cfg := e.CacheConfig()
	if cfg != (api.CacheConfig{}) {
		t.Fatalf("expected zero-value cache config, got %+v", cfg)
	}
}

// TestCacheConfig_Ugly_UnconfiguredEngineStaysDisabled verifies that an
// engine without cache middleware reports a disabled snapshot.
func TestCacheConfig_Ugly_UnconfiguredEngineStaysDisabled(t *testing.T) {
	e, err := api.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := e.CacheConfig()
	if cfg.Enabled {
		t.Fatal("expected cache config to remain disabled")
	}
	if cfg.TTL != 0 || cfg.MaxEntries != 0 || cfg.MaxBytes != 0 {
		t.Fatalf("expected zero cache settings, got %+v", cfg)
	}
}
