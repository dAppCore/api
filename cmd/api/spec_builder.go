// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"strings"

	goapi "dappco.re/go/core/api"
)

type specBuilderConfig struct {
	title                   string
	summary                 string
	description             string
	version                 string
	swaggerPath             string
	graphqlPath             string
	graphqlPlayground       bool
	ssePath                 string
	wsPath                  string
	pprofEnabled            bool
	expvarEnabled           bool
	cacheEnabled            bool
	cacheTTL                string
	cacheMaxEntries         int
	cacheMaxBytes           int
	i18nDefaultLocale       string
	i18nSupportedLocales    string
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
		Summary:                 cfg.summary,
		Description:             cfg.description,
		Version:                 cfg.version,
		SwaggerPath:             cfg.swaggerPath,
		GraphQLPath:             cfg.graphqlPath,
		GraphQLPlayground:       cfg.graphqlPlayground,
		SSEPath:                 cfg.ssePath,
		WSPath:                  cfg.wsPath,
		PprofEnabled:            cfg.pprofEnabled,
		ExpvarEnabled:           cfg.expvarEnabled,
		CacheEnabled:            cfg.cacheEnabled || strings.TrimSpace(cfg.cacheTTL) != "" || cfg.cacheMaxEntries > 0 || cfg.cacheMaxBytes > 0,
		CacheTTL:                strings.TrimSpace(cfg.cacheTTL),
		CacheMaxEntries:         cfg.cacheMaxEntries,
		CacheMaxBytes:           cfg.cacheMaxBytes,
		I18nDefaultLocale:       strings.TrimSpace(cfg.i18nDefaultLocale),
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

	builder.I18nSupportedLocales = parseLocales(cfg.i18nSupportedLocales)
	if builder.I18nDefaultLocale == "" && len(builder.I18nSupportedLocales) > 0 {
		builder.I18nDefaultLocale = "en"
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

func parseLocales(raw string) []string {
	return splitUniqueCSV(raw)
}
