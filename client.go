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
}

// OpenAPIClientOption configures a runtime OpenAPI client.
type OpenAPIClientOption func(*OpenAPIClient)

// WithSpec sets the filesystem path to the OpenAPI document.
func WithSpec(path string) OpenAPIClientOption {
	return func(c *OpenAPIClient) {
		c.specPath = path
	}
}

// WithBaseURL sets the base URL used for outgoing requests.
func WithBaseURL(baseURL string) OpenAPIClientOption {
	return func(c *OpenAPIClient) {
		c.baseURL = baseURL
	}
}

// WithBearerToken sets the Authorization bearer token used for requests.
func WithBearerToken(token string) OpenAPIClientOption {
	return func(c *OpenAPIClient) {
		c.bearerToken = token
	}
}

// WithHTTPClient sets the HTTP client used to execute requests.
func WithHTTPClient(client *http.Client) OpenAPIClientOption {
	return func(c *OpenAPIClient) {
		c.httpClient = client
	}
}

// NewOpenAPIClient constructs a runtime client for calling OpenAPI operations.
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
// include "path", "query", and "body" keys to explicitly control where the
// values are sent. When no explicit body is provided, requests with a declared
// requestBody send the remaining parameters as JSON.
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

	req, err := http.NewRequest(op.method, requestURL, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}

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
				operations[operationID] = openAPIOperation{
					method:         strings.ToUpper(method),
					pathTemplate:   pathTemplate,
					hasRequestBody: operation["requestBody"] != nil,
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

	if c.baseURL == "" && len(c.servers) > 0 {
		c.baseURL = c.servers[0]
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
			if key == "path" || key == "body" || key == "query" {
				continue
			}
			if containsString(pathKeys, key) {
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

func (c *OpenAPIClient) buildBody(op openAPIOperation, params map[string]any) (io.Reader, error) {
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
		if key == "path" || key == "query" || key == "body" {
			continue
		}
		if containsString(pathKeys, key) {
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

func encodeJSONBody(v any) (io.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
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
