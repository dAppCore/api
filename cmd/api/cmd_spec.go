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
		output                  string
		format                  string
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
	)

	cmd := cli.NewCommand("spec", "Generate OpenAPI specification", "", func(cmd *cli.Command, args []string) error {
		// Build spec from all route groups registered for CLI generation.
		builder, err := newSpecBuilder(specBuilderConfig{
			title:                   title,
			summary:                 summary,
			description:             description,
			version:                 version,
			swaggerPath:             swaggerPath,
			graphqlPath:             graphqlPath,
			graphqlPlayground:       graphqlPlayground,
			ssePath:                 ssePath,
			wsPath:                  wsPath,
			pprofEnabled:            pprofEnabled,
			expvarEnabled:           expvarEnabled,
			termsURL:                termsURL,
			contactName:             contactName,
			contactURL:              contactURL,
			contactEmail:            contactEmail,
			licenseName:             licenseName,
			licenseURL:              licenseURL,
			externalDocsDescription: externalDocsDescription,
			externalDocsURL:         externalDocsURL,
			servers:                 servers,
			securitySchemes:         securitySchemes,
		})
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
	cli.StringFlag(cmd, &title, "title", "t", "Lethean Core API", "API title in spec")
	cli.StringFlag(cmd, &summary, "summary", "", "", "OpenAPI info summary in spec")
	cli.StringFlag(cmd, &description, "description", "d", "Lethean Core API", "API description in spec")
	cli.StringFlag(cmd, &version, "version", "V", "1.0.0", "API version in spec")
	cli.StringFlag(cmd, &swaggerPath, "swagger-path", "", "", "Swagger UI path in generated spec")
	cli.StringFlag(cmd, &graphqlPath, "graphql-path", "", "", "GraphQL endpoint path in generated spec")
	cli.BoolFlag(cmd, &graphqlPlayground, "graphql-playground", "", false, "Include the GraphQL playground endpoint in generated spec")
	cli.StringFlag(cmd, &ssePath, "sse-path", "", "", "SSE endpoint path in generated spec")
	cli.StringFlag(cmd, &wsPath, "ws-path", "", "", "WebSocket endpoint path in generated spec")
	cli.BoolFlag(cmd, &pprofEnabled, "pprof", "", false, "Include pprof endpoints in generated spec")
	cli.BoolFlag(cmd, &expvarEnabled, "expvar", "", false, "Include expvar endpoint in generated spec")
	cli.StringFlag(cmd, &termsURL, "terms-of-service", "", "", "OpenAPI terms of service URL in spec")
	cli.StringFlag(cmd, &contactName, "contact-name", "", "", "OpenAPI contact name in spec")
	cli.StringFlag(cmd, &contactURL, "contact-url", "", "", "OpenAPI contact URL in spec")
	cli.StringFlag(cmd, &contactEmail, "contact-email", "", "", "OpenAPI contact email in spec")
	cli.StringFlag(cmd, &licenseName, "license-name", "", "", "OpenAPI licence name in spec")
	cli.StringFlag(cmd, &licenseURL, "license-url", "", "", "OpenAPI licence URL in spec")
	cli.StringFlag(cmd, &externalDocsDescription, "external-docs-description", "", "", "OpenAPI external documentation description in spec")
	cli.StringFlag(cmd, &externalDocsURL, "external-docs-url", "", "", "OpenAPI external documentation URL in spec")
	cli.StringFlag(cmd, &servers, "server", "S", "", "Comma-separated OpenAPI server URL(s)")
	cli.StringFlag(cmd, &securitySchemes, "security-schemes", "", "", "JSON object of custom OpenAPI security schemes")

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
