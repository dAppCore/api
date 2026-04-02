// SPDX-License-Identifier: EUPL-1.2

package api

import goapi "dappco.re/go/core/api"

type specBuilderConfig struct {
	title                   string
	description             string
	version                 string
	swaggerPath             string
	graphqlPath             string
	graphqlPlayground       bool
	ssePath                 string
	wsPath                  string
	pprofEnabled            bool
	expvarEnabled           bool
	termsURL                string
	contactName             string
	contactURL              string
	contactEmail            string
	licenseName             string
	licenseURL              string
	externalDocsDescription string
	externalDocsURL         string
	servers                 string
	securitySchemes         string
}

func newSpecBuilder(cfg specBuilderConfig) (*goapi.SpecBuilder, error) {
	builder := &goapi.SpecBuilder{
		Title:                   cfg.title,
		Description:             cfg.description,
		Version:                 cfg.version,
		SwaggerPath:             cfg.swaggerPath,
		GraphQLPath:             cfg.graphqlPath,
		GraphQLPlayground:       cfg.graphqlPlayground,
		SSEPath:                 cfg.ssePath,
		WSPath:                  cfg.wsPath,
		PprofEnabled:            cfg.pprofEnabled,
		ExpvarEnabled:           cfg.expvarEnabled,
		TermsOfService:          cfg.termsURL,
		ContactName:             cfg.contactName,
		ContactURL:              cfg.contactURL,
		ContactEmail:            cfg.contactEmail,
		Servers:                 parseServers(cfg.servers),
		LicenseName:             cfg.licenseName,
		LicenseURL:              cfg.licenseURL,
		ExternalDocsDescription: cfg.externalDocsDescription,
		ExternalDocsURL:         cfg.externalDocsURL,
	}

	if cfg.securitySchemes != "" {
		schemes, err := parseSecuritySchemes(cfg.securitySchemes)
		if err != nil {
			return nil, err
		}
		builder.SecuritySchemes = schemes
	}

	return builder, nil
}
