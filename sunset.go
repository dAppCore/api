// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"strings"
	"time"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

// SunsetOption customises the behaviour of ApiSunsetWith. Use the supplied
// constructors (e.g. WithSunsetNoticeURL) to compose the desired metadata
// without breaking the simpler ApiSunset signature.
//
//	mw := api.ApiSunsetWith("2025-06-01", "/api/v2/users",
//	    api.WithSunsetNoticeURL("https://docs.example.com/deprecation/billing"),
//	)
type SunsetOption func(*sunsetConfig)

// sunsetConfig carries optional metadata for ApiSunsetWith.
type sunsetConfig struct {
	noticeURL string
}

// WithSunsetNoticeURL adds the API-Deprecation-Notice-URL header documented
// in spec §8 to every response. The URL should point to a human-readable
// migration guide for the deprecated endpoint.
//
//	api.ApiSunsetWith("2026-04-30", "POST /api/v2/billing/invoices",
//	    api.WithSunsetNoticeURL("https://docs.api.dappco.re/deprecation/billing"),
//	)
func WithSunsetNoticeURL(url string) SunsetOption {
	return func(cfg *sunsetConfig) {
		cfg.noticeURL = url
	}
}

// ApiSunset returns middleware that marks a route or group as deprecated.
//
// The middleware appends standard deprecation headers to every response:
// Deprecation, optional Sunset, optional Link, optional API-Suggested-Replacement,
// and X-API-Warn. Existing header values are preserved so downstream middleware
// and handlers can keep their own link relations or warning metadata.
//
// Example:
//
//	rg.Use(api.ApiSunset("2025-06-01", "/api/v2/users"))
func ApiSunset(sunsetDate, replacement string) gin.HandlerFunc {
	return ApiSunsetWith(sunsetDate, replacement)
}

// ApiSunsetWith is the extensible form of ApiSunset. It accepts SunsetOption
// values to attach optional metadata such as the deprecation notice URL.
//
// Example:
//
//	rg.Use(api.ApiSunsetWith(
//	    "2026-04-30",
//	    "POST /api/v2/billing/invoices",
//	    api.WithSunsetNoticeURL("https://docs.api.dappco.re/deprecation/billing"),
//	))
func ApiSunsetWith(sunsetDate, replacement string, opts ...SunsetOption) gin.HandlerFunc {
	sunsetDate = core.Trim(sunsetDate)
	replacement = core.Trim(replacement)
	formatted := formatSunsetDate(sunsetDate)
	warning := "This endpoint is deprecated."
	if sunsetDate != "" {
		warning = "This endpoint is deprecated and will be removed on " + sunsetDate + "."
	}

	cfg := &sunsetConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	noticeURL := core.Trim(cfg.noticeURL)

	return func(c *gin.Context) {
		c.Next()

		c.Writer.Header().Add("Deprecation", "true")
		if formatted != "" {
			c.Writer.Header().Add("Sunset", formatted)
		}
		if replacement != "" {
			linkTarget := successorLinkTarget(replacement)
			c.Writer.Header().Add("Link", "<"+linkTarget+">; rel=\"successor-version\"")
			c.Writer.Header().Add("API-Suggested-Replacement", replacement)
		}
		if noticeURL != "" {
			c.Writer.Header().Add("API-Deprecation-Notice-URL", noticeURL)
		}
		c.Writer.Header().Add("X-API-Warn", warning)
	}
}

func successorLinkTarget(replacement string) string {
	replacement = core.Trim(replacement)
	if replacement == "" {
		return ""
	}

	fields := strings.Fields(replacement)
	if len(fields) >= 2 {
		switch strings.ToUpper(fields[0]) {
		case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE", "CONNECT":
			if target := strings.TrimSpace(fields[1]); target != "" {
				return target
			}
		}
	}

	return replacement
}

func formatSunsetDate(sunsetDate string) string {
	sunsetDate = core.Trim(sunsetDate)
	if sunsetDate == "" {
		return ""
	}

	if parsed, err := time.Parse(time.RFC3339, sunsetDate); err == nil {
		return parsed.UTC().Format(http.TimeFormat)
	}

	if parsed, err := time.Parse("2006-01-02", sunsetDate); err == nil {
		return parsed.UTC().Format(http.TimeFormat)
	}

	if parsed, err := http.ParseTime(sunsetDate); err == nil {
		return parsed.UTC().Format(http.TimeFormat)
	}

	return sunsetDate
}
