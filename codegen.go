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
func (g *SDKGenerator) Generate(ctx context.Context, language string) error {
	generator, ok := supportedLanguages[language]
	if !ok {
		return fmt.Errorf("unsupported language %q: supported languages are %v", language, SupportedLanguages())
	}

	if _, err := os.Stat(g.SpecPath); os.IsNotExist(err) {
		return fmt.Errorf("spec file not found: %s", g.SpecPath)
	}

	outputDir := filepath.Join(g.OutputDir, language)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	args := g.buildArgs(generator, outputDir)
	cmd := exec.CommandContext(ctx, "openapi-generator-cli", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("openapi-generator-cli failed for %s: %w", language, err)
	}

	return nil
}

// buildArgs constructs the openapi-generator-cli command arguments.
func (g *SDKGenerator) buildArgs(generator, outputDir string) []string {
	args := []string{
		"generate",
		"-i", g.SpecPath,
		"-g", generator,
		"-o", outputDir,
	}
	if g.PackageName != "" {
		args = append(args, "--additional-properties", "packageName="+g.PackageName)
	}
	return args
}

// Available checks if openapi-generator-cli is installed and accessible.
func (g *SDKGenerator) Available() bool {
	_, err := exec.LookPath("openapi-generator-cli")
	return err == nil
}

// SupportedLanguages returns the list of supported SDK target languages
// in sorted order for deterministic output.
func SupportedLanguages() []string {
	return slices.Sorted(maps.Keys(supportedLanguages))
}

// SupportedLanguagesIter returns an iterator over supported SDK target languages in sorted order.
func SupportedLanguagesIter() iter.Seq[string] {
	return slices.Values(SupportedLanguages())
}
