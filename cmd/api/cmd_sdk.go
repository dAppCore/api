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
		lang        string
		output      string
		specFile    string
		packageName string
		cfg         specBuilderConfig
	)

	cfg.title = defaultSDKTitle
	cfg.description = defaultSDKDescription
	cfg.version = defaultSDKVersion

	cmd := cli.NewCommand("sdk", "Generate client SDKs from OpenAPI spec", "", func(cmd *cli.Command, args []string) error {
		languages := splitUniqueCSV(lang)
		if len(languages) == 0 {
			return coreerr.E("sdk.Generate", "--lang is required and must include at least one non-empty language. Supported: "+strings.Join(goapi.SupportedLanguages(), ", "), nil)
		}

		gen := &goapi.SDKGenerator{
			OutputDir:   output,
			PackageName: packageName,
		}

		if !gen.Available() {
			fmt.Fprintln(os.Stderr, "openapi-generator-cli not found. Install with:")
			fmt.Fprintln(os.Stderr, "  brew install openapi-generator    (macOS)")
			fmt.Fprintln(os.Stderr, "  npm install @openapitools/openapi-generator-cli -g")
			return coreerr.E("sdk.Generate", "openapi-generator-cli not installed", nil)
		}

		// If no spec file was provided, generate one only after confirming the
		// generator is available.
		resolvedSpecFile := specFile
		if resolvedSpecFile == "" {
			builder, err := sdkSpecBuilder(cfg)
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
			resolvedSpecFile = tmpPath
		}

		gen.SpecPath = resolvedSpecFile

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
	cli.StringFlag(cmd, &specFile, "spec", "s", "", "Path to an existing OpenAPI spec (generates a temporary spec from registered route groups and the built-in tool bridge if not provided)")
	cli.StringFlag(cmd, &packageName, "package", "p", "lethean", "Package name for generated SDK")
	registerSpecBuilderFlags(cmd, &cfg)

	parent.AddCommand(cmd)
}

func sdkSpecBuilder(cfg specBuilderConfig) (*goapi.SpecBuilder, error) {
	return newSpecBuilder(specBuilderConfig{
		title:                   cfg.title,
		summary:                 cfg.summary,
		description:             cfg.description,
		version:                 cfg.version,
		swaggerPath:             cfg.swaggerPath,
		graphqlPath:             cfg.graphqlPath,
		graphqlPlayground:       cfg.graphqlPlayground,
		graphqlPlaygroundPath:   cfg.graphqlPlaygroundPath,
		ssePath:                 cfg.ssePath,
		wsPath:                  cfg.wsPath,
		pprofEnabled:            cfg.pprofEnabled,
		expvarEnabled:           cfg.expvarEnabled,
		cacheEnabled:            cfg.cacheEnabled,
		cacheTTL:                cfg.cacheTTL,
		cacheMaxEntries:         cfg.cacheMaxEntries,
		cacheMaxBytes:           cfg.cacheMaxBytes,
		i18nDefaultLocale:       cfg.i18nDefaultLocale,
		i18nSupportedLocales:    cfg.i18nSupportedLocales,
		authentikIssuer:         cfg.authentikIssuer,
		authentikClientID:       cfg.authentikClientID,
		authentikTrustedProxy:   cfg.authentikTrustedProxy,
		authentikPublicPaths:    cfg.authentikPublicPaths,
		termsURL:                cfg.termsURL,
		contactName:             cfg.contactName,
		contactURL:              cfg.contactURL,
		contactEmail:            cfg.contactEmail,
		licenseName:             cfg.licenseName,
		licenseURL:              cfg.licenseURL,
		externalDocsDescription: cfg.externalDocsDescription,
		externalDocsURL:         cfg.externalDocsURL,
		servers:                 cfg.servers,
		securitySchemes:         cfg.securitySchemes,
	})
}

func sdkSpecGroupsIter() iter.Seq[goapi.RouteGroup] {
	return specGroupsIter(goapi.NewToolBridge("/tools"))
}
