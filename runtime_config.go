// SPDX-License-Identifier: EUPL-1.2

package api

// RuntimeConfig captures the engine's current runtime-facing configuration in
// a single snapshot.
//
// It groups the existing Swagger, transport, cache, and i18n snapshots so
// callers can inspect the active engine surface without joining multiple
// method results themselves.
//
// Example:
//
//	cfg := engine.RuntimeConfig()
type RuntimeConfig struct {
	Swagger   SwaggerConfig
	Transport TransportConfig
	Cache     CacheConfig
	I18n      I18nConfig
}

// RuntimeConfig returns a stable snapshot of the engine's current runtime
// configuration.
//
// The result clones the underlying snapshots so callers can safely retain or
// modify the returned value.
//
// Example:
//
//	cfg := engine.RuntimeConfig()
func (e *Engine) RuntimeConfig() RuntimeConfig {
	if e == nil {
		return RuntimeConfig{}
	}

	return RuntimeConfig{
		Swagger:   e.SwaggerConfig(),
		Transport: e.TransportConfig(),
		Cache:     e.CacheConfig(),
		I18n:      e.I18nConfig(),
	}
}
