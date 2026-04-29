// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio" // Note: AX-6 — SSE stream line scanning
	"context"
	"dappco.re/go/api/internal/stdcompat/errors"
	"io"       // Note: AX-6 — io.Reader contract
	"net/http" // Note: AX-6 — HTTP transport boundary
	"net/url"
	"strconv"
	"time"

	core "dappco.re/go"
	coreerr "dappco.re/go/log"

	"github.com/gorilla/websocket"
)

// WebSocketClient wraps a websocket.Dialer with predictable defaults for
// connecting to the framework's WebSocket endpoints.
//
// Example:
//
//	client := api.NewWebSocketClient("wss://api.example.com/ws")
//	conn, resp, err := client.DialContext(ctx)
type WebSocketClient struct {
	URL    string
	Header http.Header
	Dialer *websocket.Dialer
}

// WebSocketClientOption configures a WebSocketClient.
type WebSocketClientOption func(*WebSocketClient)

// WithWebSocketHeaders adds request headers to the outbound WebSocket dial.
//
// Example:
//
//	client := api.NewWebSocketClient(
//	    "wss://api.example.com/ws",
//	    api.WithWebSocketHeaders(http.Header{"Authorization": {"Bearer secret"}}),
//	)
func WithWebSocketHeaders(header http.Header) WebSocketClientOption {
	return func(c *WebSocketClient) {
		if len(header) == 0 {
			return
		}
		if c.Header == nil {
			c.Header = make(http.Header, len(header))
		}
		for key, values := range header {
			if len(values) == 0 {
				continue
			}
			c.Header[key] = append([]string(nil), values...)
		}
	}
}

// WithWebSocketDialer sets the dialer used to establish the connection.
//
// Example:
//
//	client := api.NewWebSocketClient(url, api.WithWebSocketDialer(dialer))
func WithWebSocketDialer(dialer *websocket.Dialer) WebSocketClientOption {
	return func(c *WebSocketClient) {
		c.Dialer = dialer
	}
}

// NewWebSocketClient constructs a WebSocketClient with sensible defaults.
//
// Example:
//
//	client := api.NewWebSocketClient("wss://api.example.com/ws")
func NewWebSocketClient(rawURL string, opts ...WebSocketClientOption) *WebSocketClient {
	c := &WebSocketClient{
		URL:    core.Trim(rawURL),
		Header: make(http.Header),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

// DialContext opens the WebSocket connection and returns the negotiated
// connection plus the HTTP upgrade response.
//
// The client accepts ws:// and wss:// URLs. The target is validated against
// the outbound SSRF guard before the handshake is attempted.
//
// Example:
//
//	conn, resp, err := client.DialContext(ctx)
func (c *WebSocketClient) DialContext(ctx context.Context) (*websocket.Conn, *http.Response, error) {
	if c == nil {
		return nil, nil, coreerr.E("", "WebSocketClient is nil", nil)
	}

	rawURL, err := normaliseWebSocketClientURL(c.URL)
	if err != nil {
		return nil, nil, err
	}
	if err := validateOutboundWebSocketClientURL(rawURL); err != nil {
		return nil, nil, err
	}

	dialer := websocket.DefaultDialer
	if c.Dialer != nil {
		copyDialer := *c.Dialer
		dialer = &copyDialer
	} else {
		copyDialer := *websocket.DefaultDialer
		dialer = &copyDialer
	}

	header := cloneHTTPHeader(c.Header)
	return dialer.DialContext(ctx, rawURL, header)
}

// SSEEvent is a parsed Server-Sent Events message.
//
// Example:
//
//	for evt := range events {
//	    _ = evt
//	}
type SSEEvent struct {
	ID    string
	Event string
	Data  string
	Retry time.Duration
}

// SSEClient is a small client helper for consuming text/event-stream
// endpoints exposed by SSEBroker.
//
// Example:
//
//	client := api.NewSSEClient("http://localhost:8080/events")
//	events, err := client.Events(ctx)
type SSEClient struct {
	URL    string
	Header http.Header
	Client *http.Client
}

// SSEClientOption configures an SSEClient.
type SSEClientOption func(*SSEClient)

// WithSSEHeaders adds request headers to the outbound SSE request.
func WithSSEHeaders(header http.Header) SSEClientOption {
	return func(c *SSEClient) {
		if len(header) == 0 {
			return
		}
		if c.Header == nil {
			c.Header = make(http.Header, len(header))
		}
		for key, values := range header {
			if len(values) == 0 {
				continue
			}
			c.Header[key] = append([]string(nil), values...)
		}
	}
}

// WithSSEHTTPClient sets the HTTP client used to establish the SSE stream.
func WithSSEHTTPClient(client *http.Client) SSEClientOption {
	return func(c *SSEClient) {
		c.Client = client
	}
}

// NewSSEClient constructs an SSEClient with sensible defaults.
//
// Example:
//
//	client := api.NewSSEClient("http://localhost:8080/events")
func NewSSEClient(rawURL string, opts ...SSEClientOption) *SSEClient {
	c := &SSEClient{
		URL:    core.Trim(rawURL),
		Header: make(http.Header),
		Client: http.DefaultClient,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	if c.Client == nil {
		c.Client = http.DefaultClient
	}
	return c
}

// Connect opens the SSE stream and returns the underlying HTTP response.
//
// Example:
//
//	resp, err := client.Connect(ctx)
func (c *SSEClient) Connect(ctx context.Context) (*http.Response, error) {
	if c == nil {
		return nil, coreerr.E("", "SSEClient is nil", nil)
	}

	rawURL := core.Trim(c.URL)
	if rawURL == "" {
		return nil, coreerr.E("", "SSEClient URL is required", nil)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = cloneHTTPHeader(c.Header)
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("Accept", "text/event-stream")

	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := doHTTPClientRequest(client, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, coreerr.E("", core.Sprintf("unexpected SSE status %d", resp.StatusCode), nil)
	}
	return resp, nil
}

// Events connects to the stream and returns a channel of parsed SSE events.
// The channel is closed when the stream ends or the context is cancelled.
//
// Example:
//
//	events, err := client.Events(ctx)
//	if err != nil { ... }
//	for evt := range events { _ = evt }
func (c *SSEClient) Events(ctx context.Context) (<-chan SSEEvent, error) {
	resp, err := c.Connect(ctx)
	if err != nil {
		return nil, err
	}

	out := make(chan SSEEvent)
	go func() {
		defer close(out)
		defer resp.Body.Close()
		parseSSEStream(ctx, resp.Body, out)
	}()

	return out, nil
}

func parseSSEStream(ctx context.Context, body io.Reader, out chan<- SSEEvent) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var current SSEEvent
	var dataLines []string
	emit := func() bool {
		if current.Event == "" && current.ID == "" && len(dataLines) == 0 {
			current = SSEEvent{}
			return true
		}
		current.Data = core.Join("\n", dataLines...)
		select {
		case out <- current:
		case <-ctx.Done():
			return false
		}
		current = SSEEvent{}
		dataLines = dataLines[:0]
		return true
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			if !emit() {
				return
			}
			continue
		}
		if core.HasPrefix(line, ":") {
			continue
		}

		parts := core.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		field, value := parts[0], parts[1]
		value = core.TrimPrefix(value, " ")

		switch field {
		case "event":
			current.Event = value
		case "data":
			dataLines = append(dataLines, value)
		case "id":
			current.ID = value
		case "retry":
			if ms, err := time.ParseDuration(value + "ms"); err == nil {
				current.Retry = ms
			}
		}
	}

	if !emit() {
		return
	}
}

func normaliseWebSocketClientURL(rawURL string) (string, error) {
	rawURL = core.Trim(rawURL)
	if rawURL == "" {
		return "", core.E("", "WebSocketClient URL is required", nil)
	}

	parsedResult := core.URLParse(rawURL)
	if !parsedResult.OK {
		if err, ok := parsedResult.Value.(error); ok {
			return "", invalidWebSocketClientURLError(err)
		}
		return "", invalidWebSocketClientURLError(nil)
	}
	parsed, ok := parsedResult.Value.(*url.URL)
	if !ok || parsed == nil {
		return "", invalidWebSocketClientURLError(nil)
	}

	scheme := core.Lower(parsed.Scheme)
	if scheme == "" {
		return "", invalidWebSocketClientURLError(nil)
	}

	switch scheme {
	case "ws", "wss":
		if parsed.Host == "" {
			return "", invalidWebSocketClientURLError(nil)
		}
		if err := validateWebSocketClientPort(parsed); err != nil {
			return "", err
		}
		normalized := parsed.String()
		if normalized == "" {
			return "", invalidWebSocketClientURLError(nil)
		}
		return normalized, nil
	default:
		return "", wrapBlocked(core.Sprintf("disallowed websocket scheme: %s", scheme))
	}
}

func invalidWebSocketClientURLError(err error) error {
	return core.E("", "invalid WebSocketClient URL", err)
}

func validateWebSocketClientPort(parsed *url.URL) error {
	if parsed == nil {
		return invalidWebSocketClientURLError(nil)
	}
	port := parsed.Port()
	if port == "" {
		if core.HasSuffix(parsed.Host, ":") {
			return invalidWebSocketClientURLError(nil)
		}
		return nil
	}
	n, err := strconv.Atoi(port)
	if err != nil || n > 65535 {
		return invalidWebSocketClientURLError(err)
	}
	return nil
}

func validateOutboundWebSocketClientURL(rawURL string) error {
	guardURL, err := outboundWebSocketGuardURL(rawURL)
	if err != nil {
		return err
	}
	return validateOutboundURL(guardURL)
}

func outboundWebSocketGuardURL(rawURL string) (string, error) {
	parsedResult := core.URLParse(rawURL)
	if !parsedResult.OK {
		if err, ok := parsedResult.Value.(error); ok {
			return "", err
		}
		return "", coreerr.E("", "invalid WebSocketClient URL", nil)
	}
	parsed, ok := parsedResult.Value.(*url.URL)
	if !ok || parsed == nil {
		return "", coreerr.E("", "invalid WebSocketClient URL", nil)
	}

	guardURL := *parsed
	switch core.Lower(guardURL.Scheme) {
	case "ws":
		guardURL.Scheme = "http"
	case "wss":
		guardURL.Scheme = "https"
	default:
		return "", wrapBlocked(core.Sprintf("disallowed websocket scheme: %s", guardURL.Scheme))
	}
	return guardURL.String(), nil
}

func cloneHTTPHeader(header http.Header) http.Header {
	if len(header) == 0 {
		return nil
	}
	out := make(http.Header, len(header))
	for key, values := range header {
		if len(values) == 0 {
			continue
		}
		out[key] = append([]string(nil), values...)
	}
	return out
}

// doHTTPClientRequest is the singular choke point for outbound HTTP from
// SSEClient.Connect and OpenAPIClient.Call. It validates the request URL and
// any followed redirects against the deny-by-default outbound policy (see
// ssrf_guard.go) before invoking client.Do. Cerberus mechanism review attached
// to Mantis #318.
func doHTTPClientRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	if client == nil {
		client = http.DefaultClient
	}

	if req != nil && req.URL != nil {
		if err := validateOutboundURL(req.URL.String()); err != nil {
			return nil, err
		}
	}

	requestClient := clientWithOutboundRedirectGuard(client)

	//#nosec G107 G704 -- initial URL and followed redirects validated by validateOutboundURL deny-by-default policy.
	resp, err := requestClient.Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				return nil, errors.Join(err, closeErr)
			}
		}

		return nil, err
	}

	return resp, nil
}

func clientWithOutboundRedirectGuard(client *http.Client) *http.Client {
	guarded := *client
	checkRedirect := client.CheckRedirect
	guarded.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if checkRedirect != nil {
			if err := checkRedirect(req, via); err != nil {
				return err
			}
		} else if len(via) >= 10 {
			return coreerr.E("", "stopped after 10 redirects", nil)
		}

		if req != nil && req.URL != nil {
			if err := validateOutboundURL(req.URL.String()); err != nil {
				return err
			}
		}

		return nil
	}
	return &guarded
}
