// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"context"
	core "dappco.re/go"
	"slices"
	"testing"

	api "dappco.re/go/api"
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
	if !core.Contains(err.Error(), "unsupported language") {
		t.Fatalf("expected error to contain 'unsupported language', got: %v", err)
	}
}

func TestSDKGenerator_Bad_MissingSpec(t *testing.T) {
	gen := &api.SDKGenerator{
		SpecPath:  core.PathJoin(t.TempDir(), "nonexistent.json"),
		OutputDir: t.TempDir(),
	}

	err := gen.Generate(context.Background(), "go")
	if err == nil {
		t.Fatal("expected error for missing spec file, got nil")
	}
	if !core.Contains(err.Error(), "spec file not found") {
		t.Fatalf("expected error to contain 'spec file not found', got: %v", err)
	}
}

func TestSDKGenerator_Bad_EmptySpecPath(t *testing.T) {
	gen := &api.SDKGenerator{
		OutputDir: t.TempDir(),
	}

	err := gen.Generate(context.Background(), "go")
	if err == nil {
		t.Fatal("expected error for empty spec path, got nil")
	}
	if !core.Contains(err.Error(), "spec path is required") {
		t.Fatalf("expected error to contain 'spec path is required', got: %v", err)
	}
}

func TestSDKGenerator_Bad_EmptyOutputDir(t *testing.T) {
	specDir := t.TempDir()
	specPath := core.PathJoin(specDir, "spec.json")
	if err := coreWriteFile(specPath, []byte(`{"openapi":"3.1.0"}`), 0o644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	gen := &api.SDKGenerator{
		SpecPath: specPath,
	}

	err := gen.Generate(context.Background(), "go")
	if err == nil {
		t.Fatal("expected error for empty output directory, got nil")
	}
	if !core.Contains(err.Error(), "output directory is required") {
		t.Fatalf("expected error to contain 'output directory is required', got: %v", err)
	}
}

func TestSDKGenerator_Bad_NilContext(t *testing.T) {
	gen := &api.SDKGenerator{
		SpecPath:  core.PathJoin(t.TempDir(), "nonexistent.json"),
		OutputDir: t.TempDir(),
	}

	err := gen.Generate(nil, "go")
	if err == nil {
		t.Fatal("expected error for nil context, got nil")
	}
	if !core.Contains(err.Error(), "context is nil") {
		t.Fatalf("expected error to contain 'context is nil', got: %v", err)
	}
}

func TestSDKGenerator_Bad_NilReceiver(t *testing.T) {
	var gen *api.SDKGenerator

	err := gen.Generate(context.Background(), "go")
	if err == nil {
		t.Fatal("expected error for nil generator, got nil")
	}
	if !core.Contains(err.Error(), "generator is nil") {
		t.Fatalf("expected error to contain 'generator is nil', got: %v", err)
	}
}

func TestSDKGenerator_Bad_MissingGenerator(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	specDir := t.TempDir()
	specPath := core.PathJoin(specDir, "spec.json")
	if err := coreWriteFile(specPath, []byte(`{"openapi":"3.1.0"}`), 0o644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	outputDir := core.PathJoin(t.TempDir(), "nested", "sdk")
	gen := &api.SDKGenerator{
		SpecPath:  specPath,
		OutputDir: outputDir,
	}

	err := gen.Generate(context.Background(), "go")
	if err == nil {
		t.Fatal("expected error when openapi-generator-cli is missing, got nil")
	}
	if !core.Contains(err.Error(), "openapi-generator-cli not installed") {
		t.Fatalf("expected missing-generator error, got: %v", err)
	}

	if _, statErr := coreStat(core.PathJoin(outputDir, "go")); !core.IsNotExist(statErr) {
		t.Fatalf("expected output directory not to be created when generator is missing, got err=%v", statErr)
	}
}

func TestSDKGenerator_Good_OutputDirCreated(t *testing.T) {
	oldPath := core.Getenv("PATH")

	// Provide a fake openapi-generator-cli so Generate reaches the exec step
	// without depending on the host environment.
	binDir := t.TempDir()
	binPath := core.PathJoin(binDir, "openapi-generator-cli")
	script := []byte("#!/bin/sh\nexit 1\n")
	if err := coreWriteFile(binPath, script, 0o755); err != nil {
		t.Fatalf("failed to write fake generator: %v", err)
	}
	t.Setenv("PATH", binDir+string(core.PathListSeparator)+oldPath)

	// Write a minimal spec file so we pass the file-exists check.
	specDir := t.TempDir()
	specPath := core.PathJoin(specDir, "spec.json")
	if err := coreWriteFile(specPath, []byte(`{"openapi":"3.1.0"}`), 0o644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	outputDir := core.PathJoin(t.TempDir(), "nested", "sdk")
	gen := &api.SDKGenerator{
		SpecPath:  specPath,
		OutputDir: outputDir,
	}

	// Generate will fail at the exec step, but the output directory should have
	// been created before the CLI returned its non-zero status.
	_ = gen.Generate(context.Background(), "go")

	expected := core.PathJoin(outputDir, "go")
	info, err := coreStat(expected)
	if err != nil {
		t.Fatalf("expected output directory %s to exist, got error: %v", expected, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", expected)
	}
}

func TestSDKGenerator_Good_Available(t *testing.T) {
	gen := &api.SDKGenerator{}
	available := gen.Available()
	if available {
		t.Log("openapi-generator-cli is available")
	} else {
		t.Log("openapi-generator-cli is unavailable")
	}
}

// TestSDKGenerator_Generate_PackageNameRejected_Bad verifies the regex-validation
// hardening from Mantis #322 — PackageName containing flag-injection characters
// is rejected before exec.CommandContext is reached.
func TestSDKGenerator_Generate_PackageNameRejected_Bad(t *testing.T) {
	tmp := t.TempDir()
	specPath := core.PathJoin(tmp, "spec.yaml")
	if err := coreWriteFile(specPath, []byte("openapi: 3.0.0\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	rejects := []string{
		"foo --extra=evil",  // space + flag injection
		"foo;rm -rf /",      // command separator
		"foo bar",           // bare space
		"--shell-injection", // leading dash
		"foo$(whoami)",      // command substitution
	}
	for _, name := range rejects {
		t.Run(name, func(t *testing.T) {
			gen := &api.SDKGenerator{
				SpecPath:    specPath,
				OutputDir:   tmp,
				PackageName: name,
			}
			err := gen.Generate(context.Background(), "go")
			if err == nil {
				t.Errorf("expected rejection for PackageName=%q, got nil error", name)
				return
			}
			if !core.Contains(err.Error(), "package name") {
				t.Errorf("expected rejection error containing 'package name', got %q", err.Error())
			}
		})
	}
}

// TestSDKGenerator_Generate_PackageNameAccepted_Good verifies legitimate names
// pass the regex; any subsequent error must NOT be the regex-rejection.
func TestSDKGenerator_Generate_PackageNameAccepted_Good(t *testing.T) {
	accepts := []string{
		"foo",
		"FooBar",
		"foo_bar",
		"foo-bar",
		"Foo123",
		"a",
	}
	tmp := t.TempDir()
	specPath := core.PathJoin(tmp, "spec.yaml")
	if err := coreWriteFile(specPath, []byte("openapi: 3.0.0\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	for _, name := range accepts {
		t.Run(name, func(t *testing.T) {
			gen := &api.SDKGenerator{
				SpecPath:    specPath,
				OutputDir:   tmp,
				PackageName: name,
			}
			err := gen.Generate(context.Background(), "go")
			// Likely fails because openapi-generator-cli isn't installed in
			// CI; the error MUST NOT be the regex-rejection ("package name
			// X rejected").
			if err != nil && core.Contains(err.Error(), "package name") &&
				core.Contains(err.Error(), "rejected") {
				t.Errorf("name %q was unexpectedly rejected by regex: %v", name, err)
			}
		})
	}
}

