// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"dappco.re/go/cli/pkg/cli"
	core "dappco.re/go/core"

	api "dappco.re/go/api"
)

// TestCmdSdk_AddSDKCommand_Good verifies the sdk command registers under
// the expected api/sdk path with an executable Action.
func TestCmdSdk_AddSDKCommand_Good(t *testing.T) {
	c := core.New()
	addSDKCommand(c)

	for _, path := range []string{"api/sdk", "build/sdk"} {
		r := c.Command(path)
		if !r.OK {
			t.Fatalf("expected %s command to be registered", path)
		}
		cmd, ok := r.Value.(*core.Command)
		if !ok {
			t.Fatalf("expected *core.Command for %s, got %T", path, r.Value)
		}
		if cmd.Action == nil {
			t.Fatalf("expected non-nil Action on %s", path)
		}
		if cmd.Description == "" {
			t.Fatalf("expected Description on %s", path)
		}
	}
}

// TestCmdSdk_SdkAction_Bad_RequiresLanguage rejects invocations that omit
// the --lang flag entirely.
func TestCmdSdk_SdkAction_Bad_RequiresLanguage(t *testing.T) {
	opts := core.NewOptions()
	r := sdkAction(opts)
	if r.OK {
		t.Fatal("expected sdk action to fail without --lang")
	}
}

// TestCmdSdk_SdkAction_Bad_EmptyLanguageList rejects --lang values that
// resolve to no real language entries after splitting and trimming.
func TestCmdSdk_SdkAction_Bad_EmptyLanguageList(t *testing.T) {
	opts := core.NewOptions(
		core.Option{Key: "lang", Value: " , , "},
	)
	r := sdkAction(opts)
	if r.OK {
		t.Fatal("expected sdk action to fail when --lang only contains separators")
	}
}

// TestCmdSdk_SdkAction_Good_InvokesGeneratorForUniqueLanguages verifies the
// happy path using a fake openapi-generator-cli binary on PATH so the action
// can be exercised deterministically without external dependencies.
func TestCmdSdk_SdkAction_Good_InvokesGeneratorForUniqueLanguages(t *testing.T) {
	workDir := t.TempDir()

	binDir := filepath.Join(workDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("failed to create fake bin dir: %v", err)
	}

	logFile := filepath.Join(workDir, "generator-args.log")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$SDK_ACTION_LOG\"\nexit 0\n"
	if err := os.WriteFile(filepath.Join(binDir, "openapi-generator-cli"), []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake generator: %v", err)
	}

	path := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+path)
	t.Setenv("SDK_ACTION_LOG", logFile)

	initSDKActionCLITest(t)

	opts := core.NewOptions(
		core.Option{Key: "lang", Value: " go , python , go "},
		core.Option{Key: "output", Value: filepath.Join(workDir, "sdk")},
	)

	r := sdkAction(opts)
	if !r.OK {
		t.Fatalf("expected sdk action to succeed, got %v", r.Value)
	}

	for _, lang := range []string{"go", "python"} {
		if _, err := os.Stat(filepath.Join(workDir, "sdk", lang)); err != nil {
			t.Fatalf("expected output directory for %s: %v", lang, err)
		}
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected generator log to exist: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 generator invocations, got %d: %q", len(lines), string(data))
	}
	if !strings.Contains(lines[0], "-g go") || !strings.Contains(lines[0], "packageName=lethean") {
		t.Fatalf("expected default package name and go generator in first invocation, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "-g python") || !strings.Contains(lines[1], "packageName=lethean") {
		t.Fatalf("expected default package name and python generator in second invocation, got %q", lines[1])
	}
}

// TestCmdSdk_TempFile_Bad_PreExistingSymlink verifies the generated spec uses
// exclusive temp-file creation instead of the legacy predictable core.ID path.
func TestCmdSdk_TempFile_Bad_PreExistingSymlink(t *testing.T) {
	workDir := t.TempDir()
	tmpDir := filepath.Join(workDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	targetPath := filepath.Join(workDir, "do-not-delete.json")
	if err := os.WriteFile(targetPath, []byte("sentinel"), 0o600); err != nil {
		t.Fatalf("failed to write symlink target: %v", err)
	}

	legacyPath := filepath.Join(tmpDir, "openapi-id-1-deadbe.json")
	if err := os.Symlink(targetPath, legacyPath); err != nil {
		t.Fatalf("failed to create pre-existing symlink: %v", err)
	}

	binDir := filepath.Join(workDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("failed to create fake bin dir: %v", err)
	}

	specLog := filepath.Join(workDir, "spec-path.log")
	script := "#!/bin/sh\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"-i\" ]; then\n    shift\n    if [ -L \"$1\" ]; then exit 2; fi\n    if [ ! -f \"$1\" ]; then exit 3; fi\n    printf '%s\\n' \"$1\" > \"$SDK_SPEC_LOG\"\n    exit 0\n  fi\n  shift\ndone\nexit 1\n"
	if err := os.WriteFile(filepath.Join(binDir, "openapi-generator-cli"), []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake generator: %v", err)
	}

	path := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+path)
	t.Setenv("SDK_SPEC_LOG", specLog)
	t.Setenv("TMPDIR", tmpDir)

	initSDKActionCLITest(t)

	opts := core.NewOptions(
		core.Option{Key: "lang", Value: "go"},
		core.Option{Key: "output", Value: filepath.Join(workDir, "sdk")},
	)

	r := sdkAction(opts)
	if !r.OK {
		t.Fatalf("expected sdk action to succeed, got %v", r.Value)
	}

	data, err := os.ReadFile(specLog)
	if err != nil {
		t.Fatalf("expected generator spec log to exist: %v", err)
	}
	specPath := strings.TrimSpace(string(data))
	if specPath == legacyPath {
		t.Fatal("expected generated spec path not to reuse pre-existing symlink")
	}
	if !strings.HasPrefix(specPath, tmpDir+string(os.PathSeparator)+"openapi-") {
		t.Fatalf("expected temp spec under %s, got %q", tmpDir, specPath)
	}
	if !strings.HasSuffix(specPath, ".json") {
		t.Fatalf("expected temp spec to keep .json suffix, got %q", specPath)
	}
	if _, err := os.Lstat(specPath); !os.IsNotExist(err) {
		t.Fatalf("expected temp spec to be deleted after sdk action, got %v", err)
	}

	info, err := os.Lstat(legacyPath)
	if err != nil {
		t.Fatalf("expected pre-existing symlink to remain: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected %s to remain a symlink, got mode %s", legacyPath, info.Mode())
	}
	contents, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("expected symlink target to remain readable: %v", err)
	}
	if string(contents) != "sentinel" {
		t.Fatalf("expected symlink target contents to remain unchanged, got %q", string(contents))
	}
}

func initSDKActionCLITest(t *testing.T) {
	t.Helper()
	// Shutdown cancels the package-global context without clearing it, so these
	// SDK action tests leave the test runtime initialized for the process.
	if err := cli.Init(cli.Options{AppName: "core-api-test"}); err != nil {
		t.Fatalf("failed to initialise CLI runtime: %v", err)
	}
}

// TestCmdSdk_SdkSpecGroupsIter_Good_IncludesToolBridge verifies the SDK
// builder always exposes the bundled tools route group.
func TestCmdSdk_SdkSpecGroupsIter_Good_IncludesToolBridge(t *testing.T) {
	snapshot := api.RegisteredSpecGroups()
	api.ResetSpecGroups()
	t.Cleanup(func() {
		api.ResetSpecGroups()
		api.RegisterSpecGroups(snapshot...)
	})

	groups := collectRouteGroups(sdkSpecGroupsIter())
	if len(groups) == 0 {
		t.Fatal("expected at least the bundled tools bridge")
	}

	found := false
	for _, g := range groups {
		if g.BasePath() == defaultSpecToolBridgePath {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected %s route group in sdk spec iterator", defaultSpecToolBridgePath)
	}
}

// TestCmdSdk_SdkSpecGroupsIter_Good_PopulatesBundledToolDescriptors verifies
// that the synthesized SDK bridge mirrors registered tool descriptors.
func TestCmdSdk_SdkSpecGroupsIter_Good_PopulatesBundledToolDescriptors(t *testing.T) {
	snapshot := api.RegisteredSpecGroups()
	api.ResetSpecGroups()
	t.Cleanup(func() {
		api.ResetSpecGroups()
		api.RegisterSpecGroups(snapshot...)
	})

	source := api.NewToolBridge("/source-tools")
	source.Add(api.ToolDescriptor{
		Name:        "ping",
		Description: "Ping tool",
		Group:       "system",
	}, func(*gin.Context) {})
	api.RegisterSpecGroups(source)

	groups := collectRouteGroups(sdkSpecGroupsIter())
	for _, g := range groups {
		if g.BasePath() != defaultSpecToolBridgePath {
			continue
		}

		toolSource, ok := g.(interface {
			Tools() []api.ToolDescriptor
		})
		if !ok {
			t.Fatalf("expected bundled bridge to expose tools, got %T", g)
		}

		tools := toolSource.Tools()
		if len(tools) != 1 {
			t.Fatalf("expected 1 bundled tool descriptor, got %d", len(tools))
		}
		if tools[0].Name != "ping" {
			t.Fatalf("expected bundled tool descriptor to be copied from registry, got %q", tools[0].Name)
		}
		return
	}

	t.Fatalf("expected %s route group in sdk spec iterator", defaultSpecToolBridgePath)
}

// TestCmdSdk_SdkConfigFromOptions_Ugly_FallsBackToSDKDefaults exercises the
// SDK-specific default fallbacks for empty title/description/version.
func TestCmdSdk_SdkConfigFromOptions_Ugly_FallsBackToSDKDefaults(t *testing.T) {
	opts := core.NewOptions()
	cfg := sdkConfigFromOptions(opts)
	if cfg.title != defaultSDKTitle {
		t.Fatalf("expected default SDK title, got %q", cfg.title)
	}
	if cfg.description != defaultSDKDescription {
		t.Fatalf("expected default SDK description, got %q", cfg.description)
	}
	if cfg.version != defaultSDKVersion {
		t.Fatalf("expected default SDK version, got %q", cfg.version)
	}
}
