// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ApiSunset returns middleware that marks a route or group as deprecated.
//
// The middleware adds standard deprecation headers to every response:
// Deprecation, optional Sunset, optional Link, and X-API-Warn.
//
// Example:
//
//	rg.Use(api.ApiSunset("2025-06-01", "/api/v2/users"))
func ApiSunset(sunsetDate, replacement string) gin.HandlerFunc {
	sunsetDate = strings.TrimSpace(sunsetDate)
	replacement = strings.TrimSpace(replacement)
	formatted := formatSunsetDate(sunsetDate)
	warning := "This endpoint is deprecated."
	if sunsetDate != "" {
		warning = "This endpoint is deprecated and will be removed on " + sunsetDate + "."
	}

	return func(c *gin.Context) {
		c.Next()

		c.Header("Deprecation", "true")
		if formatted != "" {
			c.Header("Sunset", formatted)
		}
		if replacement != "" {
			c.Writer.Header().Add("Link", "<"+replacement+">; rel=\"successor-version\"")
		}
		c.Header("X-API-Warn", warning)
	}
}

func formatSunsetDate(sunsetDate string) string {
	sunsetDate = strings.TrimSpace(sunsetDate)
	if sunsetDate == "" {
		return ""
	}
	if strings.Contains(sunsetDate, ",") {
		return sunsetDate
	}

	parsed, err := time.Parse("2006-01-02", sunsetDate)
	if err != nil {
		return sunsetDate
	}

	return parsed.UTC().Format(http.TimeFormat)
}
