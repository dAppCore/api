// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"io/fs"
	"iter"
	"maps"
	// Note: AX-6 - retained for inheriting stdout/stderr when invoking the SDK generator; filesystem checks below use core.Fs.
	"os"
	// Note: AX-6 - retained for the subprocess boundary because SDKGenerator has no Core instance with registered process.run.
	"os/exec"
	// Note: AX-6 - compiled regexp anchors PackageName validation for command-argument safety.
	"regexp"
	"slices"

	core "dappco.re/go"
	coreerr "dappco.re/go/log"
)

// packageNameRe constrains SDKGenerator.PackageName to identifier-shaped
// values so it cannot smuggle additional CLI flags through
// --additional-properties packageName=<value>. Defence-in-depth per
// Cerberus mechanism review on Mantis #322 — current callsite is operator-
// only via cmd/api/cmd_sdk.go, but future consumers binding request input
// to this field would re-open the flag-injection surface without it.
var packageNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._-]*$`)

// Supported SDK target languages.
var supportedLanguages = map[string]string{
	"go":               "go",
	"typescript-fetch": "typescript-fetch",
	"typescript-axios": "typescript-axios",
	"python":           "python",
	"java":             "java",
	"csharp":           "csharp-netcore",
	"ruby":             "ruby",
	"swift":            "swift5",
	"kotlin":           "kotlin",
	"rust":             "rust",
	"php":              "php",
}

// SDKGenerator wraps openapi-generator-cli for SDK generation.
//
// Example:
//
//	gen := &api.SDKGenerator{SpecPath: "./openapi.yaml", OutputDir: "./sdk", PackageName: "service"}
type SDKGenerator struct {
	// SpecPath is the path to the OpenAPI spec file (JSON or YAML).
	SpecPath string

	// OutputDir is the base directory for generated SDK output.
	OutputDir string

	// PackageName is the name used for the generated package/module.
	PackageName string
}

// Generate creates an SDK for the given language using openapi-generator-cli.
// The language must be one of the supported languages returned by SupportedLanguages().
//
// Example:
//
//	err := gen.Generate(context.Background(), "go")
func (g *SDKGenerator) Generate(ctx context.Context, language string) error {
	if g == nil {
		return coreerr.E("SDKGenerator.Generate", "generator is nil", nil)
	}
	if ctx == nil {
		return coreerr.E("SDKGenerator.Generate", "context is nil", nil)
	}

	language = core.Trim(language)
	generator, ok := supportedLanguages[language]
	if !ok {
		return coreerr.E("SDKGenerator.Generate", core.Sprintf("unsupported language %q: supported languages are %v", language, SupportedLanguages()), nil)
	}

	specPath := core.Trim(g.SpecPath)
	if specPath == "" {
		return coreerr.E("SDKGenerator.Generate", "spec path is required", nil)
	}
	localFS := (&core.Fs{}).NewUnrestricted()
	if result := localFS.Stat(specPath); !result.OK {
		err, _ := result.Value.(error)
		if core.Is(err, fs.ErrNotExist) {
			return coreerr.E("SDKGenerator.Generate", "spec file not found: "+specPath, nil)
		}
		return coreerr.E("SDKGenerator.Generate", "stat spec file", err)
	}

	outputBase := core.Trim(g.OutputDir)
	if outputBase == "" {
		return coreerr.E("SDKGenerator.Generate", "output directory is required", nil)
	}

	if g.PackageName != "" && !packageNameRe.MatchString(g.PackageName) {
		return coreerr.E("SDKGenerator.Generate",
			core.Sprintf("package name %q rejected: must match %s", g.PackageName, packageNameRe.String()), nil)
	}

	if !g.Available() {
		return coreerr.E("SDKGenerator.Generate", "openapi-generator-cli not installed", nil)
	}

	outputDir := core.Path(outputBase, language)
	if !core.PathIsAbs(outputBase) {
		outputDir = core.Path(core.Env("DIR_CWD"), outputBase, language)
	}
	if result := localFS.EnsureDir(outputDir); !result.OK {
		err, _ := result.Value.(error)
		return coreerr.E("SDKGenerator.Generate", "create output directory", err)
	}

	args := g.buildArgs(specPath, generator, outputDir)
	// SDKGenerator is a standalone helper with no Core runtime handle. Without
	// a registered process.run action, core.Process cannot execute this command.
	// Command name is a string literal (zero attacker-influence). Args are
	// constructed from a closed allowlist of generator names (supportedLanguages)
	// and operator-supplied spec/output paths. Current callsite is operator-only
	// via cmd/api/cmd_sdk.go. PackageName is regex-validated above to prevent
	// flag-injection through --additional-properties. Cerberus mechanism review
	// attached to Mantis #322.
	//#nosec G204 -- command literal; args from closed allowlist + operator config + validated PackageName.
	cmd := exec.CommandContext(ctx, "openapi-generator-cli", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return coreerr.E("SDKGenerator.Generate", "openapi-generator-cli failed for "+language, err)
	}

	return nil
}

// buildArgs constructs the openapi-generator-cli command arguments.
func (g *SDKGenerator) buildArgs(specPath, generator, outputDir string) []string {
	args := []string{
		"generate",
		"-i", specPath,
		"-g", generator,
		"-o", outputDir,
	}
	if g.PackageName != "" {
		args = append(args, "--additional-properties", "packageName="+g.PackageName)
	}
	return args
}

// Available checks if openapi-generator-cli is installed and accessible.
//
// Example:
//
//	if !gen.Available() {
//		t.Fatal("openapi-generator-cli is required")
//	}
func (g *SDKGenerator) Available() bool {
	return core.App{}.Find("openapi-generator-cli", "openapi-generator-cli").OK
}

// SupportedLanguages returns the list of supported SDK target languages
// in sorted order for deterministic output.
//
// Example:
//
//	langs := api.SupportedLanguages()
func SupportedLanguages() []string {
	return slices.Sorted(maps.Keys(supportedLanguages))
}

// SupportedLanguagesIter returns an iterator over supported SDK target languages in sorted order.
//
// Example:
//
//	for lang := range api.SupportedLanguagesIter() {
//		_ = lang
//	}
func SupportedLanguagesIter() iter.Seq[string] {
	return slices.Values(SupportedLanguages())
}
