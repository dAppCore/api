// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"log/slog"
	"net/http"
	"time"

	coretest "dappco.re/go"
	apistream "dappco.re/go/api/pkg/stream"

	"github.com/gin-gonic/gin"
)

type ax7RouteGroup struct {
	name     string
	basePath string
}

func (g ax7RouteGroup) Name() string     { return g.name }
func (g ax7RouteGroup) BasePath() string { return g.basePath }
func (g ax7RouteGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
}

type ax7StreamHandler struct {
	protocol string
	path     string
}

type ax7StreamGroup struct {
	name     string
	handlers []ax7StreamHandler
}

func (g ax7StreamGroup) Name() string { return g.name }
func (g ax7StreamGroup) Handlers() []apistream.Handler {
	out := make([]apistream.Handler, 0, len(g.handlers))
	for _, h := range g.handlers {
		out = append(out, apistream.Handler{
			Protocol: apistream.Protocol(h.protocol),
			Method:   http.MethodGet,
			Path:     h.path,
			Handle:   func(*gin.Context) {},
		})
	}
	return out
}
func (g ax7StreamGroup) Register(apistream.Registrar) {}

type ax7NilStreamGroup struct{}

func (*ax7NilStreamGroup) Name() string                  { return "" }
func (*ax7NilStreamGroup) Handlers() []apistream.Handler { return nil }
func (*ax7NilStreamGroup) Register(apistream.Registrar)  {}

func TestAX7_New_Good(t *coretest.T) {
	e, err := New(WithAddr(":9090"))
	coretest.RequireNoError(t, err)
	coretest.AssertNotNil(t, e)
	coretest.AssertEqual(t, ":9090", e.addr)
}

func TestAX7_New_Bad(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	coretest.AssertNotNil(t, e)
	coretest.AssertEqual(t, defaultAddr, e.addr)
}

func TestAX7_New_Ugly(t *coretest.T) {
	e, err := New(WithChatCompletions(NewModelResolver()))
	coretest.RequireNoError(t, err)
	coretest.AssertNotNil(t, e.chatCompletionsResolver)
	coretest.AssertEqual(t, defaultChatCompletionsPath, e.chatCompletionsPath)
}

func TestAX7_Engine_Addr_Good(t *coretest.T) {
	e, err := New(WithAddr(":8181"))
	coretest.RequireNoError(t, err)
	got := e.Addr()
	coretest.AssertEqual(t, ":8181", got)
}

func TestAX7_Engine_Addr_Bad(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	got := e.Addr()
	coretest.AssertEqual(t, defaultAddr, got)
}

func TestAX7_Engine_Addr_Ugly(t *coretest.T) {
	e := &Engine{addr: ""}
	got := e.Addr()
	coretest.AssertEqual(t, "", got)
	coretest.AssertEmpty(t, got)
}

func TestAX7_Engine_Groups_Good(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.Register(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	groups := e.Groups()
	coretest.AssertLen(t, groups, 1)
}

func TestAX7_Engine_Groups_Bad(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	groups := e.Groups()
	coretest.AssertEmpty(t, groups)
	coretest.AssertLen(t, groups, 0)
}

func TestAX7_Engine_Groups_Ugly(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.Register(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	groups := e.Groups()
	groups[0] = nil
	coretest.AssertNotNil(t, e.Groups()[0])
}

func TestAX7_Engine_GroupsIter_Good(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.Register(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	count := 0
	for range e.GroupsIter() {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_Engine_GroupsIter_Bad(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	count := 0
	for range e.GroupsIter() {
		count++
	}
	coretest.AssertEqual(t, 0, count)
}

func TestAX7_Engine_GroupsIter_Ugly(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.Register(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	iter := e.GroupsIter()
	e.Register(ax7RouteGroup{name: "beta", basePath: "/beta"})
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_Engine_Register_Good(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.Register(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	coretest.AssertLen(t, e.groups, 1)
	coretest.AssertEqual(t, "alpha", e.groups[0].Name())
}

func TestAX7_Engine_Register_Bad(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.Register(nil)
	coretest.AssertLen(t, e.groups, 0)
	coretest.AssertEmpty(t, e.groups)
}

func TestAX7_Engine_Register_Ugly(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	var group *ax7RouteGroup
	e.Register(group)
	coretest.AssertLen(t, e.groups, 0)
}

func TestAX7_Engine_RegisterStreamGroup_Good(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.RegisterStreamGroup(ax7StreamGroup{name: "events", handlers: []ax7StreamHandler{{protocol: "websocket", path: "/ws"}}})
	coretest.AssertLen(t, e.streamGroups, 1)
}

func TestAX7_Engine_RegisterStreamGroup_Bad(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.RegisterStreamGroup(nil)
	coretest.AssertLen(t, e.streamGroups, 0)
}

func TestAX7_Engine_RegisterStreamGroup_Ugly(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	var group *ax7NilStreamGroup
	e.RegisterStreamGroup(group)
	coretest.AssertLen(t, e.streamGroups, 0)
}

func TestAX7_Engine_Channels_Good(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.RegisterStreamGroup(ax7StreamGroup{name: "events", handlers: []ax7StreamHandler{{protocol: "websocket", path: "/ws"}}})
	channels := e.Channels()
	coretest.AssertEqual(t, []string{"/ws"}, channels)
}

func TestAX7_Engine_Channels_Bad(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	channels := e.Channels()
	coretest.AssertEmpty(t, channels)
	coretest.AssertLen(t, channels, 0)
}

func TestAX7_Engine_Channels_Ugly(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.RegisterStreamGroup(ax7StreamGroup{name: "mixed", handlers: []ax7StreamHandler{{protocol: "sse", path: "/events"}}})
	channels := e.Channels()
	coretest.AssertEmpty(t, channels)
}

func TestAX7_Engine_ChannelsIter_Good(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.RegisterStreamGroup(ax7StreamGroup{name: "events", handlers: []ax7StreamHandler{{protocol: "websocket", path: "/ws"}}})
	var channels []string
	for ch := range e.ChannelsIter() {
		channels = append(channels, ch)
	}
	coretest.AssertEqual(t, []string{"/ws"}, channels)
}

func TestAX7_Engine_ChannelsIter_Bad(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	var channels []string
	for ch := range e.ChannelsIter() {
		channels = append(channels, ch)
	}
	coretest.AssertEmpty(t, channels)
}

func TestAX7_Engine_ChannelsIter_Ugly(t *coretest.T) {
	e, err := New()
	coretest.RequireNoError(t, err)
	e.RegisterStreamGroup(ax7StreamGroup{name: "events", handlers: []ax7StreamHandler{{protocol: "websocket", path: "/ws"}}})
	iter := e.ChannelsIter()
	e.RegisterStreamGroup(ax7StreamGroup{name: "later", handlers: []ax7StreamHandler{{protocol: "websocket", path: "/later"}}})
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_WithAddr_Good(t *coretest.T) {
	e := &Engine{}
	WithAddr(":9090")(e)
	coretest.AssertEqual(t, ":9090", e.addr)
	coretest.AssertNotEmpty(t, e.addr)
}

func TestAX7_WithAddr_Bad(t *coretest.T) {
	e := &Engine{addr: defaultAddr}
	WithAddr("")(e)
	coretest.AssertEqual(t, "", e.addr)
	coretest.AssertEmpty(t, e.addr)
}

func TestAX7_WithAddr_Ugly(t *coretest.T) {
	e := &Engine{}
	WithAddr("  :9090  ")(e)
	coretest.AssertEqual(t, "  :9090  ", e.addr)
	coretest.AssertContains(t, e.addr, " ")
}

func TestAX7_WithHTTP3_Good(t *coretest.T) {
	e := &Engine{}
	WithHTTP3(" :9443 ")(e)
	coretest.AssertTrue(t, e.http3Enabled)
	coretest.AssertEqual(t, ":9443", e.http3Addr)
}

func TestAX7_WithHTTP3_Bad(t *coretest.T) {
	e := &Engine{}
	WithHTTP3("")(e)
	coretest.AssertTrue(t, e.http3Enabled)
	coretest.AssertEqual(t, "", e.http3Addr)
}

func TestAX7_WithHTTP3_Ugly(t *coretest.T) {
	e := &Engine{}
	WithHTTP3("\t:9444\n")(e)
	coretest.AssertTrue(t, e.http3Enabled)
	coretest.AssertEqual(t, ":9444", e.http3Addr)
}

func TestAX7_WithBearerAuth_Good(t *coretest.T) {
	e := &Engine{}
	WithBearerAuth("secret")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithBearerAuth_Bad(t *coretest.T) {
	e := &Engine{}
	WithBearerAuth("")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithBearerAuth_Ugly(t *coretest.T) {
	e := &Engine{swaggerPath: "/docs", openAPISpecPath: "/openapi.json"}
	WithBearerAuth("secret")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertEqual(t, "/docs", e.swaggerPath)
}

func TestAX7_WithRequestID_Good(t *coretest.T) {
	e := &Engine{}
	WithRequestID()(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithRequestID_Bad(t *coretest.T) {
	e := &Engine{}
	WithRequestID()(e)
	WithRequestID()(e)
	coretest.AssertLen(t, e.middlewares, 2)
}

func TestAX7_WithRequestID_Ugly(t *coretest.T) {
	e := &Engine{middlewares: []gin.HandlerFunc{}}
	WithRequestID()(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithResponseMeta_Good(t *coretest.T) {
	e := &Engine{}
	WithResponseMeta()(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithResponseMeta_Bad(t *coretest.T) {
	e := &Engine{}
	WithResponseMeta()(e)
	WithResponseMeta()(e)
	coretest.AssertLen(t, e.middlewares, 2)
}

func TestAX7_WithResponseMeta_Ugly(t *coretest.T) {
	e := &Engine{middlewares: []gin.HandlerFunc{func(*gin.Context) {}}}
	WithResponseMeta()(e)
	coretest.AssertLen(t, e.middlewares, 2)
	coretest.AssertNotNil(t, e.middlewares[1])
}

func TestAX7_WithCORS_Good(t *coretest.T) {
	e := &Engine{}
	WithCORS("https://example.com")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithCORS_Bad(t *coretest.T) {
	e := &Engine{}
	coretest.AssertPanics(t, func() {
		WithCORS()(e)
	})
	coretest.AssertLen(t, e.middlewares, 0)
}

func TestAX7_WithCORS_Ugly(t *coretest.T) {
	e := &Engine{}
	WithCORS("*", "https://example.com")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithMiddleware_Good(t *coretest.T) {
	e := &Engine{}
	mw := func(*gin.Context) {}
	WithMiddleware(mw)(e)
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithMiddleware_Bad(t *coretest.T) {
	e := &Engine{}
	WithMiddleware()(e)
	coretest.AssertLen(t, e.middlewares, 0)
	coretest.AssertEmpty(t, e.middlewares)
}

func TestAX7_WithMiddleware_Ugly(t *coretest.T) {
	e := &Engine{}
	WithMiddleware(nil)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNil(t, e.middlewares[0])
}

func TestAX7_WithStatic_Good(t *coretest.T) {
	e := &Engine{}
	WithStatic("/assets", ".")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithStatic_Bad(t *coretest.T) {
	e := &Engine{}
	WithStatic("", "")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithStatic_Ugly(t *coretest.T) {
	e := &Engine{}
	WithStatic("assets/", t.TempDir())(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithWSHandler_Good(t *coretest.T) {
	e := &Engine{}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	WithWSHandler(handler)(e)
	coretest.AssertNotNil(t, e.wsHandler)
	coretest.AssertNotNil(t, handler)
}

func TestAX7_WithWSHandler_Bad(t *coretest.T) {
	e := &Engine{}
	WithWSHandler(nil)(e)
	coretest.AssertNil(t, e.wsHandler)
	coretest.AssertEqual(t, "", e.wsPath)
}

func TestAX7_WithWSHandler_Ugly(t *coretest.T) {
	e := &Engine{wsHandler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	WithWSHandler(handler)(e)
	coretest.AssertNotNil(t, e.wsHandler)
	coretest.AssertNotNil(t, handler)
}

func TestAX7_WithWebSocket_Good(t *coretest.T) {
	e := &Engine{}
	handler := func(*gin.Context) {}
	WithWebSocket(handler)(e)
	coretest.AssertNotNil(t, e.wsGinHandler)
}

func TestAX7_WithWebSocket_Bad(t *coretest.T) {
	e := &Engine{}
	WithWebSocket(nil)(e)
	coretest.AssertNil(t, e.wsGinHandler)
	coretest.AssertEqual(t, "", e.wsPath)
}

func TestAX7_WithWebSocket_Ugly(t *coretest.T) {
	e := &Engine{wsGinHandler: func(*gin.Context) {}}
	WithWebSocket(func(*gin.Context) {})(e)
	coretest.AssertNotNil(t, e.wsGinHandler)
}

func TestAX7_WithWSPath_Good(t *coretest.T) {
	e := &Engine{}
	WithWSPath("socket/")(e)
	coretest.AssertEqual(t, "/socket", e.wsPath)
	coretest.AssertTrue(t, coretest.HasPrefix(e.wsPath, "/"))
}

func TestAX7_WithWSPath_Bad(t *coretest.T) {
	e := &Engine{}
	WithWSPath("")(e)
	coretest.AssertEqual(t, defaultWSPath, e.wsPath)
	coretest.AssertNotEmpty(t, e.wsPath)
}

func TestAX7_WithWSPath_Ugly(t *coretest.T) {
	e := &Engine{}
	WithWSPath("///")(e)
	coretest.AssertEqual(t, defaultWSPath, e.wsPath)
	coretest.AssertNotEmpty(t, e.wsPath)
}

func TestAX7_WithAuthentik_Good(t *coretest.T) {
	e := &Engine{}
	WithAuthentik(AuthentikConfig{TrustedProxy: true, PublicPaths: []string{"admin/"}})(e)
	coretest.AssertTrue(t, e.authentikConfig.TrustedProxy)
	coretest.AssertEqual(t, []string{"/admin"}, e.authentikConfig.PublicPaths)
}

func TestAX7_WithAuthentik_Bad(t *coretest.T) {
	e := &Engine{}
	WithAuthentik(AuthentikConfig{})(e)
	coretest.AssertFalse(t, e.authentikConfig.TrustedProxy)
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithAuthentik_Ugly(t *coretest.T) {
	e := &Engine{}
	WithAuthentik(AuthentikConfig{PublicPaths: []string{"", " /docs/ ", "docs"}})(e)
	coretest.AssertEqual(t, []string{"/docs"}, e.authentikConfig.PublicPaths)
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithSunset_Good(t *coretest.T) {
	e := &Engine{}
	WithSunset("2026-12-31", "/v2")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSunset_Bad(t *coretest.T) {
	e := &Engine{}
	WithSunset("", "")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSunset_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSunset(" 2026-12-31 ", " /v2 ")(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSwagger_Good(t *coretest.T) {
	e := &Engine{}
	WithSwagger(" Service ", " API ", " 1.0 ")(e)
	coretest.AssertTrue(t, e.swaggerEnabled)
	coretest.AssertEqual(t, "Service", e.swaggerTitle)
}

func TestAX7_WithSwagger_Bad(t *coretest.T) {
	e := &Engine{}
	WithSwagger("", "", "")(e)
	coretest.AssertTrue(t, e.swaggerEnabled)
	coretest.AssertEqual(t, "", e.swaggerTitle)
}

func TestAX7_WithSwagger_Ugly(t *coretest.T) {
	e := &Engine{swaggerTitle: "old"}
	WithSwagger("new", "desc", "2")(e)
	coretest.AssertEqual(t, "new", e.swaggerTitle)
	coretest.AssertEqual(t, "2", e.swaggerVersion)
}

func TestAX7_WithSwaggerSummary_Good(t *coretest.T) {
	e := &Engine{}
	WithSwaggerSummary(" Summary ")(e)
	coretest.AssertEqual(t, "Summary", e.swaggerSummary)
	coretest.AssertNotEmpty(t, e.swaggerSummary)
}

func TestAX7_WithSwaggerSummary_Bad(t *coretest.T) {
	e := &Engine{swaggerSummary: "existing"}
	WithSwaggerSummary("")(e)
	coretest.AssertEqual(t, "existing", e.swaggerSummary)
	coretest.AssertNotEmpty(t, e.swaggerSummary)
}

func TestAX7_WithSwaggerSummary_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSwaggerSummary("\tNested API\n")(e)
	coretest.AssertEqual(t, "Nested API", e.swaggerSummary)
}

func TestAX7_WithSwaggerPath_Good(t *coretest.T) {
	e := &Engine{}
	WithSwaggerPath("docs/")(e)
	coretest.AssertEqual(t, "/docs", e.swaggerPath)
	coretest.AssertTrue(t, coretest.HasPrefix(e.swaggerPath, "/"))
}

func TestAX7_WithSwaggerPath_Bad(t *coretest.T) {
	e := &Engine{}
	WithSwaggerPath("")(e)
	coretest.AssertEqual(t, defaultSwaggerPath, e.swaggerPath)
	coretest.AssertNotEmpty(t, e.swaggerPath)
}

func TestAX7_WithSwaggerPath_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSwaggerPath("///")(e)
	coretest.AssertEqual(t, defaultSwaggerPath, e.swaggerPath)
	coretest.AssertNotEmpty(t, e.swaggerPath)
}

func TestAX7_WithSwaggerTermsOfService_Good(t *coretest.T) {
	e := &Engine{}
	WithSwaggerTermsOfService(" https://example.com/terms ")(e)
	coretest.AssertEqual(t, "https://example.com/terms", e.swaggerTermsOfService)
	coretest.AssertContains(t, e.swaggerTermsOfService, "terms")
}

func TestAX7_WithSwaggerTermsOfService_Bad(t *coretest.T) {
	e := &Engine{swaggerTermsOfService: "keep"}
	WithSwaggerTermsOfService("")(e)
	coretest.AssertEqual(t, "keep", e.swaggerTermsOfService)
	coretest.AssertNotEmpty(t, e.swaggerTermsOfService)
}

func TestAX7_WithSwaggerTermsOfService_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSwaggerTermsOfService("\t/ref\n")(e)
	coretest.AssertEqual(t, "/ref", e.swaggerTermsOfService)
	coretest.AssertTrue(t, coretest.HasPrefix(e.swaggerTermsOfService, "/"))
}

func TestAX7_WithSwaggerContact_Good(t *coretest.T) {
	e := &Engine{}
	WithSwaggerContact(" Support ", " https://example.com ", " help@example.com ")(e)
	coretest.AssertEqual(t, "Support", e.swaggerContactName)
	coretest.AssertEqual(t, "help@example.com", e.swaggerContactEmail)
}

func TestAX7_WithSwaggerContact_Bad(t *coretest.T) {
	e := &Engine{swaggerContactName: "keep"}
	WithSwaggerContact("", "", "")(e)
	coretest.AssertEqual(t, "keep", e.swaggerContactName)
	coretest.AssertEqual(t, "", e.swaggerContactURL)
}

func TestAX7_WithSwaggerContact_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSwaggerContact("\tOps\n", "\t/docs\n", "\tops@example.com\n")(e)
	coretest.AssertEqual(t, "Ops", e.swaggerContactName)
	coretest.AssertEqual(t, "/docs", e.swaggerContactURL)
}

func TestAX7_WithSwaggerServers_Good(t *coretest.T) {
	e := &Engine{}
	WithSwaggerServers(" https://api.example.com ", "https://api.example.com")(e)
	coretest.AssertEqual(t, []string{"https://api.example.com"}, e.swaggerServers)
}

func TestAX7_WithSwaggerServers_Bad(t *coretest.T) {
	e := &Engine{}
	WithSwaggerServers("", " ")(e)
	coretest.AssertEmpty(t, e.swaggerServers)
	coretest.AssertLen(t, e.swaggerServers, 0)
}

func TestAX7_WithSwaggerServers_Ugly(t *coretest.T) {
	e := &Engine{swaggerServers: []string{"https://old.example.com"}}
	WithSwaggerServers("https://new.example.com")(e)
	coretest.AssertEqual(t, []string{"https://old.example.com", "https://new.example.com"}, e.swaggerServers)
}

func TestAX7_WithSwaggerLicense_Good(t *coretest.T) {
	e := &Engine{}
	WithSwaggerLicense(" EUPL-1.2 ", " https://example.com/license ")(e)
	coretest.AssertEqual(t, "EUPL-1.2", e.swaggerLicenseName)
	coretest.AssertEqual(t, "https://example.com/license", e.swaggerLicenseURL)
}

func TestAX7_WithSwaggerLicense_Bad(t *coretest.T) {
	e := &Engine{swaggerLicenseName: "keep"}
	WithSwaggerLicense("", "")(e)
	coretest.AssertEqual(t, "keep", e.swaggerLicenseName)
	coretest.AssertEqual(t, "", e.swaggerLicenseURL)
}

func TestAX7_WithSwaggerLicense_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSwaggerLicense("\tMIT\n", "\t/LICENSE\n")(e)
	coretest.AssertEqual(t, "MIT", e.swaggerLicenseName)
	coretest.AssertEqual(t, "/LICENSE", e.swaggerLicenseURL)
}

func TestAX7_WithSwaggerSecuritySchemes_Good(t *coretest.T) {
	e := &Engine{}
	WithSwaggerSecuritySchemes(map[string]any{" bearer ": map[string]any{"type": "http"}})(e)
	coretest.AssertLen(t, e.swaggerSecuritySchemes, 1)
	coretest.AssertNotNil(t, e.swaggerSecuritySchemes["bearer"])
}

func TestAX7_WithSwaggerSecuritySchemes_Bad(t *coretest.T) {
	e := &Engine{}
	WithSwaggerSecuritySchemes(nil)(e)
	coretest.AssertNil(t, e.swaggerSecuritySchemes)
	coretest.AssertLen(t, e.swaggerSecuritySchemes, 0)
}

func TestAX7_WithSwaggerSecuritySchemes_Ugly(t *coretest.T) {
	e := &Engine{}
	scheme := map[string]any{"type": "apiKey"}
	WithSwaggerSecuritySchemes(map[string]any{"": scheme, "key": scheme})(e)
	scheme["type"] = "mutated"
	coretest.AssertEqual(t, "apiKey", e.swaggerSecuritySchemes["key"].(map[string]any)["type"])
}

func TestAX7_WithSwaggerExternalDocs_Good(t *coretest.T) {
	e := &Engine{}
	WithSwaggerExternalDocs(" Docs ", " https://example.com/docs ")(e)
	coretest.AssertEqual(t, "Docs", e.swaggerExternalDocsDescription)
	coretest.AssertEqual(t, "https://example.com/docs", e.swaggerExternalDocsURL)
}

func TestAX7_WithSwaggerExternalDocs_Bad(t *coretest.T) {
	e := &Engine{swaggerExternalDocsDescription: "keep"}
	WithSwaggerExternalDocs("", "")(e)
	coretest.AssertEqual(t, "keep", e.swaggerExternalDocsDescription)
	coretest.AssertEqual(t, "", e.swaggerExternalDocsURL)
}

func TestAX7_WithSwaggerExternalDocs_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSwaggerExternalDocs("\tGuide\n", "\t/docs\n")(e)
	coretest.AssertEqual(t, "Guide", e.swaggerExternalDocsDescription)
	coretest.AssertEqual(t, "/docs", e.swaggerExternalDocsURL)
}

func TestAX7_WithPprof_Good(t *coretest.T) {
	e := &Engine{}
	WithPprof()(e)
	coretest.AssertTrue(t, e.pprofEnabled)
	coretest.AssertFalse(t, e.expvarEnabled)
}

func TestAX7_WithPprof_Bad(t *coretest.T) {
	e := &Engine{pprofEnabled: true}
	WithPprof()(e)
	coretest.AssertTrue(t, e.pprofEnabled)
	coretest.AssertEqual(t, true, e.pprofEnabled)
}

func TestAX7_WithPprof_Ugly(t *coretest.T) {
	e := &Engine{}
	WithPprof()(e)
	WithPprof()(e)
	coretest.AssertTrue(t, e.pprofEnabled)
}

func TestAX7_WithExpvar_Good(t *coretest.T) {
	e := &Engine{}
	WithExpvar()(e)
	coretest.AssertTrue(t, e.expvarEnabled)
	coretest.AssertFalse(t, e.pprofEnabled)
}

func TestAX7_WithExpvar_Bad(t *coretest.T) {
	e := &Engine{expvarEnabled: true}
	WithExpvar()(e)
	coretest.AssertTrue(t, e.expvarEnabled)
	coretest.AssertEqual(t, true, e.expvarEnabled)
}

func TestAX7_WithExpvar_Ugly(t *coretest.T) {
	e := &Engine{}
	WithExpvar()(e)
	WithExpvar()(e)
	coretest.AssertTrue(t, e.expvarEnabled)
}

func TestAX7_WithSecure_Good(t *coretest.T) {
	e := &Engine{}
	WithSecure()(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSecure_Bad(t *coretest.T) {
	e := &Engine{}
	WithSecure()(e)
	WithSecure()(e)
	coretest.AssertLen(t, e.middlewares, 2)
}

func TestAX7_WithSecure_Ugly(t *coretest.T) {
	e := &Engine{middlewares: []gin.HandlerFunc{func(*gin.Context) {}}}
	WithSecure()(e)
	coretest.AssertLen(t, e.middlewares, 2)
}

func TestAX7_WithGzip_Good(t *coretest.T) {
	e := &Engine{}
	WithGzip()(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithGzip_Bad(t *coretest.T) {
	e := &Engine{}
	WithGzip(-99)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithGzip_Ugly(t *coretest.T) {
	e := &Engine{}
	WithGzip(9, 1)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithBrotli_Good(t *coretest.T) {
	e := &Engine{}
	WithBrotli()(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithBrotli_Bad(t *coretest.T) {
	e := &Engine{}
	WithBrotli(-99)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithBrotli_Ugly(t *coretest.T) {
	e := &Engine{}
	WithBrotli(BrotliBestCompression, BrotliBestSpeed)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSlog_Good(t *coretest.T) {
	e := &Engine{}
	WithSlog(slog.Default())(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSlog_Bad(t *coretest.T) {
	e := &Engine{}
	WithSlog(nil)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSlog_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSlog(slog.New(slog.NewTextHandler(coretest.NewBuffer(), nil)))(e)
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithTimeout_Good(t *coretest.T) {
	e := &Engine{}
	WithTimeout(time.Second)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithTimeout_Bad(t *coretest.T) {
	e := &Engine{}
	WithTimeout(0)(e)
	coretest.AssertLen(t, e.middlewares, 0)
	coretest.AssertEmpty(t, e.middlewares)
}

func TestAX7_WithTimeout_Ugly(t *coretest.T) {
	e := &Engine{}
	WithTimeout(-time.Second)(e)
	coretest.AssertLen(t, e.middlewares, 0)
	coretest.AssertEmpty(t, e.middlewares)
}

func TestAX7_WithCache_Good(t *coretest.T) {
	e := &Engine{}
	WithCache(time.Minute, 10)(e)
	coretest.AssertEqual(t, time.Minute, e.cacheTTL)
	coretest.AssertEqual(t, 10, e.cacheMaxEntries)
}

func TestAX7_WithCache_Bad(t *coretest.T) {
	e := &Engine{}
	WithCache(0)(e)
	coretest.AssertEqual(t, time.Duration(0), e.cacheTTL)
	coretest.AssertLen(t, e.middlewares, 0)
}

func TestAX7_WithCache_Ugly(t *coretest.T) {
	e := &Engine{}
	WithCache(time.Minute, 10, 2048)(e)
	coretest.AssertEqual(t, 10, e.cacheMaxEntries)
	coretest.AssertEqual(t, 2048, e.cacheMaxBytes)
}

func TestAX7_WithCacheLimits_Good(t *coretest.T) {
	e := &Engine{}
	WithCacheLimits(time.Minute, 10, 2048)(e)
	coretest.AssertEqual(t, time.Minute, e.cacheTTL)
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithCacheLimits_Bad(t *coretest.T) {
	e := &Engine{}
	WithCacheLimits(time.Minute, 0, 0)(e)
	coretest.AssertEqual(t, time.Duration(0), e.cacheTTL)
	coretest.AssertLen(t, e.middlewares, 0)
}

func TestAX7_WithCacheLimits_Ugly(t *coretest.T) {
	e := &Engine{}
	WithCacheLimits(-time.Minute, 10, 2048)(e)
	coretest.AssertEqual(t, time.Duration(0), e.cacheTTL)
	coretest.AssertLen(t, e.middlewares, 0)
}

func TestAX7_WithRateLimit_Good(t *coretest.T) {
	e := &Engine{}
	WithRateLimit(100)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithRateLimit_Bad(t *coretest.T) {
	e := &Engine{}
	WithRateLimit(0)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithRateLimit_Ugly(t *coretest.T) {
	e := &Engine{}
	WithRateLimit(-10)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSessions_Good(t *coretest.T) {
	e := &Engine{}
	WithSessions("session", []byte("secret"))(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSessions_Bad(t *coretest.T) {
	e := &Engine{}
	WithSessions("", nil)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSessions_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSessions(" spaced ", []byte(""))(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithAuthz_Good(t *coretest.T) {
	e := &Engine{}
	WithAuthz(nil)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithAuthz_Bad(t *coretest.T) {
	e := &Engine{}
	coretest.AssertNotPanics(t, func() {
		WithAuthz(nil)(e)
	})
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithAuthz_Ugly(t *coretest.T) {
	e := &Engine{middlewares: []gin.HandlerFunc{}}
	WithAuthz(nil)(e)
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithHTTPSign_Good(t *coretest.T) {
	e := &Engine{}
	WithHTTPSign(nil)(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithHTTPSign_Bad(t *coretest.T) {
	e := &Engine{}
	coretest.AssertNotPanics(t, func() {
		WithHTTPSign(nil)(e)
	})
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithHTTPSign_Ugly(t *coretest.T) {
	e := &Engine{}
	WithHTTPSign(nil)(e)
	WithHTTPSign(nil)(e)
	coretest.AssertLen(t, e.middlewares, 2)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithSSE_Good(t *coretest.T) {
	e := &Engine{}
	broker := NewSSEBroker()
	WithSSE(broker)(e)
	coretest.AssertEqual(t, broker, e.sseBroker)
}

func TestAX7_WithSSE_Bad(t *coretest.T) {
	e := &Engine{}
	WithSSE(nil)(e)
	coretest.AssertNil(t, e.sseBroker)
	coretest.AssertEqual(t, "", e.ssePath)
}

func TestAX7_WithSSE_Ugly(t *coretest.T) {
	e := &Engine{sseBroker: NewSSEBroker()}
	replacement := NewSSEBroker()
	WithSSE(replacement)(e)
	coretest.AssertEqual(t, replacement, e.sseBroker)
}

func TestAX7_WithSSEPath_Good(t *coretest.T) {
	e := &Engine{}
	WithSSEPath("stream/")(e)
	coretest.AssertEqual(t, "/stream", e.ssePath)
	coretest.AssertTrue(t, coretest.HasPrefix(e.ssePath, "/"))
}

func TestAX7_WithSSEPath_Bad(t *coretest.T) {
	e := &Engine{}
	WithSSEPath("")(e)
	coretest.AssertEqual(t, defaultSSEPath, e.ssePath)
	coretest.AssertNotEmpty(t, e.ssePath)
}

func TestAX7_WithSSEPath_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSSEPath("///")(e)
	coretest.AssertEqual(t, defaultSSEPath, e.ssePath)
	coretest.AssertNotEmpty(t, e.ssePath)
}

func TestAX7_WithLocation_Good(t *coretest.T) {
	e := &Engine{}
	WithLocation()(e)
	coretest.AssertLen(t, e.middlewares, 1)
	coretest.AssertNotNil(t, e.middlewares[0])
}

func TestAX7_WithLocation_Bad(t *coretest.T) {
	e := &Engine{}
	WithLocation()(e)
	WithLocation()(e)
	coretest.AssertLen(t, e.middlewares, 2)
}

func TestAX7_WithLocation_Ugly(t *coretest.T) {
	e := &Engine{middlewares: []gin.HandlerFunc{}}
	WithLocation()(e)
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithGraphQL_Good(t *coretest.T) {
	e := &Engine{}
	WithGraphQL(nil, WithPlayground(), WithGraphQLPath("/gql"))(e)
	coretest.AssertNotNil(t, e.graphql)
	coretest.AssertEqual(t, "/gql", e.graphql.path)
}

func TestAX7_WithGraphQL_Bad(t *coretest.T) {
	e := &Engine{}
	WithGraphQL(nil)(e)
	coretest.AssertNotNil(t, e.graphql)
	coretest.AssertEqual(t, defaultGraphQLPath, e.graphql.path)
}

func TestAX7_WithGraphQL_Ugly(t *coretest.T) {
	e := &Engine{}
	WithGraphQL(nil, WithGraphQLPath("///"))(e)
	coretest.AssertNotNil(t, e.graphql)
	coretest.AssertEqual(t, defaultGraphQLPath, e.graphql.path)
}

func TestAX7_WithChatCompletions_Good(t *coretest.T) {
	e := &Engine{}
	resolver := NewModelResolver()
	WithChatCompletions(resolver)(e)
	coretest.AssertEqual(t, resolver, e.chatCompletionsResolver)
}

func TestAX7_WithChatCompletions_Bad(t *coretest.T) {
	e := &Engine{}
	WithChatCompletions(nil)(e)
	coretest.AssertNil(t, e.chatCompletionsResolver)
	coretest.AssertEqual(t, "", e.chatCompletionsPath)
}

func TestAX7_WithChatCompletions_Ugly(t *coretest.T) {
	e := &Engine{chatCompletionsResolver: NewModelResolver()}
	resolver := NewModelResolver()
	WithChatCompletions(resolver)(e)
	coretest.AssertEqual(t, resolver, e.chatCompletionsResolver)
}

func TestAX7_WithChatCompletionsPath_Good(t *coretest.T) {
	e := &Engine{}
	WithChatCompletionsPath("chat/")(e)
	coretest.AssertEqual(t, "/chat", e.chatCompletionsPath)
	coretest.AssertTrue(t, coretest.HasPrefix(e.chatCompletionsPath, "/"))
}

func TestAX7_WithChatCompletionsPath_Bad(t *coretest.T) {
	e := &Engine{}
	WithChatCompletionsPath("")(e)
	coretest.AssertEqual(t, defaultChatCompletionsPath, e.chatCompletionsPath)
	coretest.AssertNotEmpty(t, e.chatCompletionsPath)
}

func TestAX7_WithChatCompletionsPath_Ugly(t *coretest.T) {
	e := &Engine{}
	WithChatCompletionsPath("///")(e)
	coretest.AssertEqual(t, defaultChatCompletionsPath, e.chatCompletionsPath)
	coretest.AssertNotEmpty(t, e.chatCompletionsPath)
}

func TestAX7_WithSDKGen_Good(t *coretest.T) {
	e := &Engine{}
	WithSDKGen()(e)
	coretest.AssertTrue(t, e.sdkGenEnabled)
	coretest.AssertEqual(t, true, e.sdkGenEnabled)
}

func TestAX7_WithSDKGen_Bad(t *coretest.T) {
	e := &Engine{sdkGenEnabled: true}
	WithSDKGen()(e)
	coretest.AssertTrue(t, e.sdkGenEnabled)
	coretest.AssertEqual(t, true, e.sdkGenEnabled)
}

func TestAX7_WithSDKGen_Ugly(t *coretest.T) {
	e := &Engine{}
	WithSDKGen()(e)
	WithSDKGen()(e)
	coretest.AssertTrue(t, e.sdkGenEnabled)
}

func TestAX7_WithOpenAPISpec_Good(t *coretest.T) {
	e := &Engine{}
	WithOpenAPISpec()(e)
	coretest.AssertTrue(t, e.openAPISpecEnabled)
	coretest.AssertEqual(t, true, e.openAPISpecEnabled)
}

func TestAX7_WithOpenAPISpec_Bad(t *coretest.T) {
	e := &Engine{openAPISpecEnabled: true}
	WithOpenAPISpec()(e)
	coretest.AssertTrue(t, e.openAPISpecEnabled)
	coretest.AssertEqual(t, true, e.openAPISpecEnabled)
}

func TestAX7_WithOpenAPISpec_Ugly(t *coretest.T) {
	e := &Engine{}
	WithOpenAPISpec()(e)
	WithOpenAPISpec()(e)
	coretest.AssertTrue(t, e.openAPISpecEnabled)
}

func TestAX7_WithOpenAPISpecPath_Good(t *coretest.T) {
	e := &Engine{}
	WithOpenAPISpecPath("openapi.json")(e)
	coretest.AssertTrue(t, e.openAPISpecEnabled)
	coretest.AssertEqual(t, "/openapi.json", e.openAPISpecPath)
}

func TestAX7_WithOpenAPISpecPath_Bad(t *coretest.T) {
	e := &Engine{}
	WithOpenAPISpecPath("")(e)
	coretest.AssertTrue(t, e.openAPISpecEnabled)
	coretest.AssertEqual(t, defaultOpenAPISpecPath, e.openAPISpecPath)
}

func TestAX7_WithOpenAPISpecPath_Ugly(t *coretest.T) {
	e := &Engine{}
	WithOpenAPISpecPath("///")(e)
	coretest.AssertTrue(t, e.openAPISpecEnabled)
	coretest.AssertEqual(t, "///", e.openAPISpecPath)
}
