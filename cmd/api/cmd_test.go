// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"testing"

	"forge.lthn.ai/core/cli/pkg/cli"
)

func TestAPISpecCmd_Good_CommandStructure(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	apiCmd, _, err := root.Find([]string{"api"})
	if err != nil {
		t.Fatalf("api command not found: %v", err)
	}

	specCmd, _, err := apiCmd.Find([]string{"spec"})
	if err != nil {
		t.Fatalf("spec subcommand not found: %v", err)
	}
	if specCmd.Use != "spec" {
		t.Fatalf("expected Use=spec, got %s", specCmd.Use)
	}
}

func TestAPISpecCmd_Good_JSON(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	apiCmd, _, err := root.Find([]string{"api"})
	if err != nil {
		t.Fatalf("api command not found: %v", err)
	}

	specCmd, _, err := apiCmd.Find([]string{"spec"})
	if err != nil {
		t.Fatalf("spec subcommand not found: %v", err)
	}

	// Verify flags exist
	if specCmd.Flag("format") == nil {
		t.Fatal("expected --format flag on spec command")
	}
	if specCmd.Flag("output") == nil {
		t.Fatal("expected --output flag on spec command")
	}
	if specCmd.Flag("title") == nil {
		t.Fatal("expected --title flag on spec command")
	}
	if specCmd.Flag("version") == nil {
		t.Fatal("expected --version flag on spec command")
	}
}

func TestAPISDKCmd_Bad_NoLang(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	root.SetArgs([]string{"api", "sdk"})
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --lang not provided")
	}
}

func TestAPISDKCmd_Good_ValidatesLanguage(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	apiCmd, _, err := root.Find([]string{"api"})
	if err != nil {
		t.Fatalf("api command not found: %v", err)
	}

	sdkCmd, _, err := apiCmd.Find([]string{"sdk"})
	if err != nil {
		t.Fatalf("sdk subcommand not found: %v", err)
	}

	// Verify flags exist
	if sdkCmd.Flag("lang") == nil {
		t.Fatal("expected --lang flag on sdk command")
	}
	if sdkCmd.Flag("output") == nil {
		t.Fatal("expected --output flag on sdk command")
	}
	if sdkCmd.Flag("spec") == nil {
		t.Fatal("expected --spec flag on sdk command")
	}
	if sdkCmd.Flag("package") == nil {
		t.Fatal("expected --package flag on sdk command")
	}
}
