// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"iter"
	"os"

	core "dappco.re/go/core"
	"dappco.re/go/core/cli/pkg/cli"

	goapi "dappco.re/go/core/api"
	coreio "dappco.re/go/core/io"
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
		return core.Result{Value: cli.Err("--lang is required and must include at least one non-empty language"), OK: false}
	}

	gen := &goapi.SDKGenerator{
		OutputDir:   output,
		PackageName: packageName,
	}

	if !gen.Available() {
		cli.Error("openapi-generator-cli not found. Install with:")
		cli.Print("  brew install openapi-generator    (macOS)")
		cli.Print("  npm install @openapitools/openapi-generator-cli -g")
		return core.Result{Value: cli.Err("openapi-generator-cli not installed"), OK: false}
	}

	resolvedSpecFile := specFile
	if resolvedSpecFile == "" {
		cfg := sdkConfigFromOptions(opts)
		builder, err := sdkSpecBuilder(cfg)
		if err != nil {
			return core.Result{Value: err, OK: false}
		}
		groups := sdkSpecGroupsIter()

		tmpFile, err := os.CreateTemp("", "openapi-*.json")
		if err != nil {
			return core.Result{Value: cli.Wrap(err, "create temp spec file"), OK: false}
		}
		tmpPath := tmpFile.Name()
		if err := tmpFile.Close(); err != nil {
			_ = coreio.Local.Delete(tmpPath)
			return core.Result{Value: cli.Wrap(err, "close temp spec file"), OK: false}
		}
		defer coreio.Local.Delete(tmpPath)

		if err := goapi.ExportSpecToFileIter(tmpPath, "json", builder, groups); err != nil {
			return core.Result{Value: cli.Wrap(err, "generate spec"), OK: false}
		}
		resolvedSpecFile = tmpPath
	}

	gen.SpecPath = resolvedSpecFile

	for _, l := range languages {
		cli.Dim("Generating " + l + " SDK...")
		if err := gen.Generate(cli.Context(), l); err != nil {
			return core.Result{Value: cli.Wrap(err, "generate "+l), OK: false}
		}
		cli.Dim("  Done: " + output + "/" + l + "/")
	}

	return core.Result{OK: true}
}

func sdkSpecBuilder(cfg specBuilderConfig) (*goapi.SpecBuilder, error) {
	return newSpecBuilder(cfg)
}

func sdkSpecGroupsIter() iter.Seq[goapi.RouteGroup] {
	return specGroupsIter(goapi.NewToolBridge(defaultSpecToolBridgePath))
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
