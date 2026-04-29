// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio" // Note: AX-6 - Gin ResponseWriter Hijack contract requires bufio.ReadWriter; no core primitive.
	"io"    // Note: AX-6 - HTTP body reads, EOF sentinel, and body reset are structural stream boundaries.
	"iter"  // Note: AX-6 - iter.Seq is the public lazy iteration shape for bridge snapshots.
	"math"
	"math/big"
	"net"
	"net/http" // Note: AX-6 - Gin bridge handlers must emit HTTP status codes and wrap request bodies directly.
	"reflect"  // Note: AX-6 - reflect is structural for JSON schema enum equality; no core primitive.
	"regexp"
	"slices" // Note: AX-6 - deterministic snapshot cloning needs slices.Clone; no core primitive.

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
)

// ToolDescriptor describes a tool that can be exposed as a REST endpoint.
//
// Example:
//
//	desc := api.ToolDescriptor{Name: "ping", Description: "Ping the service"}
type ToolDescriptor struct {
	Name         string         // Tool name, e.g. "file_read" (becomes POST path segment)
	Description  string         // Human-readable description
	Group        string         // OpenAPI tag group, e.g. "files"
	InputSchema  map[string]any // JSON Schema for request body
	OutputSchema map[string]any // JSON Schema for response data (optional)
	// TransformerIn remaps the external request DTO into the handler-facing
	// DTO before the tool handler reads the request body.
	TransformerIn any
	// TransformerOut remaps the handler-facing response DTO into the external
	// response DTO inside the standard OK() envelope.
	TransformerOut any
}

// ToolBridge converts tool descriptors into REST endpoints and OpenAPI paths.
// It implements both RouteGroup and DescribableGroup.
//
// Example:
//
//	bridge := api.NewToolBridge("/mcp")
type ToolBridge struct {
	basePath string
	name     string
	tools    []boundTool
}

type boundTool struct {
	descriptor ToolDescriptor
	handler    gin.HandlerFunc
}

var _ RouteGroup = (*ToolBridge)(nil)
var _ DescribableGroup = (*ToolBridge)(nil)

var toolNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
var mcpServerIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,63}$`)
var regexPatternCache core.SyncMap

// NewToolBridge creates a bridge that mounts tool endpoints at basePath.
//
// Example:
//
//	bridge := api.NewToolBridge("/mcp")
func NewToolBridge(basePath string) *ToolBridge {
	return &ToolBridge{
		basePath: normaliseToolBridgePath(basePath),
		name:     "tools",
	}
}

// Add registers a tool with its HTTP handler.
//
// Example:
//
//	bridge.Add(api.ToolDescriptor{Name: "ping", Description: "Ping the service"}, handler)
func (b *ToolBridge) Add(desc ToolDescriptor, handler gin.HandlerFunc) {
	if !isValidToolName(desc.Name) {
		panic(core.E("ToolBridge.Add", "invalid tool name", nil))
	}
	if pipeline, err := compileTransformerPipeline(transformerDirectionOut, desc.TransformerOut); err != nil {
		panic(err)
	} else if len(pipeline) > 0 {
		handler = wrapTransformerOutHandler(handler, pipeline)
	}
	if validator := newToolInputValidator(desc.OutputSchema); validator != nil {
		handler = wrapToolResponseHandler(handler, validator)
	}
	if pipeline, err := compileTransformerPipeline(transformerDirectionIn, desc.TransformerIn); err != nil {
		panic(err)
	} else if len(pipeline) > 0 {
		handler = wrapTransformerInHandler(handler, pipeline)
	}
	if validator := newToolInputValidator(desc.InputSchema); validator != nil {
		handler = wrapToolHandler(handler, validator)
	}
	b.tools = append(b.tools, boundTool{descriptor: desc, handler: handler})
}

// Name returns the bridge identifier.
//
// Example:
//
//	name := bridge.Name()
func (b *ToolBridge) Name() string { return b.name }

// BasePath returns the URL prefix for all tool endpoints.
//
// Example:
//
//	path := bridge.BasePath()
func (b *ToolBridge) BasePath() string { return b.basePath }

// RegisterRoutes mounts GET / (tool listing) and POST /{tool_name} for each
// registered tool. The listing endpoint returns a JSON envelope containing the
// registered tool descriptors and mirrors RFC.endpoints.md §"ToolBridge" so
// clients can discover every tool exposed on the bridge in a single call.
//
// Example:
//
//	bridge.RegisterRoutes(rg)
//	// GET  /{basePath}/        -> tool catalogue
//	// POST /{basePath}/{name}  -> dispatch tool
func (b *ToolBridge) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", b.listHandler())
	rg.GET("/", b.listHandler())
	for _, t := range b.tools {
		rg.POST("/"+t.descriptor.Name, t.handler)
	}
}

// listHandler returns a Gin handler that serves the tool catalogue at the
// bridge's base path. The response envelope matches RFC.endpoints.md — an
// array of tool descriptors with their name, description, group, and the
// declared input/output JSON schemas.
//
//	GET /v1/tools -> {"success": true, "data": [{"name": "ping", ...}]}
func (b *ToolBridge) listHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, OK(b.Tools()))
	}
}

// Describe returns OpenAPI route descriptions for all registered tools plus a
// GET entry describing the tool listing endpoint at the bridge's base path.
//
// Example:
//
//	descs := bridge.Describe()
func (b *ToolBridge) Describe() []RouteDescription {
	tools := b.snapshotTools()
	descs := make([]RouteDescription, 0, len(tools)+1)
	descs = append(descs, describeToolList(b.name))
	for _, tool := range tools {
		descs = append(descs, describeTool(tool.descriptor, b.name))
	}
	return descs
}

// DescribeIter returns an iterator over OpenAPI route descriptions for all
// registered tools plus a leading GET entry for the tool listing endpoint.
//
// Example:
//
//	for rd := range bridge.DescribeIter() {
//		_ = rd
//	}
func (b *ToolBridge) DescribeIter() iter.Seq[RouteDescription] {
	tools := b.snapshotTools()
	defaultTag := b.name
	return func(yield func(RouteDescription) bool) {
		if !yield(describeToolList(defaultTag)) {
			return
		}
		for _, tool := range tools {
			if !yield(describeTool(tool.descriptor, defaultTag)) {
				return
			}
		}
	}
}

// Tools returns all registered tool descriptors.
//
// Example:
//
//	descs := bridge.Tools()
func (b *ToolBridge) Tools() []ToolDescriptor {
	tools := b.snapshotTools()
	descs := make([]ToolDescriptor, len(tools))
	for i, t := range tools {
		descs[i] = t.descriptor
	}
	return descs
}

// ToolsIter returns an iterator over all registered tool descriptors.
//
// Example:
//
//	for desc := range bridge.ToolsIter() {
//		_ = desc
//	}
func (b *ToolBridge) ToolsIter() iter.Seq[ToolDescriptor] {
	tools := b.snapshotTools()
	return func(yield func(ToolDescriptor) bool) {
		for _, tool := range tools {
			if !yield(tool.descriptor) {
				return
			}
		}
	}
}

// IsValidMCPServerID reports whether id is safe to use as an MCP HTTP bridge
// server_id path segment or filesystem-backed lookup key.
func IsValidMCPServerID(id string) bool {
	return isValidMCPServerID(id)
}

func (b *ToolBridge) snapshotTools() []boundTool {
	if len(b.tools) == 0 {
		return nil
	}
	return slices.Clone(b.tools)
}

func describeTool(desc ToolDescriptor, defaultTag string) RouteDescription {
	tags := cleanTags([]string{desc.Group})
	if len(tags) == 0 {
		tags = []string{defaultTag}
	}
	return RouteDescription{
		Method:         "POST",
		Path:           "/" + desc.Name,
		Summary:        desc.Description,
		Description:    desc.Description,
		Tags:           tags,
		RequestBody:    desc.InputSchema,
		Response:       desc.OutputSchema,
		TransformerIn:  desc.TransformerIn,
		TransformerOut: desc.TransformerOut,
	}
}

// describeToolList returns the RouteDescription for GET {basePath}/ —
// the tool catalogue listing documented in RFC.endpoints.md.
//
//	rd := describeToolList("tools")
//	// rd.Method == "GET" && rd.Path == "/"
func describeToolList(defaultTag string) RouteDescription {
	tags := cleanTags([]string{defaultTag})
	if len(tags) == 0 {
		tags = []string{"tools"}
	}
	return RouteDescription{
		Method:      "GET",
		Path:        "/",
		Summary:     "List available tools",
		Description: "Returns every tool registered on the bridge, including its declared input and output JSON schemas.",
		Tags:        tags,
		StatusCode:  http.StatusOK,
		Response: map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":         map[string]any{"type": "string"},
					"description":  map[string]any{"type": "string"},
					"group":        map[string]any{"type": "string"},
					"inputSchema":  map[string]any{"type": "object", "additionalProperties": true},
					"outputSchema": map[string]any{"type": "object", "additionalProperties": true},
				},
				"required": []string{"name"},
			},
		},
	}
}

func isValidToolName(name string) bool {
	name = core.Trim(name)
	if name == "" {
		return false
	}

	// Keep the path segment safe even when callers try to smuggle in traversal
	// syntax or path separators through a name that otherwise matches the regex.
	if containsAnyByte(name, `/\*?`) || core.Contains(name, "..") {
		return false
	}

	return toolNamePattern.MatchString(name)
}

func isValidMCPServerID(id string) bool {
	if id == "" {
		return false
	}

	if core.Contains(id, "\x00") || containsAnyByte(id, `/\`) || core.Contains(id, "..") {
		return false
	}

	return mcpServerIDPattern.MatchString(id)
}

// normaliseToolBridgePath coerces the bridge base path into a stable form.
// A blank value maps to "/" so the bridge still has a valid mount point.
func normaliseToolBridgePath(path string) string {
	path = core.Trim(path)
	if path == "" {
		return "/"
	}

	path = "/" + trimSlashes(path)
	if path == "/" {
		return "/"
	}

	return path
}

func containsAnyByte(s, chars string) bool {
	for i := 0; i < len(s); i++ {
		for j := 0; j < len(chars); j++ {
			if s[i] == chars[j] {
				return true
			}
		}
	}
	return false
}

// maxToolRequestBodyBytes is the maximum request body size accepted by the
// tool bridge handler. Requests larger than this are rejected with 413.
const maxToolRequestBodyBytes = 10 << 20 // 10 MiB

func wrapToolHandler(handler gin.HandlerFunc, validator *toolInputValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		limited := http.MaxBytesReader(c.Writer, c.Request.Body, maxToolRequestBodyBytes)
		body, err := io.ReadAll(limited)
		if err != nil {
			status := http.StatusBadRequest
			msg := "Unable to read request body"
			if err.Error() == "http: request body too large" {
				status = http.StatusRequestEntityTooLarge
				msg = "Request body exceeds the maximum allowed size"
			}
			c.AbortWithStatusJSON(status, FailWithDetails(
				"invalid_request_body",
				msg,
				map[string]any{"error": err.Error()},
			))
			return
		}

		if err := validator.Validate(body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, FailWithDetails(
				"invalid_request_body",
				"Request body does not match the declared tool schema",
				map[string]any{"error": err.Error()},
			))
			return
		}

		c.Request.Body = io.NopCloser(core.NewBuffer(body))
		handler(c)
	}
}

func wrapToolResponseHandler(handler gin.HandlerFunc, validator *toolInputValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		recorder := newToolResponseRecorder(c.Writer)
		c.Writer = recorder

		handler(c)

		if recorder.Status() >= 200 && recorder.Status() < 300 {
			if err := validator.ValidateResponse(recorder.body); err != nil {
				recorder.reset()
				recorder.writeErrorResponse(http.StatusInternalServerError, FailWithDetails(
					"invalid_tool_response",
					"Tool response does not match the declared output schema",
					map[string]any{"error": err.Error()},
				))
				return
			}
		}

		recorder.commit()
	}
}

type toolInputValidator struct {
	schema map[string]any
}

func newToolInputValidator(schema map[string]any) *toolInputValidator {
	if len(schema) == 0 {
		return nil
	}
	return &toolInputValidator{schema: schema}
}

func (v *toolInputValidator) Validate(body []byte) (
	_ error,
) {
	if core.Trim(string(body)) == "" {
		return core.E("ToolBridge.Validate", "request body is required", nil)
	}

	payload, err := decodeJSONValuePreserveNumbers(body)
	if err != nil {
		return core.E("ToolBridge.Validate", "invalid JSON", err)
	}

	return validateSchemaNode(payload, v.schema, "")
}

func (v *toolInputValidator) ValidateResponse(body []byte) (
	_ error,
) {
	if len(v.schema) == 0 {
		return nil
	}

	decoded, err := decodeJSONValuePreserveNumbers(body)
	if err != nil {
		return core.E("ToolBridge.ValidateResponse", "invalid JSON response", err)
	}
	envelope, ok := decoded.(map[string]any)
	if !ok {
		return core.E("ToolBridge.ValidateResponse", "response envelope must be an object", nil)
	}

	success, _ := envelope["success"].(bool)
	if !success {
		return core.E("ToolBridge.ValidateResponse", "response is missing a successful envelope", nil)
	}

	// data is serialised with omitempty, so a nil/zero-value payload from
	// constructors like OK(nil) or OK(false) will omit the key entirely.
	// Treat a missing data key as a valid nil payload for successful responses.
	data, ok := envelope["data"]
	if !ok {
		return nil
	}

	encoded, err := marshalCoreJSON(data)
	if err != nil {
		return core.E("ToolBridge.ValidateResponse", "encode response data", err)
	}

	payload, err := decodeJSONValuePreserveNumbers(encoded)
	if err != nil {
		return core.E("ToolBridge.ValidateResponse", "decode response data", err)
	}

	return validateSchemaNode(payload, v.schema, "")
}

func validateSchemaNode(value any, schema map[string]any, path string) (
	_ error,
) {
	if len(schema) == 0 {
		return nil
	}

	schemaType, _ := schema["type"].(string)
	if schemaType != "" {
		switch schemaType {
		case "object":
			obj, ok := value.(map[string]any)
			if !ok {
				return typeError(path, "object", value)
			}

			for _, name := range stringList(schema["required"]) {
				if _, ok := obj[name]; !ok {
					return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s is missing required field %q", displayPath(path), name), nil)
				}
			}

			for name, rawChild := range schemaMap(schema["properties"]) {
				childSchema, ok := rawChild.(map[string]any)
				if !ok {
					continue
				}
				childValue, ok := obj[name]
				if !ok {
					continue
				}
				if err := validateSchemaNode(childValue, childSchema, joinPath(path, name)); err != nil {
					return err
				}
			}

			if additionalProperties, ok := schema["additionalProperties"].(bool); ok && !additionalProperties {
				properties := schemaMap(schema["properties"])
				for name := range obj {
					if properties != nil {
						if _, ok := properties[name]; ok {
							continue
						}
					}
					return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s contains unknown field %q", displayPath(path), name), nil)
				}
			}
			if err := validateObjectConstraints(obj, schema, path); err != nil {
				return err
			}
		case "array":
			arr, ok := value.([]any)
			if !ok {
				return typeError(path, "array", value)
			}
			if items := schemaMap(schema["items"]); len(items) > 0 {
				for i, item := range arr {
					if err := validateSchemaNode(item, items, joinPath(path, core.Itoa(i))); err != nil {
						return err
					}
				}
			}
			if err := validateArrayConstraints(arr, schema, path); err != nil {
				return err
			}
		case "string":
			str, ok := value.(string)
			if !ok {
				return typeError(path, "string", value)
			}
			if err := validateStringConstraints(str, schema, path); err != nil {
				return err
			}
		case "boolean":
			if _, ok := value.(bool); !ok {
				return typeError(path, "boolean", value)
			}
		case "integer":
			if !isIntegerValue(value) {
				return typeError(path, "integer", value)
			}
			if err := validateNumericConstraints(value, schema, path); err != nil {
				return err
			}
		case "number":
			if !isNumberValue(value) {
				return typeError(path, "number", value)
			}
			if err := validateNumericConstraints(value, schema, path); err != nil {
				return err
			}
		}
	}

	if schemaType == "" && (len(schemaMap(schema["properties"])) > 0 || schema["required"] != nil || schema["additionalProperties"] != nil) {
		props := schemaMap(schema["properties"])
		return validateSchemaNode(value, map[string]any{
			"type":                 "object",
			"properties":           props,
			"required":             schema["required"],
			"additionalProperties": schema["additionalProperties"],
		}, path)
	}

	if rawEnum, ok := schema["enum"]; ok {
		if !enumContains(value, rawEnum) {
			return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must be one of the declared enum values", displayPath(path)), nil)
		}
	}

	if err := validateSchemaCombinators(value, schema, path); err != nil {
		return err
	}

	return nil
}

func validateSchemaCombinators(value any, schema map[string]any, path string) (
	_ error,
) {
	if subschemas := schemaObjects(schema["allOf"]); len(subschemas) > 0 {
		for _, subschema := range subschemas {
			if err := validateSchemaNode(value, subschema, path); err != nil {
				return err
			}
		}
	}

	if subschemas := schemaObjects(schema["anyOf"]); len(subschemas) > 0 {
		for _, subschema := range subschemas {
			if err := validateSchemaNode(value, subschema, path); err == nil {
				goto anyOfMatched
			}
		}
		return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must match at least one schema in anyOf", displayPath(path)), nil)
	}

anyOfMatched:
	if subschemas := schemaObjects(schema["oneOf"]); len(subschemas) > 0 {
		matches := 0
		for _, subschema := range subschemas {
			if err := validateSchemaNode(value, subschema, path); err == nil {
				matches++
			}
		}
		if matches != 1 {
			if matches == 0 {
				return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must match exactly one schema in oneOf", displayPath(path)), nil)
			}
			return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s matches multiple schemas in oneOf", displayPath(path)), nil)
		}
	}

	if subschema, ok := schema["not"].(map[string]any); ok && subschema != nil {
		if err := validateSchemaNode(value, subschema, path); err == nil {
			return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must not match the forbidden schema", displayPath(path)), nil)
		}
	}

	return nil
}

func validateStringConstraints(value string, schema map[string]any, path string) (
	_ error,
) {
	length := core.RuneCount(value)
	if minLength, ok := schemaInt(schema["minLength"]); ok && length < minLength {
		return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must be at least %d characters long", displayPath(path), minLength), nil)
	}
	if maxLength, ok := schemaInt(schema["maxLength"]); ok && length > maxLength {
		return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must be at most %d characters long", displayPath(path), maxLength), nil)
	}
	if pattern, ok := schema["pattern"].(string); ok && pattern != "" {
		re, err := compiledPattern(pattern)
		if err != nil {
			return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s has an invalid pattern %q", displayPath(path), pattern), err)
		}
		if !re.MatchString(value) {
			return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s does not match pattern %q", displayPath(path), pattern), nil)
		}
	}
	return nil
}

func validateNumericConstraints(value any, schema map[string]any, path string) (
	_ error,
) {
	if minimum, ok := schemaFloat(schema["minimum"]); ok && numericLessThan(value, minimum) {
		return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must be greater than or equal to %v", displayPath(path), minimum), nil)
	}
	if maximum, ok := schemaFloat(schema["maximum"]); ok && numericGreaterThan(value, maximum) {
		return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must be less than or equal to %v", displayPath(path), maximum), nil)
	}
	return nil
}

func validateArrayConstraints(value []any, schema map[string]any, path string) (
	_ error,
) {
	if minItems, ok := schemaInt(schema["minItems"]); ok && len(value) < minItems {
		return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must contain at least %d items", displayPath(path), minItems), nil)
	}
	if maxItems, ok := schemaInt(schema["maxItems"]); ok && len(value) > maxItems {
		return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must contain at most %d items", displayPath(path), maxItems), nil)
	}
	return nil
}

func validateObjectConstraints(value map[string]any, schema map[string]any, path string) (
	_ error,
) {
	if minProps, ok := schemaInt(schema["minProperties"]); ok && len(value) < minProps {
		return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must contain at least %d properties", displayPath(path), minProps), nil)
	}
	if maxProps, ok := schemaInt(schema["maxProperties"]); ok && len(value) > maxProps {
		return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must contain at most %d properties", displayPath(path), maxProps), nil)
	}
	return nil
}

func schemaInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		if v > math.MaxInt64 {
			return 0, false
		}
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		if v > math.MaxInt64 {
			return 0, false
		}
		return int(v), true
	case float64:
		if v == float64(int(v)) {
			return int(v), true
		}
	case jsonNumber:
		if n, err := v.Int64(); err == nil {
			return int(n), true
		}
	}
	return 0, false
}

func schemaFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case jsonNumber:
		if n, err := v.Float64(); err == nil {
			return n, true
		}
	}
	return 0, false
}

func numericLessThan(value any, limit float64) bool {
	if integral, ok := integralNumericValue(value); ok {
		if cmp, ok := compareIntegralNumericToFloat64(integral, limit); ok {
			return cmp < 0
		}
	}
	if n, ok := numericValue(value); ok {
		return n < limit
	}
	return false
}

func numericGreaterThan(value any, limit float64) bool {
	if integral, ok := integralNumericValue(value); ok {
		if cmp, ok := compareIntegralNumericToFloat64(integral, limit); ok {
			return cmp > 0
		}
	}
	if n, ok := numericValue(value); ok {
		return n > limit
	}
	return false
}

type toolResponseRecorder struct {
	gin.ResponseWriter
	headers     http.Header
	body        []byte
	status      int
	wroteHeader bool
}

func newToolResponseRecorder(w gin.ResponseWriter) *toolResponseRecorder {
	headers := make(http.Header)
	for k, vals := range w.Header() {
		headers[k] = append([]string(nil), vals...)
	}
	return &toolResponseRecorder{
		ResponseWriter: w,
		headers:        headers,
		status:         http.StatusOK,
	}
}

func (w *toolResponseRecorder) Header() http.Header {
	return w.headers
}

func (w *toolResponseRecorder) WriteHeader(code int) {
	w.status = code
	w.wroteHeader = true
}

func (w *toolResponseRecorder) WriteHeaderNow() {
	w.wroteHeader = true
}

func (w *toolResponseRecorder) Write(data []byte) (
	int,
	error,
) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	w.body = append(w.body, data...)
	return len(data), nil
}

func (w *toolResponseRecorder) WriteString(s string) (
	int,
	error,
) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	w.body = append(w.body, s...)
	return len(s), nil
}

func (w *toolResponseRecorder) Flush() {
}

func (w *toolResponseRecorder) Status() int {
	if w.wroteHeader {
		return w.status
	}
	return http.StatusOK
}

func (w *toolResponseRecorder) Size() int {
	return len(w.body)
}

func (w *toolResponseRecorder) Written() bool {
	return w.wroteHeader
}

func (w *toolResponseRecorder) Hijack() (
	net.Conn,
	*bufio.ReadWriter,
	error,
) {
	return nil, nil, core.E("ToolBridge.ResponseRecorder", "response hijacking is not supported by ToolBridge output validation", nil)
}

func (w *toolResponseRecorder) commit() {
	for k := range w.ResponseWriter.Header() {
		w.ResponseWriter.Header().Del(k)
	}
	for k, vals := range w.headers {
		for _, v := range vals {
			w.ResponseWriter.Header().Add(k, v)
		}
	}
	w.ResponseWriter.WriteHeader(w.Status())
	_, _ = w.ResponseWriter.Write(w.body)
}

func (w *toolResponseRecorder) reset() {
	w.headers = make(http.Header)
	w.body = nil
	w.status = http.StatusInternalServerError
	w.wroteHeader = false
}

func (w *toolResponseRecorder) writeErrorResponse(status int, resp Response[any]) {
	data, err := marshalCoreJSON(resp)
	if err != nil {
		w.status = http.StatusInternalServerError
		w.wroteHeader = true
		http.Error(w.ResponseWriter, "internal server error", http.StatusInternalServerError)
		return
	}

	// Keep recorder state aligned with the replacement response so that
	// Status(), Written(), Header() and Size() all reflect the error
	// response. Post-handler middleware and metrics must observe correct
	// values, not stale state from the reset() call above.
	w.status = status
	w.wroteHeader = true
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	w.headers.Set("Content-Type", "application/json")
	w.body = append(w.body[:0], data...)
	w.commit()
}

func typeError(path, want string, value any) (
	_ error,
) {
	return core.E("ToolBridge.ValidateSchema", core.Sprintf("%s must be %s, got %s", displayPath(path), want, describeJSONValue(value)), nil)
}

func displayPath(path string) string {
	if path == "" {
		return "request body"
	}
	return "request body." + path
}

func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func schemaMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	m, _ := value.(map[string]any)
	return m
}

func schemaObjects(value any) []map[string]any {
	switch raw := value.(type) {
	case []any:
		out := make([]map[string]any, 0, len(raw))
		for _, item := range raw {
			if schema := schemaMap(item); schema != nil {
				out = append(out, schema)
			}
		}
		return out
	case []map[string]any:
		return append([]map[string]any(nil), raw...)
	default:
		return nil
	}
}

func stringList(value any) []string {
	switch raw := value.(type) {
	case []any:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			name, ok := item.(string)
			if !ok {
				continue
			}
			out = append(out, name)
		}
		return out
	case []string:
		return append([]string(nil), raw...)
	default:
		return nil
	}
}

func isIntegerValue(value any) bool {
	switch v := value.(type) {
	case jsonNumber:
		_, err := v.Int64()
		return err == nil
	case float64:
		return v == float64(int64(v))
	default:
		return false
	}
}

func isNumberValue(value any) bool {
	switch value.(type) {
	case jsonNumber, float64:
		return true
	default:
		return false
	}
}

func enumContains(value any, rawEnum any) bool {
	items := enumValues(rawEnum)
	for _, candidate := range items {
		if valuesEqual(value, candidate) {
			return true
		}
	}
	return false
}

func enumValues(rawEnum any) []any {
	switch values := rawEnum.(type) {
	case []any:
		out := make([]any, 0, len(values))
		for _, value := range values {
			out = append(out, value)
		}
		return out
	case []string:
		out := make([]any, 0, len(values))
		for _, value := range values {
			out = append(out, value)
		}
		return out
	default:
		return nil
	}
}

func valuesEqual(left, right any) bool {
	if isNumericValue(left) && isNumericValue(right) {
		if li, lok := integralNumericValue(left); lok {
			if ri, rok := integralNumericValue(right); rok {
				return li.Cmp(ri) == 0
			}
		}
		lv, lok := numericValue(left)
		rv, rok := numericValue(right)
		return lok && rok && lv == rv
	}
	return reflect.DeepEqual(left, right)
}

func isNumericValue(value any) bool {
	switch value.(type) {
	case jsonNumber, float64, float32, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

func numericValue(value any) (float64, bool) {
	switch v := value.(type) {
	case jsonNumber:
		n, err := v.Float64()
		return n, err == nil
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func compiledPattern(pattern string) (
	*regexp.Regexp,
	error,
) {
	if cached, ok := regexPatternCache.Load(pattern); ok {
		return cached.(*regexp.Regexp), nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	if actual, loaded := regexPatternCache.LoadOrStore(pattern, re); loaded {
		return actual.(*regexp.Regexp), nil
	}

	return re, nil
}

func integralNumericValue(value any) (*big.Int, bool) {
	switch v := value.(type) {
	case int:
		return big.NewInt(int64(v)), true
	case int8:
		return big.NewInt(int64(v)), true
	case int16:
		return big.NewInt(int64(v)), true
	case int32:
		return big.NewInt(int64(v)), true
	case int64:
		return big.NewInt(v), true
	case uint:
		return new(big.Int).SetUint64(uint64(v)), true
	case uint8:
		return new(big.Int).SetUint64(uint64(v)), true
	case uint16:
		return new(big.Int).SetUint64(uint64(v)), true
	case uint32:
		return new(big.Int).SetUint64(uint64(v)), true
	case uint64:
		return new(big.Int).SetUint64(v), true
	case float32:
		return integralBigIntFromFloat64(float64(v))
	case float64:
		return integralBigIntFromFloat64(v)
	case jsonNumber:
		text := core.Trim(v.String())
		if text == "" || containsAnyByte(text, ".eE") {
			return nil, false
		}
		n := new(big.Int)
		if _, ok := n.SetString(text, 10); ok {
			return n, true
		}
	}
	return nil, false
}

func integralBigIntFromFloat64(v float64) (*big.Int, bool) {
	if !isIntegralFloat64(v) {
		return nil, false
	}
	i, _ := new(big.Float).SetFloat64(v).Int(nil)
	if i == nil {
		return nil, false
	}
	return i, true
}

func compareIntegralNumericToFloat64(value *big.Int, limit float64) (int, bool) {
	if !isIntegralFloat64(limit) {
		return 0, false
	}

	// gosec:disable G115 -- BitLen() returns int >= 0; +1 stays positive and within
	// int range (BitLen on a *big.Int can't reach math.MaxInt minus one in
	// practice). Cast to uint cannot overflow.
	precision := uint(value.BitLen() + 1)
	if precision < 64 {
		precision = 64
	}

	left := new(big.Float).SetPrec(precision).SetInt(value)
	right := new(big.Float).SetPrec(precision).SetFloat64(limit)
	return left.Cmp(right), true
}

func isIntegralFloat64(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && math.Trunc(v) == v
}

func describeJSONValue(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case bool:
		return "boolean"
	case jsonNumber, float64:
		return "number"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return core.Sprintf("%T", value)
	}
}
