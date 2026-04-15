// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"testing"

	core "dappco.re/go/core"

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
