// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"fmt"
	"iter"
	"os"
	"strings"

	"forge.lthn.ai/core/cli/pkg/cli"

	coreio "dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"

	goapi "dappco.re/go/core/api"
)

const (
	defaultSDKTitle       = "Lethean Core API"
	defaultSDKDescription = "Lethean Core API"
	defaultSDKVersion     = "1.0.0"
)

func addSDKCommand(parent *cli.Command) {
	var (
		lang                    string
		output                  string
		specFile                string
		packageName             string
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
	)

	cmd := cli.NewCommand("sdk", "Generate client SDKs from OpenAPI spec", "", func(cmd *cli.Command, args []string) error {
		languages := splitUniqueCSV(lang)
		if len(languages) == 0 {
			return coreerr.E("sdk.Generate", "--lang is required and must include at least one non-empty language. Supported: "+strings.Join(goapi.SupportedLanguages(), ", "), nil)
		}

		// If no spec file provided, generate one to a temp file.
		if specFile == "" {
			builder, err := sdkSpecBuilder(title, description, version, swaggerPath, graphqlPath, graphqlPlayground, ssePath, wsPath, pprofEnabled, expvarEnabled, termsURL, contactName, contactURL, contactEmail, licenseName, licenseURL, externalDocsDescription, externalDocsURL, servers, securitySchemes)
			if err != nil {
				return err
			}
			groups := sdkSpecGroupsIter()

			tmpFile, err := os.CreateTemp("", "openapi-*.json")
			if err != nil {
				return coreerr.E("sdk.Generate", "create temp spec file", err)
			}
			tmpPath := tmpFile.Name()
			if err := tmpFile.Close(); err != nil {
				_ = coreio.Local.Delete(tmpPath)
				return coreerr.E("sdk.Generate", "close temp spec file", err)
			}
			defer coreio.Local.Delete(tmpPath)

			if err := goapi.ExportSpecToFileIter(tmpPath, "json", builder, groups); err != nil {
				return coreerr.E("sdk.Generate", "generate spec", err)
			}
			specFile = tmpPath
		}

		gen := &goapi.SDKGenerator{
			SpecPath:    specFile,
			OutputDir:   output,
			PackageName: packageName,
		}

		if !gen.Available() {
			fmt.Fprintln(os.Stderr, "openapi-generator-cli not found. Install with:")
			fmt.Fprintln(os.Stderr, "  brew install openapi-generator    (macOS)")
			fmt.Fprintln(os.Stderr, "  npm install @openapitools/openapi-generator-cli -g")
			return coreerr.E("sdk.Generate", "openapi-generator-cli not installed", nil)
		}

		// Generate for each language.
		for _, l := range languages {
			fmt.Fprintf(os.Stderr, "Generating %s SDK...\n", l)
			if err := gen.Generate(cli.Context(), l); err != nil {
				return coreerr.E("sdk.Generate", "generate "+l, err)
			}
			fmt.Fprintf(os.Stderr, "  Done: %s/%s/\n", output, l)
		}

		return nil
	})

	cli.StringFlag(cmd, &lang, "lang", "l", "", "Target language(s), comma-separated (e.g. go,python,typescript-fetch)")
	cli.StringFlag(cmd, &output, "output", "o", "./sdk", "Output directory for generated SDKs")
	cli.StringFlag(cmd, &specFile, "spec", "s", "", "Path to existing OpenAPI spec (generates from MCP tools if not provided)")
	cli.StringFlag(cmd, &packageName, "package", "p", "lethean", "Package name for generated SDK")
	cli.StringFlag(cmd, &title, "title", "t", defaultSDKTitle, "API title in generated spec")
	cli.StringFlag(cmd, &description, "description", "d", defaultSDKDescription, "API description in generated spec")
	cli.StringFlag(cmd, &version, "version", "V", defaultSDKVersion, "API version in generated spec")
	cli.StringFlag(cmd, &swaggerPath, "swagger-path", "", "", "Swagger UI path in generated spec")
	cli.StringFlag(cmd, &graphqlPath, "graphql-path", "", "", "GraphQL endpoint path in generated spec")
	cli.BoolFlag(cmd, &graphqlPlayground, "graphql-playground", "", false, "Include the GraphQL playground endpoint in generated spec")
	cli.StringFlag(cmd, &ssePath, "sse-path", "", "", "SSE endpoint path in generated spec")
	cli.StringFlag(cmd, &wsPath, "ws-path", "", "", "WebSocket endpoint path in generated spec")
	cli.BoolFlag(cmd, &pprofEnabled, "pprof", "", false, "Include pprof endpoints in generated spec")
	cli.BoolFlag(cmd, &expvarEnabled, "expvar", "", false, "Include expvar endpoint in generated spec")
	cli.StringFlag(cmd, &termsURL, "terms-of-service", "", "", "OpenAPI terms of service URL in generated spec")
	cli.StringFlag(cmd, &contactName, "contact-name", "", "", "OpenAPI contact name in generated spec")
	cli.StringFlag(cmd, &contactURL, "contact-url", "", "", "OpenAPI contact URL in generated spec")
	cli.StringFlag(cmd, &contactEmail, "contact-email", "", "", "OpenAPI contact email in generated spec")
	cli.StringFlag(cmd, &licenseName, "license-name", "", "", "OpenAPI licence name in generated spec")
	cli.StringFlag(cmd, &licenseURL, "license-url", "", "", "OpenAPI licence URL in generated spec")
	cli.StringFlag(cmd, &externalDocsDescription, "external-docs-description", "", "", "OpenAPI external documentation description in generated spec")
	cli.StringFlag(cmd, &externalDocsURL, "external-docs-url", "", "", "OpenAPI external documentation URL in generated spec")
	cli.StringFlag(cmd, &servers, "server", "S", "", "Comma-separated OpenAPI server URL(s)")
	cli.StringFlag(cmd, &securitySchemes, "security-schemes", "", "", "JSON object of custom OpenAPI security schemes")

	parent.AddCommand(cmd)
}

func sdkSpecBuilder(title, description, version, swaggerPath, graphqlPath string, graphqlPlayground bool, ssePath, wsPath string, pprofEnabled, expvarEnabled bool, termsURL, contactName, contactURL, contactEmail, licenseName, licenseURL, externalDocsDescription, externalDocsURL, servers, securitySchemes string) (*goapi.SpecBuilder, error) {
	return newSpecBuilder(specBuilderConfig{
		title:                   title,
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
}

func sdkSpecGroupsIter() iter.Seq[goapi.RouteGroup] {
	return specGroupsIter(goapi.NewToolBridge("/tools"))
}
