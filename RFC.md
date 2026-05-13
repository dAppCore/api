# API RFC Notes

## Handler Metadata Example

```go
type createWidgetHandler struct{}

func (h *createWidgetHandler) Describe() api.RouteDescription {
	return api.RouteDescription{
		StatusCode: http.StatusCreated,
		RequestBody: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		},
		Response: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string"},
			},
		},
	}
}

func (h *createWidgetHandler) OperationID() string { return "widgets_create" }
func (h *createWidgetHandler) Tags() []string      { return []string{"widgets"} }
func (h *createWidgetHandler) Summary() string     { return "Create widget" }
func (h *createWidgetHandler) Description() string { return "Creates a widget." }

func (h *createWidgetHandler) Render() api.RenderHints {
	return api.RenderHints{
		Kind: "form",
		Fields: []api.FieldHint{
			{Name: "name", Label: "Name", Type: "text", Required: true},
		},
		Actions: []api.ActionHint{
			{Name: "preview", Label: "Preview", Method: http.MethodGet},
		},
	}
}

func (g *widgetsGroup) Describe() []api.RouteDescription {
	handler := &createWidgetHandler{}
	return []api.RouteDescription{
		{
			Method:  http.MethodPost,
			Path:    "/",
			Handler: handler,
		},
	}
}
```

When a `RouteDescription` carries a handler that implements `api.Describable`
and/or `api.Renderable`, `SpecBuilder` uses that metadata to populate the
OpenAPI `operationId`, `tags`, `summary`, `description`, and the
`x-render-hints` vendor extension.
