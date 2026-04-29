// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"crypto/tls"
	errors "dappco.re/go/api/internal/stdcompat/coreerrors"
	strings "dappco.re/go/api/internal/stdcompat/corestrings"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"dappco.re/go"

	"github.com/gorilla/websocket"
)

type trackingReadCloser struct {
	io.ReadCloser
	closed *bool
}

func (t trackingReadCloser) Close() error {
	*t.closed = true
	return t.ReadCloser.Close()
}

type trackingRoundTripper struct {
	base   http.RoundTripper
	closed *bool
}

func (t trackingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil || resp == nil || resp.Body == nil {
		return resp, err
	}

	resp.Body = trackingReadCloser{
		ReadCloser: resp.Body,
		closed:     t.closed,
	}

	return resp, err
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func publicServerURL(t *testing.T, srv *httptest.Server) string {
	t.Helper()

	parsed, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	parsed.Host = "93.184.216.34"
	return parsed.String()
}

func localServerTransport(t *testing.T, srv *httptest.Server) http.RoundTripper {
	t.Helper()

	targetAddr := srv.Listener.Addr().String()
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, targetAddr)
	}
	return transport
}

func publicSSEClient(t *testing.T, srv *httptest.Server, opts ...SSEClientOption) *SSEClient {
	t.Helper()

	options := []SSEClientOption{
		WithSSEHTTPClient(&http.Client{Transport: localServerTransport(t, srv)}),
	}
	options = append(options, opts...)
	return NewSSEClient(publicServerURL(t, srv), options...)
}

func TestTransportClient_WithWebSocketHeaders_Good_CopiesValues(t *testing.T) {
	client := &WebSocketClient{}
	source := http.Header{
		"Authorization": {"Bearer secret"},
		"X-Trace":       {"abc", "def"},
	}

	WithWebSocketHeaders(source)(client)

	if client.Header == nil {
		t.Fatal("expected header map to be initialised")
	}
	if got := client.Header.Values("Authorization"); len(got) != 1 || got[0] != "Bearer secret" {
		t.Fatalf("expected Authorization header to be copied, got %v", got)
	}
	if got := client.Header.Values("X-Trace"); len(got) != 2 || got[0] != "abc" || got[1] != "def" {
		t.Fatalf("expected multi-value header to be copied, got %v", got)
	}
}

func TestTransportClient_WithWebSocketHeaders_Bad_IgnoresEmptyInput(t *testing.T) {
	client := &WebSocketClient{Header: http.Header{"X-Existing": {"keep"}}}

	WithWebSocketHeaders(http.Header{})(client)

	if got := client.Header.Get("X-Existing"); got != "keep" {
		t.Fatalf("expected existing header to remain untouched, got %q", got)
	}
}

func TestTransportClient_WithWebSocketHeaders_Ugly_ClonesSourceSlices(t *testing.T) {
	client := &WebSocketClient{}
	source := http.Header{"Authorization": {"Bearer secret"}}

	WithWebSocketHeaders(source)(client)
	source["Authorization"][0] = "mutated"

	if got := client.Header.Get("Authorization"); got != "Bearer secret" {
		t.Fatalf("expected copied header to be isolated from source mutation, got %q", got)
	}
}

func TestTransportClient_WithWebSocketDialer_Good_AssignsDialer(t *testing.T) {
	client := &WebSocketClient{}
	dialer := &websocket.Dialer{HandshakeTimeout: time.Second}

	WithWebSocketDialer(dialer)(client)

	if client.Dialer != dialer {
		t.Fatal("expected dialer pointer to be assigned")
	}
}

func TestTransportClient_NewWebSocketClient_Good_TrimsURLAndAllocatesHeader(t *testing.T) {
	client := NewWebSocketClient("  ws://example.invalid/ws  ")

	if client.URL != "ws://example.invalid/ws" {
		t.Fatalf("expected trimmed URL, got %q", client.URL)
	}
	if client.Header == nil {
		t.Fatal("expected header map to be initialised")
	}
}

func TestTransportClient_NewWebSocketClient_Bad_LeavesEmptyURLUsable(t *testing.T) {
	client := NewWebSocketClient("   ", nil)

	if client.URL != "" {
		t.Fatalf("expected empty URL after trimming whitespace, got %q", client.URL)
	}
	if client.Header == nil {
		t.Fatal("expected header map to be initialised even for empty input")
	}
}

func TestTransportClient_NewWebSocketClient_Ugly_IgnoresNilOptions(t *testing.T) {
	client := NewWebSocketClient("ws://example.invalid/ws", nil)

	if client.URL != "ws://example.invalid/ws" {
		t.Fatalf("expected URL to remain unchanged, got %q", client.URL)
	}
}

func TestTransportClient_DialContext_Good_DialsWSSURLAndSendsHeaders(t *testing.T) {
	prevResolveHost := resolveHost
	t.Cleanup(func() { resolveHost = prevResolveHost })
	resolveHost = func(host string) ([]net.IP, error) {
		if host != "public.example.com" {
			return []net.IP{net.IPv4(127, 0, 0, 1)}, nil
		}
		return []net.IP{net.IPv4(93, 184, 216, 34)}, nil
	}

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	sawHeaderCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawHeaderCh <- r.Header.Get("Authorization")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			errCh <- err
			return
		}
		defer conn.Close()
		_ = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
	}))
	defer srv.Close()

	targetAddr := srv.Listener.Addr().String()
	var dialedAddr string
	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Test server uses a self-signed certificate.
		NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialedAddr = addr
			return (&net.Dialer{}).DialContext(ctx, network, targetAddr)
		},
	}

	client := NewWebSocketClient(
		"wss://public.example.com/ws",
		WithWebSocketHeaders(http.Header{"Authorization": {"Bearer secret"}}),
		WithWebSocketDialer(dialer),
	)

	conn, resp, err := client.DialContext(context.Background())
	if err != nil {
		t.Fatalf("DialContext failed: %v", err)
	}
	defer conn.Close()

	if resp == nil || resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected websocket upgrade response, got %+v", resp)
	}
	if dialedAddr != "public.example.com:443" {
		t.Fatalf("expected mock dialer to receive public.example.com:443, got %q", dialedAddr)
	}
	var sawHeader string
	select {
	case sawHeader = <-sawHeaderCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for websocket request")
	}
	if sawHeader != "Bearer secret" {
		t.Fatalf("expected Authorization header to reach server, got %q", sawHeader)
	}
	select {
	case err := <-errCh:
		t.Fatalf("upgrade failed: %v", err)
	default:
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if string(msg) != "hello" {
		t.Fatalf("expected message %q, got %q", "hello", string(msg))
	}
}

func TestTransportClient_DialContext_Bad_BlocksSSRFWebSocketTargets(t *testing.T) {
	prevResolveHost := resolveHost
	t.Cleanup(func() { resolveHost = prevResolveHost })
	resolveHost = func(host string) ([]net.IP, error) {
		switch host {
		case "localhost", "hostname-resolving-to-loopback.example":
			return []net.IP{net.IPv4(127, 0, 0, 1)}, nil
		default:
			return []net.IP{net.IPv4(93, 184, 216, 34)}, nil
		}
	}

	cases := []struct {
		name string
		raw  string
	}{
		{name: "metadata", raw: "ws://169.254.169.254/latest/meta-data/"},
		{name: "localhost", raw: "wss://localhost/ws"},
		{name: "rfc1918", raw: "wss://10.0.0.1/ws"},
		{name: "dns-loopback", raw: "ws://hostname-resolving-to-loopback.example/ws"},
		{name: "scheme", raw: "file:///etc/passwd"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			dialCalls := 0
			dialer := &websocket.Dialer{
				NetDialContext: func(context.Context, string, string) (net.Conn, error) {
					dialCalls++
					return nil, errors.New("dial should not be called")
				},
				NetDialTLSContext: func(context.Context, string, string) (net.Conn, error) {
					dialCalls++
					return nil, errors.New("dial should not be called")
				},
			}

			conn, resp, err := NewWebSocketClient(tt.raw, WithWebSocketDialer(dialer)).DialContext(context.Background())
			if err == nil {
				if conn != nil {
					_ = conn.Close()
				}
				if resp != nil && resp.Body != nil {
					_ = resp.Body.Close()
				}
				t.Fatal("expected websocket target to be blocked")
			}
			if !errors.Is(err, errOutboundURLBlocked) {
				t.Fatalf("expected errOutboundURLBlocked, got %v", err)
			}
			if dialCalls != 0 {
				t.Fatalf("expected guard to block before dialing, got %d dial attempts", dialCalls)
			}
		})
	}
}

func TestTransportClient_DialContext_Bad_RejectsNilReceiverAndUnsupportedScheme(t *testing.T) {
	var client *WebSocketClient
	if _, _, err := client.DialContext(context.Background()); err == nil {
		t.Fatal("expected nil receiver to fail")
	}

	for _, raw := range []string{"ftp://example.invalid/ws", "http://example.invalid/ws", "https://example.invalid/ws"} {
		client = NewWebSocketClient(raw)
		if _, _, err := client.DialContext(context.Background()); err == nil {
			t.Fatalf("expected unsupported scheme to fail for %s", raw)
		}
	}
}

func TestTransportClient_normaliseWebSocketClientURL_Bad_ReturnsErrorsForMalformedInput(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{name: "control-char", raw: "ws://example.invalid/a\nb"},
		{name: "malformed-escape", raw: "ws://example.invalid/%2"},
		{name: "empty", raw: ""},
		{name: "no-scheme", raw: "example.invalid/ws"},
		{name: "port-out-of-range", raw: "ws://example.invalid:65536/ws"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("expected malformed URL error, got panic: %v", recovered)
				}
			}()

			normalized, err := normaliseWebSocketClientURL(tt.raw)
			if err == nil {
				t.Fatalf("expected malformed URL to fail, got normalized URL %q", normalized)
			}
			var typed *core.Err
			if !errors.As(err, &typed) {
				t.Fatalf("expected typed core error, got %T: %v", err, err)
			}
		})
	}
}

func TestTransportClient_NewSSEClient_Good_TrimsURLAndDefaultsHTTPClient(t *testing.T) {
	client := NewSSEClient("  http://example.invalid/events  ")

	if client.URL != "http://example.invalid/events" {
		t.Fatalf("expected trimmed URL, got %q", client.URL)
	}
	if client.Client == nil {
		t.Fatal("expected HTTP client to default to http.DefaultClient")
	}
}

func TestTransportClient_WithSSEHeaders_Good_CopiesValues(t *testing.T) {
	client := &SSEClient{}
	source := http.Header{"X-Request-ID": {"abc"}}

	WithSSEHeaders(source)(client)

	if got := client.Header["X-Request-ID"]; len(got) != 1 || got[0] != "abc" {
		t.Fatalf("expected X-Request-ID header to be copied, got %v", got)
	}
}

func TestTransportClient_WithSSEHTTPClient_Good_AssignsClient(t *testing.T) {
	client := &SSEClient{}
	httpClient := &http.Client{Timeout: time.Second}

	WithSSEHTTPClient(httpClient)(client)

	if client.Client != httpClient {
		t.Fatal("expected HTTP client pointer to be assigned")
	}
}

func TestTransportClient_Connect_Good_SetsAcceptHeaderAndReturnsResponse(t *testing.T) {
	var sawAccept string
	var sawToken string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAccept = r.Header.Get("Accept")
		sawToken = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "event: ping\ndata: hello\n\n")
	}))
	defer srv.Close()

	client := publicSSEClient(
		t,
		srv,
		WithSSEHeaders(http.Header{"Authorization": {"Bearer secret"}}),
	)

	resp, err := client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer resp.Body.Close()

	if sawAccept != "text/event-stream" {
		t.Fatalf("expected Accept header to request SSE, got %q", sawAccept)
	}
	if sawToken != "Bearer secret" {
		t.Fatalf("expected Authorization header to reach server, got %q", sawToken)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 response, got %d", resp.StatusCode)
	}
}

func TestTransportClient_Connect_Bad_RejectsNonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, "temporarily unavailable")
	}))
	defer srv.Close()

	client := publicSSEClient(t, srv)
	if _, err := client.Connect(context.Background()); err == nil {
		t.Fatal("expected non-200 SSE response to fail")
	}
}

func TestTransportClient_Connect_Bad_ClosesResponseBodyOnRedirectError(t *testing.T) {
	var closed bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/redirect-target")
		w.WriteHeader(http.StatusFound)
		_, _ = io.WriteString(w, "redirecting")
	}))
	defer srv.Close()

	client := NewSSEClient(publicServerURL(t, srv), WithSSEHTTPClient(&http.Client{
		Transport: trackingRoundTripper{
			base:   localServerTransport(t, srv),
			closed: &closed,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("redirect blocked")
		},
	}))

	if _, err := client.Connect(context.Background()); err == nil {
		t.Fatal("expected redirect error")
	}
	if !closed {
		t.Fatal("expected redirect response body to be closed")
	}
}

func TestTransportClient_Events_Good_ParsesStream(t *testing.T) {
	payload := strings.Join([]string{
		": comment",
		"id: 7",
		"event: update",
		"data: line one",
		"data: line two",
		"retry: 1500",
		"",
		"data: final",
		"",
	}, "\n")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		_, _ = io.WriteString(w, payload)
		flusher.Flush()
	}))
	defer srv.Close()

	client := publicSSEClient(t, srv)
	events, err := client.Events(context.Background())
	if err != nil {
		t.Fatalf("Events failed: %v", err)
	}

	var got []SSEEvent
	for evt := range events {
		got = append(got, evt)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d: %#v", len(got), got)
	}
	if got[0].ID != "7" || got[0].Event != "update" || got[0].Data != "line one\nline two" || got[0].Retry != 1500*time.Millisecond {
		t.Fatalf("unexpected first event: %#v", got[0])
	}
	if got[1].Data != "final" {
		t.Fatalf("expected second event data %q, got %#v", "final", got[1])
	}
}

func TestTransportClient_Events_Bad_ContextCancelledClosesChannel(t *testing.T) {
	started := make(chan struct{}, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		_, _ = io.WriteString(w, "data: one\n\n")
		flusher.Flush()
		started <- struct{}{}
		<-r.Context().Done()
	}))
	defer srv.Close()

	client := publicSSEClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := client.Events(ctx)
	if err != nil {
		t.Fatalf("Events failed: %v", err)
	}

	first, ok := <-events
	if !ok {
		t.Fatal("expected one event before cancellation")
	}
	if first.Data != "one" {
		t.Fatalf("expected first event data %q, got %#v", "one", first)
	}

	<-started
	cancel()
	select {
	case _, ok := <-events:
		if ok {
			t.Fatal("expected channel to close after cancellation")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event channel to close")
	}
}

func TestTransportClient_Connect_Ugly_RejectsNilReceiver(t *testing.T) {
	var client *SSEClient
	if _, err := client.Connect(context.Background()); err == nil {
		t.Fatal("expected nil receiver to fail")
	}
}

func TestTransportClient_NewSSEClient_Bad_LeavesEmptyURLUsable(t *testing.T) {
	client := NewSSEClient("   ", nil)

	if client.URL != "" {
		t.Fatalf("expected empty URL after trimming whitespace, got %q", client.URL)
	}
	if client.Client == nil {
		t.Fatal("expected HTTP client to default even when URL is empty")
	}
}

func TestTransportClient_NewSSEClient_Ugly_ClonesSourceHeaders(t *testing.T) {
	source := http.Header{"X-Request-ID": {"abc"}}
	client := NewSSEClient("http://example.invalid/events", WithSSEHeaders(source))
	source["X-Request-ID"][0] = "mutated"

	if got := client.Header["X-Request-ID"]; len(got) != 1 || got[0] != "abc" {
		t.Fatalf("expected header copy to be isolated from source mutation, got %v", got)
	}
}

func TestTransportClient_NewWebSocketClient_Ugly_ClonesOptions(t *testing.T) {
	dialer := &websocket.Dialer{HandshakeTimeout: 2 * time.Second}
	client := NewWebSocketClient("ws://example.invalid/ws", WithWebSocketDialer(dialer))

	if client.Dialer != dialer {
		t.Fatal("expected dialer to be retained")
	}
}

func TestTransportClient_Events_Bad_PropagatesConnectError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "bad gateway")
	}))
	defer srv.Close()

	client := publicSSEClient(t, srv)
	if _, err := client.Events(context.Background()); err == nil {
		t.Fatal("expected Events to fail when Connect fails")
	}
}

func TestTransportClient_DialContext_Ugly_CleansBlankURL(t *testing.T) {
	client := NewWebSocketClient("   ")
	if _, _, err := client.DialContext(context.Background()); err == nil {
		t.Fatal("expected blank URL to fail")
	}
}

func TestTransportClient_Events_Good_ClosesReaderOnEOF(t *testing.T) {
	body := strings.NewReader("event: done\ndata: ok\n\n")
	events := make(chan SSEEvent, 1)
	parseSSEStream(context.Background(), body, events)

	select {
	case evt := <-events:
		if evt.Event != "done" || evt.Data != "ok" {
			t.Fatalf("unexpected parsed event: %#v", evt)
		}
	default:
		t.Fatal("expected one parsed event")
	}
}

func TestTransportClient_Connect_Good_EmptyHeadersDoNotPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := publicSSEClient(t, srv, WithSSEHeaders(http.Header{}))
	resp, err := client.Connect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
}

func TestTransportClient_NewSSEClient_Good_IgnoresNilOptions(t *testing.T) {
	client := NewSSEClient("http://example.invalid/events", nil)
	if client.Client == nil {
		t.Fatal("expected default HTTP client")
	}
}

func TestTransportClient_doHTTPClientRequest_Bad_BlocksRedirectToMetadata(t *testing.T) {
	var attempts int
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts > 1 {
				t.Fatalf("unexpected follow-up request to %s", req.URL.String())
			}

			return &http.Response{
				StatusCode: http.StatusFound,
				Status:     "302 Found",
				Header: http.Header{
					"Location": {"http://169.254.169.254/latest/meta-data/iam/security-credentials/"},
				},
				Body:    io.NopCloser(strings.NewReader("redirecting")),
				Request: req,
			}, nil
		}),
	}
	req, err := http.NewRequest(http.MethodGet, "http://93.184.216.34/start", nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	resp, err := doHTTPClientRequest(client, req)
	if err == nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		t.Fatal("expected metadata redirect to be blocked")
	}
	if !errors.Is(err, errOutboundURLBlocked) {
		t.Fatalf("expected errOutboundURLBlocked, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected only the initial request, got %d attempts", attempts)
	}
}
