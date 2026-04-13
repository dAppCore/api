// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"strings"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

// defaultWSPath is the URL path where the WebSocket endpoint is mounted.
const defaultWSPath = "/ws"

// wrapWSHandler adapts a standard http.Handler to a Gin handler for the WebSocket route.
// The underlying handler is responsible for upgrading the connection to WebSocket.
func wrapWSHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// normaliseWSPath coerces custom WebSocket paths into a stable form.
// The path always begins with a single slash and never ends with one.
func normaliseWSPath(path string) string {
	path = core.Trim(path)
	if path == "" {
		return defaultWSPath
	}

	path = "/" + strings.Trim(path, "/")
	if path == "/" {
		return defaultWSPath
	}

	return path
}

// resolveWSPath returns the configured WebSocket path or the default path
// when no override has been provided.
func resolveWSPath(path string) string {
	if core.Trim(path) == "" {
		return defaultWSPath
	}
	return normaliseWSPath(path)
}
