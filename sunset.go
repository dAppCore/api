// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"time"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

// ApiSunset returns middleware that marks a route or group as deprecated.
//
// The middleware appends standard deprecation headers to every response:
// Deprecation, optional Sunset, optional Link, and X-API-Warn. Existing header
// values are preserved so downstream middleware and handlers can keep their own
// link relations or warning metadata.
//
// Example:
//
//	rg.Use(api.ApiSunset("2025-06-01", "/api/v2/users"))
func ApiSunset(sunsetDate, replacement string) gin.HandlerFunc {
	sunsetDate = core.Trim(sunsetDate)
	replacement = core.Trim(replacement)
	formatted := formatSunsetDate(sunsetDate)
	warning := "This endpoint is deprecated."
	if sunsetDate != "" {
		warning = "This endpoint is deprecated and will be removed on " + sunsetDate + "."
	}

	return func(c *gin.Context) {
		c.Next()

		c.Writer.Header().Add("Deprecation", "true")
		if formatted != "" {
			c.Writer.Header().Add("Sunset", formatted)
		}
		if replacement != "" {
			c.Writer.Header().Add("Link", "<"+replacement+">; rel=\"successor-version\"")
		}
		c.Writer.Header().Add("X-API-Warn", warning)
	}
}

func formatSunsetDate(sunsetDate string) string {
	sunsetDate = core.Trim(sunsetDate)
	if sunsetDate == "" {
		return ""
	}
	if core.Contains(sunsetDate, ",") {
		return sunsetDate
	}

	parsed, err := time.Parse("2006-01-02", sunsetDate)
	if err != nil {
		return sunsetDate
	}

	return parsed.UTC().Format(http.TimeFormat)
}
