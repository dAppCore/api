// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"strings"
	"time"

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
	graphqlPlaygroundPath   string
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
	cacheTTLValid := parsePositiveDuration(cacheTTL)

	builder := &goapi.SpecBuilder{
		Title:                   strings.TrimSpace(cfg.title),
		Summary:                 strings.TrimSpace(cfg.summary),
		Description:             strings.TrimSpace(cfg.description),
		Version:                 strings.TrimSpace(cfg.version),
		SwaggerEnabled:          swaggerPath != "",
		SwaggerPath:             swaggerPath,
		GraphQLEnabled:          graphqlPath != "" || cfg.graphqlPlayground,
		GraphQLPath:             graphqlPath,
		GraphQLPlayground:       cfg.graphqlPlayground,
		GraphQLPlaygroundPath:   strings.TrimSpace(cfg.graphqlPlaygroundPath),
		SSEEnabled:              ssePath != "",
		SSEPath:                 ssePath,
		WSEnabled:               wsPath != "",
		WSPath:                  wsPath,
		PprofEnabled:            cfg.pprofEnabled,
		ExpvarEnabled:           cfg.expvarEnabled,
		CacheEnabled:            cfg.cacheEnabled || cacheTTLValid,
		CacheTTL:                cacheTTL,
		CacheMaxEntries:         cfg.cacheMaxEntries,
		CacheMaxBytes:           cfg.cacheMaxBytes,
		I18nDefaultLocale:       strings.TrimSpace(cfg.i18nDefaultLocale),
		TermsOfService:          strings.TrimSpace(cfg.termsURL),
		ContactName:             strings.TrimSpace(cfg.contactName),
		ContactURL:              strings.TrimSpace(cfg.contactURL),
		ContactEmail:            strings.TrimSpace(cfg.contactEmail),
		Servers:                 parseServers(cfg.servers),
		LicenseName:             strings.TrimSpace(cfg.licenseName),
		LicenseURL:              strings.TrimSpace(cfg.licenseURL),
		ExternalDocsDescription: strings.TrimSpace(cfg.externalDocsDescription),
		ExternalDocsURL:         strings.TrimSpace(cfg.externalDocsURL),
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

func parsePositiveDuration(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}

	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return false
	}

	return true
}
