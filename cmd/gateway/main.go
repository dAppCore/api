// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	os "dappco.re/go/api/internal/stdcompat/coreos"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	core "dappco.re/go"
	coreapi "dappco.re/go/api"
	coreio "dappco.re/go/io"
	miner "dappco.re/go/miner"
	minerapi "dappco.re/go/miner/pkg/api"
	process "dappco.re/go/process"
	proxy "dappco.re/go/proxy"
	"dappco.re/go/scm/marketplace"
	scmapi "dappco.re/go/scm/pkg/api"
	"dappco.re/go/scm/repos"
	"dappco.re/go/ws"
	"github.com/gin-gonic/gin"
)

const (
	defaultGatewayBind = "0.0.0.0:8080"
	envGatewayBind     = "CORE_GATEWAY_BIND"
	envGatewayEnable   = "CORE_GATEWAY_ENABLE"
)

type providerFactory func(*gatewayDeps) coreapi.RouteGroup

type providerSpec struct {
	Name        string
	BasePath    string
	Description string
	Aliases     []string
	New         providerFactory
}

type gatewayDeps struct {
	core    *core.Core
	hub     *ws.Hub
	logger  *slog.Logger
	cleanup []func(context.Context)
}

type processRouteGroup struct {
	service *process.Service
}

func (g processRouteGroup) Name() string {
	return "process"
}

func (g processRouteGroup) BasePath() string {
	return "/api/process"
}

func (g processRouteGroup) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}
	rg.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, coreapi.OK(map[string]any{
			"provider": "process",
			"ready":    g.service != nil,
		}))
	})
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if wantsHelp(args) {
		printHelp(stdout)
		return 0
	}

	logger := slog.New(slog.NewTextHandler(stderr, nil))
	c := core.New()
	defer c.ServiceShutdown(context.Background())

	bind := core.Trim(os.Getenv(envGatewayBind))
	if bind == "" {
		bind = defaultGatewayBind
	}

	engine, err := coreapi.New(coreapi.WithAddr(bind))
	if err != nil {
		logger.Error("gateway engine init failed", "err", err)
		return 1
	}

	hub := ws.NewHub()
	deps := &gatewayDeps{
		core:   c,
		hub:    hub,
		logger: logger,
	}
	defer runCleanup(deps, logger)

	specs := gatewayProviderSpecs()
	enabled := selectedProviders(os.Getenv(envGatewayEnable))
	warnUnknownProviders(logger, specs, enabled)
	for _, spec := range specs {
		if !providerEnabled(spec, enabled) {
			continue
		}
		registerProvider(logger, engine, spec, func() coreapi.RouteGroup {
			return spec.New(deps)
		})
	}

	stopSignals := forwardSignalsToCore(c, logger)
	defer stopSignals()

	logger.Info("core gateway listening", "bind", bind, "providers", registeredProviderNames(engine.Groups()))
	if err := engine.Serve(c.Context()); err != nil {
		logger.Error("core gateway stopped with error", "err", err)
		return 1
	}
	return 0
}

func gatewayProviderSpecs() []providerSpec {
	return []providerSpec{
		{
			Name:        "brain",
			BasePath:    "/api/brain",
			Description: "core/agent brain provider",
			New: func(deps *gatewayDeps) coreapi.RouteGroup {
				return brainRouteGroup{
					name:     "brain",
					basePath: "/api/brain",
				}
			},
		},
		{
			Name:        "brain-mcp",
			BasePath:    "/api/mcp/brain",
			Description: "core/mcp brain provider variant",
			Aliases:     []string{"brain_mcp", "mcp-brain", "mcp"},
			New: func(deps *gatewayDeps) coreapi.RouteGroup {
				return brainRouteGroup{
					name:     "brain-mcp",
					basePath: "/api/mcp/brain",
				}
			},
		},
		{
			Name:        "scm",
			BasePath:    "/scm",
			Description: "go-scm repository and marketplace provider",
			New: func(deps *gatewayDeps) coreapi.RouteGroup {
				index := &marketplace.Index{Version: 1, Modules: []marketplace.Module{}}
				installer := marketplace.NewInstaller(coreio.Local, ".core/modules")
				registry := &repos.Registry{Version: 1, Repos: map[string]*repos.Repo{}}
				return scmapi.NewProvider(index, installer, registry, deps.hub)
			},
		},
		{
			Name:        "process",
			BasePath:    "/api/process",
			Description: "go-process daemon and process provider",
			New: func(deps *gatewayDeps) coreapi.RouteGroup {
				factory := process.NewService(process.Options{})
				result := factory(deps.core)
				if !result.OK {
					panic(result.Error())
				}
				service, ok := result.Value.(*process.Service)
				if !ok {
					panic(core.Sprintf("process service factory returned %T", result.Value))
				}
				deps.cleanup = append(deps.cleanup, func(ctx context.Context) {
					if r := service.OnShutdown(ctx); !r.OK {
						slog.Default().Warn("process service shutdown failed", "err", r.Error())
					}
				})
				return processRouteGroup{service: service}
			},
		},
		{
			Name:        "build",
			BasePath:    "/api/v1/build",
			Description: "go-build build, release, and SDK provider",
			New: func(deps *gatewayDeps) coreapi.RouteGroup {
				_ = deps
				return buildRouteGroup{projectDir: "."}
			},
		},
		{
			Name:        "miner",
			BasePath:    "",
			Description: "go-miner mining operations provider",
			New: func(deps *gatewayDeps) coreapi.RouteGroup {
				service := miner.NewServiceWithCore(deps.core)
				deps.cleanup = append(deps.cleanup, func(ctx context.Context) {
					if r := service.OnShutdown(ctx); !r.OK {
						slog.Default().Warn("miner service shutdown failed", "err", r.Error())
					}
				})
				return minerRouteGroup{provider: minerapi.NewProvider(service)}
			},
		},
		{
			Name:        "proxy",
			BasePath:    "/1",
			Description: "go-proxy monitoring provider",
			New: func(deps *gatewayDeps) coreapi.RouteGroup {
				_ = deps
				return newProxyRouteGroup()
			},
		},
	}
}

func registerProvider(logger *slog.Logger, engine *coreapi.Engine, spec providerSpec, factory func() coreapi.RouteGroup) (registered bool) {
	if logger == nil {
		logger = slog.Default()
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Error("provider registration failed", "provider", spec.Name, "panic", recovered)
			registered = false
		}
	}()

	if engine == nil {
		logger.Error("provider registration skipped", "provider", spec.Name, "err", "engine is nil")
		return false
	}
	if factory == nil {
		logger.Error("provider registration skipped", "provider", spec.Name, "err", "factory is nil")
		return false
	}

	group := factory()
	if isNilRouteGroup(group) {
		logger.Warn("provider registration skipped", "provider", spec.Name, "err", "provider is nil")
		return false
	}

	engine.Register(group)
	logger.Info("provider registered", "provider", spec.Name, "routeGroup", group.Name(), "basePath", displayBasePath(group.BasePath()))
	return true
}

func isNilRouteGroup(group coreapi.RouteGroup) bool {
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

func wantsHelp(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "-h", "--help", "help":
			return true
		}
	}
	return false
}

func printHelp(w io.Writer) {
	core.Print(w, "core-gateway mounts Core service providers on one API engine.")
	core.Print(w, "")
	core.Print(w, "Usage:")
	core.Print(w, "  core-gateway [--help]")
	core.Print(w, "")
	core.Print(w, "Environment:")
	core.Print(w, "  %s       listen address (default %s)", envGatewayBind, defaultGatewayBind)
	core.Print(w, "  CORE_GATEWAY_ENABLE     comma-separated provider names to mount (default all)")
	core.Print(w, "")
	core.Print(w, "Providers:")
	for _, spec := range gatewayProviderSpecs() {
		core.Print(w, "  %-10s %-16s %s", spec.Name, displayBasePath(spec.BasePath), spec.Description)
	}
}

func selectedProviders(raw string) map[string]bool {
	raw = core.Trim(raw)
	if raw == "" {
		return nil
	}
	selected := make(map[string]bool)
	for _, part := range core.Split(raw, ",") {
		name := canonicalProviderName(part)
		if name != "" {
			selected[name] = true
		}
	}
	return selected
}

func providerEnabled(spec providerSpec, selected map[string]bool) bool {
	if selected == nil {
		return true
	}
	if selected[canonicalProviderName(spec.Name)] {
		return true
	}
	for _, alias := range spec.Aliases {
		if selected[canonicalProviderName(alias)] {
			return true
		}
	}
	return false
}

func canonicalProviderName(name string) string {
	name = core.Lower(core.Trim(name))
	name = core.Replace(name, "_", "-")
	return name
}

func warnUnknownProviders(logger *slog.Logger, specs []providerSpec, selected map[string]bool) {
	if logger == nil || selected == nil {
		return
	}
	known := make(map[string]bool)
	for _, spec := range specs {
		known[canonicalProviderName(spec.Name)] = true
		for _, alias := range spec.Aliases {
			known[canonicalProviderName(alias)] = true
		}
	}
	for name := range selected {
		if !known[name] {
			logger.Warn("unknown provider requested", "provider", name, "env", envGatewayEnable)
		}
	}
}

func registeredProviderNames(groups []coreapi.RouteGroup) []string {
	names := make([]string, 0, len(groups))
	for _, group := range groups {
		if isNilRouteGroup(group) {
			continue
		}
		names = append(names, group.Name())
	}
	return names
}

func displayBasePath(path string) string {
	if core.Trim(path) == "" {
		return "(root)"
	}
	return path
}

func forwardSignalsToCore(c *core.Core, logger *slog.Logger) func() {
	return func() {
		if c != nil {
			c.ServiceShutdown(context.Background())
		}
		if logger != nil {
			logger.Debug("gateway signal bridge stopped")
		}
	}
}

func runCleanup(deps *gatewayDeps, logger *slog.Logger) {
	if deps == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for i := len(deps.cleanup) - 1; i >= 0; i-- {
		func(cleanup func(context.Context)) {
			defer func() {
				if recovered := recover(); recovered != nil && logger != nil {
					logger.Error("provider cleanup failed", "panic", recovered)
				}
			}()
			cleanup(ctx)
		}(deps.cleanup[i])
	}
}

type brainRouteGroup struct {
	name     string
	basePath string
}

func (g brainRouteGroup) Name() string {
	return g.name
}

func (g brainRouteGroup) BasePath() string {
	return g.basePath
}

func (g brainRouteGroup) Channels() []string {
	prefix := "brain"
	if g.name == "brain-mcp" {
		prefix = "brain.mcp"
	}
	return []string{
		prefix + ".remember.complete",
		prefix + ".recall.complete",
		prefix + ".forget.complete",
		prefix + ".list.complete",
	}
}

func (g brainRouteGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/remember", g.forwardBridgeMessage("brain_remember"))
	rg.POST("/recall", g.forwardBridgeMessage("brain_recall"))
	rg.POST("/forget", g.forwardBridgeMessage("brain_forget"))
	rg.GET("/list", g.forwardBridgeMessage("brain_list"))
	rg.GET("/status", g.status)
}

func (g brainRouteGroup) Describe() []coreapi.RouteDescription {
	return []coreapi.RouteDescription{
		{Method: http.MethodPost, Path: "/remember", Summary: "Store a memory", Tags: []string{g.name}},
		{Method: http.MethodPost, Path: "/recall", Summary: "Search memories", Tags: []string{g.name}},
		{Method: http.MethodPost, Path: "/forget", Summary: "Remove a memory", Tags: []string{g.name}},
		{Method: http.MethodGet, Path: "/list", Summary: "List memories", Tags: []string{g.name}},
		{Method: http.MethodGet, Path: "/status", Summary: "Brain bridge status", Tags: []string{g.name}},
	}
}

func (g brainRouteGroup) forwardBridgeMessage(messageType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = messageType
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "brain bridge unavailable"})
	}
}

func (g brainRouteGroup) status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"connected": false})
}

type buildRouteGroup struct {
	projectDir string
}

func (g buildRouteGroup) Name() string {
	return "build"
}

func (g buildRouteGroup) BasePath() string {
	return "/api/v1/build"
}

func (g buildRouteGroup) Channels() []string {
	return []string{
		"build.started",
		"build.complete",
		"build.failed",
		"release.started",
		"release.complete",
		"workflow.generated",
		"sdk.generated",
	}
}

func (g buildRouteGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/config", g.unavailable)
	rg.GET("/discover", g.unavailable)
	rg.POST("", g.unavailable)
	rg.POST("/build", g.unavailable)
	rg.GET("/artifacts", g.unavailable)
	rg.GET("/events", g.unavailable)
	rg.GET("/release/version", g.unavailable)
	rg.GET("/release/changelog", g.unavailable)
	rg.POST("/release", g.unavailable)
	rg.POST("/release/workflow", g.unavailable)
	rg.GET("/sdk/diff", g.unavailable)
	rg.POST("/sdk", g.unavailable)
	rg.POST("/sdk/generate", g.unavailable)
}

func (g buildRouteGroup) Describe() []coreapi.RouteDescription {
	return []coreapi.RouteDescription{
		{Method: http.MethodGet, Path: "/config", Summary: "Read build configuration", Tags: []string{"build"}},
		{Method: http.MethodGet, Path: "/discover", Summary: "Detect project type", Tags: []string{"build"}},
		{Method: http.MethodPost, Path: "/", Summary: "Trigger a build", Tags: []string{"build"}},
		{Method: http.MethodPost, Path: "/build", Summary: "Trigger a build", Tags: []string{"build"}},
		{Method: http.MethodGet, Path: "/artifacts", Summary: "List build artifacts", Tags: []string{"build"}},
		{Method: http.MethodGet, Path: "/events", Summary: "Subscribe to build events", Tags: []string{"build"}},
		{Method: http.MethodGet, Path: "/release/version", Summary: "Get current version", Tags: []string{"build"}},
		{Method: http.MethodGet, Path: "/release/changelog", Summary: "Generate changelog", Tags: []string{"build"}},
		{Method: http.MethodPost, Path: "/release", Summary: "Trigger release pipeline", Tags: []string{"build"}},
		{Method: http.MethodPost, Path: "/release/workflow", Summary: "Generate release workflow", Tags: []string{"build"}},
		{Method: http.MethodGet, Path: "/sdk/diff", Summary: "Read SDK diff", Tags: []string{"build"}},
		{Method: http.MethodPost, Path: "/sdk", Summary: "Generate SDK", Tags: []string{"build"}},
		{Method: http.MethodPost, Path: "/sdk/generate", Summary: "Generate SDK", Tags: []string{"build"}},
	}
}

func (g buildRouteGroup) unavailable(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":       "build provider unavailable in this worktree",
		"project_dir": g.projectDir,
	})
}

type minerRouteGroup struct {
	provider *minerapi.Provider
}

func (g minerRouteGroup) Name() string {
	return "miner"
}

func (g minerRouteGroup) BasePath() string {
	return ""
}

func (g minerRouteGroup) RegisterRoutes(rg *gin.RouterGroup) {
	if g.provider == nil || rg == nil {
		return
	}
	for _, route := range g.provider.RouteRegistrations() {
		route := route
		rg.Handle(core.Upper(route.Method), route.Path, func(c *gin.Context) {
			params := make(map[string]string, len(c.Params))
			for _, param := range c.Params {
				params[param.Key] = param.Value
			}

			var body []byte
			if c.Request != nil && c.Request.Body != nil {
				var err error
				body, err = io.ReadAll(c.Request.Body)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
			}

			value, err := route.Handler(c.Request.Context(), params, body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if value == nil {
				c.Status(http.StatusNoContent)
				return
			}
			c.JSON(http.StatusOK, value)
		})
	}
}

func (g minerRouteGroup) Describe() []coreapi.RouteDescription {
	if g.provider == nil {
		return nil
	}
	registrations := g.provider.RouteRegistrations()
	descriptions := make([]coreapi.RouteDescription, 0, len(registrations))
	for _, registration := range registrations {
		descriptions = append(descriptions, coreapi.RouteDescription{
			Method: registration.Method,
			Path:   registration.Path,
			Tags:   []string{"miner"},
		})
	}
	return descriptions
}

type proxyRouteHandler struct {
	path    string
	handler func(http.ResponseWriter, *http.Request)
	render  func() any
}

type proxyRouteGroup struct {
	proxy    *proxy.Proxy
	handlers []proxyRouteHandler
}

func (g *proxyRouteGroup) Name() string {
	return "proxy"
}

func (g *proxyRouteGroup) BasePath() string {
	return ""
}

func (g *proxyRouteGroup) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if core.Trim(pattern) == "" || handler == nil {
		return
	}
	g.handlers = append(g.handlers, proxyRouteHandler{path: pattern, handler: handler})
}

func (g *proxyRouteGroup) RegisterRoutes(rg *gin.RouterGroup) {
	if g == nil {
		return
	}
	for _, route := range g.handlers {
		route := route
		if route.handler != nil {
			rg.GET(route.path, gin.WrapF(route.handler))
			continue
		}
		rg.GET(route.path, func(c *gin.Context) {
			if g.proxy == nil || route.render == nil {
				c.Status(http.StatusServiceUnavailable)
				return
			}
			if status, ok := g.proxy.AllowMonitoringRequest(c.Request); !ok {
				switch status {
				case http.StatusMethodNotAllowed:
					c.Header("Allow", http.MethodGet)
				case http.StatusUnauthorized:
					c.Header("WWW-Authenticate", "Bearer")
				}
				c.Status(status)
				return
			}
			c.Header("Content-Type", "application/json")
			c.String(http.StatusOK, core.JSONMarshalString(route.render())+"\n")
		})
	}
}

func (g *proxyRouteGroup) Describe() []coreapi.RouteDescription {
	if g == nil {
		return nil
	}
	descriptions := make([]coreapi.RouteDescription, 0, len(g.handlers))
	for _, route := range g.handlers {
		descriptions = append(descriptions, coreapi.RouteDescription{
			Method: "GET",
			Path:   route.path,
			Tags:   []string{"proxy"},
		})
	}
	return descriptions
}

func newProxyRouteGroup() coreapi.RouteGroup {
	instance, result := proxy.New(&proxy.Config{
		Mode: "simple",
		Bind: []proxy.BindAddr{
			{Host: "127.0.0.1", Port: 0},
		},
		Pools: []proxy.PoolConfig{
			{URL: "127.0.0.1:1", Enabled: true},
		},
		Workers: proxy.WorkersByRigID,
	})
	if !result.OK {
		panic(result.Error)
	}
	group := &proxyRouteGroup{
		proxy: instance,
		handlers: []proxyRouteHandler{
			{path: "/1/summary", render: func() any { return instance.SummaryDocument() }},
			{path: "/1/workers", render: func() any { return instance.WorkersDocument() }},
			{path: "/1/miners", render: func() any { return instance.MinersDocument() }},
		},
	}
	if len(group.handlers) == 0 {
		return nil
	}
	return group
}
