// SPDX-License-Identifier: EUPL-1.2

// Package api provides a Gin-based REST framework with OpenAPI generation.
// Subsystems implement RouteGroup to register their own endpoints.
package api

import (
	"context"
	"iter"
	"net/http"
	"reflect"
	"slices"
	"time"

	core "dappco.re/go/core"
	apistream "dappco.re/go/api/pkg/stream"

	"github.com/gin-contrib/expvar"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

const defaultAddr = ":8080"

// shutdownTimeout is the maximum duration to wait for in-flight requests
// to complete during graceful shutdown.
const shutdownTimeout = 10 * time.Second

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
	groups                         []RouteGroup
	streamGroups                   []apistream.StreamGroup
	middlewares                    []gin.HandlerFunc
	chatCompletionsResolver        *ModelResolver
	chatCompletionsPath            string
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
func New(opts ...Option) (*Engine, error) {
	e := &Engine{
		addr: defaultAddr,
	}
	for _, opt := range opts {
		opt(e)
	}
	// Apply calibrated defaults for optional subsystems.
	if e.chatCompletionsResolver != nil && core.Trim(e.chatCompletionsPath) == "" {
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
func (e *Engine) Serve(ctx context.Context) error {
	srv := &http.Server{
		Addr:    e.addr,
		Handler: e.build(),
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

// build creates a configured Gin engine with recovery middleware,
// user-supplied middleware, the health endpoint, and all registered route groups.
func (e *Engine) build() *gin.Engine {
	r := gin.New()
	r.Use(recoveryMiddleware())

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

	// Mount the local OpenAI-compatible chat completion endpoint when configured.
	if e.chatCompletionsResolver != nil {
		h := newChatCompletionsHandler(e.chatCompletionsResolver)
		r.POST(e.chatCompletionsPath, h.ServeHTTP)
	}

	// Mount each registered group at its base path.
	for _, g := range e.groups {
		if isNilRouteGroup(g) {
			continue
		}
		rg := r.Group(g.BasePath())
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
