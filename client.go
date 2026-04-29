// SPDX-License-Identifier: EUPL-1.2

package api

import (
	// Note: AX-6 — byte-slice JSON whitespace checks have no core byte-trim primitive.
	"dappco.re/go/api/internal/stdcompat/bytes"
	// Note: AX-6 — io.Reader API and HTTP body reads are structural stream boundaries.
	"io"
	// Note: AX-6 — iter.Seq is the public lazy iteration shape for operation/server snapshots.
	"iter"
	// Note: AX-6 — OpenAPIClient owns an outbound HTTP client boundary; no core HTTP client primitive.
	"net/http"
	// Note: AX-6 — URL joining, query values, and parsed URL fields are structural client boundary details.
	"net/url"
	// Note: AX-6 — reflect is structural for HTTP client param marshaling.
	"reflect"
	// Note: AX-6 — deterministic ordering and snapshot cloning need slices sort/clone helpers.
	"slices"

	core "dappco.re/go"

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
	specReader  io.Reader
	baseURL     string
	bearerToken string
	httpClient  *http.Client

	once       core.Once
	operations map[string]openAPIOperation
	servers    []string
	loadErr    error
}

// OpenAPIOperation snapshots the public metadata for a single loaded OpenAPI
// operation.
//
// Example:
//
//	ops, err := client.Operations()
//	if err == nil && len(ops) > 0 {
//		fmt.Println(ops[0].OperationID, ops[0].PathTemplate)
//	}
type OpenAPIOperation struct {
	OperationID    string
	Method         string
	PathTemplate   string
	HasRequestBody bool
	Parameters     []OpenAPIParameter
}

// OpenAPIParameter snapshots a single OpenAPI parameter definition.
//
// Example:
//
//	op, err := client.Operations()
//	if err == nil && len(op) > 0 && len(op[0].Parameters) > 0 {
//		fmt.Println(op[0].Parameters[0].Name, op[0].Parameters[0].In)
//	}
type OpenAPIParameter struct {
	Name     string
	In       string
	Required bool
	Schema   map[string]any
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
	name     string
	in       string
	required bool
	schema   map[string]any
}

// OpenAPIClientOption configures a runtime OpenAPI client.
//
// Example:
//
//	client := api.NewOpenAPIClient(api.WithSpec("./openapi.yaml"))
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

// WithSpecReader sets an in-memory or streamed OpenAPI document source.
// It is read once the first time the client loads its spec.
//
// Example:
//
//	client := api.NewOpenAPIClient(api.WithSpecReader(core.NewReader(spec)))
func WithSpecReader(reader io.Reader) OpenAPIClientOption {
	return func(c *OpenAPIClient) {
		c.specReader = reader
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

// Operations returns a snapshot of the operations loaded from the OpenAPI
// document.
//
// Example:
//
//	ops, err := client.Operations()
func (c *OpenAPIClient) Operations() ([]OpenAPIOperation, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	operations := make([]OpenAPIOperation, 0, len(c.operations))
	for operationID, op := range c.operations {
		operations = append(operations, snapshotOpenAPIOperation(operationID, op))
	}
	slices.SortStableFunc(operations, func(a, b OpenAPIOperation) int {
		if a.OperationID != b.OperationID {
			if a.OperationID < b.OperationID {
				return -1
			}
			return 1
		}
		if a.Method != b.Method {
			if a.Method < b.Method {
				return -1
			}
			return 1
		}
		if a.PathTemplate < b.PathTemplate {
			return -1
		}
		if a.PathTemplate > b.PathTemplate {
			return 1
		}
		return 0
	})
	return operations, nil
}

// OperationsIter returns an iterator over the loaded OpenAPI operations.
//
// Example:
//
//	ops, err := client.OperationsIter()
//	if err != nil {
//		panic(err)
//	}
//	for op := range ops {
//		fmt.Println(op.OperationID, op.PathTemplate)
//	}
func (c *OpenAPIClient) OperationsIter() (iter.Seq[OpenAPIOperation], error) {
	operations, err := c.Operations()
	if err != nil {
		return nil, err
	}

	return func(yield func(OpenAPIOperation) bool) {
		for _, op := range operations {
			if !yield(op) {
				return
			}
		}
	}, nil
}

// Servers returns a snapshot of the server URLs discovered from the OpenAPI
// document.
//
// Example:
//
//	servers, err := client.Servers()
func (c *OpenAPIClient) Servers() ([]string, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	return slices.Clone(c.servers), nil
}

// ServersIter returns an iterator over the server URLs discovered from the
// OpenAPI document.
//
// Example:
//
//	servers, err := client.ServersIter()
//	if err != nil {
//		panic(err)
//	}
//	for server := range servers {
//		fmt.Println(server)
//	}
func (c *OpenAPIClient) ServersIter() (iter.Seq[string], error) {
	servers, err := c.Servers()
	if err != nil {
		return nil, err
	}

	return func(yield func(string) bool) {
		for _, server := range servers {
			if !yield(server) {
				return
			}
		}
	}, nil
}

// Call invokes the operation with the given operationId.
//
// The params argument may be a map, struct, or nil. For convenience, a map may
// include `path`, "query", "header", "cookie", and "body" keys to explicitly
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
		return nil, core.E("OpenAPIClient.Call", core.Sprintf("operation %q not found in OpenAPI spec", operationID), nil)
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
		bodyReader = core.NewBuffer(body)
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

	resp, err := doHTTPClientRequest(c.httpClient, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, core.E("OpenAPIClient.Call", core.Sprintf("openapi call %s returned %s: %s", operationID, resp.Status, core.Trim(string(payload))), nil)
	}

	if op.responseSchema != nil && len(bytes.TrimSpace(payload)) > 0 {
		if err := validateOpenAPIResponse(payload, op.responseSchema, operationID); err != nil {
			return nil, err
		}
	}

	if len(bytes.TrimSpace(payload)) == 0 {
		return nil, nil
	}

	decoded, err := decodeJSONValuePreserveNumbers(payload)
	if err != nil {
		return string(payload), nil
	}

	if envelope, ok := decoded.(map[string]any); ok {
		if success, ok := envelope["success"].(bool); ok {
			if !success {
				if errObj, ok := envelope["error"].(map[string]any); ok {
					return nil, core.E("OpenAPIClient.Call", core.Sprintf("openapi call %s failed: %v", operationID, errObj), nil)
				}
				return nil, core.E("OpenAPIClient.Call", core.Sprintf("openapi call %s failed", operationID), nil)
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
	var (
		data []byte
		err  error
	)

	switch {
	case c.specReader != nil:
		data, err = io.ReadAll(c.specReader)
	case c.specPath != "":
		cfs := (&core.Fs{}).NewUnrestricted()
		r := cfs.Read(c.specPath)
		if !r.OK {
			return core.E("OpenAPIClient.loadSpec", "read spec", r.Value.(error))
		}
		data = []byte(r.Value.(string))
	default:
		return core.E("OpenAPIClient.loadSpec", "spec path or reader is required", nil)
	}

	if err != nil {
		return core.E("OpenAPIClient.loadSpec", "read spec", err)
	}

	var spec map[string]any
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return core.E("OpenAPIClient.loadSpec", "parse spec", err)
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
					method:         core.Upper(method),
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

func snapshotOpenAPIOperation(operationID string, op openAPIOperation) OpenAPIOperation {
	parameters := make([]OpenAPIParameter, len(op.parameters))
	for i, param := range op.parameters {
		parameters[i] = OpenAPIParameter{
			Name:     param.name,
			In:       param.in,
			Required: param.required,
			Schema:   cloneOpenAPIObject(param.schema),
		}
	}

	return OpenAPIOperation{
		OperationID:    operationID,
		Method:         core.Upper(op.method),
		PathTemplate:   op.pathTemplate,
		HasRequestBody: op.hasRequestBody,
		Parameters:     parameters,
	}
}

func (c *OpenAPIClient) buildURL(op openAPIOperation, params map[string]any) (string, error) {
	base := c.baseURL
	for core.HasSuffix(base, "/") {
		base = core.TrimSuffix(base, "/")
	}
	if base == "" {
		return "", core.E("OpenAPIClient.buildURL", "base URL is required", nil)
	}

	path := op.pathTemplate
	pathKeys := pathParameterNames(path)
	pathValues := map[string]any{}
	if explicitPath, ok := nestedMap(params, `path`); ok {
		pathValues = explicitPath
	} else {
		pathValues = params
	}

	if err := validateRequiredParameters(op, params, pathKeys); err != nil {
		return "", err
	}
	if err := validateParameterValues(op, params); err != nil {
		return "", err
	}

	for _, key := range pathKeys {
		if value, ok := pathValues[key]; ok {
			placeholder := "{" + key + "}"
			path = core.Replace(path, placeholder, core.URLPathEscape(core.Sprint(value)))
		}
	}

	if core.Contains(path, "{") {
		return "", core.E("OpenAPIClient.buildURL", core.Sprintf("missing path parameters for %q", op.pathTemplate), nil)
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
	for key, value := range params {
		if key == `path` || key == "body" || key == "query" || key == "header" || key == "cookie" {
			continue
		}
		if containsString(pathKeys, key) {
			continue
		}
		location := operationParameterLocation(op, key)
		if location != "query" && !(location == "" && (op.method == http.MethodGet || (op.method == http.MethodHead && !op.hasRequestBody))) {
			continue
		}
		if _, exists := query[key]; exists {
			continue
		}
		appendQueryValue(query, key, value)
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
		if key == `path` || key == "query" || key == "body" || key == "header" || key == "cookie" {
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
		if key == `path` || key == "query" || key == "body" || key == "header" || key == "cookie" {
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
		if key == `path` || key == "query" || key == "body" || key == "header" || key == "cookie" {
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
			headers.Add(key, core.Sprint(item))
		}
		return
	}

	rv := reflect.ValueOf(value)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && !(rv.Type().Elem().Kind() == reflect.Uint8) {
		for i := 0; i < rv.Len(); i++ {
			headers.Add(key, core.Sprint(rv.Index(i).Interface()))
		}
		return
	}

	headers.Set(key, core.Sprint(value))
}

func applyCookieValues(req *http.Request, values map[string]any) {
	for key, value := range values {
		applyCookieValue(req, key, value)
	}
}

// applyCookieValue attaches an outbound request cookie for the given key.
// All four AddCookie sites construct cookies for an OUTBOUND http.Request —
// they are written to the Cookie request header (RFC 6265 §5.4). Secure /
// HttpOnly / SameSite are response-only Set-Cookie attributes (§5.2) and
// have no effect on outbound request cookies. G124 false-positive — verified
// no path echoes these into a server-side http.SetCookie.
// Cerberus mechanism review attached to Mantis #321.
func applyCookieValue(req *http.Request, key string, value any) {
	switch v := value.(type) {
	case nil:
		return
	case []string:
		for _, item := range v {

			//#nosec G124 -- outbound request cookie, not Set-Cookie response.

			req.AddCookie(&http.Cookie{Name: key, Value: item})
		}
		return
	case []any:
		for _, item := range v {

			//#nosec G124 -- outbound request cookie, not Set-Cookie response.

			req.AddCookie(&http.Cookie{Name: key, Value: core.Sprint(item)})
		}
		return
	}

	rv := reflect.ValueOf(value)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && !(rv.Type().Elem().Kind() == reflect.Uint8) {
		for i := 0; i < rv.Len(); i++ {

			//#nosec G124 -- outbound request cookie, not Set-Cookie response.

			req.AddCookie(&http.Cookie{Name: key, Value: core.Sprint(rv.Index(i).Interface())})
		}
		return
	}

	//#nosec G124 -- outbound request cookie, not Set-Cookie response.

	req.AddCookie(&http.Cookie{Name: key, Value: core.Sprint(value)})
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
		required, _ := param["required"].(bool)
		schema, _ := param["schema"].(map[string]any)
		params = append(params, openAPIParameter{name: name, in: in, required: required, schema: schema})
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

func validateParameterValues(op openAPIOperation, params map[string]any) error {
	for _, param := range op.parameters {
		if len(param.schema) == 0 {
			continue
		}

		if nested, ok := nestedMap(params, param.in); ok {
			if value, exists := nested[param.name]; exists {
				if err := validateParameterValue(param, value); err != nil {
					return err
				}
				continue
			}
		}

		if value, exists := params[param.name]; exists {
			if err := validateParameterValue(param, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateParameterValue(param openAPIParameter, value any) error {
	if value == nil {
		return nil
	}

	data, err := marshalCoreJSON(value)
	if err != nil {
		return core.E("OpenAPIClient.validateParameterValue", core.Sprintf("marshal %s parameter %q", param.in, param.name), err)
	}
	if err := validateOpenAPISchema(data, param.schema, core.Sprintf("%s parameter %q", param.in, param.name)); err != nil {
		return err
	}
	return nil
}

func validateRequiredParameters(op openAPIOperation, params map[string]any, pathKeys []string) error {
	for _, param := range op.parameters {
		if !param.required {
			continue
		}
		if parameterProvided(params, param.name, param.in) {
			continue
		}
		return core.E("OpenAPIClient.buildURL", core.Sprintf("missing required %s parameter %q", param.in, param.name), nil)
	}
	return nil
}

func parameterProvided(params map[string]any, name, location string) bool {
	if nested, ok := nestedMap(params, location); ok {
		if value, exists := nested[name]; exists && value != nil {
			return true
		}
	}

	if value, exists := params[name]; exists {
		if value != nil {
			return true
		}
	}

	return false
}

func encodeJSONBody(v any) ([]byte, error) {
	return marshalCoreJSON(v)
}

func normaliseParams(params any) (map[string]any, error) {
	if params == nil {
		return map[string]any{}, nil
	}

	if m, ok := params.(map[string]any); ok {
		return m, nil
	}

	data, err := marshalCoreJSON(params)
	if err != nil {
		return nil, core.E("OpenAPIClient.normaliseParams", "marshal params", err)
	}

	decoded, err := decodeJSONValuePreserveNumbers(data)
	if err != nil {
		return nil, core.E("OpenAPIClient.normaliseParams", "decode params", err)
	}
	out, ok := decoded.(map[string]any)
	if !ok {
		return nil, core.E("OpenAPIClient.normaliseParams", "params must encode to an object", nil)
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

	data, err := marshalCoreJSON(raw)
	if err != nil {
		return nil, false
	}
	decoded, err := decodeJSONValuePreserveNumbers(data)
	if err != nil {
		return nil, false
	}
	m, ok = decoded.(map[string]any)
	if !ok {
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
		parts := core.SplitN(pathTemplate[i+1:], "}", 2)
		if len(parts) < 2 {
			break
		}
		name := parts[0]
		if name != "" {
			names = append(names, name)
		}
		i += len(name) + 1
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

	query.Add(key, core.Sprint(value))
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

	payload, err := decodeJSONValuePreserveNumbers(body)
	if err != nil {
		return core.E("OpenAPIClient.validateOpenAPISchema", core.Sprintf("validate %s: invalid JSON", label), err)
	}

	if err := validateSchemaNode(payload, schema, ""); err != nil {
		return core.E("OpenAPIClient.validateOpenAPISchema", core.Sprintf("validate %s", label), err)
	}

	return nil
}

func validateOpenAPIResponse(payload []byte, schema map[string]any, operationID string) error {
	decoded, err := decodeJSONValuePreserveNumbers(payload)
	if err != nil {
		return core.E("OpenAPIClient.validateOpenAPIResponse", core.Sprintf("openapi call %s returned invalid JSON", operationID), err)
	}

	if err := validateSchemaNode(decoded, schema, ""); err != nil {
		return core.E("OpenAPIClient.validateOpenAPIResponse", core.Sprintf("openapi call %s response does not match spec", operationID), err)
	}

	return nil
}
