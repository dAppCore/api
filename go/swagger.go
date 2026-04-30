// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http" // Note: AX-6 - structural HTTP status boundary for Gin handlers; no core primitive.

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
	"slices"
)

// swaggerSeq provides unique instance names so multiple Engine instances
// (common in tests) do not collide in the global swag registry.
var swaggerSeq core.AtomicUint64

// defaultSwaggerPath is the URL path where the Swagger UI is mounted.
const defaultSwaggerPath = "/swagger"

// swaggerSpec wraps SpecBuilder to satisfy the swag.Spec interface.
// The spec is built once on first access and cached.
type swaggerSpec struct {
	builder *SpecBuilder
	groups  []RouteGroup
	once    core.Once
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

	path = "/" + trimSlashes(path)
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

// defaultOpenAPISpecPath is the URL path where the raw OpenAPI 3.1 JSON
// document is served per RFC.endpoints.md — "GET /v1/openapi.json".
const defaultOpenAPISpecPath = "/v1/openapi.json"

// registerOpenAPISpec mounts a GET handler at the configured spec path that
// serves the generated OpenAPI 3.1 JSON document. The document is built once
// and reused for every subsequent request so callers pay the generation cost
// a single time.
//
//	registerOpenAPISpec(r, engine)
//	// GET /v1/openapi.json -> application/json openapi document
func registerOpenAPISpec(g *gin.Engine, e *Engine) {
	path := resolveOpenAPISpecPath(e.openAPISpecPath)
	spec := newSwaggerSpec(e.OpenAPISpecBuilder(), e.Groups())
	g.GET(path, func(c *gin.Context) {
		doc := spec.ReadDoc()
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.String(http.StatusOK, doc)
	})
}

// normaliseOpenAPISpecPath coerces custom spec URL overrides into a stable
// form. The returned path always begins with a single slash and never ends
// with one, matching the shape of the other transport path helpers.
//
//	normaliseOpenAPISpecPath("openapi.json") // "/openapi.json"
func normaliseOpenAPISpecPath(path string) string {
	path = core.Trim(path)
	if path == "" {
		return defaultOpenAPISpecPath
	}

	path = "/" + trimSlashes(path)
	if path == "/" {
		return defaultOpenAPISpecPath
	}

	return path
}

// resolveOpenAPISpecPath returns the configured OpenAPI spec URL or the
// RFC default when no override is provided.
//
//	resolveOpenAPISpecPath("") // "/v1/openapi.json"
func resolveOpenAPISpecPath(path string) string {
	if core.Trim(path) == "" {
		return defaultOpenAPISpecPath
	}
	return normaliseOpenAPISpecPath(path)
}
