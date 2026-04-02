// SPDX-License-Identifier: EUPL-1.2

// Package api provides a Gin-based REST framework with OpenAPI generation.
// Subsystems implement RouteGroup to register their own endpoints.
package api

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"reflect"
	"slices"
	"time"

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
	middlewares                    []gin.HandlerFunc
	wsHandler                      http.Handler
	sseBroker                      *SSEBroker
	swaggerEnabled                 bool
	swaggerTitle                   string
	swaggerDesc                    string
	swaggerVersion                 string
	swaggerTermsOfService          string
	swaggerServers                 []string
	swaggerContactName             string
	swaggerContactURL              string
	swaggerContactEmail            string
	swaggerLicenseName             string
	swaggerLicenseURL              string
	swaggerExternalDocsDescription string
	swaggerExternalDocsURL         string
	pprofEnabled                   bool
	expvarEnabled                  bool
	ssePath                        string
	graphql                        *graphqlConfig
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
func (e *Engine) Groups() []RouteGroup {
	return slices.Clone(e.groups)
}

// GroupsIter returns an iterator over all registered route groups.
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

// Channels returns all WebSocket channel names from registered StreamGroups.
// Groups that do not implement StreamGroup are silently skipped.
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
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	// Built-in health check.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, OK("healthy"))
	})

	// Mount each registered group at its base path.
	for _, g := range e.groups {
		if isNilRouteGroup(g) {
			continue
		}
		rg := r.Group(g.BasePath())
		g.RegisterRoutes(rg)
	}

	// Mount WebSocket handler if configured.
	if e.wsHandler != nil {
		r.GET("/ws", wrapWSHandler(e.wsHandler))
	}

	// Mount SSE endpoint if configured.
	if e.sseBroker != nil {
		r.GET(resolveSSEPath(e.ssePath), e.sseBroker.Handler())
	}

	// Mount GraphQL endpoint if configured.
	if e.graphql != nil {
		mountGraphQL(r, e.graphql)
	}

	// Mount Swagger UI if enabled.
	if e.swaggerEnabled {
		ssePath := ""
		if e.sseBroker != nil {
			ssePath = resolveSSEPath(e.ssePath)
		}
		registerSwagger(
			r,
			e.swaggerTitle,
			e.swaggerDesc,
			e.swaggerVersion,
			func() string {
				if e.graphql == nil {
					return ""
				}
				return e.graphql.path
			}(),
			ssePath,
			e.swaggerTermsOfService,
			e.swaggerContactName,
			e.swaggerContactURL,
			e.swaggerContactEmail,
			e.swaggerServers,
			e.swaggerLicenseName,
			e.swaggerLicenseURL,
			e.swaggerExternalDocsDescription,
			e.swaggerExternalDocsURL,
			e.groups,
		)
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
	if group == nil {
		return true
	}

	value := reflect.ValueOf(group)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
