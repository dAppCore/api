// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"fmt"
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
		title       string
		description string
		version     string
		servers     string
	)

	cmd := cli.NewCommand("sdk", "Generate client SDKs from OpenAPI spec", "", func(cmd *cli.Command, args []string) error {
		languages := splitUniqueCSV(lang)
		if len(languages) == 0 {
			return coreerr.E("sdk.Generate", "--lang is required and must include at least one non-empty language. Supported: "+strings.Join(goapi.SupportedLanguages(), ", "), nil)
		}

		// If no spec file provided, generate one to a temp file.
		if specFile == "" {
			builder := sdkSpecBuilder(title, description, version, servers)
			groups := sdkSpecGroups()

			tmpFile, err := os.CreateTemp("", "openapi-*.json")
			if err != nil {
				return coreerr.E("sdk.Generate", "create temp spec file", err)
			}
			defer coreio.Local.Delete(tmpFile.Name())

			if err := goapi.ExportSpec(tmpFile, "json", builder, groups); err != nil {
				tmpFile.Close()
				return coreerr.E("sdk.Generate", "generate spec", err)
			}
			tmpFile.Close()
			specFile = tmpFile.Name()
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
			if err := gen.Generate(context.Background(), l); err != nil {
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
	cli.StringFlag(cmd, &servers, "server", "S", "", "Comma-separated OpenAPI server URL(s)")

	parent.AddCommand(cmd)
}

func sdkSpecBuilder(title, description, version, servers string) *goapi.SpecBuilder {
	return &goapi.SpecBuilder{
		Title:       title,
		Description: description,
		Version:     version,
		Servers:     parseServers(servers),
	}
}

func sdkSpecGroups() []goapi.RouteGroup {
	bridge := goapi.NewToolBridge("/tools")
	return append(goapi.RegisteredSpecGroups(), bridge)
}
