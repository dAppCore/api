// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"forge.lthn.ai/core/cli/pkg/cli"

	api "dappco.re/go/core/api"
)

type specCmdStubGroup struct{}

func (specCmdStubGroup) Name() string                       { return "registered" }
func (specCmdStubGroup) BasePath() string                   { return "/registered" }
func (specCmdStubGroup) RegisterRoutes(rg *gin.RouterGroup) {}
func (specCmdStubGroup) Describe() []api.RouteDescription {
	return []api.RouteDescription{
		{
			Method:  "GET",
			Path:    "/ping",
			Summary: "Ping registered group",
			Tags:    []string{"registered"},
			Response: map[string]any{
				"type": "string",
			},
		},
	}
}

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
	if specCmd.Flag("description") == nil {
		t.Fatal("expected --description flag on spec command")
	}
	if specCmd.Flag("version") == nil {
		t.Fatal("expected --version flag on spec command")
	}
	if specCmd.Flag("server") == nil {
		t.Fatal("expected --server flag on spec command")
	}
}

func TestAPISpecCmd_Good_CustomDescription(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{"api", "spec", "--description", "Custom API description", "--output", outputFile})
	root.SetErr(new(bytes.Buffer))

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("expected spec file to be written: %v", err)
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("expected valid JSON spec, got error: %v", err)
	}

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in generated spec")
	}
	if info["description"] != "Custom API description" {
		t.Fatalf("expected custom description, got %v", info["description"])
	}
}

func TestAPISpecCmd_Good_ServerFlagAddsServers(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{"api", "spec", "--server", "https://api.example.com, /, https://api.example.com, ", "--output", outputFile})
	root.SetErr(new(bytes.Buffer))

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("expected spec file to be written: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("expected valid JSON spec, got error: %v", err)
	}

	servers, ok := spec["servers"].([]any)
	if !ok {
		t.Fatalf("expected servers array in generated spec, got %T", spec["servers"])
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if servers[0].(map[string]any)["url"] != "https://api.example.com" {
		t.Fatalf("expected first server to be https://api.example.com, got %v", servers[0])
	}
	if servers[1].(map[string]any)["url"] != "/" {
		t.Fatalf("expected second server to be /, got %v", servers[1])
	}
}

func TestAPISpecCmd_Good_RegisteredSpecGroups(t *testing.T) {
	api.RegisterSpecGroups(specCmdStubGroup{})

	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{"api", "spec", "--output", outputFile})
	root.SetErr(new(bytes.Buffer))

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("expected spec file to be written: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("expected valid JSON spec, got error: %v", err)
	}

	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatalf("expected paths object in generated spec, got %T", spec["paths"])
	}

	if _, ok := paths["/registered/ping"]; !ok {
		t.Fatal("expected registered route group path in generated spec")
	}
}

func TestAPISDKCmd_Bad_EmptyLanguages(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	root.SetArgs([]string{"api", "sdk", "--lang", " , , "})
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --lang only contains empty values")
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
	if sdkCmd.Flag("title") == nil {
		t.Fatal("expected --title flag on sdk command")
	}
	if sdkCmd.Flag("description") == nil {
		t.Fatal("expected --description flag on sdk command")
	}
	if sdkCmd.Flag("version") == nil {
		t.Fatal("expected --version flag on sdk command")
	}
	if sdkCmd.Flag("server") == nil {
		t.Fatal("expected --server flag on sdk command")
	}
}

func TestAPISDKCmd_Good_TempSpecUsesMetadataFlags(t *testing.T) {
	snapshot := api.RegisteredSpecGroups()
	api.ResetSpecGroups()
	t.Cleanup(func() {
		api.ResetSpecGroups()
		api.RegisterSpecGroups(snapshot...)
	})

	api.RegisterSpecGroups(specCmdStubGroup{})

	builder := sdkSpecBuilder("Custom SDK API", "Custom SDK description", "9.9.9", "https://api.example.com, /, https://api.example.com")
	groups := sdkSpecGroups()

	outputFile := t.TempDir() + "/spec.json"
	if err := api.ExportSpecToFile(outputFile, "json", builder, groups); err != nil {
		t.Fatalf("unexpected error writing temp spec: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("expected spec file to be written: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("expected valid JSON spec, got error: %v", err)
	}

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in generated spec")
	}
	if info["title"] != "Custom SDK API" {
		t.Fatalf("expected custom title, got %v", info["title"])
	}
	if info["description"] != "Custom SDK description" {
		t.Fatalf("expected custom description, got %v", info["description"])
	}
	if info["version"] != "9.9.9" {
		t.Fatalf("expected custom version, got %v", info["version"])
	}

	servers, ok := spec["servers"].([]any)
	if !ok {
		t.Fatalf("expected servers array in generated spec, got %T", spec["servers"])
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if servers[0].(map[string]any)["url"] != "https://api.example.com" {
		t.Fatalf("expected first server to be https://api.example.com, got %v", servers[0])
	}
	if servers[1].(map[string]any)["url"] != "/" {
		t.Fatalf("expected second server to be /, got %v", servers[1])
	}
}
