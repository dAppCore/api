// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"slices"
)

// OpenAPISpecBuilder returns a SpecBuilder populated from the engine's current
// Swagger and transport metadata.
//
// Example:
//
//	builder := engine.OpenAPISpecBuilder()
func (e *Engine) OpenAPISpecBuilder() *SpecBuilder {
	if e == nil {
		return &SpecBuilder{}
	}

	transport := e.TransportConfig()
	builder := &SpecBuilder{
		Title:                   e.swaggerTitle,
		Description:             e.swaggerDesc,
		Version:                 e.swaggerVersion,
		TermsOfService:          e.swaggerTermsOfService,
		ContactName:             e.swaggerContactName,
		ContactURL:              e.swaggerContactURL,
		ContactEmail:            e.swaggerContactEmail,
		Servers:                 slices.Clone(e.swaggerServers),
		LicenseName:             e.swaggerLicenseName,
		LicenseURL:              e.swaggerLicenseURL,
		ExternalDocsDescription: e.swaggerExternalDocsDescription,
		ExternalDocsURL:         e.swaggerExternalDocsURL,
	}

	builder.SwaggerPath = transport.SwaggerPath
	builder.GraphQLPath = transport.GraphQLPath
	builder.GraphQLPlayground = transport.GraphQLPlayground
	builder.WSPath = transport.WSPath
	builder.SSEPath = transport.SSEPath
	builder.PprofEnabled = transport.PprofEnabled
	builder.ExpvarEnabled = transport.ExpvarEnabled

	return builder
}
