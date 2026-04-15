// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient wraps a websocket.Dialer with predictable defaults for
// connecting to the framework's WebSocket endpoints.
//
// Example:
//
//	client := api.NewWebSocketClient("ws://localhost:8080/ws")
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
//	client := api.NewWebSocketClient("ws://localhost:8080/ws")
func NewWebSocketClient(rawURL string, opts ...WebSocketClientOption) *WebSocketClient {
	c := &WebSocketClient{
		URL:    strings.TrimSpace(rawURL),
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
// The client accepts ws://, wss://, http://, and https:// URLs. HTTP schemes
// are converted to their WebSocket equivalents automatically.
//
// Example:
//
//	conn, resp, err := client.DialContext(ctx)
func (c *WebSocketClient) DialContext(ctx context.Context) (*websocket.Conn, *http.Response, error) {
	if c == nil {
		return nil, nil, errors.New("WebSocketClient is nil")
	}

	rawURL, err := normaliseWebSocketClientURL(c.URL)
	if err != nil {
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
//	    fmt.Println(evt.Event, evt.Data)
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
		URL:    strings.TrimSpace(rawURL),
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
		return nil, errors.New("SSEClient is nil")
	}

	rawURL := strings.TrimSpace(c.URL)
	if rawURL == "" {
		return nil, errors.New("SSEClient URL is required")
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

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("unexpected SSE status %d", resp.StatusCode)
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
		current.Data = strings.Join(dataLines, "\n")
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
		if strings.HasPrefix(line, ":") {
			continue
		}

		field, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimPrefix(value, " ")

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

	_ = emit()
}

func normaliseWebSocketClientURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("WebSocketClient URL is required")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	switch parsed.Scheme {
	case "ws", "wss":
		return parsed.String(), nil
	case "http":
		parsed.Scheme = "ws"
		return parsed.String(), nil
	case "https":
		parsed.Scheme = "wss"
		return parsed.String(), nil
	default:
		return "", fmt.Errorf("unsupported websocket URL scheme %q", parsed.Scheme)
	}
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
