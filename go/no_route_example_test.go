// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	api "dappco.re/go/api"
	"github.com/gin-gonic/gin"
)

// ExampleWithNoRoute shows the canonical SPA-host pattern: every
// request whose path doesn't match a registered route is served the
// SPA's index.html, letting the client-side router take over.
func ExampleWithNoRoute() {
	gin.SetMode(gin.ReleaseMode)
	e, _ := api.New(api.WithNoRoute(func(c *gin.Context) {
		// Real SPA host calls c.File("dist/index.html"); the example
		// writes a string so the test output stays deterministic.
		c.String(http.StatusOK, "<!doctype html><html>SPA shell</html>")
	}))

	w := httptest.NewRecorder()
	e.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/anywhere/the/client/router/cares-about", nil))
	fmt.Println(w.Code, w.Body.String())
	// Output: 200 <!doctype html><html>SPA shell</html>
}
