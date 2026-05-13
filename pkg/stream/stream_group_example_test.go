// SPDX-License-Identifier: EUPL-1.2

package stream_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	api "dappco.re/go/api"
	"dappco.re/go/api/pkg/stream"

	"github.com/gin-gonic/gin"
)

func ExampleNewGroup() {
	gin.SetMode(gin.TestMode)

	engine, _ := api.New()
	engine.RegisterStreamGroup(stream.NewGroup(
		"system",
		stream.SSE("/events", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/event-stream", []byte("data: ready\n\n"))
		}),
	))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/events", nil)
	engine.Handler().ServeHTTP(rec, req)

	_, _ = io.WriteString(os.Stdout, strings.TrimSpace(rec.Body.String()))
	// Output: data: ready
}
