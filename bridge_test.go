// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── ToolBridge ─────────────────────────────────────────────────────────

func TestToolBridge_Good_RegisterAndServe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file",
		Group:       "files",
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("result1"))
	})
	bridge.Add(api.ToolDescriptor{
		Name:        "file_write",
		Description: "Write a file",
		Group:       "files",
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("result2"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	// POST /tools/file_read
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/tools/file_read", nil)
	engine.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 for file_read, got %d", w1.Code)
	}
	var resp1 api.Response[string]
	if err := json.Unmarshal(w1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp1.Data != "result1" {
		t.Fatalf("expected Data=%q, got %q", "result1", resp1.Data)
	}

	// POST /tools/file_write
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/tools/file_write", nil)
	engine.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for file_write, got %d", w2.Code)
	}
	var resp2 api.Response[string]
	if err := json.Unmarshal(w2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp2.Data != "result2" {
		t.Fatalf("expected Data=%q, got %q", "result2", resp2.Data)
	}
}

func TestToolBridge_Good_BasePath(t *testing.T) {
	bridge := api.NewToolBridge("/api/v1/tools")

	if bridge.BasePath() != "/api/v1/tools" {
		t.Fatalf("expected BasePath=%q, got %q", "/api/v1/tools", bridge.BasePath())
	}
	if bridge.Name() != "tools" {
		t.Fatalf("expected Name=%q, got %q", "tools", bridge.Name())
	}
}

func TestToolBridge_Good_Describe(t *testing.T) {
	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file from disk",
		Group:       "files",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"content": map[string]any{"type": "string"},
			},
		},
	}, func(c *gin.Context) {})
	bridge.Add(api.ToolDescriptor{
		Name:        "metrics_query",
		Description: "Query metrics data",
		Group:       "metrics",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		},
	}, func(c *gin.Context) {})

	// Verify DescribableGroup interface satisfaction.
	var dg api.DescribableGroup = bridge
	descs := dg.Describe()

	if len(descs) != 2 {
		t.Fatalf("expected 2 descriptions, got %d", len(descs))
	}

	// First tool.
	if descs[0].Method != "POST" {
		t.Fatalf("expected descs[0].Method=%q, got %q", "POST", descs[0].Method)
	}
	if descs[0].Path != "/file_read" {
		t.Fatalf("expected descs[0].Path=%q, got %q", "/file_read", descs[0].Path)
	}
	if descs[0].Summary != "Read a file from disk" {
		t.Fatalf("expected descs[0].Summary=%q, got %q", "Read a file from disk", descs[0].Summary)
	}
	if len(descs[0].Tags) != 1 || descs[0].Tags[0] != "files" {
		t.Fatalf("expected descs[0].Tags=[files], got %v", descs[0].Tags)
	}
	if descs[0].RequestBody == nil {
		t.Fatal("expected descs[0].RequestBody to be non-nil")
	}
	if descs[0].Response == nil {
		t.Fatal("expected descs[0].Response to be non-nil")
	}

	// Second tool.
	if descs[1].Path != "/metrics_query" {
		t.Fatalf("expected descs[1].Path=%q, got %q", "/metrics_query", descs[1].Path)
	}
	if len(descs[1].Tags) != 1 || descs[1].Tags[0] != "metrics" {
		t.Fatalf("expected descs[1].Tags=[metrics], got %v", descs[1].Tags)
	}
	if descs[1].Response != nil {
		t.Fatalf("expected descs[1].Response to be nil, got %v", descs[1].Response)
	}
}

func TestToolBridge_Good_ValidatesRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file from disk",
		Group:       "files",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
			"required": []any{"path"},
		},
	}, func(c *gin.Context) {
		var payload map[string]any
		if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
			t.Fatalf("handler could not read validated body: %v", err)
		}
		c.JSON(http.StatusOK, api.OK(payload["path"]))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/file_read", bytes.NewBufferString(`{"path":"/tmp/file.txt"}`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Data != "/tmp/file.txt" {
		t.Fatalf("expected validated payload to reach handler, got %q", resp.Data)
	}
}

func TestToolBridge_Good_ValidatesResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file from disk",
		Group:       "files",
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
			"required": []any{"path"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK(map[string]any{"path": "/tmp/file.txt"}))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/file_read", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[map[string]any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success=true")
	}
	if resp.Data["path"] != "/tmp/file.txt" {
		t.Fatalf("expected validated response data to reach client, got %v", resp.Data["path"])
	}
}

func TestToolBridge_Bad_InvalidResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file from disk",
		Group:       "files",
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
			"required": []any{"path"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK(map[string]any{"path": 123}))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/file_read", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected Success=false")
	}
	if resp.Error == nil || resp.Error.Code != "invalid_tool_response" {
		t.Fatalf("expected invalid_tool_response error, got %#v", resp.Error)
	}
}

func TestToolBridge_Bad_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file from disk",
		Group:       "files",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
			"required": []any{"path"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("should not run"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/file_read", bytes.NewBufferString(`{"path":123}`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected Success=false")
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request_body" {
		t.Fatalf("expected invalid_request_body error, got %#v", resp.Error)
	}
}

func TestToolBridge_Good_ToolsAccessor(t *testing.T) {
	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{Name: "alpha", Description: "Tool A", Group: "a"}, func(c *gin.Context) {})
	bridge.Add(api.ToolDescriptor{Name: "beta", Description: "Tool B", Group: "b"}, func(c *gin.Context) {})
	bridge.Add(api.ToolDescriptor{Name: "gamma", Description: "Tool C", Group: "c"}, func(c *gin.Context) {})

	tools := bridge.Tools()
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}

	expected := []string{"alpha", "beta", "gamma"}
	for i, want := range expected {
		if tools[i].Name != want {
			t.Fatalf("expected tools[%d].Name=%q, got %q", i, want, tools[i].Name)
		}
	}
}

func TestToolBridge_Bad_EmptyBridge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	bridge := api.NewToolBridge("/tools")

	// RegisterRoutes should not panic with no tools.
	engine := gin.New()
	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	// Describe should return empty slice.
	descs := bridge.Describe()
	if len(descs) != 0 {
		t.Fatalf("expected 0 descriptions, got %d", len(descs))
	}

	// Tools should return empty slice.
	tools := bridge.Tools()
	if len(tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(tools))
	}
}

func TestToolBridge_Good_IntegrationWithEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, err := api.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "ping",
		Description: "Ping tool",
		Group:       "util",
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("pong"))
	})

	e.Register(bridge)

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success=true")
	}
	if resp.Data != "pong" {
		t.Fatalf("expected Data=%q, got %q", "pong", resp.Data)
	}
}
