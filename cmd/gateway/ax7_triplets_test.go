// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"net/http"
	"net/http/httptest"

	. "dappco.re/go"

	"github.com/gin-gonic/gin"
)

func TestAX7_RouteGroup_Name_Good(t *T) {
	group := brainRouteGroup{name: "brain", basePath: "/api/brain"}
	name := group.Name()
	AssertEqual(t, "brain", name)
	AssertNotEmpty(t, name)
}

func TestAX7_RouteGroup_Name_Bad(t *T) {
	group := brainRouteGroup{}
	name := group.Name()
	AssertEqual(t, "", name)
	AssertEmpty(t, name)
}

func TestAX7_RouteGroup_Name_Ugly(t *T) {
	group := &proxyRouteGroup{}
	name := group.Name()
	AssertEqual(t, "proxy", name)
	AssertNotEqual(t, "brain", name)
}

func TestAX7_RouteGroup_BasePath_Good(t *T) {
	group := brainRouteGroup{name: "brain", basePath: "/api/brain"}
	path := group.BasePath()
	AssertEqual(t, "/api/brain", path)
	AssertTrue(t, HasPrefix(path, "/"))
}

func TestAX7_RouteGroup_BasePath_Bad(t *T) {
	group := minerRouteGroup{}
	path := group.BasePath()
	AssertEqual(t, "", path)
	AssertEmpty(t, path)
}

func TestAX7_RouteGroup_BasePath_Ugly(t *T) {
	group := buildRouteGroup{projectDir: "."}
	path := group.BasePath()
	AssertEqual(t, "/api/v1/build", path)
	AssertContains(t, path, "build")
}

func TestAX7_RouteGroup_Channels_Good(t *T) {
	group := buildRouteGroup{projectDir: "."}
	channels := group.Channels()
	AssertLen(t, channels, 7)
	AssertContains(t, channels, "build.complete")
}

func TestAX7_RouteGroup_Channels_Bad(t *T) {
	group := brainRouteGroup{name: "", basePath: ""}
	channels := group.Channels()
	AssertLen(t, channels, 4)
	AssertContains(t, channels, "brain.recall.complete")
}

func TestAX7_RouteGroup_Channels_Ugly(t *T) {
	group := brainRouteGroup{name: "brain-mcp", basePath: "/mcp"}
	channels := group.Channels()
	AssertLen(t, channels, 4)
	AssertContains(t, channels[0], "brain.mcp")
}

func TestAX7_RouteGroup_HandleFunc_Good(t *T) {
	group := &proxyRouteGroup{}
	group.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	AssertLen(t, group.handlers, 1)
	AssertEqual(t, "/health", group.handlers[0].path)
}

func TestAX7_RouteGroup_HandleFunc_Bad(t *T) {
	group := &proxyRouteGroup{}
	group.HandleFunc("", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	AssertLen(t, group.handlers, 0)
	AssertEmpty(t, group.handlers)
}

func TestAX7_RouteGroup_HandleFunc_Ugly(t *T) {
	group := &proxyRouteGroup{}
	group.HandleFunc("/nil", nil)
	AssertLen(t, group.handlers, 0)
	AssertEmpty(t, group.Describe())
}

func TestAX7_RouteGroup_RegisterRoutes_Good(t *T) {
	gin.SetMode(gin.TestMode)
	group := &proxyRouteGroup{}
	group.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusAccepted) })
	router := gin.New()
	group.RegisterRoutes(&router.RouterGroup)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	AssertEqual(t, http.StatusAccepted, rec.Code)
}

func TestAX7_RouteGroup_RegisterRoutes_Bad(t *T) {
	var group *proxyRouteGroup
	router := gin.New()
	AssertNotPanics(t, func() {
		group.RegisterRoutes(&router.RouterGroup)
	})
	AssertNil(t, group)
}

func TestAX7_RouteGroup_RegisterRoutes_Ugly(t *T) {
	gin.SetMode(gin.TestMode)
	group := buildRouteGroup{projectDir: "/tmp/project"}
	router := gin.New()
	group.RegisterRoutes(&router.RouterGroup)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/config", nil))
	AssertEqual(t, http.StatusNotImplemented, rec.Code)
}

func TestAX7_RouteGroup_Describe_Good(t *T) {
	group := buildRouteGroup{projectDir: "."}
	descriptions := group.Describe()
	AssertLen(t, descriptions, 13)
	AssertEqual(t, "/config", descriptions[0].Path)
}

func TestAX7_RouteGroup_Describe_Bad(t *T) {
	group := minerRouteGroup{}
	descriptions := group.Describe()
	AssertNil(t, descriptions)
	AssertEmpty(t, descriptions)
}

func TestAX7_RouteGroup_Describe_Ugly(t *T) {
	group := &proxyRouteGroup{}
	group.HandleFunc("/metrics", func(http.ResponseWriter, *http.Request) {})
	descriptions := group.Describe()
	AssertLen(t, descriptions, 1)
	AssertEqual(t, "GET", descriptions[0].Method)
}
