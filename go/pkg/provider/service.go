// SPDX-License-Identifier: EUPL-1.2

// Service registration for the provider package. Wraps the *Registry
// as a self-contained Core subsystem so it can be supervised by
// go-process (in-binary or out-of-process) and queried over IPC for
// provider discovery.
//
// Providers themselves are interface-typed and held by direct method
// use on svc.Registry; the Service action surface is read-only
// introspection (list, names, by-name lookup, capability filters)
// since *gin.Engine and Provider implementations are not
// IPC-serialisable.
//
//	c, _ := core.New(
//	    core.WithName("api.provider", provider.NewService(provider.ProviderConfig{})),
//	)
//	svc := core.MustServiceFor[*provider.Service](c, "api.provider")
//	svc.Registry.Add(myProvider)
//	r := c.Action("api.provider.list").Run(ctx, core.Options{})

package provider

import (
	"context"

	core "dappco.re/go"
)

// ProviderConfig configures the provider service. Empty config gives
// an empty Registry — providers are added via direct method use on
// svc.Registry after registration.
//
// Usage example: `cfg := provider.ProviderConfig{}`
type ProviderConfig struct{}

// Service is the registerable handle for the provider package — embeds
// *core.ServiceRuntime[ProviderConfig] for typed options access and
// holds the live *Registry for direct method use (provider Add /
// MountAll / etc. are not IPC-serialisable so stay direct).
//
// Usage example: `svc := core.MustServiceFor[*provider.Service](c, "api.provider"); svc.Registry.Add(p)`
type Service struct {
	*core.ServiceRuntime[ProviderConfig]
	// Registry holds the live registered providers — interface-typed
	// adds and the gin-engine MountAll call stay on this handle since
	// neither crosses an IPC boundary.
	// Usage example: `svc.Registry.Add(myProvider)`
	Registry *Registry
	registrations core.Once
}

// NewService returns a factory that constructs an empty Registry and
// produces a *Service ready for c.Service() registration.
//
// Usage example: `c, _ := core.New(core.WithName("api.provider", provider.NewService(provider.ProviderConfig{})))`
func NewService(config ProviderConfig) func(*core.Core) core.Result {
	return func(c *core.Core) core.Result {
		return core.Ok(&Service{
			ServiceRuntime: core.NewServiceRuntime(c, config),
			Registry:       NewRegistry(),
		})
	}
}

// Register builds the provider service with default ProviderConfig and
// returns the service Result directly — the imperative-style
// alternative to NewService for consumers wiring services without
// WithName options.
//
// Usage example: `r := provider.Register(c); svc := r.Value.(*provider.Service)`
func Register(c *core.Core) core.Result {
	return NewService(ProviderConfig{})(c)
}

// OnStartup registers the provider action handlers on the attached
// Core. Implements core.Startable. Idempotent via core.Once.
//
// Usage example: `r := svc.OnStartup(ctx)`
func (s *Service) OnStartup(context.Context) core.Result {
	if s == nil {
		return core.Ok(nil)
	}
	s.registrations.Do(func() {
		c := s.Core()
		if c == nil {
			return
		}
		c.Action("api.provider.list", s.handleList)
		c.Action("api.provider.names", s.handleNames)
		c.Action("api.provider.count", s.handleCount)
		c.Action("api.provider.info", s.handleInfo)
		c.Action("api.provider.streamable_names", s.handleStreamableNames)
		c.Action("api.provider.describable_names", s.handleDescribableNames)
		c.Action("api.provider.renderable_names", s.handleRenderableNames)
		c.Action("api.provider.spec_files", s.handleSpecFiles)
	})
	return core.Ok(nil)
}

// OnShutdown is a no-op — the Registry holds no closable resources.
// Implements core.Stoppable.
//
// Usage example: `r := svc.OnShutdown(ctx)`
func (s *Service) OnShutdown(context.Context) core.Result {
	return core.Ok(nil)
}

// handleList — `api.provider.list` action handler. Returns the full
// []ProviderInfo summary for every registered provider in r.Value.
//
// Usage example: `r := c.Action("api.provider.list").Run(ctx, core.Options{})`
func (s *Service) handleList(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Registry == nil {
		return core.Fail(core.E("api.provider.list", "service not initialised", nil))
	}
	return core.Ok(s.Registry.Info())
}

// handleNames — `api.provider.names` action handler. Returns []string
// of all registered provider names in r.Value.
//
// Usage example: `r := c.Action("api.provider.names").Run(ctx, core.Options{})`
func (s *Service) handleNames(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Registry == nil {
		return core.Fail(core.E("api.provider.names", "service not initialised", nil))
	}
	providers := s.Registry.List()
	names := make([]string, 0, len(providers))
	for _, p := range providers {
		names = append(names, p.Name())
	}
	return core.Ok(names)
}

// handleCount — `api.provider.count` action handler. Returns int count
// of registered providers in r.Value.
//
// Usage example: `r := c.Action("api.provider.count").Run(ctx, core.Options{})`
func (s *Service) handleCount(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Registry == nil {
		return core.Fail(core.E("api.provider.count", "service not initialised", nil))
	}
	return core.Ok(s.Registry.Len())
}

// handleInfo — `api.provider.info` action handler. Reads opts.name
// and returns the matching ProviderInfo (with Channels/Element/SpecFile
// populated by capability) in r.Value, or nil if not found.
//
// Usage example: `r := c.Action("api.provider.info").Run(ctx, core.NewOptions(core.Option{Key: "name", Value: "brain"}))`
func (s *Service) handleInfo(_ core.Context, opts core.Options) core.Result {
	if s == nil || s.Registry == nil {
		return core.Fail(core.E("api.provider.info", "service not initialised", nil))
	}
	name := opts.String("name")
	for _, info := range s.Registry.Info() {
		if info.Name == name {
			return core.Ok(info)
		}
	}
	return core.Ok(nil)
}

// handleStreamableNames — `api.provider.streamable_names` action
// handler. Returns []string of provider names that implement the
// Streamable interface in r.Value.
//
// Usage example: `r := c.Action("api.provider.streamable_names").Run(ctx, core.Options{})`
func (s *Service) handleStreamableNames(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Registry == nil {
		return core.Fail(core.E("api.provider.streamable_names", "service not initialised", nil))
	}
	streamables := s.Registry.Streamable()
	names := make([]string, 0, len(streamables))
	for _, p := range streamables {
		names = append(names, p.Name())
	}
	return core.Ok(names)
}

// handleDescribableNames — `api.provider.describable_names` action
// handler. Returns []string of provider names that implement the
// Describable interface in r.Value.
//
// Usage example: `r := c.Action("api.provider.describable_names").Run(ctx, core.Options{})`
func (s *Service) handleDescribableNames(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Registry == nil {
		return core.Fail(core.E("api.provider.describable_names", "service not initialised", nil))
	}
	describables := s.Registry.Describable()
	names := make([]string, 0, len(describables))
	for _, p := range describables {
		names = append(names, p.Name())
	}
	return core.Ok(names)
}

// handleRenderableNames — `api.provider.renderable_names` action
// handler. Returns []string of provider names that implement the
// Renderable interface in r.Value.
//
// Usage example: `r := c.Action("api.provider.renderable_names").Run(ctx, core.Options{})`
func (s *Service) handleRenderableNames(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Registry == nil {
		return core.Fail(core.E("api.provider.renderable_names", "service not initialised", nil))
	}
	renderables := s.Registry.Renderable()
	names := make([]string, 0, len(renderables))
	for _, p := range renderables {
		names = append(names, p.Name())
	}
	return core.Ok(names)
}

// handleSpecFiles — `api.provider.spec_files` action handler. Returns
// []string of OpenAPI spec file paths declared by registered providers
// (deduplicated, sorted) in r.Value.
//
// Usage example: `r := c.Action("api.provider.spec_files").Run(ctx, core.Options{})`
func (s *Service) handleSpecFiles(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Registry == nil {
		return core.Fail(core.E("api.provider.spec_files", "service not initialised", nil))
	}
	return core.Ok(s.Registry.SpecFiles())
}
