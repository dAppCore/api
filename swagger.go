// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

// swaggerSeq provides unique instance names so multiple Engine instances
// (common in tests) do not collide in the global swag registry.
var swaggerSeq atomic.Uint64

// swaggerSpec wraps SpecBuilder to satisfy the swag.Spec interface.
// The spec is built once on first access and cached.
type swaggerSpec struct {
	builder *SpecBuilder
	groups  []RouteGroup
	once    sync.Once
	doc     string
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
func registerSwagger(g *gin.Engine, title, description, version, contactName, contactURL, contactEmail string, servers []string, licenseName, licenseURL string, groups []RouteGroup) {
	spec := &swaggerSpec{
		builder: &SpecBuilder{
			Title:        title,
			Description:  description,
			Version:      version,
			ContactName:  contactName,
			ContactURL:   contactURL,
			ContactEmail: contactEmail,
			Servers:      servers,
			LicenseName:  licenseName,
			LicenseURL:   licenseURL,
		},
		groups: groups,
	}
	name := fmt.Sprintf("swagger_%d", swaggerSeq.Add(1))
	swag.Register(name, spec)
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName(name)))
}
