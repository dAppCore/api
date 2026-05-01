// SPDX-License-Identifier: EUPL-1.2

// Service registration for the api package. Wraps the package *Engine
// as a self-contained Core subsystem so it can be supervised by
// go-process (in-binary or out-of-process) and queried over IPC.
//
// RouteGroups, Providers, and StreamGroups are registered through
// direct method use on svc.Engine (interface-typed values + gin
// handlers don't cross IPC boundaries); the action surface is
// read-only introspection of mounted addresses, groups, and channels.
//
//	c, _ := core.New(
//	    core.WithName("api", api.NewService(api.ApiConfig{Addr: ":8081"})),
//	)
//	svc := core.MustServiceFor[*api.Service](c, "api")
//	svc.Engine.Register(myProvider)
//	r := c.Action("api.groups").Run(ctx, core.Options{})

package api

import (
	"context"

	core "dappco.re/go"
)

// ApiConfig configures the api service. Empty config gives an Engine
// bound to the package default listen address (":8080") with no
// optional subsystems enabled.
//
// Usage example: `cfg := api.ApiConfig{Addr: ":8081"}`
type ApiConfig struct {
	// Addr is the TCP listen address. Empty = use the package default
	// (":8080"). Forwarded to api.WithAddr at construction time.
	// Usage example: `cfg := api.ApiConfig{Addr: ":8081"}`
	Addr string
}

// Service is the registerable handle for the api package — embeds
// *core.ServiceRuntime[ApiConfig] for typed options access and holds
// the live *Engine for direct method use (Register / RegisterStreamGroup
// / Serve all stay on this handle since interface-typed group values
// and gin handlers aren't IPC-serialisable).
//
// Usage example: `svc := core.MustServiceFor[*api.Service](c, "api"); svc.Engine.Register(p)`
type Service struct {
	*core.ServiceRuntime[ApiConfig]
	// Engine is the live *Engine — RouteGroup / StreamGroup
	// registration and Serve calls stay on this handle since neither
	// crosses an IPC boundary.
	// Usage example: `svc.Engine.Register(myProvider)`
	Engine *Engine
	registrations core.Once
}

// NewService returns a factory that constructs an *Engine seeded from
// config and produces a *Service ready for c.Service() registration.
//
// Usage example: `c, _ := core.New(core.WithName("api", api.NewService(api.ApiConfig{Addr: ":8081"})))`
func NewService(config ApiConfig) func(*core.Core) core.Result {
	return func(c *core.Core) core.Result {
		opts := []Option{}
		if config.Addr != "" {
			opts = append(opts, WithAddr(config.Addr))
		}
		engine, err := New(opts...)
		if err != nil {
			return core.Fail(core.E("api.NewService", "construct engine", err))
		}
		return core.Ok(&Service{
			ServiceRuntime: core.NewServiceRuntime(c, config),
			Engine:         engine,
		})
	}
}

// Register builds the api service with default ApiConfig (Engine bound
// to ":8080", no optional subsystems) and returns the service Result
// directly — the imperative-style alternative to NewService for
// consumers wiring services without WithName options.
//
// Usage example: `r := api.Register(c); svc := r.Value.(*api.Service)`
func Register(c *core.Core) core.Result {
	return NewService(ApiConfig{})(c)
}

// OnStartup registers the api action handlers on the attached Core.
// Implements core.Startable. Idempotent via core.Once.
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
		c.Action("api.addr", s.handleAddr)
		c.Action("api.groups", s.handleGroups)
		c.Action("api.group_count", s.handleGroupCount)
		c.Action("api.channels", s.handleChannels)
	})
	return core.Ok(nil)
}

// OnShutdown is a no-op — Engine.Serve owns its own listener lifecycle
// via the context passed to it; the Service does not start the engine.
// Implements core.Stoppable.
//
// Usage example: `r := svc.OnShutdown(ctx)`
func (s *Service) OnShutdown(context.Context) core.Result {
	return core.Ok(nil)
}

// handleAddr — `api.addr` action handler. Returns the Engine's
// configured listen address in r.Value.
//
// Usage example: `r := c.Action("api.addr").Run(ctx, core.Options{})`
func (s *Service) handleAddr(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Engine == nil {
		return core.Fail(core.E("api.addr", "service not initialised", nil))
	}
	return core.Ok(s.Engine.Addr())
}

// handleGroups — `api.groups` action handler. Returns []string of all
// registered RouteGroup names (Provider.Name() / RouteGroup.Name()) in
// r.Value.
//
// Usage example: `r := c.Action("api.groups").Run(ctx, core.Options{})`
func (s *Service) handleGroups(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Engine == nil {
		return core.Fail(core.E("api.groups", "service not initialised", nil))
	}
	groups := s.Engine.Groups()
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.Name())
	}
	return core.Ok(names)
}

// handleGroupCount — `api.group_count` action handler. Returns int
// count of registered RouteGroups in r.Value.
//
// Usage example: `r := c.Action("api.group_count").Run(ctx, core.Options{})`
func (s *Service) handleGroupCount(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Engine == nil {
		return core.Fail(core.E("api.group_count", "service not initialised", nil))
	}
	return core.Ok(len(s.Engine.Groups()))
}

// handleChannels — `api.channels` action handler. Returns []string of
// all WebSocket channel names declared by registered StreamGroups in
// r.Value.
//
// Usage example: `r := c.Action("api.channels").Run(ctx, core.Options{})`
func (s *Service) handleChannels(_ core.Context, _ core.Options) core.Result {
	if s == nil || s.Engine == nil {
		return core.Fail(core.E("api.channels", "service not initialised", nil))
	}
	return core.Ok(s.Engine.Channels())
}
