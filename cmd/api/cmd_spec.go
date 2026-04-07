// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"forge.lthn.ai/core/cli/pkg/cli"

	goapi "dappco.re/go/core/api"
)

func addSpecCommand(parent *cli.Command) {
	var (
		output string
		format string
		cfg    specBuilderConfig
	)

	cfg.title = "Lethean Core API"
	cfg.description = "Lethean Core API"
	cfg.version = "1.0.0"

	cmd := cli.NewCommand("spec", "Generate OpenAPI specification", "", func(cmd *cli.Command, args []string) error {
		// Build spec from all route groups registered for CLI generation.
		builder, err := newSpecBuilder(cfg)
		if err != nil {
			return err
		}

		bridge := goapi.NewToolBridge("/tools")
		groups := specGroupsIter(bridge)

		if output != "" {
			if err := goapi.ExportSpecToFileIter(output, format, builder, groups); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Spec written to %s\n", output)
			return nil
		}

		return goapi.ExportSpecIter(os.Stdout, format, builder, groups)
	})

	cli.StringFlag(cmd, &output, "output", "o", "", "Write spec to file instead of stdout")
	cli.StringFlag(cmd, &format, "format", "f", "json", "Output format: json or yaml")
	registerSpecBuilderFlags(cmd, &cfg)

	parent.AddCommand(cmd)
}

func parseServers(raw string) []string {
	return splitUniqueCSV(raw)
}

func parseSecuritySchemes(raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var schemes map[string]any
	if err := json.Unmarshal([]byte(raw), &schemes); err != nil {
		return nil, cli.Err("invalid security schemes JSON: %w", err)
	}
	return schemes, nil
}

func registerSpecBuilderFlags(cmd *cli.Command, cfg *specBuilderConfig) {
	cli.StringFlag(cmd, &cfg.title, "title", "t", cfg.title, "API title in spec")
	cli.StringFlag(cmd, &cfg.summary, "summary", "", cfg.summary, "OpenAPI info summary in spec")
	cli.StringFlag(cmd, &cfg.description, "description", "d", cfg.description, "API description in spec")
	cli.StringFlag(cmd, &cfg.version, "version", "V", cfg.version, "API version in spec")
	cli.StringFlag(cmd, &cfg.swaggerPath, "swagger-path", "", "", "Swagger UI path in generated spec")
	cli.StringFlag(cmd, &cfg.graphqlPath, "graphql-path", "", "", "GraphQL endpoint path in generated spec")
	cli.BoolFlag(cmd, &cfg.graphqlPlayground, "graphql-playground", "", false, "Include the GraphQL playground endpoint in generated spec")
	cli.StringFlag(cmd, &cfg.graphqlPlaygroundPath, "graphql-playground-path", "", "", "GraphQL playground path in generated spec")
	cli.StringFlag(cmd, &cfg.ssePath, "sse-path", "", "", "SSE endpoint path in generated spec")
	cli.StringFlag(cmd, &cfg.wsPath, "ws-path", "", "", "WebSocket endpoint path in generated spec")
	cli.BoolFlag(cmd, &cfg.pprofEnabled, "pprof", "", false, "Include pprof endpoints in generated spec")
	cli.BoolFlag(cmd, &cfg.expvarEnabled, "expvar", "", false, "Include expvar endpoint in generated spec")
	cli.BoolFlag(cmd, &cfg.cacheEnabled, "cache", "", false, "Include cache metadata in generated spec")
	cli.StringFlag(cmd, &cfg.cacheTTL, "cache-ttl", "", "", "Cache TTL in generated spec")
	cli.IntFlag(cmd, &cfg.cacheMaxEntries, "cache-max-entries", "", 0, "Cache max entries in generated spec")
	cli.IntFlag(cmd, &cfg.cacheMaxBytes, "cache-max-bytes", "", 0, "Cache max bytes in generated spec")
	cli.StringFlag(cmd, &cfg.i18nDefaultLocale, "i18n-default-locale", "", "", "Default locale in generated spec")
	cli.StringFlag(cmd, &cfg.i18nSupportedLocales, "i18n-supported-locales", "", "", "Comma-separated supported locales in generated spec")
	cli.StringFlag(cmd, &cfg.authentikIssuer, "authentik-issuer", "", "", "Authentik issuer URL in generated spec")
	cli.StringFlag(cmd, &cfg.authentikClientID, "authentik-client-id", "", "", "Authentik client ID in generated spec")
	cli.BoolFlag(cmd, &cfg.authentikTrustedProxy, "authentik-trusted-proxy", "", false, "Mark Authentik proxy headers as trusted in generated spec")
	cli.StringFlag(cmd, &cfg.authentikPublicPaths, "authentik-public-paths", "", "", "Comma-separated public paths in generated spec")
	cli.StringFlag(cmd, &cfg.termsURL, "terms-of-service", "", "", "OpenAPI terms of service URL in spec")
	cli.StringFlag(cmd, &cfg.contactName, "contact-name", "", "", "OpenAPI contact name in spec")
	cli.StringFlag(cmd, &cfg.contactURL, "contact-url", "", "", "OpenAPI contact URL in spec")
	cli.StringFlag(cmd, &cfg.contactEmail, "contact-email", "", "", "OpenAPI contact email in spec")
	cli.StringFlag(cmd, &cfg.licenseName, "license-name", "", "", "OpenAPI licence name in spec")
	cli.StringFlag(cmd, &cfg.licenseURL, "license-url", "", "", "OpenAPI licence URL in spec")
	cli.StringFlag(cmd, &cfg.externalDocsDescription, "external-docs-description", "", "", "OpenAPI external documentation description in spec")
	cli.StringFlag(cmd, &cfg.externalDocsURL, "external-docs-url", "", "", "OpenAPI external documentation URL in spec")
	cli.StringFlag(cmd, &cfg.servers, "server", "S", "", "Comma-separated OpenAPI server URL(s)")
	cli.StringFlag(cmd, &cfg.securitySchemes, "security-schemes", "", "", "JSON object of custom OpenAPI security schemes")
}
