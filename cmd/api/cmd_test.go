// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"encoding/json"
	"iter"
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

func collectRouteGroups(groups iter.Seq[api.RouteGroup]) []api.RouteGroup {
	out := make([]api.RouteGroup, 0)
	for group := range groups {
		out = append(out, group)
	}
	return out
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
	if specCmd.Flag("summary") == nil {
		t.Fatal("expected --summary flag on spec command")
	}
	if specCmd.Flag("description") == nil {
		t.Fatal("expected --description flag on spec command")
	}
	if specCmd.Flag("version") == nil {
		t.Fatal("expected --version flag on spec command")
	}
	if specCmd.Flag("swagger-path") == nil {
		t.Fatal("expected --swagger-path flag on spec command")
	}
	if specCmd.Flag("graphql-path") == nil {
		t.Fatal("expected --graphql-path flag on spec command")
	}
	if specCmd.Flag("graphql-playground") == nil {
		t.Fatal("expected --graphql-playground flag on spec command")
	}
	if specCmd.Flag("sse-path") == nil {
		t.Fatal("expected --sse-path flag on spec command")
	}
	if specCmd.Flag("ws-path") == nil {
		t.Fatal("expected --ws-path flag on spec command")
	}
	if specCmd.Flag("pprof") == nil {
		t.Fatal("expected --pprof flag on spec command")
	}
	if specCmd.Flag("expvar") == nil {
		t.Fatal("expected --expvar flag on spec command")
	}
	if specCmd.Flag("terms-of-service") == nil {
		t.Fatal("expected --terms-of-service flag on spec command")
	}
	if specCmd.Flag("contact-name") == nil {
		t.Fatal("expected --contact-name flag on spec command")
	}
	if specCmd.Flag("contact-url") == nil {
		t.Fatal("expected --contact-url flag on spec command")
	}
	if specCmd.Flag("contact-email") == nil {
		t.Fatal("expected --contact-email flag on spec command")
	}
	if specCmd.Flag("license-name") == nil {
		t.Fatal("expected --license-name flag on spec command")
	}
	if specCmd.Flag("license-url") == nil {
		t.Fatal("expected --license-url flag on spec command")
	}
	if specCmd.Flag("external-docs-description") == nil {
		t.Fatal("expected --external-docs-description flag on spec command")
	}
	if specCmd.Flag("external-docs-url") == nil {
		t.Fatal("expected --external-docs-url flag on spec command")
	}
	if specCmd.Flag("server") == nil {
		t.Fatal("expected --server flag on spec command")
	}
	if specCmd.Flag("security-schemes") == nil {
		t.Fatal("expected --security-schemes flag on spec command")
	}
}

func TestAPISpecCmd_Good_CustomDescription(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{"api", "spec", "--description", "Custom API description", "--swagger-path", "/docs", "--output", outputFile})
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

	if got := spec["x-swagger-ui-path"]; got != "/docs" {
		t.Fatalf("expected x-swagger-ui-path=/docs, got %v", got)
	}

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in generated spec")
	}
	if info["description"] != "Custom API description" {
		t.Fatalf("expected custom description, got %v", info["description"])
	}
}

func TestAPISpecCmd_Good_SummaryPopulatesSpecInfo(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--summary", "Short API overview",
		"--output", outputFile,
	})
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

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in generated spec")
	}
	if info["summary"] != "Short API overview" {
		t.Fatalf("expected summary to be preserved, got %v", info["summary"])
	}
}

func TestAPISpecCmd_Good_GraphQLPlaygroundFlagPopulatesSpecPaths(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--graphql-path", "/graphql",
		"--graphql-playground",
		"--output", outputFile,
	})
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
		t.Fatal("expected paths object in generated spec")
	}
	if _, ok := paths["/graphql/playground"]; !ok {
		t.Fatal("expected GraphQL playground path in generated spec")
	}
}

func TestAPISpecCmd_Good_ContactFlagsPopulateSpecInfo(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--contact-name", "API Support",
		"--contact-url", "https://example.com/support",
		"--contact-email", "support@example.com",
		"--output", outputFile,
	})
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

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in generated spec")
	}

	contact, ok := info["contact"].(map[string]any)
	if !ok {
		t.Fatal("expected contact metadata in generated spec")
	}
	if contact["name"] != "API Support" {
		t.Fatalf("expected contact name API Support, got %v", contact["name"])
	}
	if contact["url"] != "https://example.com/support" {
		t.Fatalf("expected contact url to be preserved, got %v", contact["url"])
	}
	if contact["email"] != "support@example.com" {
		t.Fatalf("expected contact email to be preserved, got %v", contact["email"])
	}
}

func TestAPISpecCmd_Good_SecuritySchemesFlagPopulatesSpecComponents(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--security-schemes", `{"apiKeyAuth":{"type":"apiKey","in":"header","name":"X-API-Key"}}`,
		"--output", outputFile,
	})
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

	securitySchemes, ok := spec["components"].(map[string]any)["securitySchemes"].(map[string]any)
	if !ok {
		t.Fatal("expected securitySchemes object in generated spec")
	}
	apiKeyAuth, ok := securitySchemes["apiKeyAuth"].(map[string]any)
	if !ok {
		t.Fatal("expected apiKeyAuth security scheme in generated spec")
	}
	if apiKeyAuth["type"] != "apiKey" {
		t.Fatalf("expected apiKeyAuth.type=apiKey, got %v", apiKeyAuth["type"])
	}
	if apiKeyAuth["in"] != "header" {
		t.Fatalf("expected apiKeyAuth.in=header, got %v", apiKeyAuth["in"])
	}
	if apiKeyAuth["name"] != "X-API-Key" {
		t.Fatalf("expected apiKeyAuth.name=X-API-Key, got %v", apiKeyAuth["name"])
	}
}

func TestSpecGroupsIter_Good_DeduplicatesExtraBridge(t *testing.T) {
	snapshot := api.RegisteredSpecGroups()
	api.ResetSpecGroups()
	t.Cleanup(func() {
		api.ResetSpecGroups()
		api.RegisterSpecGroups(snapshot...)
	})

	group := specCmdStubGroup{}
	api.RegisterSpecGroups(group)

	var groups []api.RouteGroup
	for g := range specGroupsIter(group) {
		groups = append(groups, g)
	}

	if len(groups) != 1 {
		t.Fatalf("expected duplicate extra group to be skipped, got %d groups", len(groups))
	}
	if groups[0].Name() != group.Name() || groups[0].BasePath() != group.BasePath() {
		t.Fatalf("expected original group to be preserved, got %s at %s", groups[0].Name(), groups[0].BasePath())
	}
}

func TestAPISpecCmd_Good_TermsOfServiceFlagPopulatesSpecInfo(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--terms-of-service", "https://example.com/terms",
		"--output", outputFile,
	})
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

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in generated spec")
	}
	if info["termsOfService"] != "https://example.com/terms" {
		t.Fatalf("expected termsOfService to be preserved, got %v", info["termsOfService"])
	}
}

func TestAPISpecCmd_Good_ExternalDocsFlagsPopulateSpec(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--external-docs-description", "Developer guide",
		"--external-docs-url", "https://example.com/docs",
		"--output", outputFile,
	})
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

	externalDocs, ok := spec["externalDocs"].(map[string]any)
	if !ok {
		t.Fatal("expected externalDocs metadata in generated spec")
	}
	if externalDocs["description"] != "Developer guide" {
		t.Fatalf("expected externalDocs description Developer guide, got %v", externalDocs["description"])
	}
	if externalDocs["url"] != "https://example.com/docs" {
		t.Fatalf("expected externalDocs url to be preserved, got %v", externalDocs["url"])
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

func TestAPISpecCmd_Good_LicenseFlagsPopulateSpecInfo(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--license-name", "EUPL-1.2",
		"--license-url", "https://eupl.eu/1.2/en/",
		"--output", outputFile,
	})
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

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in generated spec")
	}

	license, ok := info["license"].(map[string]any)
	if !ok {
		t.Fatal("expected license metadata in generated spec")
	}
	if license["name"] != "EUPL-1.2" {
		t.Fatalf("expected licence name EUPL-1.2, got %v", license["name"])
	}
	if license["url"] != "https://eupl.eu/1.2/en/" {
		t.Fatalf("expected licence url to be preserved, got %v", license["url"])
	}
}

func TestAPISpecCmd_Good_GraphQLPathPopulatesSpec(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--graphql-path", "/gql",
		"--output", outputFile,
	})
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

	if _, ok := paths["/gql"]; !ok {
		t.Fatal("expected GraphQL path to be included in generated spec")
	}
}

func TestAPISpecCmd_Good_SSEPathPopulatesSpec(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--sse-path", "/events",
		"--output", outputFile,
	})
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

	if _, ok := paths["/events"]; !ok {
		t.Fatal("expected SSE path to be included in generated spec")
	}
}

func TestAPISpecCmd_Good_RuntimePathsPopulatedSpec(t *testing.T) {
	root := &cli.Command{Use: "root"}
	AddAPICommands(root)

	outputFile := t.TempDir() + "/spec.json"
	root.SetArgs([]string{
		"api", "spec",
		"--ws-path", "/ws",
		"--pprof",
		"--expvar",
		"--output", outputFile,
	})
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

	if _, ok := paths["/ws"]; !ok {
		t.Fatal("expected WebSocket path to be included in generated spec")
	}
	if _, ok := paths["/debug/pprof"]; !ok {
		t.Fatal("expected pprof path to be included in generated spec")
	}
	if _, ok := paths["/debug/vars"]; !ok {
		t.Fatal("expected expvar path to be included in generated spec")
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
	if sdkCmd.Flag("swagger-path") == nil {
		t.Fatal("expected --swagger-path flag on sdk command")
	}
	if sdkCmd.Flag("graphql-path") == nil {
		t.Fatal("expected --graphql-path flag on sdk command")
	}
	if sdkCmd.Flag("sse-path") == nil {
		t.Fatal("expected --sse-path flag on sdk command")
	}
	if sdkCmd.Flag("ws-path") == nil {
		t.Fatal("expected --ws-path flag on sdk command")
	}
	if sdkCmd.Flag("pprof") == nil {
		t.Fatal("expected --pprof flag on sdk command")
	}
	if sdkCmd.Flag("expvar") == nil {
		t.Fatal("expected --expvar flag on sdk command")
	}
	if sdkCmd.Flag("terms-of-service") == nil {
		t.Fatal("expected --terms-of-service flag on sdk command")
	}
	if sdkCmd.Flag("contact-name") == nil {
		t.Fatal("expected --contact-name flag on sdk command")
	}
	if sdkCmd.Flag("contact-url") == nil {
		t.Fatal("expected --contact-url flag on sdk command")
	}
	if sdkCmd.Flag("contact-email") == nil {
		t.Fatal("expected --contact-email flag on sdk command")
	}
	if sdkCmd.Flag("license-name") == nil {
		t.Fatal("expected --license-name flag on sdk command")
	}
	if sdkCmd.Flag("license-url") == nil {
		t.Fatal("expected --license-url flag on sdk command")
	}
	if sdkCmd.Flag("server") == nil {
		t.Fatal("expected --server flag on sdk command")
	}
	if sdkCmd.Flag("security-schemes") == nil {
		t.Fatal("expected --security-schemes flag on sdk command")
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

	builder, err := sdkSpecBuilder(specBuilderConfig{
		title:             "Custom SDK API",
		summary:           "Custom SDK overview",
		description:       "Custom SDK description",
		version:           "9.9.9",
		swaggerPath:       "/docs",
		graphqlPath:       "/gql",
		graphqlPlayground: true,
		ssePath:           "/events",
		wsPath:            "/ws",
		pprofEnabled:      true,
		expvarEnabled:     true,
		termsURL:          "https://example.com/terms",
		contactName:       "SDK Support",
		contactURL:        "https://example.com/support",
		contactEmail:      "support@example.com",
		licenseName:       "EUPL-1.2",
		licenseURL:        "https://eupl.eu/1.2/en/",
		servers:           "https://api.example.com, /, https://api.example.com",
		securitySchemes:   `{"apiKeyAuth":{"type":"apiKey","in":"header","name":"X-API-Key"}}`,
	})
	if err != nil {
		t.Fatalf("unexpected error building sdk spec: %v", err)
	}
	groups := collectRouteGroups(sdkSpecGroupsIter())

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
	if info["summary"] != "Custom SDK overview" {
		t.Fatalf("expected custom summary, got %v", info["summary"])
	}
	if info["version"] != "9.9.9" {
		t.Fatalf("expected custom version, got %v", info["version"])
	}

	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatalf("expected paths object in generated spec, got %T", spec["paths"])
	}
	if _, ok := paths["/gql"]; !ok {
		t.Fatal("expected GraphQL path to be included in generated spec")
	}
	if got := builder.SwaggerPath; got != "/docs" {
		t.Fatalf("expected swagger path to be preserved in sdk spec builder, got %v", got)
	}
	if _, ok := paths["/gql/playground"]; !ok {
		t.Fatal("expected GraphQL playground path to be included in generated spec")
	}
	if _, ok := paths["/events"]; !ok {
		t.Fatal("expected SSE path to be included in generated spec")
	}
	if _, ok := paths["/ws"]; !ok {
		t.Fatal("expected WebSocket path to be included in generated spec")
	}
	if _, ok := paths["/debug/pprof"]; !ok {
		t.Fatal("expected pprof path to be included in generated spec")
	}
	if _, ok := paths["/debug/vars"]; !ok {
		t.Fatal("expected expvar path to be included in generated spec")
	}

	if info["termsOfService"] != "https://example.com/terms" {
		t.Fatalf("expected termsOfService to be preserved, got %v", info["termsOfService"])
	}

	contact, ok := info["contact"].(map[string]any)
	if !ok {
		t.Fatal("expected contact metadata in generated spec")
	}
	if contact["name"] != "SDK Support" {
		t.Fatalf("expected contact name SDK Support, got %v", contact["name"])
	}
	if contact["url"] != "https://example.com/support" {
		t.Fatalf("expected contact url to be preserved, got %v", contact["url"])
	}
	if contact["email"] != "support@example.com" {
		t.Fatalf("expected contact email to be preserved, got %v", contact["email"])
	}

	license, ok := info["license"].(map[string]any)
	if !ok {
		t.Fatal("expected licence metadata in generated spec")
	}
	if license["name"] != "EUPL-1.2" {
		t.Fatalf("expected licence name EUPL-1.2, got %v", license["name"])
	}
	if license["url"] != "https://eupl.eu/1.2/en/" {
		t.Fatalf("expected licence url to be preserved, got %v", license["url"])
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

	securitySchemes, ok := spec["components"].(map[string]any)["securitySchemes"].(map[string]any)
	if !ok {
		t.Fatal("expected securitySchemes in generated spec")
	}
	if _, ok := securitySchemes["apiKeyAuth"].(map[string]any); !ok {
		t.Fatalf("expected apiKeyAuth security scheme in generated spec, got %v", securitySchemes)
	}
}

func TestAPISDKCmd_Good_SpecGroupsDeduplicateToolBridge(t *testing.T) {
	snapshot := api.RegisteredSpecGroups()
	api.ResetSpecGroups()
	t.Cleanup(func() {
		api.ResetSpecGroups()
		api.RegisterSpecGroups(snapshot...)
	})

	api.RegisterSpecGroups(api.NewToolBridge("/tools"))

	groups := collectRouteGroups(sdkSpecGroupsIter())
	if len(groups) != 1 {
		t.Fatalf("expected the built-in tools bridge to be deduplicated, got %d groups", len(groups))
	}
	if groups[0].BasePath() != "/tools" {
		t.Fatalf("expected the remaining group to be /tools, got %s", groups[0].BasePath())
	}
}
