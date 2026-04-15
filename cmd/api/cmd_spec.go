// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"os"

	core "dappco.re/go/core"
	"dappco.re/go/core/cli/pkg/cli"

	goapi "dappco.re/go/core/api"
)

const defaultSpecToolBridgePath = "/v1/tools"

func addSpecCommand(c *core.Core) {
	c.Command("api/spec", core.Command{
		Description: "Generate OpenAPI specification",
		Action:      specAction,
	})
}

func specAction(opts core.Options) core.Result {
	cfg := specConfigFromOptions(opts)
	output := opts.String("output")
	format := opts.String("format")
	if format == "" {
		format = "json"
	}

	builder, err := newSpecBuilder(cfg)
	if err != nil {
		return core.Result{Value: err, OK: false}
	}

	bridge := goapi.NewToolBridge(defaultSpecToolBridgePath)
	groups := specGroupsIter(bridge)

	if output != "" {
		if err := goapi.ExportSpecToFileIter(output, format, builder, groups); err != nil {
			return core.Result{Value: cli.Wrap(err, "write spec"), OK: false}
		}
		cli.Dim("Spec written to " + output)
		return core.Result{OK: true}
	}

	if err := goapi.ExportSpecIter(os.Stdout, format, builder, groups); err != nil {
		return core.Result{Value: cli.Wrap(err, "render spec"), OK: false}
	}
	return core.Result{OK: true}
}

func parseServers(raw string) []string {
	return splitUniqueCSV(raw)
}

func parseSecuritySchemes(raw string) (map[string]any, error) {
	raw = core.Trim(raw)
	if raw == "" {
		return nil, nil
	}

	var schemes map[string]any
	if err := json.Unmarshal([]byte(raw), &schemes); err != nil {
		return nil, cli.Wrap(err, "invalid security schemes JSON")
	}
	return schemes, nil
}

// specConfigFromOptions extracts a specBuilderConfig from the CLI options bag.
// Callers supply flags via `--key=value` on the command line; the CLI parser
// converts them to the option keys read here.
func specConfigFromOptions(opts core.Options) specBuilderConfig {
	cfg := specBuilderConfig{
		title:                   stringOr(opts.String("title"), "Lethean Core API"),
		summary:                 opts.String("summary"),
		description:             stringOr(opts.String("description"), "Lethean Core API"),
		version:                 stringOr(opts.String("version"), "1.0.0"),
		swaggerPath:             opts.String("swagger-path"),
		graphqlPath:             opts.String("graphql-path"),
		graphqlPlayground:       opts.Bool("graphql-playground"),
		graphqlPlaygroundPath:   opts.String("graphql-playground-path"),
		ssePath:                 opts.String("sse-path"),
		wsPath:                  opts.String("ws-path"),
		pprofEnabled:            opts.Bool("pprof"),
		expvarEnabled:           opts.Bool("expvar"),
		openAPISpecEnabled:      opts.Bool("openapi-spec"),
		openAPISpecPath:         opts.String("openapi-spec-path"),
		chatCompletionsEnabled:  opts.Bool("chat-completions"),
		chatCompletionsPath:     opts.String("chat-completions-path"),
		cacheEnabled:            opts.Bool("cache"),
		cacheTTL:                opts.String("cache-ttl"),
		cacheMaxEntries:         opts.Int("cache-max-entries"),
		cacheMaxBytes:           opts.Int("cache-max-bytes"),
		i18nDefaultLocale:       opts.String("i18n-default-locale"),
		i18nSupportedLocales:    opts.String("i18n-supported-locales"),
		authentikIssuer:         opts.String("authentik-issuer"),
		authentikClientID:       opts.String("authentik-client-id"),
		authentikTrustedProxy:   opts.Bool("authentik-trusted-proxy"),
		authentikPublicPaths:    opts.String("authentik-public-paths"),
		termsURL:                opts.String("terms-of-service"),
		contactName:             opts.String("contact-name"),
		contactURL:              opts.String("contact-url"),
		contactEmail:            opts.String("contact-email"),
		licenseName:             opts.String("license-name"),
		licenseURL:              opts.String("license-url"),
		externalDocsDescription: opts.String("external-docs-description"),
		externalDocsURL:         opts.String("external-docs-url"),
		servers:                 opts.String("server"),
		securitySchemes:         opts.String("security-schemes"),
	}
	return cfg
}

func stringOr(v, fallback string) string {
	if core.Trim(v) == "" {
		return fallback
	}
	return v
}
