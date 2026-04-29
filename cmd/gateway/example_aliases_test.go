// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"net/http"

	coreapi "dappco.re/go/api"
	"github.com/gin-gonic/gin"
)

type RouteGroup struct{}

func (RouteGroup) Name() string {
	return ""
}

func (RouteGroup) BasePath() string {
	return ""
}

func (RouteGroup) Channels() []string {
	return nil
}

func (RouteGroup) RegisterRoutes(*gin.RouterGroup) {}

func (RouteGroup) Describe() []coreapi.RouteDescription {
	return nil
}

func (RouteGroup) HandleFunc(string, func(http.ResponseWriter, *http.Request)) {}
