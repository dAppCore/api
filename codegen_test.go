// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	api "dappco.re/go/core/api"
)

// ── SDKGenerator tests ─────────────────────────────────────────────────────

func TestSDKGenerator_Good_SupportedLanguages(t *testing.T) {
	langs := api.SupportedLanguages()
	if len(langs) == 0 {
		t.Fatal("expected at least one supported language")
	}

	expected := []string{"go", "typescript-fetch", "python", "java", "csharp"}
	for _, lang := range expected {
		if !slices.Contains(langs, lang) {
			t.Errorf("expected %q in supported languages, got %v", lang, langs)
		}
	}
}

func TestSDKGenerator_Bad_UnsupportedLanguage(t *testing.T) {
	gen := &api.SDKGenerator{
		SpecPath:  "spec.json",
		OutputDir: t.TempDir(),
	}

	err := gen.Generate(context.Background(), "brainfuck")
	if err == nil {
		t.Fatal("expected error for unsupported language, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported language") {
		t.Fatalf("expected error to contain 'unsupported language', got: %v", err)
	}
}

func TestSDKGenerator_Bad_MissingSpec(t *testing.T) {
	gen := &api.SDKGenerator{
		SpecPath:  filepath.Join(t.TempDir(), "nonexistent.json"),
		OutputDir: t.TempDir(),
	}

	err := gen.Generate(context.Background(), "go")
	if err == nil {
		t.Fatal("expected error for missing spec file, got nil")
	}
	if !strings.Contains(err.Error(), "spec file not found") {
		t.Fatalf("expected error to contain 'spec file not found', got: %v", err)
	}
}

func TestSDKGenerator_Good_OutputDirCreated(t *testing.T) {
	// Write a minimal spec file so we pass the file-exists check.
	specDir := t.TempDir()
	specPath := filepath.Join(specDir, "spec.json")
	if err := os.WriteFile(specPath, []byte(`{"openapi":"3.1.0"}`), 0o644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	outputDir := filepath.Join(t.TempDir(), "nested", "sdk")
	gen := &api.SDKGenerator{
		SpecPath:  specPath,
		OutputDir: outputDir,
	}

	// Generate will fail at the exec step (openapi-generator-cli likely not installed),
	// but the output directory should have been created before that.
	_ = gen.Generate(context.Background(), "go")

	expected := filepath.Join(outputDir, "go")
	info, err := os.Stat(expected)
	if err != nil {
		t.Fatalf("expected output directory %s to exist, got error: %v", expected, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", expected)
	}
}

func TestSDKGenerator_Good_Available(t *testing.T) {
	gen := &api.SDKGenerator{}
	// Just verify it returns a bool and does not panic.
	_ = gen.Available()
}
