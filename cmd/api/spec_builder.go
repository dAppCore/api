// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"time"

	core "dappco.re/go/core"

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
	openAPISpecEnabled      bool
	openAPISpecPath         string
	chatCompletionsEnabled  bool
	chatCompletionsPath     string
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
	swaggerPath := core.Trim(cfg.swaggerPath)
	graphqlPath := core.Trim(cfg.graphqlPath)
	ssePath := core.Trim(cfg.ssePath)
	wsPath := core.Trim(cfg.wsPath)
	cacheTTL := core.Trim(cfg.cacheTTL)
	cacheTTLValid := parsePositiveDuration(cacheTTL)

	openAPISpecPath := core.Trim(cfg.openAPISpecPath)
	chatCompletionsPath := core.Trim(cfg.chatCompletionsPath)
	builder := &goapi.SpecBuilder{
		Title:                   core.Trim(cfg.title),
		Summary:                 core.Trim(cfg.summary),
		Description:             core.Trim(cfg.description),
		Version:                 core.Trim(cfg.version),
		SwaggerEnabled:          swaggerPath != "",
		SwaggerPath:             swaggerPath,
		GraphQLEnabled:          graphqlPath != "" || cfg.graphqlPlayground,
		GraphQLPath:             graphqlPath,
		GraphQLPlayground:       cfg.graphqlPlayground,
		GraphQLPlaygroundPath:   core.Trim(cfg.graphqlPlaygroundPath),
		SSEEnabled:              ssePath != "",
		SSEPath:                 ssePath,
		WSEnabled:               wsPath != "",
		WSPath:                  wsPath,
		PprofEnabled:            cfg.pprofEnabled,
		ExpvarEnabled:           cfg.expvarEnabled,
		ChatCompletionsEnabled:  cfg.chatCompletionsEnabled || chatCompletionsPath != "",
		ChatCompletionsPath:     chatCompletionsPath,
		OpenAPISpecEnabled:      cfg.openAPISpecEnabled || openAPISpecPath != "",
		OpenAPISpecPath:         openAPISpecPath,
		CacheEnabled:            cfg.cacheEnabled || cacheTTLValid,
		CacheTTL:                cacheTTL,
		CacheMaxEntries:         cfg.cacheMaxEntries,
		CacheMaxBytes:           cfg.cacheMaxBytes,
		I18nDefaultLocale:       core.Trim(cfg.i18nDefaultLocale),
		TermsOfService:          core.Trim(cfg.termsURL),
		ContactName:             core.Trim(cfg.contactName),
		ContactURL:              core.Trim(cfg.contactURL),
		ContactEmail:            core.Trim(cfg.contactEmail),
		Servers:                 parseServers(cfg.servers),
		LicenseName:             core.Trim(cfg.licenseName),
		LicenseURL:              core.Trim(cfg.licenseURL),
		ExternalDocsDescription: core.Trim(cfg.externalDocsDescription),
		ExternalDocsURL:         core.Trim(cfg.externalDocsURL),
		AuthentikIssuer:         core.Trim(cfg.authentikIssuer),
		AuthentikClientID:       core.Trim(cfg.authentikClientID),
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
	raw = core.Trim(raw)
	if raw == "" {
		return false
	}

	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return false
	}

	return true
}
