// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	core "dappco.re/go"
	coreapi "dappco.re/go/api"
	coregrpc "dappco.re/go/api/pkg/grpc"
	coreio "dappco.re/go/io"
	process "dappco.re/go/process"
	proxy "dappco.re/go/proxy"
	"dappco.re/go/scm/marketplace"
	scmapi "dappco.re/go/scm/pkg/api"
	"dappco.re/go/scm/repos"
	store "dappco.re/go/store"
	"dappco.re/go/ws"
	"github.com/gin-gonic/gin"
)

const (
	defaultGatewayBind = "0.0.0.0:8080"
	envGatewayBind     = "CORE_GATEWAY_BIND"
	envGatewayEnable   = "CORE_GATEWAY_ENABLE"

	// envGatewayGRPCSocket overrides the Unix domain socket the gRPC
	// sidecar bridge (GoService) listens on. When empty the gateway uses
	// defaultGatewayGRPCSocket under the workspace .core directory.
	envGatewayGRPCSocket = "CORE_GATEWAY_GRPC_SOCKET"
	// defaultGatewayGRPCSocket is the sidecar socket path used when
	// envGatewayGRPCSocket is unset. Deno dials this to reach Go.
	defaultGatewayGRPCSocket = ".core/run/core-sidecar.sock"
	// defaultSidecarStorePath is the SQLite KV database backing the
	// GoService StoreGet/StoreSet rpcs when no override is supplied.
	defaultSidecarStorePath = ".core/run/sidecar-store.db"
	// envGatewaySidecarStore overrides defaultSidecarStorePath.
	envGatewaySidecarStore = "CORE_GATEWAY_SIDECAR_STORE"
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

	// procService is the single go-process Service shared by the HTTP
	// process provider and the gRPC sidecar GoService. It is constructed
	// once in run via ensureProcessService so both consumers exec through
	// the same daemon rather than spinning up duplicate services.
	procService *process.Service
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

// ensureProcessService returns the gateway's shared go-process Service,
// constructing it on first use and registering its shutdown cleanup
// exactly once. Both the HTTP process provider and the gRPC sidecar
// GoService call this so a single daemon backs every exec.
//
//	svc := ensureProcessService(deps)
func ensureProcessService(deps *gatewayDeps) *process.Service {
	if deps.procService != nil {
		return deps.procService
	}
	factory := process.NewService(process.Options{})
	result := factory(deps.core)
	if !result.OK {
		panic(result.Error())
	}
	service, ok := result.Value.(*process.Service)
	if !ok {
		panic(core.Sprintf("process service factory returned %T", result.Value))
	}
	deps.procService = service
	deps.cleanup = append(deps.cleanup, func(ctx context.Context) {
		if r := service.OnShutdown(ctx); !r.OK {
			slog.Default().Warn("process service shutdown failed", "err", r.Error())
		}
	})
	return service
}

// kvStoreAdapter adapts a go-store *Store to the grpc.KVStore surface
// the GoService consumes. go-store returns plain errors; the bridge
// contract is core.Result, and an absent key must read back as an empty
// value on an OK Result (RFC.grpc.md StoreGet semantics), so a
// store.NotFoundError is folded into success here rather than surfaced.
//
//	var kv coregrpc.KVStore = kvStoreAdapter{store: s}
type kvStoreAdapter struct {
	store *store.Store
}

// Get returns the value for (group, key). A missing key yields an empty
// value on an OK Result; any other backend error fails the Result.
func (a kvStoreAdapter) Get(group, key string) (string, core.Result) {
	value, err := a.store.Get(group, key)
	if err != nil {
		if core.Is(err, store.NotFoundError) {
			return "", core.Ok(nil)
		}
		return "", core.Fail(err)
	}
	return value, core.Ok(nil)
}

// Set writes value under (group, key), translating a go-store error
// into a failed Result.
func (a kvStoreAdapter) Set(group, key, value string) core.Result {
	if err := a.store.Set(group, key, value); err != nil {
		return core.Fail(err)
	}
	return core.Ok(nil)
}

// procRunnerAdapter adapts a go-process *Service to the grpc.ProcRunner
// surface. The bridge's RunOptions is a narrow wire-facing struct, so
// this maps it onto the concrete go-process RunOptions at the call site.
//
//	var r coregrpc.ProcRunner = procRunnerAdapter{service: svc}
type procRunnerAdapter struct {
	service *process.Service
}

// RunWithOptions executes a command through go-process and returns the
// captured output on the Result.
func (a procRunnerAdapter) RunWithOptions(ctx context.Context, opts coregrpc.RunOptions) core.Result {
	return a.service.RunWithOptions(ctx, process.RunOptions{
		Command: opts.Command,
		Args:    opts.Args,
		Dir:     opts.Dir,
		Env:     opts.Env,
	})
}

// startSidecarBridge wires the gRPC sidecar GoService to real Core
// subsystems (go-io Local medium, a go-store KV database, the shared
// go-process Service) and serves it on a Unix domain socket. Deno dials
// this socket for sandboxed I/O, KV state, and process execution.
//
// The bridge is additive: any failure to open the store or bind the
// socket is logged and the gateway continues serving HTTP. The server
// is stopped gracefully both when Core's context is cancelled (signal /
// shutdown) and via the cleanup stack run on exit.
func startSidecarBridge(deps *gatewayDeps) {
	logger := deps.logger
	if logger == nil {
		logger = slog.Default()
	}

	socket := core.Trim(core.Getenv(envGatewayGRPCSocket))
	if socket == "" {
		socket = defaultGatewayGRPCSocket
	}
	if r := core.MkdirAll(core.PathDir(socket), 0o755); !r.OK {
		logger.Error("sidecar bridge socket dir create failed", "path", core.PathDir(socket), "err", r.Error())
		return
	}

	goService := coregrpc.NewGoService(
		coreio.Local,
		openSidecarStore(logger),
		procRunnerAdapter{service: ensureProcessService(deps)},
	)

	srv, err := coregrpc.NewGRPCServer(
		coregrpc.WithGRPCSocket(socket),
		coregrpc.WithGRPCServices(goService),
	)
	if err != nil {
		logger.Error("sidecar bridge listen failed", "socket", socket, "err", err)
		return
	}

	// Stop gracefully on cleanup (exit path) and when Core's context is
	// cancelled (signal / ServiceShutdown). srv.Stop is idempotent.
	deps.cleanup = append(deps.cleanup, func(context.Context) { srv.Stop() })
	if deps.core != nil {
		ctx := deps.core.Context()
		deps.core.Go(func() {
			<-ctx.Done()
			srv.Stop()
		})
		deps.core.Go(func() {
			if serveErr := srv.Serve(); serveErr != nil {
				logger.Error("sidecar bridge serve stopped with error", "err", serveErr)
			}
		})
	} else {
		go func() {
			if serveErr := srv.Serve(); serveErr != nil {
				logger.Error("sidecar bridge serve stopped with error", "err", serveErr)
			}
		}()
	}

	logger.Info("sidecar bridge listening", "socket", srv.Address())
}

// openSidecarStore opens the go-store SQLite KV database backing the
// GoService and returns it wrapped in the grpc.KVStore adapter. On
// failure it logs and returns nil, leaving StoreGet/StoreSet to report
// the subsystem as unavailable rather than aborting the gateway.
func openSidecarStore(logger *slog.Logger) coregrpc.KVStore {
	path := core.Trim(core.Getenv(envGatewaySidecarStore))
	if path == "" {
		path = defaultSidecarStorePath
	}
	if r := core.MkdirAll(core.PathDir(path), 0o755); !r.OK {
		logger.Error("sidecar store dir create failed", "path", core.PathDir(path), "err", r.Error())
		return nil
	}
	s, err := store.New(path)
	if err != nil {
		logger.Error("sidecar store open failed", "path", path, "err", err)
		return nil
	}
	return kvStoreAdapter{store: s}
}

func main() {
	core.Exit(run(core.Args()[1:], core.Stdout(), core.Stderr()))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if wantsHelp(args) {
		printHelp(stdout)
		return 0
	}

	logger := slog.New(slog.NewTextHandler(stderr, nil))
	c := core.New()
	defer func() {
		if r := c.ServiceShutdown(context.Background()); !r.OK {
			logger.Error("gateway core shutdown failed", "err", r.Error())
		}
	}()

	bind := core.Trim(core.Getenv(envGatewayBind))
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
	enabled := selectedProviders(core.Getenv(envGatewayEnable))
	warnUnknownProviders(logger, specs, enabled)
	for _, spec := range specs {
		if !providerEnabled(spec, enabled) {
			continue
		}
		registerProvider(logger, engine, spec, func() coreapi.RouteGroup {
			return spec.New(deps)
		})
	}

	startSidecarBridge(deps)

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
				return processRouteGroup{service: ensureProcessService(deps)}
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
			if r := c.ServiceShutdown(context.Background()); !r.OK && logger != nil {
				logger.Error("gateway signal shutdown failed", "err", r.Error())
			}
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
