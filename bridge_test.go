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

func TestToolBridge_Good_NormalisesConfiguredBasePath(t *testing.T) {
	bridge := api.NewToolBridge(" /api/v1/tools/ ")

	if bridge.BasePath() != "/api/v1/tools" {
		t.Fatalf("expected BasePath=%q, got %q", "/api/v1/tools", bridge.BasePath())
	}
}

func TestToolBridge_Bad_BlankBasePathFallsBackToRoot(t *testing.T) {
	bridge := api.NewToolBridge("   ")

	if bridge.BasePath() != "/" {
		t.Fatalf("expected blank base path to fall back to %q, got %q", "/", bridge.BasePath())
	}
}

func TestToolBridge_Ugly_CollapsesRepeatedSlashesInBasePath(t *testing.T) {
	bridge := api.NewToolBridge("///mcp///")

	if bridge.BasePath() != "/mcp" {
		t.Fatalf("expected repeated slashes to normalise to %q, got %q", "/mcp", bridge.BasePath())
	}
}

func TestToolBridge_Ugly_RootBasePathFallsBackToRoot(t *testing.T) {
	bridge := api.NewToolBridge(" / ")

	if bridge.BasePath() != "/" {
		t.Fatalf("expected root base path to normalise to %q, got %q", "/", bridge.BasePath())
	}
}

func TestToolBridge_Bad_RejectsUnsafeToolNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	bridge := api.NewToolBridge("/tools")

	defer func() {
		if recover() == nil {
			t.Fatal("expected Add to reject an unsafe tool name")
		}
	}()

	bridge.Add(api.ToolDescriptor{
		Name:        "../health",
		Description: "Invalid tool name",
	}, func(c *gin.Context) {})
}

func TestToolBridge_Good_AcceptsSafeToolNames(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []string{
		"ping",
		"file.read",
		"file-read",
		"file_read",
		"A1",
	}

	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			engine := gin.New()
			bridge := api.NewToolBridge("/tools")
			bridge.Add(api.ToolDescriptor{
				Name:        name,
				Description: "Safe tool name",
			}, func(c *gin.Context) {
				c.JSON(http.StatusOK, api.OK(name))
			})

			rg := engine.Group(bridge.BasePath())
			bridge.RegisterRoutes(rg)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/tools/"+name, nil)
			engine.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200 for %q, got %d", name, w.Code)
			}
		})
	}
}

func TestToolBridge_Ugly_RejectsUnsafeToolNameForms(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []string{
		"",
		" ",
		"../health",
		"foo/bar",
		"foo\\bar",
		"foo*bar",
		"foo?bar",
		"foo..bar",
	}

	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			bridge := api.NewToolBridge("/tools")

			defer func() {
				if recover() == nil {
					t.Fatalf("expected Add to reject tool name %q", name)
				}
			}()

			bridge.Add(api.ToolDescriptor{
				Name:        name,
				Description: "Invalid tool name",
			}, func(c *gin.Context) {})
		})
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

	// Describe() returns the GET tool listing entry followed by every tool.
	if len(descs) != 3 {
		t.Fatalf("expected 3 descriptions (listing + 2 tools), got %d", len(descs))
	}

	// Listing entry mirrors RFC.endpoints.md — GET /v1/tools returns the catalogue.
	if descs[0].Method != "GET" {
		t.Fatalf("expected descs[0].Method=%q, got %q", "GET", descs[0].Method)
	}
	if descs[0].Path != "/" {
		t.Fatalf("expected descs[0].Path=%q, got %q", "/", descs[0].Path)
	}

	// First tool.
	if descs[1].Method != "POST" {
		t.Fatalf("expected descs[1].Method=%q, got %q", "POST", descs[1].Method)
	}
	if descs[1].Path != "/file_read" {
		t.Fatalf("expected descs[1].Path=%q, got %q", "/file_read", descs[1].Path)
	}
	if descs[1].Summary != "Read a file from disk" {
		t.Fatalf("expected descs[1].Summary=%q, got %q", "Read a file from disk", descs[1].Summary)
	}
	if len(descs[1].Tags) != 1 || descs[1].Tags[0] != "files" {
		t.Fatalf("expected descs[1].Tags=[files], got %v", descs[1].Tags)
	}
	if descs[1].RequestBody == nil {
		t.Fatal("expected descs[1].RequestBody to be non-nil")
	}
	if descs[1].Response == nil {
		t.Fatal("expected descs[1].Response to be non-nil")
	}

	// Second tool.
	if descs[2].Path != "/metrics_query" {
		t.Fatalf("expected descs[2].Path=%q, got %q", "/metrics_query", descs[2].Path)
	}
	if len(descs[2].Tags) != 1 || descs[2].Tags[0] != "metrics" {
		t.Fatalf("expected descs[2].Tags=[metrics], got %v", descs[2].Tags)
	}
	if descs[2].Response != nil {
		t.Fatalf("expected descs[2].Response to be nil, got %v", descs[2].Response)
	}
}

func TestToolBridge_Good_DescribeTrimsBlankGroup(t *testing.T) {
	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "file_read",
		Description: "Read a file from disk",
		Group:       "   ",
	}, func(c *gin.Context) {})

	descs := bridge.Describe()
	// Describe() returns the GET listing plus one tool description.
	if len(descs) != 2 {
		t.Fatalf("expected 2 descriptions (listing + tool), got %d", len(descs))
	}
	if len(descs[1].Tags) != 1 || descs[1].Tags[0] != "tools" {
		t.Fatalf("expected blank group to fall back to bridge tag, got %v", descs[1].Tags)
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
	req, _ := http.NewRequest(http.MethodPost, "/tools/file_read", bytes.NewBufferString(""))
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

func TestToolBridge_Bad_RejectsWhitespaceOnlyRequestBody(t *testing.T) {
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
	req, _ := http.NewRequest(http.MethodPost, "/tools/file_read", bytes.NewBufferString("   "))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for whitespace-only request body, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request_body" {
		t.Fatalf("expected invalid_request_body error, got %#v", resp.Error)
	}
}

func TestToolBridge_Ugly_RejectsMalformedJSONRequestBody(t *testing.T) {
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
	req, _ := http.NewRequest(http.MethodPost, "/tools/file_read", bytes.NewBufferString(`{"path":`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request_body" {
		t.Fatalf("expected invalid_request_body error, got %#v", resp.Error)
	}
}

func TestToolBridge_Ugly_RejectsOversizedRequestBody(t *testing.T) {
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
	req, _ := http.NewRequest(http.MethodPost, "/tools/file_read", bytes.NewBuffer(bytes.Repeat([]byte("a"), 10<<20+1)))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized request body, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request_body" {
		t.Fatalf("expected invalid_request_body error, got %#v", resp.Error)
	}
}

func TestToolBridge_Good_ValidatesEnumValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "publish_item",
		Description: "Publish an item",
		Group:       "items",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{
					"type": "string",
					"enum": []any{"draft", "published"},
				},
			},
			"required": []any{"status"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("published"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/publish_item", bytes.NewBufferString(`{"status":"published"}`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestToolBridge_Bad_RejectsInvalidEnumValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "publish_item",
		Description: "Publish an item",
		Group:       "items",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{
					"type": "string",
					"enum": []any{"draft", "published"},
				},
			},
			"required": []any{"status"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("published"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/publish_item", bytes.NewBufferString(`{"status":"archived"}`))
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

func TestToolBridge_Good_ValidatesSchemaCombinators(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "route_choice",
		Description: "Choose a route",
		Group:       "items",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"choice": map[string]any{
					"oneOf": []any{
						map[string]any{
							"type": "string",
							"allOf": []any{
								map[string]any{"minLength": 2},
								map[string]any{"pattern": "^[A-Z]+$"},
							},
						},
						map[string]any{
							"type":    "string",
							"pattern": "^A",
						},
					},
				},
			},
			"required": []any{"choice"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("accepted"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/route_choice", bytes.NewBufferString(`{"choice":"BC"}`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestToolBridge_Bad_RejectsAmbiguousOneOfMatches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "route_choice",
		Description: "Choose a route",
		Group:       "items",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"choice": map[string]any{
					"oneOf": []any{
						map[string]any{
							"type": "string",
							"allOf": []any{
								map[string]any{"minLength": 1},
								map[string]any{"pattern": "^[A-Z]+$"},
							},
						},
						map[string]any{
							"type":    "string",
							"pattern": "^A",
						},
					},
				},
			},
			"required": []any{"choice"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("accepted"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/route_choice", bytes.NewBufferString(`{"choice":"A"}`))
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

func TestToolBridge_Bad_RejectsAdditionalProperties(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "publish_item",
		Description: "Publish an item",
		Group:       "items",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{"type": "string"},
			},
			"required":             []any{"status"},
			"additionalProperties": false,
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("published"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/publish_item", bytes.NewBufferString(`{"status":"published","unexpected":true}`))
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

func TestToolBridge_Good_EnforcesStringConstraints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "publish_code",
		Description: "Publish a code",
		Group:       "items",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"code": map[string]any{
					"type":      "string",
					"minLength": 3,
					"maxLength": 5,
					"pattern":   "^[A-Z]+$",
				},
			},
			"required": []any{"code"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("accepted"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/publish_code", bytes.NewBufferString(`{"code":"ABC"}`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestToolBridge_Bad_RejectsNumericAndCollectionConstraints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "quota_check",
		Description: "Check quotas",
		Group:       "items",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"count": map[string]any{
					"type":    "integer",
					"minimum": 1,
					"maximum": 3,
				},
				"labels": map[string]any{
					"type":     "array",
					"minItems": 2,
					"maxItems": 4,
					"items": map[string]any{
						"type": "string",
					},
				},
				"payload": map[string]any{
					"type":                 "object",
					"minProperties":        1,
					"maxProperties":        2,
					"additionalProperties": true,
				},
			},
			"required": []any{"count", "labels", "payload"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("accepted"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/quota_check", bytes.NewBufferString(`{"count":0,"labels":["one"],"payload":{}}`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for numeric/collection constraint failure, got %d", w.Code)
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

	// Describe should return only the GET listing entry when no tools are registered.
	descs := bridge.Describe()
	if len(descs) != 1 {
		t.Fatalf("expected 1 description (tool listing), got %d", len(descs))
	}
	if descs[0].Method != "GET" || descs[0].Path != "/" {
		t.Fatalf("expected solitary description to be the tool listing, got %+v", descs[0])
	}

	// Tools should return empty slice.
	tools := bridge.Tools()
	if len(tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(tools))
	}
}

// TestToolBridge_Good_ListsRegisteredTools verifies that GET on the bridge's
// base path returns the catalogue of registered tools per RFC.endpoints.md —
// "GET /v1/tools  List available tools".
func TestToolBridge_Good_ListsRegisteredTools(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bridge := api.NewToolBridge("/v1/tools")
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
	}, func(c *gin.Context) {})
	bridge.Add(api.ToolDescriptor{
		Name:        "metrics_query",
		Description: "Query metrics data",
		Group:       "metrics",
	}, func(c *gin.Context) {})

	engine := gin.New()
	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/tools", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", w.Code, w.Body.String())
	}

	var resp api.Response[[]api.ToolDescriptor]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success=true for tool listing")
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 tool descriptors, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "file_read" {
		t.Fatalf("expected Data[0].Name=%q, got %q", "file_read", resp.Data[0].Name)
	}
	if resp.Data[1].Name != "metrics_query" {
		t.Fatalf("expected Data[1].Name=%q, got %q", "metrics_query", resp.Data[1].Name)
	}
}

// TestToolBridge_Bad_ListingRoutesWhenEmpty verifies the listing endpoint
// still serves an empty array when no tools are registered on the bridge.
func TestToolBridge_Bad_ListingRoutesWhenEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bridge := api.NewToolBridge("/tools")
	engine := gin.New()
	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/tools", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from empty listing, got %d", w.Code)
	}

	var resp api.Response[[]api.ToolDescriptor]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success=true from empty listing")
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(resp.Data))
	}
}

// TestToolBridge_Ugly_ListingCoexistsWithToolEndpoint verifies that the GET
// listing and POST /{tool_name} endpoints register on the same base path
// without colliding.
func TestToolBridge_Ugly_ListingCoexistsWithToolEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bridge := api.NewToolBridge("/v1/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "ping",
		Description: "Ping tool",
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("pong"))
	})

	engine := gin.New()
	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	// Listing still answers at the base path.
	listReq, _ := http.NewRequest(http.MethodGet, "/v1/tools", nil)
	listW := httptest.NewRecorder()
	engine.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 from listing, got %d", listW.Code)
	}

	// Tool dispatch still answers at POST {basePath}/{name}.
	toolReq, _ := http.NewRequest(http.MethodPost, "/v1/tools/ping", nil)
	toolW := httptest.NewRecorder()
	engine.ServeHTTP(toolW, toolReq)
	if toolW.Code != http.StatusOK {
		t.Fatalf("expected 200 from tool dispatch, got %d", toolW.Code)
	}
}

func TestToolBridge_Good_ValidatesArrayInputSchema(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "tags",
		Description: "Validate array input",
		InputSchema: map[string]any{
			"type":     "array",
			"items":    map[string]any{"type": "string"},
			"minItems": 2,
			"maxItems": 3,
		},
	}, func(c *gin.Context) {
		var payload []string
		if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
			t.Fatalf("handler could not read validated array body: %v", err)
		}
		c.JSON(http.StatusOK, api.OK(payload))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/tags", bytes.NewBufferString(`["alpha","beta"]`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[[]string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success=true")
	}
	if len(resp.Data) != 2 || resp.Data[0] != "alpha" || resp.Data[1] != "beta" {
		t.Fatalf("expected validated array payload to round-trip, got %v", resp.Data)
	}
}

func TestToolBridge_Bad_RejectsTooSmallArrayInputSchema(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "tags",
		Description: "Validate array input",
		InputSchema: map[string]any{
			"type":     "array",
			"items":    map[string]any{"type": "string"},
			"minItems": 2,
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("should not run"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/tags", bytes.NewBufferString(`["alpha"]`))
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

func TestToolBridge_Ugly_RejectsWrongArrayElementType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "tags",
		Description: "Validate array input",
		InputSchema: map[string]any{
			"type":  "array",
			"items": map[string]any{"type": "string"},
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("should not run"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/tags", bytes.NewBufferString(`["alpha",123]`))
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

func TestToolBridge_Good_ValidatesNumericBounds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "score",
		Description: "Validate numeric input",
		InputSchema: map[string]any{
			"type":    "number",
			"minimum": 1,
			"maximum": 10,
		},
	}, func(c *gin.Context) {
		var payload float64
		if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
			t.Fatalf("handler could not read validated numeric body: %v", err)
		}
		c.JSON(http.StatusOK, api.OK(payload))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/score", bytes.NewBufferString(`5.5`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[float64]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success=true")
	}
	if resp.Data != 5.5 {
		t.Fatalf("expected validated numeric payload to round-trip, got %v", resp.Data)
	}
}

func TestToolBridge_Bad_RejectsLargeIntegerAboveMaximum(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "quota",
		Description: "Validate large integer input",
		InputSchema: map[string]any{
			"type":    "integer",
			"maximum": 9007199254740992,
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("should not run"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/quota", bytes.NewBufferString(`9007199254740993`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for large integer maximum failure, got %d", w.Code)
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

func TestToolBridge_Bad_RejectsNumericInputBelowMinimum(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "score",
		Description: "Validate numeric input",
		InputSchema: map[string]any{
			"type":    "number",
			"minimum": 1,
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("should not run"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/score", bytes.NewBufferString(`0`))
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

func TestToolBridge_Ugly_RejectsNonNumericInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "score",
		Description: "Validate numeric input",
		InputSchema: map[string]any{
			"type": "number",
		},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("should not run"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/score", bytes.NewBufferString(`"oops"`))
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
