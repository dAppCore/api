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
	authentikIssuer         string
	authentikClientID       string
	authentikTrustedProxy   bool
	authentikPublicPaths    string
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
	swaggerPath := strings.TrimSpace(cfg.swaggerPath)
	graphqlPath := strings.TrimSpace(cfg.graphqlPath)
	ssePath := strings.TrimSpace(cfg.ssePath)
	wsPath := strings.TrimSpace(cfg.wsPath)
	cacheTTL := strings.TrimSpace(cfg.cacheTTL)

	builder := &goapi.SpecBuilder{
		Title:                   cfg.title,
		Summary:                 cfg.summary,
		Description:             cfg.description,
		Version:                 cfg.version,
		SwaggerEnabled:          swaggerPath != "",
		SwaggerPath:             swaggerPath,
		GraphQLEnabled:          graphqlPath != "" || cfg.graphqlPlayground,
		GraphQLPath:             graphqlPath,
		GraphQLPlayground:       cfg.graphqlPlayground,
		SSEEnabled:              ssePath != "",
		SSEPath:                 ssePath,
		WSEnabled:               wsPath != "",
		WSPath:                  wsPath,
		PprofEnabled:            cfg.pprofEnabled,
		ExpvarEnabled:           cfg.expvarEnabled,
		CacheEnabled:            cfg.cacheEnabled || cacheTTL != "" || cfg.cacheMaxEntries > 0 || cfg.cacheMaxBytes > 0,
		CacheTTL:                cacheTTL,
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
		AuthentikIssuer:         strings.TrimSpace(cfg.authentikIssuer),
		AuthentikClientID:       strings.TrimSpace(cfg.authentikClientID),
		AuthentikTrustedProxy:   cfg.authentikTrustedProxy,
		AuthentikPublicPaths:    normalisePublicPaths(splitUniqueCSV(cfg.authentikPublicPaths)),
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
