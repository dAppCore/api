// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"encoding/json"
	"iter"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	api "dappco.re/go/core/api"
)

// ── ExportSpec tests ─────────────────────────────────────────────────────

func TestExportSpec_Good_JSON(t *testing.T) {
	builder := &api.SpecBuilder{Title: "Test", Description: "Test API", Version: "1.0.0"}

	var buf bytes.Buffer
	if err := api.ExportSpec(&buf, "json", builder, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &spec); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if spec["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi=3.1.0, got %v", spec["openapi"])
	}

	info := spec["info"].(map[string]any)
	if info["title"] != "Test" {
		t.Fatalf("expected title=Test, got %v", info["title"])
	}
}

func TestExportSpec_Good_YAML(t *testing.T) {
	builder := &api.SpecBuilder{Title: "Test", Description: "Test API", Version: "1.0.0"}

	var buf bytes.Buffer
	if err := api.ExportSpec(&buf, "yaml", builder, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "openapi:") {
		t.Fatalf("expected YAML output to contain 'openapi:', got:\n%s", output)
	}

	var spec map[string]any
	if err := yaml.Unmarshal(buf.Bytes(), &spec); err != nil {
		t.Fatalf("output is not valid YAML: %v", err)
	}

	if spec["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi=3.1.0, got %v", spec["openapi"])
	}
}

func TestExportSpec_Bad_InvalidFormat(t *testing.T) {
	builder := &api.SpecBuilder{Title: "Test", Description: "Test API", Version: "1.0.0"}

	var buf bytes.Buffer
	err := api.ExportSpec(&buf, "xml", builder, nil)
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("expected error to contain 'unsupported format', got: %v", err)
	}
}

func TestExportSpecToFile_Good_CreatesFile(t *testing.T) {
	builder := &api.SpecBuilder{Title: "Test", Description: "Test API", Version: "1.0.0"}

	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "spec.json")

	if err := api.ExportSpecToFile(path, "json", builder, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("file content is not valid JSON: %v", err)
	}

	if spec["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi=3.1.0, got %v", spec["openapi"])
	}
}

func TestExportSpec_Good_WithToolBridge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	builder := &api.SpecBuilder{Title: "Test", Description: "Test API", Version: "1.0.0"}

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file",
		Group:       "files",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("ok"))
	})
	bridge.Add(api.ToolDescriptor{
		Name:        "metrics_query",
		Description: "Query metrics",
		Group:       "metrics",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("ok"))
	})

	var buf bytes.Buffer
	if err := api.ExportSpec(&buf, "json", builder, []api.RouteGroup{bridge}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "/tools/file_read") {
		t.Fatalf("expected output to contain /tools/file_read, got:\n%s", output)
	}
	if !strings.Contains(output, "/tools/metrics_query") {
		t.Fatalf("expected output to contain /tools/metrics_query, got:\n%s", output)
	}

	// Verify it's valid JSON.
	var spec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &spec); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify paths exist.
	paths := spec["paths"].(map[string]any)
	if _, ok := paths["/tools/file_read"]; !ok {
		t.Fatal("expected /tools/file_read path in spec")
	}
	if _, ok := paths["/tools/metrics_query"]; !ok {
		t.Fatal("expected /tools/metrics_query path in spec")
	}
}

func TestExportSpecIter_Good_WithGroupIterator(t *testing.T) {
	builder := &api.SpecBuilder{Title: "Test", Description: "Test API", Version: "1.0.0"}

	group := &specStubGroup{
		name:     "iter",
		basePath: "/iter",
		descs: []api.RouteDescription{
			{
				Method:  "GET",
				Path:    "/ping",
				Summary: "Ping iter group",
				Response: map[string]any{
					"type": "string",
				},
			},
		},
	}

	groups := iter.Seq[api.RouteGroup](func(yield func(api.RouteGroup) bool) {
		_ = yield(group)
	})

	var buf bytes.Buffer
	if err := api.ExportSpecIter(&buf, "json", builder, groups); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &spec); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	paths := spec["paths"].(map[string]any)
	if _, ok := paths["/iter/ping"]; !ok {
		t.Fatal("expected /iter/ping path in spec")
	}
}
