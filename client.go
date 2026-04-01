// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// OpenAPIClient is a small runtime client that can call operations by their
// OpenAPI operationId. It loads the spec once, resolves the HTTP method and
// path for each operation, and performs JSON request/response handling.
//
// Example:
//
//	client := api.NewOpenAPIClient(api.WithSpec("./openapi.yaml"), api.WithBaseURL("https://api.example.com"))
//	data, err := client.Call("get_health", nil)
type OpenAPIClient struct {
	specPath    string
	baseURL     string
	bearerToken string
	httpClient  *http.Client

	once       sync.Once
	operations map[string]openAPIOperation
	servers    []string
	loadErr    error
}

type openAPIOperation struct {
	method         string
	pathTemplate   string
	hasRequestBody bool
	parameters     []openAPIParameter
	requestSchema  map[string]any
	responseSchema map[string]any
}

type openAPIParameter struct {
	name string
	in   string
}

// OpenAPIClientOption configures a runtime OpenAPI client.
type OpenAPIClientOption func(*OpenAPIClient)

// WithSpec sets the filesystem path to the OpenAPI document.
//
// Example:
//
//	client := api.NewOpenAPIClient(api.WithSpec("./openapi.yaml"))
func WithSpec(path string) OpenAPIClientOption {
	return func(c *OpenAPIClient) {
		c.specPath = path
	}
}

// WithBaseURL sets the base URL used for outgoing requests.
//
// Example:
//
//	client := api.NewOpenAPIClient(api.WithBaseURL("https://api.example.com"))
func WithBaseURL(baseURL string) OpenAPIClientOption {
	return func(c *OpenAPIClient) {
		c.baseURL = baseURL
	}
}

// WithBearerToken sets the Authorization bearer token used for requests.
//
// Example:
//
//	client := api.NewOpenAPIClient(
//		api.WithBaseURL("https://api.example.com"),
//		api.WithBearerToken("secret-token"),
//	)
func WithBearerToken(token string) OpenAPIClientOption {
	return func(c *OpenAPIClient) {
		c.bearerToken = token
	}
}

// WithHTTPClient sets the HTTP client used to execute requests.
//
// Example:
//
//	client := api.NewOpenAPIClient(api.WithHTTPClient(http.DefaultClient))
func WithHTTPClient(client *http.Client) OpenAPIClientOption {
	return func(c *OpenAPIClient) {
		c.httpClient = client
	}
}

// NewOpenAPIClient constructs a runtime client for calling OpenAPI operations.
//
// Example:
//
//	client := api.NewOpenAPIClient(api.WithSpec("./openapi.yaml"))
func NewOpenAPIClient(opts ...OpenAPIClientOption) *OpenAPIClient {
	c := &OpenAPIClient{
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c
}

// Call invokes the operation with the given operationId.
//
// The params argument may be a map, struct, or nil. For convenience, a map may
// include "path", "query", "header", "cookie", and "body" keys to explicitly
// control where the values are sent. When no explicit body is provided,
// requests with a declared requestBody send the remaining parameters as JSON.
//
// Example:
//
//	data, err := client.Call("create_item", map[string]any{"name": "alpha"})
func (c *OpenAPIClient) Call(operationID string, params any) (any, error) {
	if err := c.load(); err != nil {
		return nil, err
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	op, ok := c.operations[operationID]
	if !ok {
		return nil, fmt.Errorf("operation %q not found in OpenAPI spec", operationID)
	}

	merged, err := normaliseParams(params)
	if err != nil {
		return nil, err
	}

	requestURL, err := c.buildURL(op, merged)
	if err != nil {
		return nil, err
	}

	body, err := c.buildBody(op, merged)
	if err != nil {
		return nil, err
	}

	if op.requestSchema != nil && len(body) > 0 {
		if err := validateOpenAPISchema(body, op.requestSchema, "request body"); err != nil {
			return nil, err
		}
	}

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(op.method, requestURL, bodyReader)
	if err != nil {
		return nil, err
	}
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}
	applyRequestParameters(req, op, merged)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openapi call %s returned %s: %s", operationID, resp.Status, strings.TrimSpace(string(payload)))
	}

	if op.responseSchema != nil && len(bytes.TrimSpace(payload)) > 0 {
		if err := validateOpenAPIResponse(payload, op.responseSchema, operationID); err != nil {
			return nil, err
		}
	}

	if len(bytes.TrimSpace(payload)) == 0 {
		return nil, nil
	}

	var decoded any
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	if err := dec.Decode(&decoded); err != nil {
		return string(payload), nil
	}

	if envelope, ok := decoded.(map[string]any); ok {
		if success, ok := envelope["success"].(bool); ok {
			if !success {
				if errObj, ok := envelope["error"].(map[string]any); ok {
					return nil, fmt.Errorf("openapi call %s failed: %v", operationID, errObj)
				}
				return nil, fmt.Errorf("openapi call %s failed", operationID)
			}
			if data, ok := envelope["data"]; ok {
				return data, nil
			}
		}
	}

	return decoded, nil
}

func (c *OpenAPIClient) load() error {
	c.once.Do(func() {
		c.loadErr = c.loadSpec()
	})
	return c.loadErr
}

func (c *OpenAPIClient) loadSpec() error {
	if c.specPath == "" {
		return fmt.Errorf("spec path is required")
	}

	data, err := os.ReadFile(c.specPath)
	if err != nil {
		return fmt.Errorf("read spec: %w", err)
	}

	var spec map[string]any
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("parse spec: %w", err)
	}

	operations := make(map[string]openAPIOperation)
	if paths, ok := spec["paths"].(map[string]any); ok {
		for pathTemplate, rawPathItem := range paths {
			pathItem, ok := rawPathItem.(map[string]any)
			if !ok {
				continue
			}
			for method, rawOperation := range pathItem {
				operation, ok := rawOperation.(map[string]any)
				if !ok {
					continue
				}
				operationID, _ := operation["operationId"].(string)
				if operationID == "" {
					continue
				}
				params := parseOperationParameters(operation)
				operations[operationID] = openAPIOperation{
					method:         strings.ToUpper(method),
					pathTemplate:   pathTemplate,
					hasRequestBody: operation["requestBody"] != nil,
					parameters:     params,
					requestSchema:  requestBodySchema(operation),
					responseSchema: firstSuccessResponseSchema(operation),
				}
			}
		}
	}

	c.operations = operations
	if servers, ok := spec["servers"].([]any); ok {
		for _, rawServer := range servers {
			server, ok := rawServer.(map[string]any)
			if !ok {
				continue
			}
			if u, _ := server["url"].(string); u != "" {
				c.servers = append(c.servers, u)
			}
		}
	}
	c.servers = normaliseServers(c.servers)

	if c.baseURL == "" {
		for _, server := range c.servers {
			if isAbsoluteBaseURL(server) {
				c.baseURL = server
				break
			}
		}
	}

	return nil
}

func (c *OpenAPIClient) buildURL(op openAPIOperation, params map[string]any) (string, error) {
	base := strings.TrimRight(c.baseURL, "/")
	if base == "" {
		return "", fmt.Errorf("base URL is required")
	}

	path := op.pathTemplate
	pathKeys := pathParameterNames(path)
	pathValues := map[string]any{}
	if explicitPath, ok := nestedMap(params, "path"); ok {
		pathValues = explicitPath
	} else {
		pathValues = params
	}
	for _, key := range pathKeys {
		if value, ok := pathValues[key]; ok {
			placeholder := "{" + key + "}"
			path = strings.ReplaceAll(path, placeholder, url.PathEscape(fmt.Sprint(value)))
		}
	}

	if strings.Contains(path, "{") {
		return "", fmt.Errorf("missing path parameters for %q", op.pathTemplate)
	}

	fullURL, err := url.JoinPath(base, path)
	if err != nil {
		return "", err
	}

	query := url.Values{}
	if explicitQuery, ok := nestedMap(params, "query"); ok {
		for key, value := range explicitQuery {
			appendQueryValue(query, key, value)
		}
	}
	if op.method == http.MethodGet || (op.method == http.MethodHead && !op.hasRequestBody) {
		for key, value := range params {
			if key == "path" || key == "body" || key == "query" || key == "header" || key == "cookie" {
				continue
			}
			if containsString(pathKeys, key) {
				continue
			}
			if operationParameterLocation(op, key) == "header" || operationParameterLocation(op, key) == "cookie" {
				continue
			}
			if _, exists := query[key]; exists {
				continue
			}
			appendQueryValue(query, key, value)
		}
	}

	if encoded := query.Encode(); encoded != "" {
		fullURL += "?" + encoded
	}

	return fullURL, nil
}

func (c *OpenAPIClient) buildBody(op openAPIOperation, params map[string]any) ([]byte, error) {
	if explicitBody, ok := params["body"]; ok {
		return encodeJSONBody(explicitBody)
	}

	if op.method == http.MethodGet || (op.method == http.MethodHead && !op.hasRequestBody) {
		return nil, nil
	}

	if len(params) == 0 {
		return nil, nil
	}

	pathKeys := pathParameterNames(op.pathTemplate)
	queryKeys := map[string]struct{}{}
	if explicitQuery, ok := nestedMap(params, "query"); ok {
		for key := range explicitQuery {
			queryKeys[key] = struct{}{}
		}
	}

	payload := make(map[string]any, len(params))
	for key, value := range params {
		if key == "path" || key == "query" || key == "body" || key == "header" || key == "cookie" {
			continue
		}
		if containsString(pathKeys, key) {
			continue
		}
		switch operationParameterLocation(op, key) {
		case "header", "cookie", "query":
			continue
		}
		if _, exists := queryKeys[key]; exists {
			continue
		}
		payload[key] = value
	}
	if len(payload) == 0 {
		return nil, nil
	}
	return encodeJSONBody(payload)
}

func applyRequestParameters(req *http.Request, op openAPIOperation, params map[string]any) {
	explicitHeaders, hasExplicitHeaders := nestedMap(params, "header")
	explicitCookies, hasExplicitCookies := nestedMap(params, "cookie")

	if hasExplicitHeaders {
		applyHeaderValues(req.Header, explicitHeaders)
	}

	applyTopLevelHeaderParameters(req.Header, op, params, explicitHeaders, hasExplicitHeaders)

	if hasExplicitCookies {
		applyCookieValues(req, explicitCookies)
	}
	applyTopLevelCookieParameters(req, op, params, explicitCookies, hasExplicitCookies)
}

func applyTopLevelHeaderParameters(headers http.Header, op openAPIOperation, params, explicit map[string]any, hasExplicit bool) {
	for key, value := range params {
		if key == "path" || key == "query" || key == "body" || key == "header" || key == "cookie" {
			continue
		}
		if operationParameterLocation(op, key) != "header" {
			continue
		}
		if hasExplicit {
			if _, ok := explicit[key]; ok {
				continue
			}
		}
		applyHeaderValue(headers, key, value)
	}
}

func applyTopLevelCookieParameters(req *http.Request, op openAPIOperation, params, explicit map[string]any, hasExplicit bool) {
	for key, value := range params {
		if key == "path" || key == "query" || key == "body" || key == "header" || key == "cookie" {
			continue
		}
		if operationParameterLocation(op, key) != "cookie" {
			continue
		}
		if hasExplicit {
			if _, ok := explicit[key]; ok {
				continue
			}
		}
		applyCookieValue(req, key, value)
	}
}

func applyHeaderValues(headers http.Header, values map[string]any) {
	for key, value := range values {
		applyHeaderValue(headers, key, value)
	}
}

func applyHeaderValue(headers http.Header, key string, value any) {
	switch v := value.(type) {
	case nil:
		return
	case []string:
		for _, item := range v {
			headers.Add(key, item)
		}
		return
	case []any:
		for _, item := range v {
			headers.Add(key, fmt.Sprint(item))
		}
		return
	}

	rv := reflect.ValueOf(value)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && !(rv.Type().Elem().Kind() == reflect.Uint8) {
		for i := 0; i < rv.Len(); i++ {
			headers.Add(key, fmt.Sprint(rv.Index(i).Interface()))
		}
		return
	}

	headers.Set(key, fmt.Sprint(value))
}

func applyCookieValues(req *http.Request, values map[string]any) {
	for key, value := range values {
		applyCookieValue(req, key, value)
	}
}

func applyCookieValue(req *http.Request, key string, value any) {
	switch v := value.(type) {
	case nil:
		return
	case []string:
		for _, item := range v {
			req.AddCookie(&http.Cookie{Name: key, Value: item})
		}
		return
	case []any:
		for _, item := range v {
			req.AddCookie(&http.Cookie{Name: key, Value: fmt.Sprint(item)})
		}
		return
	}

	rv := reflect.ValueOf(value)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && !(rv.Type().Elem().Kind() == reflect.Uint8) {
		for i := 0; i < rv.Len(); i++ {
			req.AddCookie(&http.Cookie{Name: key, Value: fmt.Sprint(rv.Index(i).Interface())})
		}
		return
	}

	req.AddCookie(&http.Cookie{Name: key, Value: fmt.Sprint(value)})
}

func parseOperationParameters(operation map[string]any) []openAPIParameter {
	rawParams, ok := operation["parameters"].([]any)
	if !ok {
		return nil
	}

	params := make([]openAPIParameter, 0, len(rawParams))
	for _, rawParam := range rawParams {
		param, ok := rawParam.(map[string]any)
		if !ok {
			continue
		}
		name, _ := param["name"].(string)
		in, _ := param["in"].(string)
		if name == "" || in == "" {
			continue
		}
		params = append(params, openAPIParameter{name: name, in: in})
	}

	return params
}

func operationParameterLocation(op openAPIOperation, name string) string {
	for _, param := range op.parameters {
		if param.name == name {
			return param.in
		}
	}
	return ""
}

func encodeJSONBody(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func normaliseParams(params any) (map[string]any, error) {
	if params == nil {
		return map[string]any{}, nil
	}

	if m, ok := params.(map[string]any); ok {
		return m, nil
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode params: %w", err)
	}

	return out, nil
}

func nestedMap(params map[string]any, key string) (map[string]any, bool) {
	raw, ok := params[key]
	if !ok {
		return nil, false
	}

	m, ok := raw.(map[string]any)
	if ok {
		return m, true
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return nil, false
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, false
	}
	return m, true
}

func pathParameterNames(pathTemplate string) []string {
	var names []string
	for i := 0; i < len(pathTemplate); i++ {
		if pathTemplate[i] != '{' {
			continue
		}
		end := strings.IndexByte(pathTemplate[i+1:], '}')
		if end < 0 {
			break
		}
		name := pathTemplate[i+1 : i+1+end]
		if name != "" {
			names = append(names, name)
		}
		i += end + 1
	}
	return names
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func appendQueryValue(query url.Values, key string, value any) {
	switch v := value.(type) {
	case nil:
		return
	case []byte:
		query.Add(key, string(v))
		return
	case []string:
		for _, item := range v {
			query.Add(key, item)
		}
		return
	case []any:
		for _, item := range v {
			appendQueryValue(query, key, item)
		}
		return
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			query.Add(key, string(rv.Bytes()))
			return
		}
		for i := 0; i < rv.Len(); i++ {
			appendQueryValue(query, key, rv.Index(i).Interface())
		}
		return
	}

	query.Add(key, fmt.Sprint(value))
}

func isAbsoluteBaseURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func requestBodySchema(operation map[string]any) map[string]any {
	rawRequestBody, ok := operation["requestBody"].(map[string]any)
	if !ok {
		return nil
	}

	content, ok := rawRequestBody["content"].(map[string]any)
	if !ok {
		return nil
	}

	rawJSON, ok := content["application/json"].(map[string]any)
	if !ok {
		return nil
	}

	schema, _ := rawJSON["schema"].(map[string]any)
	return schema
}

func firstSuccessResponseSchema(operation map[string]any) map[string]any {
	responses, ok := operation["responses"].(map[string]any)
	if !ok {
		return nil
	}

	for _, code := range []string{"200", "201", "202", "203", "204", "205", "206", "207", "208", "226"} {
		rawResp, ok := responses[code].(map[string]any)
		if !ok {
			continue
		}
		content, ok := rawResp["content"].(map[string]any)
		if !ok {
			continue
		}
		rawJSON, ok := content["application/json"].(map[string]any)
		if !ok {
			continue
		}
		schema, _ := rawJSON["schema"].(map[string]any)
		if len(schema) > 0 {
			return schema
		}
	}

	return nil
}

func validateOpenAPISchema(body []byte, schema map[string]any, label string) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}

	var payload any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return fmt.Errorf("validate %s: invalid JSON: %w", label, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("validate %s: expected a single JSON value", label)
	}

	if err := validateSchemaNode(payload, schema, ""); err != nil {
		return fmt.Errorf("validate %s: %w", label, err)
	}

	return nil
}

func validateOpenAPIResponse(payload []byte, schema map[string]any, operationID string) error {
	var decoded any
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	if err := dec.Decode(&decoded); err != nil {
		return fmt.Errorf("openapi call %s returned invalid JSON: %w", operationID, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("openapi call %s returned multiple JSON values", operationID)
	}

	if err := validateSchemaNode(decoded, schema, ""); err != nil {
		return fmt.Errorf("openapi call %s response does not match spec: %w", operationID, err)
	}

	return nil
}
