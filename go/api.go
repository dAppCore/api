// SPDX-License-Identifier: EUPL-1.2

// Package api provides a Gin-based REST framework with OpenAPI generation.
// Subsystems implement RouteGroup to register their own endpoints.
package api

import (
	"context"
	"iter"
	"net"      // Note: AX-6 — net.SplitHostPort/ParseIP are structural for loopback bind classification; no core primitive
	"net/http" // Note: AX-6 — structural HTTP boundary for Handler/WebSocket contracts; no core primitive
	"reflect"  // Note: AX-6 — reflect is structural for runtime nil-pointer detection in handler binding; no core primitive
	"slices"
	"time"

	core "dappco.re/go"
	apistream "dappco.re/go/api/pkg/stream"

	"github.com/gin-contrib/expvar"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

const defaultAddr = ":8080"

var (
	// ErrNonLoopbackBind is returned by Serve under strict bind mode when the
	// configured listen address is not loopback and WithPublicBind was not
	// set. Strict mode is opt-in via WithStrictBind / WithLoopbackOnly.
	ErrNonLoopbackBind = core.NewError("api: strict bind rejects non-loopback address without WithPublicBind")
	// ErrPublicBindNoBearer is returned by Serve under strict bind mode when a
	// public (non-loopback) bind is requested without a bearer credential
	// supplied via WithBearerAuth.
	ErrPublicBindNoBearer = core.NewError("api: strict bind rejects public address without WithBearerAuth")
)

// shutdownTimeout is the maximum duration to wait for in-flight requests
// to complete during graceful shutdown.
const shutdownTimeout = 10 * time.Second

const (
	serverReadHeaderTimeout = 10 * time.Second
	serverReadTimeout       = 30 * time.Second
	serverWriteTimeout      = 60 * time.Second
	serverIdleTimeout       = 120 * time.Second
)

// Engine is the central API server managing route groups and middleware.
//
// Example:
//
//	engine, err := api.New(api.WithAddr(":8081"))
//	if err != nil {
//		panic(err)
//	}
//	_ = engine.Handler()
type Engine struct {
	addr                           string
	http3Enabled                   bool
	http3Addr                      string
	groups                         []RouteGroup
	streamGroups                   []apistream.StreamGroup
	middlewares                    []gin.HandlerFunc
	chatCompletionsResolver        *ModelResolver
	chatCompletionsPath            string
	sdkGenEnabled                  bool
	cacheTTL                       time.Duration
	cacheMaxEntries                int
	cacheMaxBytes                  int
	wsHandler                      http.Handler
	wsGinHandler                   gin.HandlerFunc
	wsPath                         string
	sseBroker                      *SSEBroker
	swaggerEnabled                 bool
	swaggerTitle                   string
	swaggerSummary                 string
	swaggerDesc                    string
	swaggerVersion                 string
	swaggerPath                    string
	swaggerTermsOfService          string
	swaggerServers                 []string
	swaggerContactName             string
	swaggerContactURL              string
	swaggerContactEmail            string
	swaggerLicenseName             string
	swaggerLicenseURL              string
	swaggerSecuritySchemes         map[string]any
	swaggerExternalDocsDescription string
	swaggerExternalDocsURL         string
	authentikConfig                AuthentikConfig
	pprofEnabled                   bool
	expvarEnabled                  bool
	ssePath                        string
	graphql                        *graphqlConfig
	i18nConfig                     I18nConfig
	openAPISpecEnabled             bool
	openAPISpecPath                string
	// strictBind, when set via WithStrictBind / WithLoopbackOnly, makes
	// Serve refuse a non-loopback listen address unless publicBind is also
	// set, and refuse to serve a public bind without a bearer credential.
	// Default false preserves the historic permissive behaviour so existing
	// consumers (go-ml, go-ai, desktop, core-agent) are not broken.
	strictBind bool
	// publicBind, when set via WithPublicBind, is the explicit opt-in that
	// allows a non-loopback bind under strict mode. It carries no effect
	// when strictBind is false. A public bind still requires a bearer.
	publicBind bool
	// bearerConfigured records that a bearer credential was supplied via
	// WithBearerAuth. Strict mode refuses to serve a public listener
	// without one.
	bearerConfigured bool
	// bearerToken is the static bearer credential supplied via WithBearerAuth.
	// The chat endpoint's off-loopback gate validates the inbound request
	// against this token directly so it fails closed independently of the
	// bearer middleware's path coverage.
	bearerToken string
	// noRouteHandler is the SPA / fallback handler invoked when no
	// registered route matches the request. Set via WithNoRoute; nil
	// means gin returns 404 with its default body.
	noRouteHandler gin.HandlerFunc
	// upstreamRouter, when set via WithUpstreamRouter, mounts a selector-keyed
	// reverse proxy over a pool of HTTP upstreams at the configured paths.
	upstreamRouter *upstreamRouterConfig
	// chatRemote, when set via WithChatCompletionsRemote, adds a remote backend
	// to the chat completions endpoint (local-first dispatch).
	chatRemote *chatRemoteConfig
	// chatAllowRemote permits non-loopback chat clients when a bearer is set.
	chatAllowRemote bool
}

// New creates an Engine with the given options.
// The default listen address is ":8080".
//
// Example:
//
//	engine, err := api.New(api.WithAddr(":8081"), api.WithResponseMeta())
//	if err != nil {
//		panic(err)
//	}
func New(opts ...Option) (
	*Engine,
	error,
) {
	e := &Engine{
		addr: defaultAddr,
	}
	for _, opt := range opts {
		opt(e)
	}
	// Apply calibrated defaults for optional subsystems.
	if (e.chatCompletionsResolver != nil || e.chatRemote != nil) && core.Trim(e.chatCompletionsPath) == "" {
		e.chatCompletionsPath = defaultChatCompletionsPath
	}
	return e, nil
}

// Addr returns the configured listen address.
//
// Example:
//
//	engine, _ := api.New(api.WithAddr(":9090"))
//	addr := engine.Addr()
func (e *Engine) Addr() string {
	return e.addr
}

// Groups returns a copy of all registered route groups.
//
// Example:
//
//	groups := engine.Groups()
func (e *Engine) Groups() []RouteGroup {
	return slices.Clone(e.groups)
}

// GroupsIter returns an iterator over all registered route groups.
//
// Example:
//
//	for group := range engine.GroupsIter() {
//		_ = group
//	}
func (e *Engine) GroupsIter() iter.Seq[RouteGroup] {
	groups := slices.Clone(e.groups)
	return slices.Values(groups)
}

// Register adds a route group to the engine.
//
// Example:
//
//	engine.Register(myGroup)
func (e *Engine) Register(group RouteGroup) {
	if isNilRouteGroup(group) {
		return
	}
	e.groups = append(e.groups, group)
}

// RegisterStreamGroup adds a declarative SSE/WebSocket handler group to the engine.
//
// Example:
//
//	engine.RegisterStreamGroup(stream.NewGroup("events"))
func (e *Engine) RegisterStreamGroup(group apistream.StreamGroup) {
	if isNilStreamGroup(group) {
		return
	}
	e.streamGroups = append(e.streamGroups, group)
}

// Channels returns all WebSocket channel names from registered StreamGroups.
// Groups that do not implement StreamGroup are silently skipped.
//
// Example:
//
//	channels := engine.Channels()
func (e *Engine) Channels() []string {
	var channels []string
	for _, g := range e.groups {
		if sg, ok := g.(StreamGroup); ok {
			channels = append(channels, sg.Channels()...)
		}
	}
	for _, g := range e.streamGroups {
		for _, h := range g.Handlers() {
			if h.Protocol == apistream.ProtocolWebSocket {
				channels = append(channels, h.Path)
			}
		}
	}
	return channels
}

// ChannelsIter returns an iterator over WebSocket channel names from registered StreamGroups.
//
// Example:
//
//	for channel := range engine.ChannelsIter() {
//		_ = channel
//	}
func (e *Engine) ChannelsIter() iter.Seq[string] {
	groups := slices.Clone(e.groups)
	streamGroups := slices.Clone(e.streamGroups)
	return func(yield func(string) bool) {
		for _, g := range groups {
			if sg, ok := g.(StreamGroup); ok {
				for _, c := range sg.Channels() {
					if !yield(c) {
						return
					}
				}
			}
		}
		for _, g := range streamGroups {
			for _, h := range g.Handlers() {
				if h.Protocol != apistream.ProtocolWebSocket {
					continue
				}
				if !yield(h.Path) {
					return
				}
			}
		}
	}
}

// Handler builds the Gin engine and returns it as an http.Handler.
// Each call produces a fresh handler reflecting the current set of groups.
//
// Example:
//
//	handler := engine.Handler()
func (e *Engine) Handler() http.Handler {
	return e.build()
}

// Serve starts the HTTP server and blocks until the context is cancelled,
// then performs a graceful shutdown allowing in-flight requests to complete.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	_ = engine.Serve(ctx)
func (e *Engine) Serve(ctx context.Context) (
	_ error,
) {
	if err := e.validateBind(); err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              e.addr,
		Handler:           e.build(),
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !core.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	// Return immediately if the listener fails before shutdown is requested.
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	// Signal SSE clients first so their handlers can exit cleanly before the
	// HTTP server begins its own shutdown sequence.
	if e.sseBroker != nil {
		e.sseBroker.Drain()
	}

	// Graceful shutdown with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	// Return any listen error that occurred before shutdown.
	return <-errCh
}

// validateBind enforces the strict bind invariants before Serve binds a
// listener. It is a no-op unless WithStrictBind / WithLoopbackOnly was set, so
// existing consumers that bind non-loopback addresses are unaffected.
//
// Under strict mode:
//   - a loopback address always serves;
//   - a non-loopback address is rejected unless WithPublicBind is set;
//   - a non-loopback address with WithPublicBind set still requires a bearer
//     credential (WithBearerAuth), and is otherwise rejected.
//
// Example:
//
//	e, _ := api.New(api.WithAddr("0.0.0.0:8787"), api.WithStrictBind())
//	err := e.Serve(ctx) // err == api.ErrNonLoopbackBind
func (e *Engine) validateBind() (
	_ error,
) {
	if !e.strictBind {
		return nil
	}
	if addrIsLoopback(e.addr) {
		return nil
	}
	if !e.publicBind {
		return core.E("api.bind", e.addr, ErrNonLoopbackBind)
	}
	if !e.bearerConfigured {
		return core.E("api.bind", e.addr, ErrPublicBindNoBearer)
	}
	return nil
}

// addrIsLoopback reports whether a listen address binds only the loopback
// interface. The host portion is parsed from "host:port"; a bare ":port" or an
// unspecified host ("0.0.0.0", "::", empty) is treated as non-loopback because
// it binds all interfaces. The textual host "localhost" is treated as loopback.
//
// Example:
//
//	addrIsLoopback("127.0.0.1:8787") // true
//	addrIsLoopback("[::1]:8787")     // true
//	addrIsLoopback("localhost:8787") // true
//	addrIsLoopback("0.0.0.0:8787")   // false
//	addrIsLoopback(":8787")          // false
func addrIsLoopback(addr string) bool {
	addr = core.Trim(addr)
	if addr == "" {
		return false
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// No port separator (or malformed); treat the whole string as the host.
		host = addr
	}

	host = core.Trim(host)
	if host == "" {
		// Bare ":port" — binds every interface, not loopback.
		return false
	}
	if host == "localhost" {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		// A named host other than localhost is not a loopback guarantee.
		return false
	}
	if ip.IsUnspecified() {
		// 0.0.0.0 / :: bind all interfaces.
		return false
	}
	return ip.IsLoopback()
}

// SetNoRoute attaches or replaces the fallback handler invoked when
// no registered route matches the incoming request. Mirrors the
// WithNoRoute Option but is callable after construction — useful
// when the SPA handler is built later in the boot sequence than
// the Engine (e.g. once the embedded frontend FS is resolved).
// Pass nil to clear and restore gin's default 404.
//
// Example:
//
//	e, _ := api.New()
//	// … later, once dist/ is mounted …
//	e.SetNoRoute(spaHandler)
func (e *Engine) SetNoRoute(h gin.HandlerFunc) {
	e.noRouteHandler = h
}

// build creates a configured Gin engine with recovery middleware,
// user-supplied middleware, the health endpoint, and all registered route groups.
func (e *Engine) build() *gin.Engine {
	r := gin.New()
	r.Use(recoveryMiddleware())
	if e.http3Enabled {
		if altSvc := http3AltSvcHeader(e.resolvedHTTP3Addr()); altSvc != "" {
			r.Use(http3AltSvcMiddleware(altSvc))
		}
	}

	// Apply user-supplied middleware after recovery but before routes.
	for _, mw := range e.middlewares {
		r.Use(mw)
	}
	if policies := cacheControlPolicies(e.groups); len(policies) > 0 {
		r.Use(cacheControlMiddleware(policies))
	}

	// Built-in health check.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, OK("healthy"))
	})

	// Mount the OpenAI-compatible chat completion endpoint when a local resolver
	// and/or a remote backend is configured.
	if e.chatCompletionsResolver != nil || e.chatRemote != nil {
		path := e.chatCompletionsPath
		if core.Trim(path) == "" {
			path = defaultChatCompletionsPath
		}
		h := newChatCompletionsHandler(e.chatCompletionsResolver, e.chatRemote, e.chatAllowRemote, bearerValidator(e.bearerToken))
		r.POST(path, h.ServeHTTP)
	}

	// Mount the selector-keyed upstream router when configured.
	if e.upstreamRouter != nil {
		proxy := e.upstreamRouter.buildProxy()
		h := e.upstreamRouter.handler(proxy)
		for _, p := range e.upstreamRouter.paths {
			r.Any(p, h)
		}
	}

	if e.sdkGenEnabled {
		mountSDKGen(r)
	}

	// Mount each registered group at its base path.
	for _, g := range e.groups {
		if isNilRouteGroup(g) {
			continue
		}
		rg := r.Group(g.BasePath())
		if mw := transformerMiddlewareForGroup(g); mw != nil {
			rg.Use(mw)
		}
		g.RegisterRoutes(rg)
	}

	// Mount each registered declarative stream group at the engine root.
	for _, g := range e.streamGroups {
		if isNilStreamGroup(g) {
			continue
		}
		g.Register(r)
	}

	// Mount WebSocket handler if configured. WithWebSocket (gin-native) takes
	// precedence over WithWSHandler (http.Handler) when both are supplied so
	// the more specific gin form wins.
	switch {
	case e.wsGinHandler != nil:
		r.GET(resolveWSPath(e.wsPath), e.wsGinHandler)
	case e.wsHandler != nil:
		r.GET(resolveWSPath(e.wsPath), wrapWSHandler(e.wsHandler))
	}

	// Mount SSE endpoint if configured.
	if e.sseBroker != nil {
		sseHandler := e.sseBroker.Handler()
		ssePath := resolveSSEPath(e.ssePath)
		r.GET(ssePath, sseHandler)
		if legacyPath := resolveLegacySSEPath(e.ssePath); legacyPath != "" && legacyPath != ssePath {
			r.GET(legacyPath, sseHandler)
		}
	}

	// Mount GraphQL endpoint if configured.
	if e.graphql != nil {
		mountGraphQL(r, e.graphql)
	}

	// Mount Swagger UI if enabled.
	if e.swaggerEnabled {
		registerSwagger(r, e, e.groups)
	}

	// Mount the standalone OpenAPI JSON endpoint (RFC.endpoints.md — "GET
	// /v1/openapi.json") when explicitly enabled. Unlike Swagger UI the spec
	// document is served directly so ToolBridge consumers and SDK generators
	// can fetch the latest description without loading the UI bundle.
	if e.openAPISpecEnabled {
		registerOpenAPISpec(r, e)
	}

	// Mount pprof profiling endpoints if enabled.
	if e.pprofEnabled {
		pprof.Register(r)
	}

	// Mount expvar runtime metrics endpoint if enabled.
	if e.expvarEnabled {
		r.GET("/debug/vars", expvar.Handler())
	}

	// SPA / fallback handler. Registered last so all explicit routes
	// take precedence — Gin invokes NoRoute only when no other handler
	// matched the request path + method.
	if e.noRouteHandler != nil {
		r.NoRoute(e.noRouteHandler)
	}

	return r
}

func isNilRouteGroup(group RouteGroup) bool {
	return isNilValue(group)
}

func isNilStreamGroup(group apistream.StreamGroup) bool {
	return isNilValue(group)
}

func isNilValue(v any) bool {
	if v == nil {
		return true
	}

	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
