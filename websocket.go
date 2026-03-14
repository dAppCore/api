// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// wrapWSHandler adapts a standard http.Handler to a Gin handler for the /ws route.
// The underlying handler is responsible for upgrading the connection to WebSocket.
func wrapWSHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
