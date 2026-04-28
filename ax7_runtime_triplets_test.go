// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	coretest "dappco.re/go"
	inference "dappco.re/go/inference"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	quichttp3 "github.com/quic-go/quic-go/http3"
	"go.opentelemetry.io/otel"
)

func TestAX7_WithWebSocketHeaders_Good(t *coretest.T) {
	client := &WebSocketClient{}
	source := http.Header{"Authorization": {"Bearer secret"}, "X-Trace": {"abc", "def"}}
	WithWebSocketHeaders(source)(client)
	coretest.AssertEqual(t, "Bearer secret", client.Header.Get("Authorization"))
	coretest.AssertEqual(t, []string{"abc", "def"}, client.Header.Values("X-Trace"))
}

func TestAX7_WithWebSocketHeaders_Bad(t *coretest.T) {
	client := &WebSocketClient{Header: http.Header{"X-Existing": {"keep"}}}
	WithWebSocketHeaders(http.Header{})(client)
	coretest.AssertEqual(t, "keep", client.Header.Get("X-Existing"))
	coretest.AssertLen(t, client.Header, 1)
}

func TestAX7_WithWebSocketHeaders_Ugly(t *coretest.T) {
	source := http.Header{"Authorization": {"Bearer secret"}}
	client := NewWebSocketClient("ws://example.invalid/ws", WithWebSocketHeaders(source))
	source["Authorization"][0] = "mutated"
	coretest.AssertEqual(t, "Bearer secret", client.Header.Get("Authorization"))
	coretest.AssertEqual(t, "ws://example.invalid/ws", client.URL)
}

func TestAX7_WithWebSocketDialer_Good(t *coretest.T) {
	client := &WebSocketClient{}
	dialer := &websocket.Dialer{HandshakeTimeout: time.Second}
	WithWebSocketDialer(dialer)(client)
	coretest.AssertEqual(t, dialer, client.Dialer)
	coretest.AssertEqual(t, time.Second, client.Dialer.HandshakeTimeout)
}

func TestAX7_WithWebSocketDialer_Bad(t *coretest.T) {
	client := &WebSocketClient{Dialer: &websocket.Dialer{HandshakeTimeout: time.Second}}
	WithWebSocketDialer(nil)(client)
	coretest.AssertNil(t, client.Dialer)
	coretest.AssertNotNil(t, client)
}

func TestAX7_WithWebSocketDialer_Ugly(t *coretest.T) {
	dialer := &websocket.Dialer{HandshakeTimeout: 2 * time.Second}
	client := NewWebSocketClient(" ws://example.invalid/ws ", WithWebSocketDialer(dialer), nil)
	coretest.AssertEqual(t, dialer, client.Dialer)
	coretest.AssertEqual(t, "ws://example.invalid/ws", client.URL)
}

func TestAX7_NewWebSocketClient_Good(t *coretest.T) {
	client := NewWebSocketClient("  ws://example.invalid/ws  ")
	coretest.AssertEqual(t, "ws://example.invalid/ws", client.URL)
	coretest.AssertNotNil(t, client.Header)
	coretest.AssertNil(t, client.Dialer)
}

func TestAX7_NewWebSocketClient_Bad(t *coretest.T) {
	client := NewWebSocketClient("   ", nil)
	coretest.AssertEqual(t, "", client.URL)
	coretest.AssertNotNil(t, client.Header)
	coretest.AssertLen(t, client.Header, 0)
}

func TestAX7_NewWebSocketClient_Ugly(t *coretest.T) {
	source := http.Header{"X-Trace": {"abc"}}
	client := NewWebSocketClient("\tws://example.invalid/ws\n", WithWebSocketHeaders(source))
	source["X-Trace"][0] = "mutated"
	coretest.AssertEqual(t, "abc", client.Header.Get("X-Trace"))
	coretest.AssertEqual(t, "ws://example.invalid/ws", client.URL)
}

func TestAX7_WebSocketClient_DialContext_Good(t *coretest.T) {
	ax7PublicDNS(t)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		coretest.AssertEqual(t, "Bearer secret", r.Header.Get("Authorization"))
		conn, err := upgrader.Upgrade(w, r, nil)
		coretest.RequireNoError(t, err)
		defer conn.Close()
		coretest.RequireNoError(t, conn.WriteMessage(websocket.TextMessage, []byte("hello")))
	}))
	defer srv.Close()

	targetAddr := srv.Listener.Addr().String()
	dialer := &websocket.Dialer{NetDialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, targetAddr)
	}}
	client := NewWebSocketClient("ws://public.example.com/ws", WithWebSocketDialer(dialer), WithWebSocketHeaders(http.Header{"Authorization": {"Bearer secret"}}))
	conn, resp, err := client.DialContext(context.Background())
	coretest.RequireNoError(t, err)
	defer conn.Close()
	coretest.AssertEqual(t, http.StatusSwitchingProtocols, resp.StatusCode)
	_, msg, err := conn.ReadMessage()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "hello", string(msg))
}

func TestAX7_WebSocketClient_DialContext_Bad(t *coretest.T) {
	client := NewWebSocketClient("ws://127.0.0.1/ws")
	conn, resp, err := client.DialContext(context.Background())
	coretest.AssertError(t, err)
	coretest.AssertNil(t, conn)
	coretest.AssertNil(t, resp)
	coretest.AssertTrue(t, errors.Is(err, errOutboundURLBlocked))
}

func TestAX7_WebSocketClient_DialContext_Ugly(t *coretest.T) {
	var client *WebSocketClient
	conn, resp, err := client.DialContext(context.Background())
	coretest.AssertError(t, err)
	coretest.AssertNil(t, conn)
	coretest.AssertNil(t, resp)
	coretest.AssertContains(t, err.Error(), "nil")
}

func TestAX7_WithSSEHeaders_Good(t *coretest.T) {
	client := &SSEClient{}
	WithSSEHeaders(http.Header{"X-Request-ID": {"abc"}})(client)
	coretest.AssertEqual(t, []string{"abc"}, client.Header["X-Request-ID"])
	coretest.AssertLen(t, client.Header["X-Request-ID"], 1)
}

func TestAX7_WithSSEHeaders_Bad(t *coretest.T) {
	client := &SSEClient{Header: http.Header{"X-Existing": {"keep"}}}
	WithSSEHeaders(nil)(client)
	coretest.AssertEqual(t, "keep", client.Header.Get("X-Existing"))
	coretest.AssertLen(t, client.Header, 1)
}

func TestAX7_WithSSEHeaders_Ugly(t *coretest.T) {
	source := http.Header{"X-Request-ID": {"abc"}}
	client := NewSSEClient("http://example.invalid/events", WithSSEHeaders(source))
	source["X-Request-ID"][0] = "mutated"
	coretest.AssertEqual(t, []string{"abc"}, client.Header["X-Request-ID"])
	coretest.AssertEqual(t, "http://example.invalid/events", client.URL)
}

func TestAX7_WithSSEHTTPClient_Good(t *coretest.T) {
	httpClient := &http.Client{Timeout: time.Second}
	client := &SSEClient{}
	WithSSEHTTPClient(httpClient)(client)
	coretest.AssertEqual(t, httpClient, client.Client)
	coretest.AssertEqual(t, time.Second, client.Client.Timeout)
}

func TestAX7_WithSSEHTTPClient_Bad(t *coretest.T) {
	client := &SSEClient{Client: http.DefaultClient}
	WithSSEHTTPClient(nil)(client)
	coretest.AssertNil(t, client.Client)
	coretest.AssertNotNil(t, client)
}

func TestAX7_WithSSEHTTPClient_Ugly(t *coretest.T) {
	client := NewSSEClient("http://example.invalid/events", WithSSEHTTPClient(nil))
	coretest.AssertEqual(t, http.DefaultClient, client.Client)
	coretest.AssertEqual(t, "http://example.invalid/events", client.URL)
}

func TestAX7_NewSSEClient_Good(t *coretest.T) {
	client := NewSSEClient("  http://example.invalid/events  ")
	coretest.AssertEqual(t, "http://example.invalid/events", client.URL)
	coretest.AssertNotNil(t, client.Header)
	coretest.AssertEqual(t, http.DefaultClient, client.Client)
}

func TestAX7_NewSSEClient_Bad(t *coretest.T) {
	client := NewSSEClient("   ", nil)
	coretest.AssertEqual(t, "", client.URL)
	coretest.AssertNotNil(t, client.Header)
	coretest.AssertEqual(t, http.DefaultClient, client.Client)
}

func TestAX7_NewSSEClient_Ugly(t *coretest.T) {
	source := http.Header{"X-Request-ID": {"abc"}}
	client := NewSSEClient("\thttp://example.invalid/events\n", WithSSEHeaders(source))
	source["X-Request-ID"][0] = "mutated"
	coretest.AssertEqual(t, []string{"abc"}, client.Header["X-Request-ID"])
	coretest.AssertEqual(t, "http://example.invalid/events", client.URL)
}

func TestAX7_SSEClient_Connect_Good(t *coretest.T) {
	var sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "data: ok\n\n")
	}))
	defer srv.Close()
	client := publicSSEClient(t, srv)
	resp, err := client.Connect(context.Background())
	coretest.RequireNoError(t, err)
	defer resp.Body.Close()
	coretest.AssertEqual(t, "text/event-stream", sawAccept)
	coretest.AssertEqual(t, http.StatusOK, resp.StatusCode)
}

func TestAX7_SSEClient_Connect_Bad(t *coretest.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, "unavailable")
	}))
	defer srv.Close()
	client := publicSSEClient(t, srv)
	resp, err := client.Connect(context.Background())
	coretest.AssertError(t, err)
	coretest.AssertNil(t, resp)
}

func TestAX7_SSEClient_Connect_Ugly(t *coretest.T) {
	var client *SSEClient
	resp, err := client.Connect(context.Background())
	coretest.AssertError(t, err)
	coretest.AssertNil(t, resp)
	coretest.AssertContains(t, err.Error(), "nil")
}

func TestAX7_SSEClient_Events_Good(t *coretest.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "event: update\ndata: one\ndata: two\n\n")
	}))
	defer srv.Close()
	client := publicSSEClient(t, srv)
	events, err := client.Events(context.Background())
	coretest.RequireNoError(t, err)
	evt := <-events
	coretest.AssertEqual(t, "update", evt.Event)
	coretest.AssertEqual(t, "one\ntwo", evt.Data)
}

func TestAX7_SSEClient_Events_Bad(t *coretest.T) {
	client := NewSSEClient("")
	events, err := client.Events(context.Background())
	coretest.AssertError(t, err)
	coretest.AssertNil(t, events)
	coretest.AssertContains(t, err.Error(), "URL")
}

func TestAX7_SSEClient_Events_Ugly(t *coretest.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := publicSSEClient(t, srv)
	events, err := client.Events(ctx)
	coretest.AssertError(t, err)
	coretest.AssertNil(t, events)
}

func TestAX7_NewSSEBroker_Good(t *coretest.T) {
	broker := NewSSEBroker()
	coretest.AssertNotNil(t, broker)
	coretest.AssertNotNil(t, broker.clients)
	coretest.AssertEqual(t, 0, broker.ClientCount())
}

func TestAX7_NewSSEBroker_Bad(t *coretest.T) {
	broker := NewSSEBroker()
	broker.clients = nil
	coretest.AssertEqual(t, 0, broker.ClientCount())
	coretest.AssertNil(t, broker.clients)
}

func TestAX7_NewSSEBroker_Ugly(t *coretest.T) {
	first := NewSSEBroker()
	second := NewSSEBroker()
	first.clients[&sseClient{events: make(chan sseEvent), done: make(chan struct{})}] = struct{}{}
	coretest.AssertFalse(t, first == second)
	coretest.AssertLen(t, first.clients, 1)
	coretest.AssertLen(t, second.clients, 0)
}

func TestAX7_SSEBroker_Publish_Good(t *coretest.T) {
	broker := NewSSEBroker()
	client := &sseClient{channel: "system", events: make(chan sseEvent, 1), done: make(chan struct{})}
	broker.clients[client] = struct{}{}
	broker.Publish("system", "ready", map[string]any{"ok": true})
	evt := <-client.events
	coretest.AssertEqual(t, "ready", evt.Event)
	coretest.AssertContains(t, evt.Data, `"ok":true`)
}

func TestAX7_SSEBroker_Publish_Bad(t *coretest.T) {
	broker := NewSSEBroker()
	client := &sseClient{events: make(chan sseEvent, 1), done: make(chan struct{})}
	broker.clients[client] = struct{}{}
	broker.Publish("", "bad", func() {})
	coretest.AssertLen(t, client.events, 0)
	coretest.AssertEqual(t, 1, broker.ClientCount())
}

func TestAX7_SSEBroker_Publish_Ugly(t *coretest.T) {
	broker := NewSSEBroker()
	client := &sseClient{channel: "other", events: make(chan sseEvent, 1), done: make(chan struct{})}
	broker.clients[client] = struct{}{}
	broker.Publish("system", "ready", "ok")
	coretest.AssertLen(t, client.events, 0)
	coretest.AssertEqual(t, 1, broker.ClientCount())
}

func TestAX7_SSEBroker_Handler_Good(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	broker := NewSSEBroker()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	reqCtx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx.Request = httptest.NewRequest(http.MethodGet, "/events?channel=system", nil).WithContext(reqCtx)
	broker.Handler()(ctx)
	coretest.AssertEqual(t, http.StatusOK, rec.Code)
	coretest.AssertEqual(t, "text/event-stream", rec.Header().Get("Content-Type"))
	coretest.AssertEqual(t, 0, broker.ClientCount())
}

func TestAX7_SSEBroker_Handler_Bad(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	var broker *SSEBroker
	handler := broker.Handler()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/events", nil)
	coretest.AssertPanics(t, func() { handler(ctx) })
}

func TestAX7_SSEBroker_Handler_Ugly(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	broker := NewSSEBroker()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	reqCtx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx.Request = httptest.NewRequest(http.MethodGet, "/events?channel=a%20b", nil).WithContext(reqCtx)
	broker.Handler()(ctx)
	coretest.AssertEqual(t, http.StatusOK, rec.Code)
	coretest.AssertEqual(t, "no-cache", rec.Header().Get("Cache-Control"))
}

func TestAX7_SSEBroker_Drain_Good(t *coretest.T) {
	broker := NewSSEBroker()
	client := &sseClient{events: make(chan sseEvent), done: make(chan struct{})}
	broker.clients[client] = struct{}{}
	broker.Drain()
	_, doneOpen := <-client.done
	_, eventsOpen := <-client.events
	coretest.AssertFalse(t, doneOpen)
	coretest.AssertFalse(t, eventsOpen)
}

func TestAX7_SSEBroker_Drain_Bad(t *coretest.T) {
	broker := NewSSEBroker()
	broker.Drain()
	coretest.AssertEqual(t, 0, broker.ClientCount())
	coretest.AssertNotNil(t, broker.clients)
}

func TestAX7_SSEBroker_Drain_Ugly(t *coretest.T) {
	broker := NewSSEBroker()
	client := &sseClient{events: make(chan sseEvent), done: make(chan struct{})}
	broker.clients[client] = struct{}{}
	broker.Drain()
	broker.Drain()
	coretest.AssertEqual(t, 1, broker.ClientCount())
}

func TestAX7_NewEntitlementBridge_Good(t *coretest.T) {
	bridge := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: " https://app.example.com/ ", Token: " secret "})
	coretest.AssertEqual(t, "https://app.example.com", bridge.baseURL)
	coretest.AssertEqual(t, "secret", bridge.token)
	coretest.AssertNotNil(t, bridge.client)
}

func TestAX7_NewEntitlementBridge_Bad(t *coretest.T) {
	bridge := NewEntitlementBridge(EntitlementBridgeConfig{})
	coretest.AssertEqual(t, "", bridge.baseURL)
	coretest.AssertEqual(t, "", bridge.token)
	coretest.AssertNotNil(t, bridge.client)
}

func TestAX7_NewEntitlementBridge_Ugly(t *coretest.T) {
	httpClient := &http.Client{Timeout: 17 * time.Millisecond}
	bridge := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: "http://example.com///", HTTPClient: httpClient})
	coretest.AssertEqual(t, "http://example.com", bridge.baseURL)
	coretest.AssertEqual(t, httpClient, bridge.client)
	coretest.AssertEqual(t, 17*time.Millisecond, bridge.client.Timeout)
}

func TestAX7_EntitlementBridge_Check_Good(t *coretest.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		coretest.AssertEqual(t, "/api/v1/workspaces/ws-1/entitlements/check/billing", r.URL.Path)
		_, _ = io.WriteString(w, `{"allowed":true}`)
	}))
	defer srv.Close()
	bridge := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: srv.URL})
	allowed, err := bridge.Check(context.Background(), "ws-1", "billing", http.Header{"Authorization": {"Bearer user"}})
	coretest.RequireNoError(t, err)
	coretest.AssertTrue(t, allowed)
	coretest.AssertEqual(t, "Bearer user", sawAuth)
}

func TestAX7_EntitlementBridge_Check_Bad(t *coretest.T) {
	bridge := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: "http://example.invalid"})
	allowed, err := bridge.Check(context.Background(), "", "  ", nil)
	coretest.AssertError(t, err)
	coretest.AssertFalse(t, allowed)
	coretest.AssertContains(t, err.Error(), "feature")
}

func TestAX7_EntitlementBridge_Check_Ugly(t *coretest.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		coretest.AssertContains(t, r.URL.Path, "space id")
		_, _ = io.WriteString(w, `{"entitlement":{"allowed":false}}`)
	}))
	defer srv.Close()
	bridge := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: srv.URL, Token: "service-token"})
	allowed, err := bridge.Check(context.Background(), "space id", "feature/name", nil)
	coretest.RequireNoError(t, err)
	coretest.AssertFalse(t, allowed)
}

func TestAX7_EntitlementBridge_Callback_Good(t *coretest.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"allowed":true}`)
	}))
	defer srv.Close()
	callback := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: srv.URL}).Callback(context.Background(), "", nil)
	coretest.AssertTrue(t, callback("feature"))
	coretest.AssertFalse(t, callback(""))
}

func TestAX7_EntitlementBridge_Callback_Bad(t *coretest.T) {
	bridge := NewEntitlementBridge(EntitlementBridgeConfig{})
	callback := bridge.Callback(context.Background(), "", nil)
	coretest.AssertFalse(t, callback("feature"))
	coretest.AssertFalse(t, callback(""))
}

func TestAX7_EntitlementBridge_Callback_Ugly(t *coretest.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"allowed":false}`)
	}))
	defer srv.Close()
	callback := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: srv.URL}).Callback(context.Background(), "ws", nil)
	coretest.AssertFalse(t, callback("feature"))
	coretest.AssertFalse(t, callback("other"))
}

func TestAX7_EntitlementBridge_CallbackForRequest_Good(t *coretest.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, `{"allowed":true}`)
	}))
	defer srv.Close()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer user")
	callback := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: srv.URL}).CallbackForRequest(req, "")
	coretest.AssertTrue(t, callback("feature"))
	coretest.AssertEqual(t, "Bearer user", sawAuth)
}

func TestAX7_EntitlementBridge_CallbackForRequest_Bad(t *coretest.T) {
	bridge := NewEntitlementBridge(EntitlementBridgeConfig{})
	callback := bridge.CallbackForRequest(nil, "")
	coretest.AssertFalse(t, callback("feature"))
	coretest.AssertFalse(t, callback(""))
}

func TestAX7_EntitlementBridge_CallbackForRequest_Ugly(t *coretest.T) {
	var sawWorkspace string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawWorkspace = r.Header.Get("X-Workspace-Id")
		_, _ = io.WriteString(w, `{"allowed":true}`)
	}))
	defer srv.Close()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	callback := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: srv.URL}).CallbackForRequest(req, "ws-1")
	coretest.AssertTrue(t, callback("feature"))
	coretest.AssertEqual(t, "ws-1", sawWorkspace)
}

func TestAX7_EntitlementBridge_CallbackForGin_Good(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"allowed":true}`)
	}))
	defer srv.Close()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	callback := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: srv.URL}).CallbackForGin(ctx, "ws-1")
	coretest.AssertTrue(t, callback("feature"))
	coretest.AssertEqual(t, http.StatusOK, rec.Code)
}

func TestAX7_EntitlementBridge_CallbackForGin_Bad(t *coretest.T) {
	bridge := NewEntitlementBridge(EntitlementBridgeConfig{})
	callback := bridge.CallbackForGin(nil, "")
	coretest.AssertFalse(t, callback("feature"))
	coretest.AssertFalse(t, callback(""))
}

func TestAX7_EntitlementBridge_CallbackForGin_Ugly(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = io.WriteString(w, `{"allowed":true}`)
	}))
	defer srv.Close()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	callback := NewEntitlementBridge(EntitlementBridgeConfig{BaseURL: srv.URL}).CallbackForGin(ctx, "")
	coretest.AssertTrue(t, callback("feature name"))
	coretest.AssertContains(t, sawPath, "/api/entitlements/check/")
}

func TestAX7_WithTracing_Good(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	engine, err := New(WithTracing("trace-service"))
	coretest.RequireNoError(t, err)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	engine.Handler().ServeHTTP(rec, req)
	coretest.AssertEqual(t, http.StatusOK, rec.Code)
	coretest.AssertContains(t, rec.Body.String(), "healthy")
}

func TestAX7_WithTracing_Bad(t *coretest.T) {
	engine, err := New(WithTracing(""))
	coretest.RequireNoError(t, err)
	coretest.AssertNotEmpty(t, engine.middlewares)
	coretest.AssertNotPanics(t, func() { engine.Handler() })
}

func TestAX7_WithTracing_Ugly(t *coretest.T) {
	called := false
	engine, err := New(WithTracing("trace-service"), func(e *Engine) {
		e.middlewares = append(e.middlewares, func(c *gin.Context) { called = true; c.Next() })
	})
	coretest.RequireNoError(t, err)
	engine.Handler().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
	coretest.AssertTrue(t, called)
}

func TestAX7_NewTracerProvider_Good(t *coretest.T) {
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	tp := NewTracerProvider(ax7SpanExporter{})
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})
	coretest.AssertNotNil(t, tp)
	err := tp.Shutdown(context.Background())
	coretest.AssertNoError(t, err)
}

func TestAX7_NewTracerProvider_Bad(t *coretest.T) {
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	tp := NewTracerProvider(nil)
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})
	coretest.AssertNotNil(t, tp)
	coretest.AssertEqual(t, tp, otel.GetTracerProvider())
}

func TestAX7_NewTracerProvider_Ugly(t *coretest.T) {
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	tp := NewTracerProvider(ax7SpanExporter{})
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})
	first := tp.Shutdown(context.Background())
	second := tp.Shutdown(context.Background())
	coretest.AssertNoError(t, first)
	coretest.AssertNoError(t, second)
}

func TestAX7_Engine_Handler_Good(t *coretest.T) {
	engine, err := New()
	coretest.RequireNoError(t, err)
	rec := httptest.NewRecorder()
	engine.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	coretest.AssertEqual(t, http.StatusOK, rec.Code)
	coretest.AssertContains(t, rec.Body.String(), "healthy")
}

func TestAX7_Engine_Handler_Bad(t *coretest.T) {
	engine := &Engine{}
	rec := httptest.NewRecorder()
	engine.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))
	coretest.AssertEqual(t, http.StatusNotFound, rec.Code)
	coretest.AssertNotNil(t, engine.Handler())
}

func TestAX7_Engine_Handler_Ugly(t *coretest.T) {
	engine, err := New(WithSSE(NewSSEBroker()))
	coretest.RequireNoError(t, err)
	first := engine.Handler()
	second := engine.Handler()
	coretest.AssertNotNil(t, first)
	coretest.AssertNotEqual(t, first, second)
}

func TestAX7_Engine_Serve_Good(t *coretest.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	engine, err := New(WithAddr("127.0.0.1:0"))
	coretest.RequireNoError(t, err)
	err = engine.Serve(ctx)
	coretest.AssertNoError(t, err)
}

func TestAX7_Engine_Serve_Bad(t *coretest.T) {
	engine, err := New(WithAddr("127.0.0.1:bad"))
	coretest.RequireNoError(t, err)
	err = engine.Serve(context.Background())
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "port")
}

func TestAX7_Engine_Serve_Ugly(t *coretest.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	engine, err := New(WithAddr("127.0.0.1:0"), WithSSE(NewSSEBroker()))
	coretest.RequireNoError(t, err)
	err = engine.Serve(ctx)
	coretest.AssertNoError(t, err)
}

func TestAX7_Engine_ServeH3_Good(t *coretest.T) {
	addr := reserveHTTP3UDPAddr(t)
	serverTLS, clientTLS := testHTTP3TLSConfigs(t)
	engine, err := New(WithHTTP3(addr))
	coretest.RequireNoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- engine.ServeH3(ctx, serverTLS) }()
	transport := &quichttp3.Transport{TLSClientConfig: clientTLS}
	defer transport.Close()
	client := &http.Client{Transport: transport, Timeout: time.Second}
	var resp *http.Response
	var lastErr error
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err = client.Get("https://" + addr + "/health")
		if err == nil {
			break
		}
		lastErr = err
		select {
		case serveErr := <-errCh:
			t.Fatalf("ServeH3 exited before health check: %v", serveErr)
		default:
		}
		time.Sleep(50 * time.Millisecond)
	}
	if resp == nil {
		t.Fatalf("HTTP/3 health request failed before deadline: %v", lastErr)
	}
	resp.Body.Close()
	cancel()
	select {
	case serveErr := <-errCh:
		coretest.AssertNoError(t, serveErr)
	case <-time.After(5 * time.Second):
		t.Fatal("ServeH3 did not return after context cancellation")
	}
	coretest.AssertEqual(t, http.StatusOK, resp.StatusCode)
}

func TestAX7_Engine_ServeH3_Bad(t *coretest.T) {
	engine, err := New(WithHTTP3("127.0.0.1:9443"))
	coretest.RequireNoError(t, err)
	err = engine.ServeH3(context.Background(), nil)
	coretest.AssertTrue(t, errors.Is(err, ErrHTTP3TLSRequired))
	coretest.AssertError(t, err)
}

func TestAX7_Engine_ServeH3_Ugly(t *coretest.T) {
	engine, err := New()
	coretest.RequireNoError(t, err)
	err = engine.ServeH3(context.Background(), nil)
	coretest.AssertTrue(t, errors.Is(err, ErrHTTP3NotConfigured))
	coretest.AssertError(t, err)
}

func TestAX7_StopList_UnmarshalJSON_Good(t *coretest.T) {
	var stops chatStopList
	err := stops.UnmarshalJSON([]byte(`"END"`))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, chatStopList{"END"}, stops)
}

func TestAX7_StopList_UnmarshalJSON_Bad(t *coretest.T) {
	var stops chatStopList
	err := stops.UnmarshalJSON([]byte(`{"bad":true}`))
	coretest.AssertError(t, err)
	coretest.AssertNil(t, stops)
}

func TestAX7_StopList_UnmarshalJSON_Ugly(t *coretest.T) {
	stops := chatStopList{"keep"}
	err := stops.UnmarshalJSON([]byte(`null`))
	coretest.RequireNoError(t, err)
	coretest.AssertNil(t, stops)
}

func TestAX7_ChatMessageDelta_MarshalJSON_Good(t *coretest.T) {
	data, err := json.Marshal(ChatMessageDelta{Role: "assistant"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, `{"role":"assistant","content":""}`, string(data))
	coretest.AssertContains(t, string(data), "assistant")
}

func TestAX7_ChatMessageDelta_MarshalJSON_Bad(t *coretest.T) {
	data, err := json.Marshal(ChatMessageDelta{})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, `{}`, string(data))
	coretest.AssertNotContains(t, string(data), "content")
}

func TestAX7_ChatMessageDelta_MarshalJSON_Ugly(t *coretest.T) {
	data, err := json.Marshal(ChatMessageDelta{Content: "token"})
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, `{"content":"token"}`, string(data))
	coretest.AssertNotContains(t, string(data), "role")
}

func TestAX7_ResolutionError_Error_Good(t *coretest.T) {
	err := &modelResolutionError{msg: "missing model"}
	coretest.AssertEqual(t, "missing model", err.Error())
	coretest.AssertNotEmpty(t, err.Error())
}

func TestAX7_ResolutionError_Error_Bad(t *coretest.T) {
	var err *modelResolutionError
	got := err.Error()
	coretest.AssertEqual(t, "", got)
	coretest.AssertEmpty(t, got)
}

func TestAX7_ResolutionError_Error_Ugly(t *coretest.T) {
	err := &modelResolutionError{code: "model_loading", param: "model", msg: "loading"}
	coretest.AssertEqual(t, "loading", err.Error())
	coretest.AssertEqual(t, "model_loading", err.code)
}

func TestAX7_NewModelResolver_Good(t *coretest.T) {
	resolver := NewModelResolver()
	coretest.AssertNotNil(t, resolver.loadedByName)
	coretest.AssertNotNil(t, resolver.loadedByPath)
	coretest.AssertNotNil(t, resolver.inFlight)
}

func TestAX7_NewModelResolver_Bad(t *coretest.T) {
	resolver := NewModelResolver()
	model, err := resolver.ResolveModel("")
	coretest.AssertError(t, err)
	coretest.AssertNil(t, model)
	coretest.AssertLen(t, resolver.loadedByName, 0)
}

func TestAX7_NewModelResolver_Ugly(t *coretest.T) {
	first := NewModelResolver()
	second := NewModelResolver()
	first.loadedByName["x"] = &chatModelStub{}
	coretest.AssertLen(t, second.loadedByName, 0)
	coretest.AssertNotEqual(t, first, second)
}

func TestAX7_ModelResolver_ResolveModel_Good(t *coretest.T) {
	model := &chatModelStub{}
	resolver := NewModelResolver()
	resolver.loadedByName["lemer"] = model
	got, err := resolver.ResolveModel("  LEMER  ")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, model, got)
}

func TestAX7_ModelResolver_ResolveModel_Bad(t *coretest.T) {
	resolver := NewModelResolver()
	model, err := resolver.ResolveModel("missing-model")
	coretest.AssertError(t, err)
	coretest.AssertNil(t, model)
	coretest.AssertContains(t, err.Error(), "not found")
}

func TestAX7_ModelResolver_ResolveModel_Ugly(t *coretest.T) {
	var resolver *ModelResolver
	model, err := resolver.ResolveModel("lemer")
	coretest.AssertError(t, err)
	coretest.AssertNil(t, model)
	coretest.AssertContains(t, err.Error(), "not configured")
}

func TestAX7_NewThinkingExtractor_Good(t *coretest.T) {
	extractor := NewThinkingExtractor()
	coretest.AssertNotNil(t, extractor)
	coretest.AssertEqual(t, "assistant", extractor.currentChannel)
	coretest.AssertEqual(t, "", extractor.Content())
}

func TestAX7_NewThinkingExtractor_Bad(t *coretest.T) {
	var extractor *ThinkingExtractor
	coretest.AssertEqual(t, "", extractor.Content())
	coretest.AssertNil(t, extractor.Thinking())
	coretest.AssertNotEqual(t, NewThinkingExtractor(), extractor)
}

func TestAX7_NewThinkingExtractor_Ugly(t *coretest.T) {
	first := NewThinkingExtractor()
	second := NewThinkingExtractor()
	first.Process(inference.Token{Text: "hello"})
	coretest.AssertEqual(t, "", second.Content())
	coretest.AssertEqual(t, "hello", first.Content())
}

func TestAX7_ThinkingExtractor_Process_Good(t *coretest.T) {
	extractor := NewThinkingExtractor()
	extractor.Process(inference.Token{Text: "Hello"})
	coretest.AssertEqual(t, "Hello", extractor.Content())
	coretest.AssertNil(t, extractor.Thinking())
}

func TestAX7_ThinkingExtractor_Process_Bad(t *coretest.T) {
	var extractor *ThinkingExtractor
	coretest.AssertNotPanics(t, func() { extractor.Process(inference.Token{Text: "ignored"}) })
	coretest.AssertEqual(t, "", extractor.Content())
	coretest.AssertNil(t, extractor.Thinking())
}

func TestAX7_ThinkingExtractor_Process_Ugly(t *coretest.T) {
	extractor := NewThinkingExtractor()
	extractor.Process(inference.Token{Text: "Hello <|channel>thought plan <|channel>assistant world"})
	coretest.AssertEqual(t, "Hello  world", extractor.Content())
	coretest.AssertEqual(t, " plan ", *extractor.Thinking())
}

func TestAX7_ThinkingExtractor_Content_Good(t *coretest.T) {
	extractor := NewThinkingExtractor()
	extractor.Process(inference.Token{Text: "visible"})
	content := extractor.Content()
	coretest.AssertEqual(t, "visible", content)
	coretest.AssertNotEmpty(t, content)
}

func TestAX7_ThinkingExtractor_Content_Bad(t *coretest.T) {
	var extractor *ThinkingExtractor
	content := extractor.Content()
	coretest.AssertEqual(t, "", content)
	coretest.AssertEmpty(t, content)
}

func TestAX7_ThinkingExtractor_Content_Ugly(t *coretest.T) {
	extractor := NewThinkingExtractor()
	extractor.Process(inference.Token{Text: "<|channel>thought hidden"})
	content := extractor.Content()
	coretest.AssertEqual(t, "", content)
	coretest.AssertNotNil(t, extractor.Thinking())
}

func TestAX7_ThinkingExtractor_Thinking_Good(t *coretest.T) {
	extractor := NewThinkingExtractor()
	extractor.Process(inference.Token{Text: "<|channel>thought hidden"})
	thinking := extractor.Thinking()
	coretest.AssertNotNil(t, thinking)
	coretest.AssertEqual(t, " hidden", *thinking)
}

func TestAX7_ThinkingExtractor_Thinking_Bad(t *coretest.T) {
	extractor := NewThinkingExtractor()
	thinking := extractor.Thinking()
	coretest.AssertNil(t, thinking)
	coretest.AssertEqual(t, "", extractor.Content())
}

func TestAX7_ThinkingExtractor_Thinking_Ugly(t *coretest.T) {
	var extractor *ThinkingExtractor
	thinking := extractor.Thinking()
	coretest.AssertNil(t, thinking)
	coretest.AssertEqual(t, "", extractor.Content())
}

func TestAX7_CompletionsHandler_ServeHTTP_Good(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	handler := newChatHandlerWithModel(&chatModelStub{tokens: []inference.Token{{Text: "Hello"}}})
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = newChatLoopbackRequest(t, `{"model":"lemer","messages":[{"role":"user","content":"hi"}]}`)
	handler.ServeHTTP(ctx)
	coretest.AssertEqual(t, http.StatusOK, rec.Code)
	coretest.AssertContains(t, rec.Body.String(), "Hello")
}

func TestAX7_CompletionsHandler_ServeHTTP_Bad(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	handler := newChatCompletionsHandler(nil)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = newChatLoopbackRequest(t, `{"model":"lemer","messages":[{"role":"user","content":"hi"}]}`)
	handler.ServeHTTP(ctx)
	coretest.AssertEqual(t, http.StatusServiceUnavailable, rec.Code)
	coretest.AssertContains(t, rec.Body.String(), "not configured")
}

func TestAX7_CompletionsHandler_ServeHTTP_Ugly(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	handler := newChatHandlerWithModel(&chatModelStub{})
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"lemer","messages":[{"role":"user","content":"hi"}]}`))
	ctx.Request.RemoteAddr = "203.0.113.1:1234"
	handler.ServeHTTP(ctx)
	coretest.AssertEqual(t, http.StatusForbidden, rec.Code)
	coretest.AssertContains(t, rec.Body.String(), "loopback")
}

func TestAX7_CompletionRequestError_Error_Good(t *coretest.T) {
	err := &chatCompletionRequestError{Message: "bad request"}
	coretest.AssertEqual(t, "bad request", err.Error())
	coretest.AssertNotEmpty(t, err.Error())
}

func TestAX7_CompletionRequestError_Error_Bad(t *coretest.T) {
	var err *chatCompletionRequestError
	got := err.Error()
	coretest.AssertEqual(t, "", got)
	coretest.AssertEmpty(t, got)
}

func TestAX7_CompletionRequestError_Error_Ugly(t *coretest.T) {
	err := &chatCompletionRequestError{Status: http.StatusBadRequest, Param: "model", Message: "model is required"}
	coretest.AssertEqual(t, "model is required", err.Error())
	coretest.AssertEqual(t, "model", err.Param)
}

func TestAX7_URLError_Error_Good(t *coretest.T) {
	err := blockedURLError{reason: "metadata host"}
	coretest.AssertContains(t, err.Error(), "metadata host")
	coretest.AssertContains(t, err.Error(), errOutboundURLBlocked.Error())
}

func TestAX7_URLError_Error_Bad(t *coretest.T) {
	err := blockedURLError{}
	coretest.AssertContains(t, err.Error(), errOutboundURLBlocked.Error())
	coretest.AssertContains(t, err.Error(), ":")
}

func TestAX7_URLError_Error_Ugly(t *coretest.T) {
	err := wrapBlocked("private IP")
	coretest.AssertContains(t, err.Error(), "private IP")
	coretest.AssertTrue(t, errors.Is(err, errOutboundURLBlocked))
}

func TestAX7_URLError_Unwrap_Good(t *coretest.T) {
	err := blockedURLError{reason: "metadata host"}
	coretest.AssertEqual(t, errOutboundURLBlocked, err.Unwrap())
	coretest.AssertTrue(t, errors.Is(err, errOutboundURLBlocked))
}

func TestAX7_URLError_Unwrap_Bad(t *coretest.T) {
	err := blockedURLError{}
	coretest.AssertEqual(t, errOutboundURLBlocked, err.Unwrap())
	coretest.AssertTrue(t, errors.Is(err, errOutboundURLBlocked))
}

func TestAX7_URLError_Unwrap_Ugly(t *coretest.T) {
	err := wrapBlocked("metadata host")
	coretest.AssertTrue(t, errors.Is(err, errOutboundURLBlocked))
	coretest.AssertNotNil(t, errors.Unwrap(err))
}
