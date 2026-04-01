// SPDX-Licence-Identifier: EUPL-1.2

package provider

import (
	"iter"
	"slices"
	"sync"

	"dappco.re/go/core/api"
)

// Registry collects providers and mounts them on an api.Engine.
// It is a convenience wrapper — providers could be registered directly
// via engine.Register(), but the Registry enables discovery by consumers
// (GUI, MCP) that need to query provider capabilities.
type Registry struct {
	mu        sync.RWMutex
	providers []Provider
}

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Add registers a provider. Providers are mounted in the order they are added.
func (r *Registry) Add(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = append(r.providers, p)
}

// MountAll registers every provider with the given api.Engine.
// Each provider is passed to engine.Register(), which mounts it as a
// RouteGroup at its BasePath with all configured middleware.
func (r *Registry) MountAll(engine *api.Engine) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.providers {
		engine.Register(p)
	}
}

// List returns a copy of all registered providers.
func (r *Registry) List() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.providers)
}

// Iter returns an iterator over all registered providers.
func (r *Registry) Iter() iter.Seq[Provider] {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Values(slices.Clone(r.providers))
}

// Len returns the number of registered providers.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.providers)
}

// Get returns a provider by name, or nil if not found.
func (r *Registry) Get(name string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.providers {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// Streamable returns all providers that implement the Streamable interface.
func (r *Registry) Streamable() []Streamable {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []Streamable
	for _, p := range r.providers {
		if s, ok := p.(Streamable); ok {
			result = append(result, s)
		}
	}
	return result
}

// Describable returns all providers that implement the Describable interface.
func (r *Registry) Describable() []Describable {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []Describable
	for _, p := range r.providers {
		if d, ok := p.(Describable); ok {
			result = append(result, d)
		}
	}
	return result
}

// Renderable returns all providers that implement the Renderable interface.
func (r *Registry) Renderable() []Renderable {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []Renderable
	for _, p := range r.providers {
		if rv, ok := p.(Renderable); ok {
			result = append(result, rv)
		}
	}
	return result
}

// ProviderInfo is a serialisable summary of a registered provider.
type ProviderInfo struct {
	Name     string       `json:"name"`
	BasePath string       `json:"basePath"`
	Channels []string     `json:"channels,omitempty"`
	Element  *ElementSpec `json:"element,omitempty"`
	SpecFile string       `json:"specFile,omitempty"`
	Upstream string       `json:"upstream,omitempty"`
}

// Info returns a summary of all registered providers.
func (r *Registry) Info() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(r.providers))
	for _, p := range r.providers {
		info := ProviderInfo{
			Name:     p.Name(),
			BasePath: p.BasePath(),
		}
		if s, ok := p.(Streamable); ok {
			info.Channels = s.Channels()
		}
		if rv, ok := p.(Renderable); ok {
			elem := rv.Element()
			info.Element = &elem
		}
		if sf, ok := p.(interface{ SpecFile() string }); ok {
			info.SpecFile = sf.SpecFile()
		}
		if up, ok := p.(interface{ Upstream() string }); ok {
			info.Upstream = up.Upstream()
		}
		infos = append(infos, info)
	}
	return infos
}

// SpecFiles returns all non-empty provider OpenAPI spec file paths.
// The result is deduplicated and sorted for stable discovery output.
func (r *Registry) SpecFiles() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	files := make(map[string]struct{}, len(r.providers))
	for _, p := range r.providers {
		if sf, ok := p.(interface{ SpecFile() string }); ok {
			if path := sf.SpecFile(); path != "" {
				files[path] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(files))
	for path := range files {
		out = append(out, path)
	}

	slices.Sort(out)
	return out
}

// SpecFilesIter returns an iterator over all non-empty provider OpenAPI spec files.
func (r *Registry) SpecFilesIter() iter.Seq[string] {
	return slices.Values(r.SpecFiles())
}
