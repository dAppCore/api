// SPDX-License-Identifier: EUPL-1.2

// Service registration for the stream package. Holds a registry of
// declared StreamGroup instances so the set of mounted SSE/WebSocket
// surfaces can be inspected over IPC under go-process supervision.
//
// Group instances themselves carry gin.HandlerFunc closures that are
// not IPC-serialisable — the action surface is read-only introspection
// of names + handler metadata; mounting onto a real *gin.Engine stays
// on the direct svc.Groups handle.
//
//	c, _ := core.New(
//	    core.WithName("api.stream", stream.NewService(stream.StreamConfig{})),
//	)
//	svc := core.MustServiceFor[*stream.Service](c, "api.stream")
//	svc.Add(stream.NewGroup("system", stream.SSE("/events", handler)))
//	r := c.Action("api.stream.handlers").Run(ctx, core.NewOptions(
//	    core.Option{Key: "group", Value: "system"},
//	))

package stream

import (
	"context"

	core "dappco.re/go"
)

// StreamConfig configures the stream service. Empty config gives an
// empty StreamGroup registry — groups are added via svc.Add after
// registration.
//
// Usage example: `cfg := stream.StreamConfig{}`
type StreamConfig struct{}

// HandlerInfo is a serialisable summary of a Handler — Method/Path/
// Protocol cross IPC fine; the gin.HandlerFunc closure does not.
//
// Usage example: `infos := svc.HandlerInfos("system")`
type HandlerInfo struct {
	Protocol Protocol `json:"protocol"`
	Method   string   `json:"method"`
	Path     string   `json:"path"`
}

// Service is the registerable handle for the stream package — embeds
// *core.ServiceRuntime[StreamConfig] for typed options access and
// holds a live []StreamGroup registry for direct method use (Register
// onto a *gin.Engine stays direct since gin handlers don't cross IPC).
//
// Usage example: `svc := core.MustServiceFor[*stream.Service](c, "api.stream"); svc.Add(group)`
type Service struct {
	*core.ServiceRuntime[StreamConfig]
	mu     core.RWMutex
	groups []StreamGroup
	registrations core.Once
}

// NewService returns a factory that constructs an empty stream-group
// registry and produces a *Service ready for c.Service() registration.
//
// Usage example: `c, _ := core.New(core.WithName("api.stream", stream.NewService(stream.StreamConfig{})))`
func NewService(config StreamConfig) func(*core.Core) core.Result {
	return func(c *core.Core) core.Result {
		return core.Ok(&Service{
			ServiceRuntime: core.NewServiceRuntime(c, config),
		})
	}
}

// Register builds the stream service with default StreamConfig and
// returns the service Result directly — the imperative-style
// alternative to NewService for consumers wiring services without
// WithName options.
//
// Usage example: `r := stream.Register(c); svc := r.Value.(*stream.Service)`
func Register(c *core.Core) core.Result {
	return NewService(StreamConfig{})(c)
}

// Add registers a StreamGroup. Groups are mounted onto a Registrar in
// the order they are added.
//
// Usage example: `svc.Add(stream.NewGroup("system", stream.SSE("/events", handler)))`
func (s *Service) Add(g StreamGroup) {
	if s == nil || g == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groups = append(s.groups, g)
}

// Groups returns a defensive copy of all registered StreamGroups for
// direct method use (Register onto a *gin.Engine).
//
// Usage example: `for _, g := range svc.Groups() { g.Register(engine) }`
func (s *Service) Groups() []StreamGroup {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.groups) == 0 {
		return nil
	}
	return append([]StreamGroup(nil), s.groups...)
}

// MountAll registers every StreamGroup onto the supplied Registrar in
// insertion order.
//
// Usage example: `svc.MountAll(engine)`
func (s *Service) MountAll(reg Registrar) {
	for _, g := range s.Groups() {
		g.Register(reg)
	}
}

// HandlerInfos returns the IPC-serialisable handler metadata for the
// named group, or nil if the group is unknown.
//
// Usage example: `infos := svc.HandlerInfos("system")`
func (s *Service) HandlerInfos(name string) []HandlerInfo {
	for _, g := range s.Groups() {
		if g.Name() != name {
			continue
		}
		handlers := g.Handlers()
		out := make([]HandlerInfo, 0, len(handlers))
		for _, h := range handlers {
			out = append(out, HandlerInfo{Protocol: h.Protocol, Method: h.Method, Path: h.Path})
		}
		return out
	}
	return nil
}

// OnStartup registers the stream action handlers on the attached Core.
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
		c.Action("api.stream.groups", s.handleGroups)
		c.Action("api.stream.handlers", s.handleHandlers)
	})
	return core.Ok(nil)
}

// OnShutdown is a no-op — StreamGroups hold no closable resources.
// Implements core.Stoppable.
//
// Usage example: `r := svc.OnShutdown(ctx)`
func (s *Service) OnShutdown(context.Context) core.Result {
	return core.Ok(nil)
}

// handleGroups — `api.stream.groups` action handler. Returns []string
// of all registered StreamGroup names in r.Value.
//
// Usage example: `r := c.Action("api.stream.groups").Run(ctx, core.Options{})`
func (s *Service) handleGroups(_ core.Context, _ core.Options) core.Result {
	if s == nil {
		return core.Fail(core.E("api.stream.groups", "service not initialised", nil))
	}
	groups := s.Groups()
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.Name())
	}
	return core.Ok(names)
}

// handleHandlers — `api.stream.handlers` action handler. Reads
// opts.group and returns []HandlerInfo (Method/Path/Protocol) for the
// named group's handlers in r.Value, or nil if the group is unknown.
//
// Usage example: `r := c.Action("api.stream.handlers").Run(ctx, core.NewOptions(core.Option{Key: "group", Value: "system"}))`
func (s *Service) handleHandlers(_ core.Context, opts core.Options) core.Result {
	if s == nil {
		return core.Fail(core.E("api.stream.handlers", "service not initialised", nil))
	}
	return core.Ok(s.HandlerInfos(opts.String("group")))
}
