// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// bearerAuthMiddleware validates the Authorization: Bearer <token> header.
// Requests to paths in the skip list are allowed through without authentication.
// Returns 401 with Fail("unauthorised", ...) on missing or invalid tokens.
func bearerAuthMiddleware(token string, skip []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check whether the request path should bypass authentication.
		for _, path := range skip {
			if strings.HasPrefix(c.Request.URL.Path, path) {
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

// requestIDMiddleware ensures every response carries an X-Request-ID header.
// If the client sends one, it is preserved; otherwise a random 16-byte hex
// string is generated. The ID is also stored in the Gin context as "request_id".
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}

		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
