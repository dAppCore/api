// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// requestIDContextKey is the Gin context key used by requestIDMiddleware.
const requestIDContextKey = "request_id"

// requestStartContextKey stores when the request began so handlers can
// calculate elapsed duration for response metadata.
const requestStartContextKey = "request_start"

// recoveryMiddleware converts panics into a standard JSON error envelope.
// This keeps internal failures consistent with the rest of the framework
// and avoids Gin's default plain-text 500 response.
func recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		fmt.Fprintf(gin.DefaultErrorWriter, "[Recovery] panic recovered: %v\n", recovered)
		debug.PrintStack()
		c.AbortWithStatusJSON(http.StatusInternalServerError, Fail(
			"internal_server_error",
			"Internal server error",
		))
	})
}

// bearerAuthMiddleware validates the Authorization: Bearer <token> header.
// Requests to paths in the skip list are allowed through without authentication.
// Returns 401 with Fail("unauthorised", ...) on missing or invalid tokens.
func bearerAuthMiddleware(token string, skip func() []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check whether the request path should bypass authentication.
		for _, path := range skip() {
			if isPublicPath(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, Fail("unauthorised", "missing authorization header"))
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, Fail("unauthorised", "invalid bearer token"))
			return
		}

		c.Next()
	}
}

// isPublicPath reports whether requestPath should bypass auth for publicPath.
// It matches the exact path and any nested subpath, but not sibling prefixes
// such as /swaggerx when the public path is /swagger.
func isPublicPath(requestPath, publicPath string) bool {
	if publicPath == "" {
		return false
	}

	normalized := strings.TrimRight(publicPath, "/")
	if normalized == "" {
		normalized = "/"
	}

	if requestPath == normalized {
		return true
	}

	if normalized == "/" {
		return true
	}

	return strings.HasPrefix(requestPath, normalized+"/")
}

// requestIDMiddleware ensures every response carries an X-Request-ID header.
// If the client sends one, it is preserved; otherwise a random 16-byte hex
// string is generated. The ID is also stored in the Gin context as "request_id".
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(requestStartContextKey, time.Now())

		id := c.GetHeader("X-Request-ID")
		if id == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}

		c.Set(requestIDContextKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// GetRequestID returns the request ID assigned by requestIDMiddleware.
// Returns an empty string when the middleware was not applied.
//
// Example:
//
//	id := api.GetRequestID(c)
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(requestIDContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetRequestDuration returns the elapsed time since requestIDMiddleware started
// handling the request. Returns 0 when the middleware was not applied.
//
// Example:
//
//	d := api.GetRequestDuration(c)
func GetRequestDuration(c *gin.Context) time.Duration {
	if v, ok := c.Get(requestStartContextKey); ok {
		if started, ok := v.(time.Time); ok && !started.IsZero() {
			return time.Since(started)
		}
	}
	return 0
}

// GetRequestMeta returns request metadata collected by requestIDMiddleware.
// The returned meta includes the request ID and elapsed duration when
// available. It returns nil when neither value is available.
//
// Example:
//
//	meta := api.GetRequestMeta(c)
func GetRequestMeta(c *gin.Context) *Meta {
	meta := &Meta{}

	if id := GetRequestID(c); id != "" {
		meta.RequestID = id
	}

	if duration := GetRequestDuration(c); duration > 0 {
		meta.Duration = duration.String()
	}

	if meta.RequestID == "" && meta.Duration == "" {
		return nil
	}

	return meta
}
