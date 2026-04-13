// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
	"slices"
)

// swaggerSeq provides unique instance names so multiple Engine instances
// (common in tests) do not collide in the global swag registry.
var swaggerSeq atomic.Uint64

// defaultSwaggerPath is the URL path where the Swagger UI is mounted.
const defaultSwaggerPath = "/swagger"

// swaggerSpec wraps SpecBuilder to satisfy the swag.Spec interface.
// The spec is built once on first access and cached.
type swaggerSpec struct {
	builder *SpecBuilder
	groups  []RouteGroup
	once    sync.Once
	doc     string
}

var _ swag.Swagger = (*swaggerSpec)(nil)

func newSwaggerSpec(builder *SpecBuilder, groups []RouteGroup) *swaggerSpec {
	return &swaggerSpec{
		builder: builder,
		groups:  slices.Clone(groups),
	}
}

// ReadDoc returns the OpenAPI 3.1 JSON document for this spec.
func (s *swaggerSpec) ReadDoc() string {
	s.once.Do(func() {
		data, err := s.builder.Build(s.groups)
		if err != nil {
			s.doc = `{"openapi":"3.1.0","info":{"title":"error","version":"0.0.0"},"paths":{}}`
			return
		}
		s.doc = string(data)
	})
	return s.doc
}

// registerSwagger mounts the Swagger UI and doc.json endpoint.
func registerSwagger(g *gin.Engine, e *Engine, groups []RouteGroup) {
	swaggerPath := resolveSwaggerPath(e.swaggerPath)
	spec := newSwaggerSpec(e.OpenAPISpecBuilder(), groups)
	name := core.Sprintf("swagger_%d", swaggerSeq.Add(1))
	swag.Register(name, spec)
	g.GET(swaggerPath, func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, swaggerPath+"/")
	})
	g.GET(swaggerPath+"/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName(name)))
}

// normaliseSwaggerPath coerces custom Swagger paths into a stable form.
// The path always begins with a single slash and never ends with one.
func normaliseSwaggerPath(path string) string {
	path = core.Trim(path)
	if path == "" {
		return defaultSwaggerPath
	}

	path = "/" + strings.Trim(path, "/")
	if path == "/" {
		return defaultSwaggerPath
	}

	return path
}

// resolveSwaggerPath returns the configured Swagger path or the default path
// when no override has been provided.
func resolveSwaggerPath(path string) string {
	if core.Trim(path) == "" {
		return defaultSwaggerPath
	}
	return normaliseSwaggerPath(path)
}
