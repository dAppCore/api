// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"net/http/httptest"
	"time"

	coretest "dappco.re/go"

	"github.com/gin-gonic/gin"
)

func ax7GinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return ctx, rec
}

func TestAX7_OK_Good(t *coretest.T) {
	resp := OK("payload")
	coretest.AssertTrue(t, resp.Success)
	coretest.AssertEqual(t, "payload", resp.Data)
	coretest.AssertNil(t, resp.Error)
}

func TestAX7_OK_Bad(t *coretest.T) {
	resp := OK[any](nil)
	coretest.AssertTrue(t, resp.Success)
	coretest.AssertNil(t, resp.Data)
	coretest.AssertNil(t, resp.Meta)
}

func TestAX7_OK_Ugly(t *coretest.T) {
	resp := OK(map[string]any{"empty": ""})
	resp.Data["empty"] = "changed"
	coretest.AssertTrue(t, resp.Success)
	coretest.AssertEqual(t, "changed", resp.Data["empty"])
}

func TestAX7_Fail_Good(t *coretest.T) {
	resp := Fail("not_found", "missing")
	coretest.AssertFalse(t, resp.Success)
	coretest.AssertNotNil(t, resp.Error)
	coretest.AssertEqual(t, "not_found", resp.Error.Code)
}

func TestAX7_Fail_Bad(t *coretest.T) {
	resp := Fail("", "")
	coretest.AssertFalse(t, resp.Success)
	coretest.AssertEqual(t, "", resp.Error.Code)
	coretest.AssertEqual(t, "", resp.Error.Message)
}

func TestAX7_Fail_Ugly(t *coretest.T) {
	resp := Fail(" spaced ", " message ")
	coretest.AssertFalse(t, resp.Success)
	coretest.AssertEqual(t, " spaced ", resp.Error.Code)
	coretest.AssertEqual(t, " message ", resp.Error.Message)
}

func TestAX7_FailWithDetails_Good(t *coretest.T) {
	details := map[string]string{"field": "email"}
	resp := FailWithDetails("invalid", "bad input", details)
	coretest.AssertFalse(t, resp.Success)
	coretest.AssertEqual(t, details, resp.Error.Details)
}

func TestAX7_FailWithDetails_Bad(t *coretest.T) {
	resp := FailWithDetails("", "", nil)
	coretest.AssertFalse(t, resp.Success)
	coretest.AssertNotNil(t, resp.Error)
	coretest.AssertNil(t, resp.Error.Details)
}

func TestAX7_FailWithDetails_Ugly(t *coretest.T) {
	resp := FailWithDetails("invalid", "bad", []string{"a", "b"})
	coretest.AssertFalse(t, resp.Success)
	coretest.AssertEqual(t, []string{"a", "b"}, resp.Error.Details)
}

func TestAX7_Paginated_Good(t *coretest.T) {
	resp := Paginated([]string{"a"}, 2, 10, 25)
	coretest.AssertTrue(t, resp.Success)
	coretest.AssertEqual(t, []string{"a"}, resp.Data)
	coretest.AssertEqual(t, 25, resp.Meta.Total)
}

func TestAX7_Paginated_Bad(t *coretest.T) {
	resp := Paginated([]string{}, 0, 0, 0)
	coretest.AssertTrue(t, resp.Success)
	coretest.AssertEqual(t, 0, resp.Meta.Page)
	coretest.AssertEmpty(t, resp.Data)
}

func TestAX7_Paginated_Ugly(t *coretest.T) {
	resp := Paginated("items", -1, -10, -20)
	coretest.AssertTrue(t, resp.Success)
	coretest.AssertEqual(t, -1, resp.Meta.Page)
	coretest.AssertEqual(t, -20, resp.Meta.Total)
}

func TestAX7_AttachRequestMeta_Good(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(requestIDContextKey, "req-1")
	resp := AttachRequestMeta(ctx, OK("payload"))
	coretest.AssertNotNil(t, resp.Meta)
	coretest.AssertEqual(t, "req-1", resp.Meta.RequestID)
}

func TestAX7_AttachRequestMeta_Bad(t *coretest.T) {
	ctx, _ := ax7GinContext()
	resp := AttachRequestMeta(ctx, OK("payload"))
	coretest.AssertNil(t, resp.Meta)
	coretest.AssertEqual(t, "payload", resp.Data)
}

func TestAX7_AttachRequestMeta_Ugly(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(requestIDContextKey, "req-2")
	resp := Paginated([]int{1}, 3, 10, 30)
	resp = AttachRequestMeta(ctx, resp)
	coretest.AssertEqual(t, 3, resp.Meta.Page)
	coretest.AssertEqual(t, "req-2", resp.Meta.RequestID)
}

func TestAX7_GetRequestID_Good(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(requestIDContextKey, "req-1")
	got := GetRequestID(ctx)
	coretest.AssertEqual(t, "req-1", got)
}

func TestAX7_GetRequestID_Bad(t *coretest.T) {
	ctx, _ := ax7GinContext()
	got := GetRequestID(ctx)
	coretest.AssertEqual(t, "", got)
	coretest.AssertEmpty(t, got)
}

func TestAX7_GetRequestID_Ugly(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(requestIDContextKey, 42)
	got := GetRequestID(ctx)
	coretest.AssertEqual(t, "", got)
}

func TestAX7_GetRequestDuration_Good(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(requestStartContextKey, time.Now().Add(-time.Millisecond))
	got := GetRequestDuration(ctx)
	coretest.AssertGreater(t, got, time.Duration(0))
}

func TestAX7_GetRequestDuration_Bad(t *coretest.T) {
	ctx, _ := ax7GinContext()
	got := GetRequestDuration(ctx)
	coretest.AssertEqual(t, time.Duration(0), got)
	coretest.AssertFalse(t, got > 0)
}

func TestAX7_GetRequestDuration_Ugly(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(requestStartContextKey, "not-time")
	got := GetRequestDuration(ctx)
	coretest.AssertEqual(t, time.Duration(0), got)
}

func TestAX7_GetRequestMeta_Good(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(requestIDContextKey, "req-1")
	ctx.Set(requestStartContextKey, time.Now().Add(-time.Millisecond))
	meta := GetRequestMeta(ctx)
	coretest.AssertNotNil(t, meta)
	coretest.AssertEqual(t, "req-1", meta.RequestID)
}

func TestAX7_GetRequestMeta_Bad(t *coretest.T) {
	ctx, _ := ax7GinContext()
	meta := GetRequestMeta(ctx)
	coretest.AssertNil(t, meta)
	coretest.AssertEqual(t, "", GetRequestID(ctx))
}

func TestAX7_GetRequestMeta_Ugly(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(requestStartContextKey, time.Now().Add(-time.Millisecond))
	meta := GetRequestMeta(ctx)
	coretest.AssertNotNil(t, meta)
	coretest.AssertNotEmpty(t, meta.Duration)
}

func TestAX7_AuthentikUser_HasGroup_Good(t *coretest.T) {
	user := &AuthentikUser{Groups: []string{"admins", "editors"}}
	got := user.HasGroup("admins")
	coretest.AssertTrue(t, got)
}

func TestAX7_AuthentikUser_HasGroup_Bad(t *coretest.T) {
	user := &AuthentikUser{Groups: []string{"editors"}}
	got := user.HasGroup("admins")
	coretest.AssertFalse(t, got)
}

func TestAX7_AuthentikUser_HasGroup_Ugly(t *coretest.T) {
	user := &AuthentikUser{}
	got := user.HasGroup("")
	coretest.AssertFalse(t, got)
}

func TestAX7_GetUser_Good(t *coretest.T) {
	ctx, _ := ax7GinContext()
	user := &AuthentikUser{Username: "ada"}
	ctx.Set(authentikUserKey, user)
	got := GetUser(ctx)
	coretest.AssertEqual(t, user, got)
}

func TestAX7_GetUser_Bad(t *coretest.T) {
	ctx, _ := ax7GinContext()
	got := GetUser(ctx)
	coretest.AssertNil(t, got)
}

func TestAX7_GetUser_Ugly(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(authentikUserKey, "not-user")
	got := GetUser(ctx)
	coretest.AssertNil(t, got)
}

func TestAX7_RequireAuth_Good(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ctx.Set(authentikUserKey, &AuthentikUser{Username: "ada"})
	RequireAuth()(ctx)
	coretest.AssertFalse(t, ctx.IsAborted())
	coretest.AssertEqual(t, http.StatusOK, rec.Code)
}

func TestAX7_RequireAuth_Bad(t *coretest.T) {
	ctx, rec := ax7GinContext()
	RequireAuth()(ctx)
	coretest.AssertTrue(t, ctx.IsAborted())
	coretest.AssertEqual(t, http.StatusUnauthorized, rec.Code)
}

func TestAX7_RequireAuth_Ugly(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ctx.Set(authentikUserKey, "not-user")
	RequireAuth()(ctx)
	coretest.AssertTrue(t, ctx.IsAborted())
	coretest.AssertEqual(t, http.StatusUnauthorized, rec.Code)
}

func TestAX7_RequireGroup_Good(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ctx.Set(authentikUserKey, &AuthentikUser{Groups: []string{"admins"}})
	RequireGroup("admins")(ctx)
	coretest.AssertFalse(t, ctx.IsAborted())
	coretest.AssertEqual(t, http.StatusOK, rec.Code)
}

func TestAX7_RequireGroup_Bad(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ctx.Set(authentikUserKey, &AuthentikUser{Groups: []string{"users"}})
	RequireGroup("admins")(ctx)
	coretest.AssertTrue(t, ctx.IsAborted())
	coretest.AssertEqual(t, http.StatusForbidden, rec.Code)
}

func TestAX7_RequireGroup_Ugly(t *coretest.T) {
	ctx, rec := ax7GinContext()
	RequireGroup("")(ctx)
	coretest.AssertTrue(t, ctx.IsAborted())
	coretest.AssertEqual(t, http.StatusUnauthorized, rec.Code)
}

func TestAX7_WithI18n_Good(t *coretest.T) {
	e := &Engine{}
	WithI18n(I18nConfig{DefaultLocale: "en", Supported: []string{"fr"}})(e)
	coretest.AssertEqual(t, "en", e.i18nConfig.DefaultLocale)
	coretest.AssertEqual(t, []string{"fr"}, e.i18nConfig.Supported)
}

func TestAX7_WithI18n_Bad(t *coretest.T) {
	e := &Engine{}
	WithI18n()(e)
	coretest.AssertEqual(t, "en", e.i18nConfig.DefaultLocale)
	coretest.AssertLen(t, e.middlewares, 1)
}

func TestAX7_WithI18n_Ugly(t *coretest.T) {
	e := &Engine{}
	WithI18n(I18nConfig{Messages: map[string]map[string]string{"en": {"hello": "Hello"}}})(e)
	coretest.AssertEqual(t, "en", e.i18nConfig.DefaultLocale)
	coretest.AssertEqual(t, "Hello", e.i18nConfig.Messages["en"]["hello"])
}

func TestAX7_GetLocale_Good(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(i18nContextKey, "fr")
	got := GetLocale(ctx)
	coretest.AssertEqual(t, "fr", got)
}

func TestAX7_GetLocale_Bad(t *coretest.T) {
	ctx, _ := ax7GinContext()
	got := GetLocale(ctx)
	coretest.AssertEqual(t, "en", got)
}

func TestAX7_GetLocale_Ugly(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(i18nContextKey, 123)
	got := GetLocale(ctx)
	coretest.AssertEqual(t, "en", got)
}

func TestAX7_GetMessage_Good(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(i18nMessagesKey, map[string]string{"hello": "Bonjour"})
	msg, ok := GetMessage(ctx, "hello")
	coretest.AssertTrue(t, ok)
	coretest.AssertEqual(t, "Bonjour", msg)
}

func TestAX7_GetMessage_Bad(t *coretest.T) {
	ctx, _ := ax7GinContext()
	msg, ok := GetMessage(ctx, "missing")
	coretest.AssertFalse(t, ok)
	coretest.AssertEqual(t, "", msg)
}

func TestAX7_GetMessage_Ugly(t *coretest.T) {
	ctx, _ := ax7GinContext()
	ctx.Set(i18nContextKey, "fr-CA")
	ctx.Set(i18nCatalogKey, map[string]map[string]string{"fr": {"hello": "Bonjour"}})
	msg, ok := GetMessage(ctx, "hello")
	coretest.AssertTrue(t, ok)
	coretest.AssertEqual(t, "Bonjour", msg)
}

func TestAX7_Engine_AuthentikConfig_Good(t *coretest.T) {
	e, err := New(WithAuthentik(AuthentikConfig{Issuer: " issuer ", PublicPaths: []string{"docs/"}}))
	coretest.RequireNoError(t, err)
	cfg := e.AuthentikConfig()
	coretest.AssertEqual(t, "issuer", cfg.Issuer)
	coretest.AssertEqual(t, []string{"/docs"}, cfg.PublicPaths)
}

func TestAX7_Engine_AuthentikConfig_Bad(t *coretest.T) {
	var e *Engine
	cfg := e.AuthentikConfig()
	coretest.AssertEqual(t, AuthentikConfig{}, cfg)
}

func TestAX7_Engine_AuthentikConfig_Ugly(t *coretest.T) {
	e, err := New(WithAuthentik(AuthentikConfig{PublicPaths: []string{"docs"}}))
	coretest.RequireNoError(t, err)
	cfg := e.AuthentikConfig()
	cfg.PublicPaths[0] = "/mutated"
	coretest.AssertEqual(t, []string{"/docs"}, e.AuthentikConfig().PublicPaths)
}

func TestAX7_Engine_I18nConfig_Good(t *coretest.T) {
	e, err := New(WithI18n(I18nConfig{DefaultLocale: "en", Supported: []string{"fr"}}))
	coretest.RequireNoError(t, err)
	cfg := e.I18nConfig()
	coretest.AssertEqual(t, "en", cfg.DefaultLocale)
	coretest.AssertEqual(t, []string{"fr"}, cfg.Supported)
}

func TestAX7_Engine_I18nConfig_Bad(t *coretest.T) {
	var e *Engine
	cfg := e.I18nConfig()
	coretest.AssertEqual(t, I18nConfig{}, cfg)
}

func TestAX7_Engine_I18nConfig_Ugly(t *coretest.T) {
	e, err := New(WithI18n(I18nConfig{Supported: []string{"fr"}}))
	coretest.RequireNoError(t, err)
	cfg := e.I18nConfig()
	cfg.Supported[0] = "de"
	coretest.AssertEqual(t, []string{"fr"}, e.I18nConfig().Supported)
}

func TestAX7_Engine_GraphQLConfig_Good(t *coretest.T) {
	e, err := New(WithGraphQL(nil, WithPlayground(), WithGraphQLPath("/gql")))
	coretest.RequireNoError(t, err)
	cfg := e.GraphQLConfig()
	coretest.AssertTrue(t, cfg.Enabled)
	coretest.AssertEqual(t, "/gql/playground", cfg.PlaygroundPath)
}

func TestAX7_Engine_GraphQLConfig_Bad(t *coretest.T) {
	var e *Engine
	cfg := e.GraphQLConfig()
	coretest.AssertFalse(t, cfg.Enabled)
	coretest.AssertEqual(t, "", cfg.Path)
}

func TestAX7_Engine_GraphQLConfig_Ugly(t *coretest.T) {
	e, err := New(WithGraphQL(nil, WithGraphQLPath("///")))
	coretest.RequireNoError(t, err)
	cfg := e.GraphQLConfig()
	coretest.AssertTrue(t, cfg.Enabled)
	coretest.AssertEqual(t, defaultGraphQLPath, cfg.Path)
}

func TestAX7_Engine_CacheConfig_Good(t *coretest.T) {
	e, err := New(WithCacheLimits(time.Minute, 10, 1024))
	coretest.RequireNoError(t, err)
	cfg := e.CacheConfig()
	coretest.AssertTrue(t, cfg.Enabled)
	coretest.AssertEqual(t, 10, cfg.MaxEntries)
}

func TestAX7_Engine_CacheConfig_Bad(t *coretest.T) {
	var e *Engine
	cfg := e.CacheConfig()
	coretest.AssertFalse(t, cfg.Enabled)
	coretest.AssertEqual(t, time.Duration(0), cfg.TTL)
}

func TestAX7_Engine_CacheConfig_Ugly(t *coretest.T) {
	e, err := New(WithCacheLimits(time.Minute, 0, 2048))
	coretest.RequireNoError(t, err)
	cfg := e.CacheConfig()
	coretest.AssertTrue(t, cfg.Enabled)
	coretest.AssertEqual(t, 2048, cfg.MaxBytes)
}

func TestAX7_Engine_TransportConfig_Good(t *coretest.T) {
	e, err := New(WithSwagger("API", "", "1"), WithWSPath("/socket"), WithOpenAPISpec())
	coretest.RequireNoError(t, err)
	cfg := e.TransportConfig()
	coretest.AssertTrue(t, cfg.SwaggerEnabled)
	coretest.AssertEqual(t, "/socket", cfg.WSPath)
}

func TestAX7_Engine_TransportConfig_Bad(t *coretest.T) {
	var e *Engine
	cfg := e.TransportConfig()
	coretest.AssertFalse(t, cfg.SwaggerEnabled)
	coretest.AssertEqual(t, "", cfg.WSPath)
}

func TestAX7_Engine_TransportConfig_Ugly(t *coretest.T) {
	e, err := New(WithSSEPath("/stream"), WithChatCompletionsPath("/chat"))
	coretest.RequireNoError(t, err)
	cfg := e.TransportConfig()
	coretest.AssertEqual(t, "/stream", cfg.SSEPath)
	coretest.AssertEqual(t, "/chat", cfg.ChatCompletionsPath)
}

func TestAX7_Engine_RuntimeConfig_Good(t *coretest.T) {
	e, err := New(WithSwagger("API", "", "1"), WithI18n(I18nConfig{DefaultLocale: "en"}))
	coretest.RequireNoError(t, err)
	cfg := e.RuntimeConfig()
	coretest.AssertEqual(t, "API", cfg.Swagger.Title)
	coretest.AssertEqual(t, "en", cfg.I18n.DefaultLocale)
}

func TestAX7_Engine_RuntimeConfig_Bad(t *coretest.T) {
	var e *Engine
	cfg := e.RuntimeConfig()
	coretest.AssertFalse(t, cfg.Swagger.Enabled)
	coretest.AssertFalse(t, cfg.Cache.Enabled)
}

func TestAX7_Engine_RuntimeConfig_Ugly(t *coretest.T) {
	e, err := New(WithCacheLimits(time.Minute, 1, 0), WithAuthentik(AuthentikConfig{TrustedProxy: true}))
	coretest.RequireNoError(t, err)
	cfg := e.RuntimeConfig()
	coretest.AssertTrue(t, cfg.Cache.Enabled)
	coretest.AssertTrue(t, cfg.Authentik.TrustedProxy)
}

func TestAX7_Engine_SwaggerConfig_Good(t *coretest.T) {
	e, err := New(WithSwagger("API", "desc", "1"), WithSwaggerPath("/docs"))
	coretest.RequireNoError(t, err)
	cfg := e.SwaggerConfig()
	coretest.AssertTrue(t, cfg.Enabled)
	coretest.AssertEqual(t, "/docs", cfg.Path)
}

func TestAX7_Engine_SwaggerConfig_Bad(t *coretest.T) {
	var e *Engine
	cfg := e.SwaggerConfig()
	coretest.AssertFalse(t, cfg.Enabled)
	coretest.AssertEqual(t, "", cfg.Title)
}

func TestAX7_Engine_SwaggerConfig_Ugly(t *coretest.T) {
	e, err := New(WithSwaggerServers("https://api.example.com"))
	coretest.RequireNoError(t, err)
	cfg := e.SwaggerConfig()
	cfg.Servers[0] = "mutated"
	coretest.AssertEqual(t, []string{"https://api.example.com"}, e.SwaggerConfig().Servers)
}

func TestAX7_Engine_OpenAPISpecBuilder_Good(t *coretest.T) {
	e, err := New(WithSwagger("API", "desc", "1"), WithOpenAPISpec())
	coretest.RequireNoError(t, err)
	builder := e.OpenAPISpecBuilder()
	coretest.AssertEqual(t, "API", builder.Title)
	coretest.AssertTrue(t, builder.OpenAPISpecEnabled)
}

func TestAX7_Engine_OpenAPISpecBuilder_Bad(t *coretest.T) {
	var e *Engine
	builder := e.OpenAPISpecBuilder()
	coretest.AssertNotNil(t, builder)
	coretest.AssertEqual(t, "", builder.Title)
}

func TestAX7_Engine_OpenAPISpecBuilder_Ugly(t *coretest.T) {
	e, err := New(WithCacheLimits(time.Minute, 2, 3), WithI18n(I18nConfig{Supported: []string{"fr"}}))
	coretest.RequireNoError(t, err)
	builder := e.OpenAPISpecBuilder()
	coretest.AssertEqual(t, "1m0s", builder.CacheTTL)
	coretest.AssertEqual(t, []string{"fr"}, builder.I18nSupportedLocales)
}

func TestAX7_WithPlayground_Good(t *coretest.T) {
	cfg := &graphqlConfig{}
	WithPlayground()(cfg)
	coretest.AssertTrue(t, cfg.playground)
}

func TestAX7_WithPlayground_Bad(t *coretest.T) {
	cfg := &graphqlConfig{playground: true}
	WithPlayground()(cfg)
	coretest.AssertTrue(t, cfg.playground)
}

func TestAX7_WithPlayground_Ugly(t *coretest.T) {
	cfg := &graphqlConfig{path: "/gql"}
	WithPlayground()(cfg)
	coretest.AssertTrue(t, cfg.playground)
	coretest.AssertEqual(t, "/gql", cfg.path)
}

func TestAX7_WithGraphQLPath_Good(t *coretest.T) {
	cfg := &graphqlConfig{}
	WithGraphQLPath("gql/")(cfg)
	coretest.AssertEqual(t, "/gql", cfg.path)
}

func TestAX7_WithGraphQLPath_Bad(t *coretest.T) {
	cfg := &graphqlConfig{}
	WithGraphQLPath("")(cfg)
	coretest.AssertEqual(t, defaultGraphQLPath, cfg.path)
}

func TestAX7_WithGraphQLPath_Ugly(t *coretest.T) {
	cfg := &graphqlConfig{}
	WithGraphQLPath("///")(cfg)
	coretest.AssertEqual(t, defaultGraphQLPath, cfg.path)
}

func TestAX7_RegisterSpecGroups_Good(t *coretest.T) {
	ResetSpecGroups()
	group := ax7RouteGroup{name: "alpha", basePath: "/alpha"}
	RegisterSpecGroups(group)
	coretest.AssertLen(t, RegisteredSpecGroups(), 1)
	coretest.AssertEqual(t, "alpha", RegisteredSpecGroups()[0].Name())
}

func TestAX7_RegisterSpecGroups_Bad(t *coretest.T) {
	ResetSpecGroups()
	RegisterSpecGroups(nil)
	coretest.AssertEmpty(t, RegisteredSpecGroups())
	coretest.AssertLen(t, RegisteredSpecGroups(), 0)
}

func TestAX7_RegisterSpecGroups_Ugly(t *coretest.T) {
	ResetSpecGroups()
	group := ax7RouteGroup{name: "alpha", basePath: "/alpha"}
	RegisterSpecGroups(group, group)
	coretest.AssertLen(t, RegisteredSpecGroups(), 1)
}

func TestAX7_RegisterSpecGroupsIter_Good(t *coretest.T) {
	ResetSpecGroups()
	groups := []RouteGroup{ax7RouteGroup{name: "alpha", basePath: "/alpha"}}
	RegisterSpecGroupsIter(func(yield func(RouteGroup) bool) {
		yield(groups[0])
	})
	coretest.AssertLen(t, RegisteredSpecGroups(), 1)
}

func TestAX7_RegisterSpecGroupsIter_Bad(t *coretest.T) {
	ResetSpecGroups()
	RegisterSpecGroupsIter(nil)
	coretest.AssertEmpty(t, RegisteredSpecGroups())
	coretest.AssertLen(t, RegisteredSpecGroups(), 0)
}

func TestAX7_RegisterSpecGroupsIter_Ugly(t *coretest.T) {
	ResetSpecGroups()
	RegisterSpecGroupsIter(func(yield func(RouteGroup) bool) {
		yield(nil)
		yield(ax7RouteGroup{name: "beta", basePath: "/beta"})
	})
	coretest.AssertLen(t, RegisteredSpecGroups(), 1)
}

func TestAX7_RegisteredSpecGroups_Good(t *coretest.T) {
	ResetSpecGroups()
	RegisterSpecGroups(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	groups := RegisteredSpecGroups()
	coretest.AssertLen(t, groups, 1)
	coretest.AssertEqual(t, "alpha", groups[0].Name())
}

func TestAX7_RegisteredSpecGroups_Bad(t *coretest.T) {
	ResetSpecGroups()
	groups := RegisteredSpecGroups()
	coretest.AssertEmpty(t, groups)
	coretest.AssertLen(t, groups, 0)
}

func TestAX7_RegisteredSpecGroups_Ugly(t *coretest.T) {
	ResetSpecGroups()
	RegisterSpecGroups(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	groups := RegisteredSpecGroups()
	groups[0] = nil
	coretest.AssertNotNil(t, RegisteredSpecGroups()[0])
}

func TestAX7_RegisteredSpecGroupsIter_Good(t *coretest.T) {
	ResetSpecGroups()
	RegisterSpecGroups(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	count := 0
	for range RegisteredSpecGroupsIter() {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_RegisteredSpecGroupsIter_Bad(t *coretest.T) {
	ResetSpecGroups()
	count := 0
	for range RegisteredSpecGroupsIter() {
		count++
	}
	coretest.AssertEqual(t, 0, count)
}

func TestAX7_RegisteredSpecGroupsIter_Ugly(t *coretest.T) {
	ResetSpecGroups()
	RegisterSpecGroups(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	iter := RegisteredSpecGroupsIter()
	RegisterSpecGroups(ax7RouteGroup{name: "beta", basePath: "/beta"})
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_SpecGroupsIter_Good(t *coretest.T) {
	ResetSpecGroups()
	RegisterSpecGroups(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	var groups []RouteGroup
	for group := range SpecGroupsIter(ax7RouteGroup{name: "beta", basePath: "/beta"}) {
		groups = append(groups, group)
	}
	coretest.AssertLen(t, groups, 2)
}

func TestAX7_SpecGroupsIter_Bad(t *coretest.T) {
	ResetSpecGroups()
	var groups []RouteGroup
	for group := range SpecGroupsIter(nil) {
		groups = append(groups, group)
	}
	coretest.AssertEmpty(t, groups)
}

func TestAX7_SpecGroupsIter_Ugly(t *coretest.T) {
	ResetSpecGroups()
	group := ax7RouteGroup{name: "alpha", basePath: "/alpha"}
	RegisterSpecGroups(group)
	var groups []RouteGroup
	for item := range SpecGroupsIter(group) {
		groups = append(groups, item)
	}
	coretest.AssertLen(t, groups, 1)
}

func TestAX7_ResetSpecGroups_Good(t *coretest.T) {
	RegisterSpecGroups(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	ResetSpecGroups()
	coretest.AssertEmpty(t, RegisteredSpecGroups())
	coretest.AssertLen(t, RegisteredSpecGroups(), 0)
}

func TestAX7_ResetSpecGroups_Bad(t *coretest.T) {
	ResetSpecGroups()
	ResetSpecGroups()
	coretest.AssertEmpty(t, RegisteredSpecGroups())
}

func TestAX7_ResetSpecGroups_Ugly(t *coretest.T) {
	RegisterSpecGroups(ax7RouteGroup{name: "alpha", basePath: "/alpha"})
	ResetSpecGroups()
	RegisterSpecGroups(ax7RouteGroup{name: "beta", basePath: "/beta"})
	coretest.AssertEqual(t, "beta", RegisteredSpecGroups()[0].Name())
}

func TestAX7_TransformerInFunc_TransformIn_Good(t *coretest.T) {
	fn := TransformerInFunc[string, string](func(_ *gin.Context, in string) (string, error) { return in + "!", nil })
	got, err := fn.TransformIn(nil, "go")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "go!", got)
}

func TestAX7_TransformerInFunc_TransformIn_Bad(t *coretest.T) {
	fn := TransformerInFunc[string, string](func(_ *gin.Context, _ string) (string, error) { return "", coretest.NewError("bad") })
	got, err := fn.TransformIn(nil, "go")
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "", got)
}

func TestAX7_TransformerInFunc_TransformIn_Ugly(t *coretest.T) {
	fn := TransformerInFunc[map[string]any, map[string]any](func(_ *gin.Context, in map[string]any) (map[string]any, error) { return in, nil })
	got, err := fn.TransformIn(nil, nil)
	coretest.RequireNoError(t, err)
	coretest.AssertNil(t, got)
}

func TestAX7_TransformerOutFunc_TransformOut_Good(t *coretest.T) {
	fn := TransformerOutFunc[string, string](func(_ *gin.Context, in string) (string, error) { return in + "!", nil })
	got, err := fn.TransformOut(nil, "go")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "go!", got)
}

func TestAX7_TransformerOutFunc_TransformOut_Bad(t *coretest.T) {
	fn := TransformerOutFunc[string, string](func(_ *gin.Context, _ string) (string, error) { return "", coretest.NewError("bad") })
	got, err := fn.TransformOut(nil, "go")
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "", got)
}

func TestAX7_TransformerOutFunc_TransformOut_Ugly(t *coretest.T) {
	fn := TransformerOutFunc[map[string]any, map[string]any](func(_ *gin.Context, in map[string]any) (map[string]any, error) { return in, nil })
	got, err := fn.TransformOut(nil, nil)
	coretest.RequireNoError(t, err)
	coretest.AssertNil(t, got)
}

func TestAX7_RenameFields_Good(t *coretest.T) {
	renamer := RenameFields(map[string]string{"full_name": "name"})
	got, err := renamer.TransformIn(nil, map[string]any{"full_name": "Ada"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "Ada", got["name"])
}

func TestAX7_RenameFields_Bad(t *coretest.T) {
	renamer := RenameFields(nil)
	got, err := renamer.TransformIn(nil, map[string]any{"name": "Ada"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "Ada", got["name"])
}

func TestAX7_RenameFields_Ugly(t *coretest.T) {
	fields := map[string]string{"a": "b"}
	renamer := RenameFields(fields)
	fields["a"] = "c"
	got, err := renamer.TransformIn(nil, map[string]any{"a": 1})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 1, got["b"])
}

func TestAX7_FieldRenamer_TransformIn_Good(t *coretest.T) {
	renamer := FieldRenamer{Fields: map[string]string{"full": "name"}}
	got, err := renamer.TransformIn(nil, map[string]any{"full": "Ada"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "Ada", got["name"])
}

func TestAX7_FieldRenamer_TransformIn_Bad(t *coretest.T) {
	renamer := FieldRenamer{Fields: map[string]string{"missing": "name"}}
	got, err := renamer.TransformIn(nil, map[string]any{"full": "Ada"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "Ada", got["full"])
}

func TestAX7_FieldRenamer_TransformIn_Ugly(t *coretest.T) {
	renamer := FieldRenamer{Fields: map[string]string{"": "name", "full": ""}}
	got, err := renamer.TransformIn(nil, map[string]any{"full": "Ada"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "Ada", got["full"])
}

func TestAX7_FieldRenamer_TransformOut_Good(t *coretest.T) {
	renamer := FieldRenamer{Fields: map[string]string{"name": "full_name"}}
	got, err := renamer.TransformOut(nil, map[string]any{"name": "Ada"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "Ada", got["full_name"])
}

func TestAX7_FieldRenamer_TransformOut_Bad(t *coretest.T) {
	renamer := FieldRenamer{Fields: map[string]string{"missing": "name"}}
	got, err := renamer.TransformOut(nil, map[string]any{"name": "Ada"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "Ada", got["name"])
}

func TestAX7_FieldRenamer_TransformOut_Ugly(t *coretest.T) {
	renamer := FieldRenamer{Fields: map[string]string{"name": "name"}}
	got, err := renamer.TransformOut(nil, map[string]any{"name": "Ada"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "Ada", got["name"])
}

func TestAX7_Number_String_Good(t *coretest.T) {
	n := jsonNumber("42")
	got := n.String()
	coretest.AssertEqual(t, "42", got)
}

func TestAX7_Number_String_Bad(t *coretest.T) {
	n := jsonNumber("")
	got := n.String()
	coretest.AssertEqual(t, "", got)
}

func TestAX7_Number_String_Ugly(t *coretest.T) {
	n := jsonNumber("-1.5")
	got := n.String()
	coretest.AssertEqual(t, "-1.5", got)
}

func TestAX7_Number_Float64_Good(t *coretest.T) {
	n := jsonNumber("1.5")
	got, err := n.Float64()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 1.5, got)
}

func TestAX7_Number_Float64_Bad(t *coretest.T) {
	n := jsonNumber("nope")
	got, err := n.Float64()
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, 0.0, got)
}

func TestAX7_Number_Float64_Ugly(t *coretest.T) {
	n := jsonNumber("-0")
	got, err := n.Float64()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, -0.0, got)
}

func TestAX7_Number_Int64_Good(t *coretest.T) {
	n := jsonNumber("42")
	got, err := n.Int64()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, int64(42), got)
}

func TestAX7_Number_Int64_Bad(t *coretest.T) {
	n := jsonNumber("1.5")
	got, err := n.Int64()
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, int64(0), got)
}

func TestAX7_Number_Int64_Ugly(t *coretest.T) {
	n := jsonNumber("-7")
	got, err := n.Int64()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, int64(-7), got)
}

func TestAX7_Number_MarshalJSON_Good(t *coretest.T) {
	n := jsonNumber("42")
	data, err := n.MarshalJSON()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, []byte("42"), data)
}

func TestAX7_Number_MarshalJSON_Bad(t *coretest.T) {
	n := jsonNumber("")
	data, err := n.MarshalJSON()
	coretest.AssertError(t, err)
	coretest.AssertNil(t, data)
}

func TestAX7_Number_MarshalJSON_Ugly(t *coretest.T) {
	n := jsonNumber("-1.25")
	data, err := n.MarshalJSON()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, []byte("-1.25"), data)
}

func TestAX7_RawMessage_MarshalJSON_Good(t *coretest.T) {
	msg := jsonRawMessage(`{"ok":true}`)
	data, err := msg.MarshalJSON()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, []byte(`{"ok":true}`), data)
}

func TestAX7_RawMessage_MarshalJSON_Bad(t *coretest.T) {
	var msg jsonRawMessage
	data, err := msg.MarshalJSON()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, []byte("null"), data)
}

func TestAX7_RawMessage_MarshalJSON_Ugly(t *coretest.T) {
	msg := jsonRawMessage(`[]`)
	data, err := msg.MarshalJSON()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, []byte(`[]`), data)
}

func TestAX7_RawMessage_UnmarshalJSON_Good(t *coretest.T) {
	var msg jsonRawMessage
	err := msg.UnmarshalJSON([]byte(`{"ok":true}`))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, jsonRawMessage(`{"ok":true}`), msg)
}

func TestAX7_RawMessage_UnmarshalJSON_Bad(t *coretest.T) {
	var msg *jsonRawMessage
	err := msg.UnmarshalJSON([]byte(`{}`))
	coretest.AssertError(t, err)
	coretest.AssertNil(t, msg)
}

func TestAX7_RawMessage_UnmarshalJSON_Ugly(t *coretest.T) {
	msg := jsonRawMessage(`old`)
	err := msg.UnmarshalJSON([]byte(`[]`))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, jsonRawMessage(`[]`), msg)
}

func TestAX7_Value_UnmarshalJSON_Good(t *coretest.T) {
	var value jsonValue
	err := value.UnmarshalJSON([]byte(`{"n":42}`))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, jsonNumber("42"), value.value.(map[string]any)["n"])
}

func TestAX7_Value_UnmarshalJSON_Bad(t *coretest.T) {
	var value *jsonValue
	err := value.UnmarshalJSON([]byte(`true`))
	coretest.AssertError(t, err)
	coretest.AssertNil(t, value)
}

func TestAX7_Value_UnmarshalJSON_Ugly(t *coretest.T) {
	var value jsonValue
	err := value.UnmarshalJSON([]byte(``))
	coretest.AssertError(t, err)
	coretest.AssertNil(t, value.value)
}

func TestAX7_WebhookEvents_Good(t *coretest.T) {
	events := WebhookEvents()
	coretest.AssertContains(t, events, WebhookEventWorkspaceCreated)
	coretest.AssertLen(t, events, 8)
}

func TestAX7_WebhookEvents_Bad(t *coretest.T) {
	events := WebhookEvents()
	events[0] = "mutated"
	coretest.AssertEqual(t, WebhookEventWorkspaceCreated, WebhookEvents()[0])
}

func TestAX7_WebhookEvents_Ugly(t *coretest.T) {
	events := WebhookEvents()
	coretest.AssertContains(t, events, WebhookEventTicketReplied)
	coretest.AssertNotContains(t, events, "")
}

func TestAX7_IsKnownWebhookEvent_Good(t *coretest.T) {
	ok := IsKnownWebhookEvent(WebhookEventLinkClicked)
	coretest.AssertTrue(t, ok)
	coretest.AssertTrue(t, IsKnownWebhookEvent(WebhookEventTicketCreated))
}

func TestAX7_IsKnownWebhookEvent_Bad(t *coretest.T) {
	ok := IsKnownWebhookEvent("unknown")
	coretest.AssertFalse(t, ok)
	coretest.AssertFalse(t, IsKnownWebhookEvent(""))
}

func TestAX7_IsKnownWebhookEvent_Ugly(t *coretest.T) {
	ok := IsKnownWebhookEvent(" " + WebhookEventTicketCreated + " ")
	coretest.AssertTrue(t, ok)
	coretest.AssertFalse(t, IsKnownWebhookEvent("\tunknown\n"))
}

func TestAX7_NewWebhookSigner_Good(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	coretest.AssertNotNil(t, signer)
	coretest.AssertEqual(t, DefaultWebhookTolerance, signer.Tolerance())
}

func TestAX7_NewWebhookSigner_Bad(t *coretest.T) {
	signer := NewWebhookSigner("")
	sig := signer.Sign([]byte("payload"), 1)
	coretest.AssertNotEmpty(t, sig)
	coretest.AssertEqual(t, DefaultWebhookTolerance, signer.Tolerance())
}

func TestAX7_NewWebhookSigner_Ugly(t *coretest.T) {
	signer := NewWebhookSigner("line\nsecret")
	sig := signer.Sign([]byte("payload"), 1)
	coretest.AssertNotEmpty(t, sig)
	coretest.AssertEqual(t, DefaultWebhookTolerance, signer.Tolerance())
}

func TestAX7_NewWebhookSignerWithTolerance_Good(t *coretest.T) {
	signer := NewWebhookSignerWithTolerance("secret", time.Minute)
	coretest.AssertNotNil(t, signer)
	coretest.AssertEqual(t, time.Minute, signer.Tolerance())
}

func TestAX7_NewWebhookSignerWithTolerance_Bad(t *coretest.T) {
	signer := NewWebhookSignerWithTolerance("secret", 0)
	coretest.AssertEqual(t, DefaultWebhookTolerance, signer.Tolerance())
	coretest.AssertNotNil(t, signer)
}

func TestAX7_NewWebhookSignerWithTolerance_Ugly(t *coretest.T) {
	signer := NewWebhookSignerWithTolerance("secret", -time.Second)
	coretest.AssertEqual(t, DefaultWebhookTolerance, signer.Tolerance())
	coretest.AssertNotNil(t, signer)
}

func TestAX7_GenerateWebhookSecret_Good(t *coretest.T) {
	secret, err := GenerateWebhookSecret()
	coretest.RequireNoError(t, err)
	coretest.AssertLen(t, secret, 64)
}

func TestAX7_GenerateWebhookSecret_Bad(t *coretest.T) {
	secret, err := GenerateWebhookSecret()
	coretest.RequireNoError(t, err)
	coretest.AssertNotEqual(t, "", secret)
}

func TestAX7_GenerateWebhookSecret_Ugly(t *coretest.T) {
	first, err := GenerateWebhookSecret()
	coretest.RequireNoError(t, err)
	second, err := GenerateWebhookSecret()
	coretest.RequireNoError(t, err)
	coretest.AssertNotEqual(t, first, second)
}

func TestAX7_WebhookSigner_Tolerance_Good(t *coretest.T) {
	signer := NewWebhookSignerWithTolerance("secret", time.Second)
	got := signer.Tolerance()
	coretest.AssertEqual(t, time.Second, got)
}

func TestAX7_WebhookSigner_Tolerance_Bad(t *coretest.T) {
	var signer *WebhookSigner
	got := signer.Tolerance()
	coretest.AssertEqual(t, DefaultWebhookTolerance, got)
}

func TestAX7_WebhookSigner_Tolerance_Ugly(t *coretest.T) {
	signer := &WebhookSigner{tolerance: -time.Second}
	got := signer.Tolerance()
	coretest.AssertEqual(t, DefaultWebhookTolerance, got)
}

func TestAX7_WebhookSigner_Sign_Good(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	sig := signer.Sign([]byte("payload"), 123)
	coretest.AssertNotEmpty(t, sig)
	coretest.AssertLen(t, sig, 64)
}

func TestAX7_WebhookSigner_Sign_Bad(t *coretest.T) {
	var signer *WebhookSigner
	sig := signer.Sign([]byte("payload"), 123)
	coretest.AssertEqual(t, "", sig)
}

func TestAX7_WebhookSigner_Sign_Ugly(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	sig := signer.Sign(nil, -1)
	coretest.AssertNotEmpty(t, sig)
	coretest.AssertLen(t, sig, 64)
}

func TestAX7_WebhookSigner_SignNow_Good(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	sig, ts := signer.SignNow([]byte("payload"))
	coretest.AssertNotEmpty(t, sig)
	coretest.AssertTrue(t, ts > 0)
}

func TestAX7_WebhookSigner_SignNow_Bad(t *coretest.T) {
	var signer *WebhookSigner
	sig, ts := signer.SignNow([]byte("payload"))
	coretest.AssertEqual(t, "", sig)
	coretest.AssertTrue(t, ts > 0)
}

func TestAX7_WebhookSigner_SignNow_Ugly(t *coretest.T) {
	signer := NewWebhookSigner("")
	sig, ts := signer.SignNow(nil)
	coretest.AssertNotEmpty(t, sig)
	coretest.AssertTrue(t, ts > 0)
}

func TestAX7_WebhookSigner_Headers_Good(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	headers := signer.Headers([]byte("payload"))
	coretest.AssertNotEmpty(t, headers[WebhookSignatureHeader])
	coretest.AssertNotEmpty(t, headers[WebhookTimestampHeader])
}

func TestAX7_WebhookSigner_Headers_Bad(t *coretest.T) {
	var signer *WebhookSigner
	headers := signer.Headers([]byte("payload"))
	coretest.AssertEqual(t, "", headers[WebhookSignatureHeader])
	coretest.AssertNotEmpty(t, headers[WebhookTimestampHeader])
}

func TestAX7_WebhookSigner_Headers_Ugly(t *coretest.T) {
	signer := NewWebhookSigner("")
	headers := signer.Headers(nil)
	coretest.AssertNotEmpty(t, headers[WebhookSignatureHeader])
	coretest.AssertLen(t, headers, 2)
}

func TestAX7_WebhookSigner_Verify_Good(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	ts := time.Now().Unix()
	sig := signer.Sign([]byte("payload"), ts)
	coretest.AssertTrue(t, signer.Verify([]byte("payload"), sig, ts))
}

func TestAX7_WebhookSigner_Verify_Bad(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	ok := signer.Verify([]byte("payload"), "bad", time.Now().Unix())
	coretest.AssertFalse(t, ok)
}

func TestAX7_WebhookSigner_Verify_Ugly(t *coretest.T) {
	var signer *WebhookSigner
	ok := signer.Verify([]byte("payload"), "bad", time.Now().Unix())
	coretest.AssertFalse(t, ok)
}

func TestAX7_WebhookSigner_VerifySignatureOnly_Good(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	sig := signer.Sign([]byte("payload"), 1)
	coretest.AssertTrue(t, signer.VerifySignatureOnly([]byte("payload"), sig, 1))
}

func TestAX7_WebhookSigner_VerifySignatureOnly_Bad(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	ok := signer.VerifySignatureOnly([]byte("payload"), "bad", 1)
	coretest.AssertFalse(t, ok)
}

func TestAX7_WebhookSigner_VerifySignatureOnly_Ugly(t *coretest.T) {
	var signer *WebhookSigner
	ok := signer.VerifySignatureOnly(nil, "", 0)
	coretest.AssertFalse(t, ok)
}

func TestAX7_WebhookSigner_IsTimestampValid_Good(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	ok := signer.IsTimestampValid(time.Now().Unix())
	coretest.AssertTrue(t, ok)
}

func TestAX7_WebhookSigner_IsTimestampValid_Bad(t *coretest.T) {
	signer := NewWebhookSignerWithTolerance("secret", time.Second)
	ok := signer.IsTimestampValid(time.Now().Add(-time.Hour).Unix())
	coretest.AssertFalse(t, ok)
}

func TestAX7_WebhookSigner_IsTimestampValid_Ugly(t *coretest.T) {
	signer := NewWebhookSignerWithTolerance("secret", time.Second)
	ok := signer.IsTimestampValid(time.Now().Add(time.Hour).Unix())
	coretest.AssertFalse(t, ok)
}

func TestAX7_WebhookSigner_VerifyRequest_Good(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	payload := []byte("payload")
	headers := signer.Headers(payload)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(WebhookSignatureHeader, headers[WebhookSignatureHeader])
	req.Header.Set(WebhookTimestampHeader, headers[WebhookTimestampHeader])
	coretest.AssertTrue(t, signer.VerifyRequest(req, payload))
}

func TestAX7_WebhookSigner_VerifyRequest_Bad(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	ok := signer.VerifyRequest(req, []byte("payload"))
	coretest.AssertFalse(t, ok)
}

func TestAX7_WebhookSigner_VerifyRequest_Ugly(t *coretest.T) {
	signer := NewWebhookSigner("secret")
	ok := signer.VerifyRequest(nil, []byte("payload"))
	coretest.AssertFalse(t, ok)
}

func TestAX7_ValidateWebhookURL_Good(t *coretest.T) {
	err := ValidateWebhookURL("https://1.1.1.1/hook")
	coretest.AssertNoError(t, err)
	coretest.AssertNoError(t, ValidateWebhookURL("http://8.8.8.8/hook"))
}

func TestAX7_ValidateWebhookURL_Bad(t *coretest.T) {
	err := ValidateWebhookURL("http://127.0.0.1/hook")
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "private")
}

func TestAX7_ValidateWebhookURL_Ugly(t *coretest.T) {
	err := ValidateWebhookURL("ftp://1.1.1.1/hook")
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "HTTP")
}

func TestAX7_WithSunsetNoticeURL_Good(t *coretest.T) {
	cfg := &sunsetConfig{}
	WithSunsetNoticeURL(" https://example.com/notice ")(cfg)
	coretest.AssertEqual(t, " https://example.com/notice ", cfg.noticeURL)
}

func TestAX7_WithSunsetNoticeURL_Bad(t *coretest.T) {
	cfg := &sunsetConfig{noticeURL: "keep"}
	WithSunsetNoticeURL("")(cfg)
	coretest.AssertEqual(t, "", cfg.noticeURL)
}

func TestAX7_WithSunsetNoticeURL_Ugly(t *coretest.T) {
	cfg := &sunsetConfig{}
	WithSunsetNoticeURL("\t/docs\n")(cfg)
	coretest.AssertEqual(t, "\t/docs\n", cfg.noticeURL)
}

func TestAX7_ApiSunset_Good(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ApiSunset("2026-12-31", "/v2")(ctx)
	coretest.AssertEqual(t, "true", rec.Header().Get("Deprecation"))
	coretest.AssertContains(t, rec.Header().Get("Link"), "/v2")
}

func TestAX7_ApiSunset_Bad(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ApiSunset("", "")(ctx)
	coretest.AssertEqual(t, "true", rec.Header().Get("Deprecation"))
	coretest.AssertEqual(t, "", rec.Header().Get("Sunset"))
}

func TestAX7_ApiSunset_Ugly(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ApiSunset("not-a-date", " replacement ")(ctx)
	coretest.AssertEqual(t, "true", rec.Header().Get("Deprecation"))
	coretest.AssertContains(t, rec.Header().Get("API-Suggested-Replacement"), "replacement")
}

func TestAX7_ApiSunsetWith_Good(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ApiSunsetWith("2026-12-31", "/v2", WithSunsetNoticeURL("https://example.com/notice"))(ctx)
	coretest.AssertEqual(t, "true", rec.Header().Get("Deprecation"))
	coretest.AssertEqual(t, "https://example.com/notice", rec.Header().Get("API-Deprecation-Notice-URL"))
}

func TestAX7_ApiSunsetWith_Bad(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ApiSunsetWith("", "", nil)(ctx)
	coretest.AssertEqual(t, "true", rec.Header().Get("Deprecation"))
	coretest.AssertEqual(t, "", rec.Header().Get("API-Deprecation-Notice-URL"))
}

func TestAX7_ApiSunsetWith_Ugly(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ApiSunsetWith(" 2026-12-31 ", " /v2 ", WithSunsetNoticeURL(" /notice "))(ctx)
	coretest.AssertEqual(t, "true", rec.Header().Get("Deprecation"))
	coretest.AssertEqual(t, "/notice", rec.Header().Get("API-Deprecation-Notice-URL"))
}
