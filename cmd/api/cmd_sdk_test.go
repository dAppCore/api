// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	core "dappco.re/go/core"
	"dappco.re/go/core/cli/pkg/cli"

	api "dappco.re/go/core/api"
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

	if err := cli.Init(cli.Options{AppName: "core-api-test"}); err != nil {
		t.Fatalf("failed to initialise CLI runtime: %v", err)
	}
	t.Cleanup(cli.Shutdown)

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
