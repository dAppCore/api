// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	coreerr "dappco.re/go/core/log"
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

// NewToolBridge creates a bridge that mounts tool endpoints at basePath.
//
// Example:
//
//	bridge := api.NewToolBridge("/mcp")
func NewToolBridge(basePath string) *ToolBridge {
	return &ToolBridge{
		basePath: basePath,
		name:     "tools",
	}
}

// Add registers a tool with its HTTP handler.
//
// Example:
//
//	bridge.Add(api.ToolDescriptor{Name: "ping", Description: "Ping the service"}, handler)
func (b *ToolBridge) Add(desc ToolDescriptor, handler gin.HandlerFunc) {
	if validator := newToolInputValidator(desc.OutputSchema); validator != nil {
		handler = wrapToolResponseHandler(handler, validator)
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

// RegisterRoutes mounts POST /{tool_name} for each registered tool.
//
// Example:
//
//	bridge.RegisterRoutes(rg)
func (b *ToolBridge) RegisterRoutes(rg *gin.RouterGroup) {
	for _, t := range b.tools {
		rg.POST("/"+t.descriptor.Name, t.handler)
	}
}

// Describe returns OpenAPI route descriptions for all registered tools.
//
// Example:
//
//	descs := bridge.Describe()
func (b *ToolBridge) Describe() []RouteDescription {
	tools := b.snapshotTools()
	descs := make([]RouteDescription, 0, len(tools))
	for _, tool := range tools {
		descs = append(descs, describeTool(tool.descriptor, b.name))
	}
	return descs
}

// DescribeIter returns an iterator over OpenAPI route descriptions for all registered tools.
//
// Example:
//
//	for rd := range bridge.DescribeIter() {
//		_ = rd
//	}
func (b *ToolBridge) DescribeIter() iter.Seq[RouteDescription] {
	tools := b.snapshotTools()
	return func(yield func(RouteDescription) bool) {
		for _, tool := range tools {
			if !yield(describeTool(tool.descriptor, b.name)) {
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
		Method:      "POST",
		Path:        "/" + desc.Name,
		Summary:     desc.Description,
		Description: desc.Description,
		Tags:        tags,
		RequestBody: desc.InputSchema,
		Response:    desc.OutputSchema,
	}
}

func wrapToolHandler(handler gin.HandlerFunc, validator *toolInputValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, FailWithDetails(
				"invalid_request_body",
				"Unable to read request body",
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

		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		handler(c)
	}
}

func wrapToolResponseHandler(handler gin.HandlerFunc, validator *toolInputValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		recorder := newToolResponseRecorder(c.Writer)
		c.Writer = recorder

		handler(c)

		if recorder.Status() >= 200 && recorder.Status() < 300 {
			if err := validator.ValidateResponse(recorder.body.Bytes()); err != nil {
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

func (v *toolInputValidator) Validate(body []byte) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return coreerr.E("ToolBridge.Validate", "request body is required", nil)
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()

	var payload any
	if err := dec.Decode(&payload); err != nil {
		return coreerr.E("ToolBridge.Validate", "invalid JSON", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return coreerr.E("ToolBridge.Validate", "request body must contain a single JSON value", nil)
	}

	return validateSchemaNode(payload, v.schema, "")
}

func (v *toolInputValidator) ValidateResponse(body []byte) error {
	if len(v.schema) == 0 {
		return nil
	}

	var envelope map[string]any
	if err := json.Unmarshal(body, &envelope); err != nil {
		return coreerr.E("ToolBridge.ValidateResponse", "invalid JSON response", err)
	}

	success, _ := envelope["success"].(bool)
	if !success {
		return coreerr.E("ToolBridge.ValidateResponse", "response is missing a successful envelope", nil)
	}

	data, ok := envelope["data"]
	if !ok {
		return coreerr.E("ToolBridge.ValidateResponse", "response is missing data", nil)
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return coreerr.E("ToolBridge.ValidateResponse", "encode response data", err)
	}

	var payload any
	dec := json.NewDecoder(bytes.NewReader(encoded))
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return coreerr.E("ToolBridge.ValidateResponse", "decode response data", err)
	}

	return validateSchemaNode(payload, v.schema, "")
}

func validateSchemaNode(value any, schema map[string]any, path string) error {
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
					return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s is missing required field %q", displayPath(path), name), nil)
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
					return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s contains unknown field %q", displayPath(path), name), nil)
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
					if err := validateSchemaNode(item, items, joinPath(path, strconv.Itoa(i))); err != nil {
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
			return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must be one of the declared enum values", displayPath(path)), nil)
		}
	}

	if err := validateSchemaCombinators(value, schema, path); err != nil {
		return err
	}

	return nil
}

func validateSchemaCombinators(value any, schema map[string]any, path string) error {
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
		return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must match at least one schema in anyOf", displayPath(path)), nil)
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
				return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must match exactly one schema in oneOf", displayPath(path)), nil)
			}
			return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s matches multiple schemas in oneOf", displayPath(path)), nil)
		}
	}

	if subschema, ok := schema["not"].(map[string]any); ok && subschema != nil {
		if err := validateSchemaNode(value, subschema, path); err == nil {
			return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must not match the forbidden schema", displayPath(path)), nil)
		}
	}

	return nil
}

func validateStringConstraints(value string, schema map[string]any, path string) error {
	length := utf8.RuneCountInString(value)
	if minLength, ok := schemaInt(schema["minLength"]); ok && length < minLength {
		return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must be at least %d characters long", displayPath(path), minLength), nil)
	}
	if maxLength, ok := schemaInt(schema["maxLength"]); ok && length > maxLength {
		return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must be at most %d characters long", displayPath(path), maxLength), nil)
	}
	if pattern, ok := schema["pattern"].(string); ok && pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s has an invalid pattern %q", displayPath(path), pattern), err)
		}
		if !re.MatchString(value) {
			return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s does not match pattern %q", displayPath(path), pattern), nil)
		}
	}
	return nil
}

func validateNumericConstraints(value any, schema map[string]any, path string) error {
	if minimum, ok := schemaFloat(schema["minimum"]); ok && numericLessThan(value, minimum) {
		return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must be greater than or equal to %v", displayPath(path), minimum), nil)
	}
	if maximum, ok := schemaFloat(schema["maximum"]); ok && numericGreaterThan(value, maximum) {
		return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must be less than or equal to %v", displayPath(path), maximum), nil)
	}
	return nil
}

func validateArrayConstraints(value []any, schema map[string]any, path string) error {
	if minItems, ok := schemaInt(schema["minItems"]); ok && len(value) < minItems {
		return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must contain at least %d items", displayPath(path), minItems), nil)
	}
	if maxItems, ok := schemaInt(schema["maxItems"]); ok && len(value) > maxItems {
		return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must contain at most %d items", displayPath(path), maxItems), nil)
	}
	return nil
}

func validateObjectConstraints(value map[string]any, schema map[string]any, path string) error {
	if minProps, ok := schemaInt(schema["minProperties"]); ok && len(value) < minProps {
		return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must contain at least %d properties", displayPath(path), minProps), nil)
	}
	if maxProps, ok := schemaInt(schema["maxProperties"]); ok && len(value) > maxProps {
		return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must contain at most %d properties", displayPath(path), maxProps), nil)
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
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	case float64:
		if v == float64(int(v)) {
			return int(v), true
		}
	case json.Number:
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
	case json.Number:
		if n, err := v.Float64(); err == nil {
			return n, true
		}
	}
	return 0, false
}

func numericLessThan(value any, limit float64) bool {
	if n, ok := numericValue(value); ok {
		return n < limit
	}
	return false
}

func numericGreaterThan(value any, limit float64) bool {
	if n, ok := numericValue(value); ok {
		return n > limit
	}
	return false
}

type toolResponseRecorder struct {
	gin.ResponseWriter
	headers     http.Header
	body        bytes.Buffer
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

func (w *toolResponseRecorder) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(data)
}

func (w *toolResponseRecorder) WriteString(s string) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.WriteString(s)
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
	return w.body.Len()
}

func (w *toolResponseRecorder) Written() bool {
	return w.wroteHeader
}

func (w *toolResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, coreerr.E("ToolBridge.ResponseRecorder", "response hijacking is not supported by ToolBridge output validation", nil)
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
	_, _ = w.ResponseWriter.Write(w.body.Bytes())
}

func (w *toolResponseRecorder) reset() {
	w.headers = make(http.Header)
	w.body.Reset()
	w.status = http.StatusInternalServerError
	w.wroteHeader = false
}

func (w *toolResponseRecorder) writeErrorResponse(status int, resp Response[any]) {
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w.ResponseWriter, "internal server error", http.StatusInternalServerError)
		return
	}

	w.ResponseWriter.Header().Set("Content-Type", "application/json")
	w.ResponseWriter.WriteHeader(status)
	_, _ = w.ResponseWriter.Write(data)
}

func typeError(path, want string, value any) error {
	return coreerr.E("ToolBridge.ValidateSchema", fmt.Sprintf("%s must be %s, got %s", displayPath(path), want, describeJSONValue(value)), nil)
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
	case json.Number:
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
	case json.Number, float64:
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
		lv, lok := numericValue(left)
		rv, rok := numericValue(right)
		return lok && rok && lv == rv
	}
	return reflect.DeepEqual(left, right)
}

func isNumericValue(value any) bool {
	switch value.(type) {
	case json.Number, float64, float32, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

func numericValue(value any) (float64, bool) {
	switch v := value.(type) {
	case json.Number:
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

func describeJSONValue(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case bool:
		return "boolean"
	case json.Number, float64:
		return "number"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return fmt.Sprintf("%T", value)
	}
}
