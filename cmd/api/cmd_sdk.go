// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"iter"
	"os" // Note: AX-6 - os.CreateTemp provides O_CREATE|O_EXCL temp-file creation; no core primitive exists.

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"

	goapi "dappco.re/go/api"
	coreio "dappco.re/go/io"
)

const (
	defaultSDKTitle       = "Lethean Core API"
	defaultSDKDescription = "Lethean Core API"
	defaultSDKVersion     = "1.0.0"
)

func addSDKCommand(c *core.Core) {
	cmd := core.Command{
		Description: "Generate client SDKs from OpenAPI spec",
		Action:      sdkAction,
	}
	c.Command("api/sdk", cmd)
	c.Command("build/sdk", cmd)
}

func sdkAction(opts core.Options) core.Result {
	lang := opts.String("lang")
	output := opts.String("output")
	if output == "" {
		output = "./sdk"
	}
	specFile := opts.String("spec")
	packageName := opts.String("package")
	if packageName == "" {
		packageName = "lethean"
	}

	languages := splitUniqueCSV(lang)
	if len(languages) == 0 {
		return core.Fail(cli.Err("--lang is required and must include at least one non-empty language"))
	}

	gen := &goapi.SDKGenerator{
		OutputDir:   output,
		PackageName: packageName,
	}

	if !gen.Available() {
		cli.Error("openapi-generator-cli not found. Install with:")
		cli.Print("  brew install openapi-generator    (macOS)")
		cli.Print("  npm install @openapitools/openapi-generator-cli -g")
		return core.Fail(cli.Err("openapi-generator-cli not installed"))
	}

	resolvedSpecFile := specFile
	if resolvedSpecFile == "" {
		cfg := sdkConfigFromOptions(opts)
		builder, err := sdkSpecBuilder(cfg)
		if err != nil {
			return core.Fail(err)
		}
		groups := sdkSpecGroupsIter()

		tmpFile, err := os.CreateTemp("", "openapi-*.json")
		if err != nil {
			return core.Fail(cli.Wrap(err, "create temp spec file"))
		}
		tmpPath := tmpFile.Name()
		defer coreio.Local.Delete(tmpPath)

		if err := goapi.ExportSpecIter(tmpFile, "json", builder, groups); err != nil {
			if closeErr := tmpFile.Close(); closeErr != nil {
				return core.Fail(cli.Wrap(closeErr, "close temp spec file after generate spec failure"))
			}
			return core.Fail(cli.Wrap(err, "generate spec"))
		}
		if err := tmpFile.Close(); err != nil {
			return core.Fail(cli.Wrap(err, "close temp spec file"))
		}
		resolvedSpecFile = tmpPath
	}

	gen.SpecPath = resolvedSpecFile

	for _, l := range languages {
		cli.Dim("Generating " + l + " SDK...")
		if err := gen.Generate(cli.Context(), l); err != nil {
			return core.Fail(cli.Wrap(err, "generate "+l))
		}
		cli.Dim("  Done: " + output + "/" + l + "/")
	}

	return core.Ok(nil)
}

func sdkSpecBuilder(cfg specBuilderConfig) (*goapi.SpecBuilder, error) {
	return newSpecBuilder(cfg)
}

func sdkSpecGroupsIter() iter.Seq[goapi.RouteGroup] {
	return specGroupsIter(specToolBridge(defaultSpecToolBridgePath))
}

// sdkConfigFromOptions mirrors specConfigFromOptions but falls back to
// SDK-specific title/description/version defaults.
func sdkConfigFromOptions(opts core.Options) specBuilderConfig {
	cfg := specConfigFromOptions(opts)
	cfg.title = stringOr(opts.String("title"), defaultSDKTitle)
	cfg.description = stringOr(opts.String("description"), defaultSDKDescription)
	cfg.version = stringOr(opts.String("version"), defaultSDKVersion)
	return cfg
}
