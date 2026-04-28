// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http" // Note: AX-6 — structural HTTP status boundary for Gin handlers; no core primitive.

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
)

const defaultSDKGenPath = "/v1/sdk/generate"

// SDKGenRequest is the request body for POST /v1/sdk/generate.
type SDKGenRequest struct {
	Language string `json:"language"`
}

// SDKGenResponse is the successful response body for generated SDK artifacts.
type SDKGenResponse struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

func mountSDKGen(r *gin.Engine) {
	r.POST(defaultSDKGenPath, func(c *gin.Context) {
		var req SDKGenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, Fail("invalid_request", "Invalid SDK generation request"))
			return
		}

		language := normaliseSDKGenLanguage(req.Language)
		if language == "" {
			c.JSON(http.StatusBadRequest, FailWithDetails(
				"unsupported_sdk_language",
				"Unsupported SDK language",
				map[string]any{"supported": []string{"go", "php", "ts"}},
			))
			return
		}

		c.JSON(http.StatusNotImplemented, FailWithDetails(
			"sdk_generation_unavailable",
			"SDK generation endpoint is registered but no artifact backend is configured",
			map[string]any{"language": language},
		))
	})
}

func normaliseSDKGenLanguage(language string) string {
	switch core.Lower(core.Trim(language)) {
	case "go":
		return "go"
	case "php":
		return "php"
	case "ts", "typescript":
		return "ts"
	default:
		return ""
	}
}
