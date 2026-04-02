// SPDX-License-Identifier: EUPL-1.2

package api

import "slices"

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

	if e.graphql != nil {
		builder.GraphQLPath = e.graphql.path
		builder.GraphQLPlayground = e.graphql.playground
	}
	if e.sseBroker != nil {
		builder.SSEPath = resolveSSEPath(e.ssePath)
	}
	builder.PprofEnabled = e.pprofEnabled
	builder.ExpvarEnabled = e.expvarEnabled

	return builder
}
