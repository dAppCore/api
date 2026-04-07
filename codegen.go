// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	coreio "dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
)

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

	language = strings.TrimSpace(language)
	generator, ok := supportedLanguages[language]
	if !ok {
		return coreerr.E("SDKGenerator.Generate", fmt.Sprintf("unsupported language %q: supported languages are %v", language, SupportedLanguages()), nil)
	}

	specPath := strings.TrimSpace(g.SpecPath)
	if specPath == "" {
		return coreerr.E("SDKGenerator.Generate", "spec path is required", nil)
	}
	if _, err := os.Stat(specPath); err != nil {
		if os.IsNotExist(err) {
			return coreerr.E("SDKGenerator.Generate", "spec file not found: "+specPath, nil)
		}
		return coreerr.E("SDKGenerator.Generate", "stat spec file", err)
	}

	outputBase := strings.TrimSpace(g.OutputDir)
	if outputBase == "" {
		return coreerr.E("SDKGenerator.Generate", "output directory is required", nil)
	}

	if !g.Available() {
		return coreerr.E("SDKGenerator.Generate", "openapi-generator-cli not installed", nil)
	}

	outputDir := filepath.Join(outputBase, language)
	if err := coreio.Local.EnsureDir(outputDir); err != nil {
		return coreerr.E("SDKGenerator.Generate", "create output directory", err)
	}

	args := g.buildArgs(specPath, generator, outputDir)
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
	_, err := exec.LookPath("openapi-generator-cli")
	return err == nil
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
//		fmt.Println(lang)
//	}
func SupportedLanguagesIter() iter.Seq[string] {
	return slices.Values(SupportedLanguages())
}
