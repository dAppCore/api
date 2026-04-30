// SPDX-License-Identifier: EUPL-1.2

package provider_test

import (
	"net/http"
	"net/http/httptest"

	core "dappco.re/go"
	"github.com/gin-gonic/gin"
)

func ExampleRegistry_MountAll_cacheControl() {
	gin.SetMode(gin.TestMode)

	handler := mountProviderHandler(&cacheControlProvider{
		basePath:         "/api/cache",
		withDescriptions: true,
	})

	cacheable := httptest.NewRecorder()
	handler.ServeHTTP(cacheable, httptest.NewRequest(http.MethodGet, "/api/cache/items/123", nil))
	core.Println(cacheable.Header().Get("Cache-Control"))

	ephemeral := httptest.NewRecorder()
	handler.ServeHTTP(ephemeral, httptest.NewRequest(http.MethodPost, "/api/cache/sessions", nil))
	core.Println(ephemeral.Header().Get("Cache-Control"))

	// Output:
	// public, max-age=300
	// no-store
}
