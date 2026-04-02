// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"

	goapi "dappco.re/go/core/api"
)

func addSpecCommand(parent *cli.Command) {
	var (
		output                  string
		format                  string
		title                   string
		description             string
		version                 string
		graphqlPath             string
		ssePath                 string
		termsURL                string
		contactName             string
		contactURL              string
		contactEmail            string
		licenseName             string
		licenseURL              string
		externalDocsDescription string
		externalDocsURL         string
		servers                 string
	)

	cmd := cli.NewCommand("spec", "Generate OpenAPI specification", "", func(cmd *cli.Command, args []string) error {
		// Build spec from all route groups registered for CLI generation.
		builder := &goapi.SpecBuilder{
			Title:                   title,
			Description:             description,
			Version:                 version,
			GraphQLPath:             graphqlPath,
			SSEPath:                 ssePath,
			TermsOfService:          termsURL,
			ContactName:             contactName,
			ContactURL:              contactURL,
			ContactEmail:            contactEmail,
			Servers:                 parseServers(servers),
			LicenseName:             licenseName,
			LicenseURL:              licenseURL,
			ExternalDocsDescription: externalDocsDescription,
			ExternalDocsURL:         externalDocsURL,
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
	cli.StringFlag(cmd, &description, "description", "d", "Lethean Core API", "API description in spec")
	cli.StringFlag(cmd, &version, "version", "V", "1.0.0", "API version in spec")
	cli.StringFlag(cmd, &graphqlPath, "graphql-path", "", "", "GraphQL endpoint path in generated spec")
	cli.StringFlag(cmd, &ssePath, "sse-path", "", "", "SSE endpoint path in generated spec")
	cli.StringFlag(cmd, &termsURL, "terms-of-service", "", "", "OpenAPI terms of service URL in spec")
	cli.StringFlag(cmd, &contactName, "contact-name", "", "", "OpenAPI contact name in spec")
	cli.StringFlag(cmd, &contactURL, "contact-url", "", "", "OpenAPI contact URL in spec")
	cli.StringFlag(cmd, &contactEmail, "contact-email", "", "", "OpenAPI contact email in spec")
	cli.StringFlag(cmd, &licenseName, "license-name", "", "", "OpenAPI licence name in spec")
	cli.StringFlag(cmd, &licenseURL, "license-url", "", "", "OpenAPI licence URL in spec")
	cli.StringFlag(cmd, &externalDocsDescription, "external-docs-description", "", "", "OpenAPI external documentation description in spec")
	cli.StringFlag(cmd, &externalDocsURL, "external-docs-url", "", "", "OpenAPI external documentation URL in spec")
	cli.StringFlag(cmd, &servers, "server", "S", "", "Comma-separated OpenAPI server URL(s)")

	parent.AddCommand(cmd)
}

func parseServers(raw string) []string {
	return splitUniqueCSV(raw)
}
