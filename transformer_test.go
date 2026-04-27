// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	core "dappco.re/go/core"

	api "dappco.re/go/api"

	"github.com/gin-gonic/gin"
)

type transformerExternalUser struct {
	FullName string `json:"full_name"`
}

type transformerInternalUser struct {
	Name string `json:"name"`
}

type transformerBridgeIn struct{}

func (transformerBridgeIn) TransformIn(_ *gin.Context, in transformerExternalUser) (transformerInternalUser, error) {
	if in.FullName == "" {
		return transformerInternalUser{}, core.E("transformerBridgeIn", "full_name is required", nil)
	}
	return transformerInternalUser{Name: in.FullName}, nil
}

type transformerBridgeOut struct{}

func (transformerBridgeOut) TransformOut(_ *gin.Context, out transformerInternalUser) (transformerExternalUser, error) {
	return transformerExternalUser{FullName: out.Name}, nil
}

func TestTransformer_Good_ToolBridgeRemapsInboundAndOutboundDTOs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "create_user",
		Description: "Create a user",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"full_name": map[string]any{"type": "string"},
			},
			"required":             []any{"full_name"},
			"additionalProperties": false,
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"full_name": map[string]any{"type": "string"},
			},
			"required":             []any{"full_name"},
			"additionalProperties": false,
		},
		TransformerIn:  transformerBridgeIn{},
		TransformerOut: transformerBridgeOut{},
	}, func(c *gin.Context) {
		var payload transformerInternalUser
		if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
			t.Fatalf("handler could not decode transformed payload: %v", err)
		}
		if payload.Name != "Ada Lovelace" {
			t.Fatalf("expected internal name, got %q", payload.Name)
		}
		c.JSON(http.StatusOK, api.OK(payload))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/create_user", bytes.NewBufferString(`{"full_name":"Ada Lovelace"}`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp api.Response[map[string]any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success=true")
	}
	if resp.Data["full_name"] != "Ada Lovelace" {
		t.Fatalf("expected external full_name, got %v", resp.Data)
	}
	if _, ok := resp.Data["name"]; ok {
		t.Fatalf("expected internal name to be hidden, got %v", resp.Data)
	}
}

func TestTransformer_Bad_ToolBridgeValidatesExternalPayloadBeforeTransform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	bridge := api.NewToolBridge("/tools")
	bridge.Add(api.ToolDescriptor{
		Name:        "create_user",
		Description: "Create a user",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"full_name": map[string]any{"type": "string"},
			},
			"required":             []any{"full_name"},
			"additionalProperties": false,
		},
		TransformerIn: transformerBridgeIn{},
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("should not run"))
	})

	rg := engine.Group(bridge.BasePath())
	bridge.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tools/create_user", bytes.NewBufferString(`{"name":"Ada Lovelace"}`))
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Success {
		t.Fatal("expected Success=false")
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request_body" {
		t.Fatalf("expected invalid_request_body, got %#v", resp.Error)
	}
}

type transformerRouteGroup struct{}

func (transformerRouteGroup) Name() string     { return "transformer-users" }
func (transformerRouteGroup) BasePath() string { return "/users" }
func (transformerRouteGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", func(c *gin.Context) {
		var payload transformerInternalUser
		if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
			c.JSON(http.StatusBadRequest, api.Fail("invalid_body", err.Error()))
			return
		}
		c.JSON(http.StatusOK, api.OK(payload))
	})
}

func (transformerRouteGroup) Describe() []api.RouteDescription {
	return []api.RouteDescription{
		{
			Method: "POST",
			Path:   "/",
			RequestBody: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"full_name": map[string]any{"type": "string"},
				},
				"required":             []any{"full_name"},
				"additionalProperties": false,
			},
			Response: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"full_name": map[string]any{"type": "string"},
				},
				"required":             []any{"full_name"},
				"additionalProperties": false,
			},
			TransformerIn:  api.RenameFields(map[string]string{"full_name": "name"}),
			TransformerOut: api.RenameFields(map[string]string{"name": "full_name"}),
		},
	}
}

func TestTransformer_Good_EngineRouteDescriptionRemapsDTOs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, err := api.New()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.Register(transformerRouteGroup{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"full_name":"Grace Hopper"}`))
	engine.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp api.Response[map[string]any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success=true")
	}
	if resp.Data["full_name"] != "Grace Hopper" {
		t.Fatalf("expected outbound field rename, got %v", resp.Data)
	}
	if _, ok := resp.Data["name"]; ok {
		t.Fatalf("expected internal name field to be absent, got %v", resp.Data)
	}
}

func TestTransformer_Bad_EngineTransformerErrorReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, err := api.New()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.Register(errorTransformerRouteGroup{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/error-users", bytes.NewBufferString(`{"full_name":"Alan Turing"}`))
	engine.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Success {
		t.Fatal("expected Success=false")
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request_body" {
		t.Fatalf("expected invalid_request_body, got %#v", resp.Error)
	}
}

type errorTransformerRouteGroup struct{}

func (errorTransformerRouteGroup) Name() string     { return "error-transformer-users" }
func (errorTransformerRouteGroup) BasePath() string { return "/error-users" }
func (errorTransformerRouteGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.OK("should not run"))
	})
}

func (errorTransformerRouteGroup) Describe() []api.RouteDescription {
	return []api.RouteDescription{
		{
			Method: "POST",
			Path:   "/",
			TransformerIn: api.TransformerInFunc[transformerExternalUser, transformerInternalUser](
				func(_ *gin.Context, _ transformerExternalUser) (transformerInternalUser, error) {
					return transformerInternalUser{}, core.E("errorTransformer", "rejected", nil)
				},
			),
		},
	}
}
