// SPDX-License-Identifier: EUPL-1.2

package stream_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	api "dappco.re/go/api"
	"dappco.re/go/api/pkg/stream"

	"github.com/gin-gonic/gin"
)

func TestStreamGroup_Good_RoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	group := stream.NewGroup(
		"events",
		stream.SSE("/events", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/event-stream", []byte("data: ready\n\n"))
		}),
		stream.WebSocket("/ws", func(c *gin.Context) {
			c.Header("Upgrade", "websocket")
			c.Status(http.StatusSwitchingProtocols)
		}),
	)

	handlers := group.Handlers()
	if len(handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %d", len(handlers))
	}
	if handlers[0].Protocol != stream.ProtocolSSE {
		t.Fatalf("expected first protocol %q, got %q", stream.ProtocolSSE, handlers[0].Protocol)
	}
	if handlers[0].Method != http.MethodGet {
		t.Fatalf("expected first method %q, got %q", http.MethodGet, handlers[0].Method)
	}
	if handlers[0].Path != "/events" {
		t.Fatalf("expected first path %q, got %q", "/events", handlers[0].Path)
	}
	if handlers[1].Protocol != stream.ProtocolWebSocket {
		t.Fatalf("expected second protocol %q, got %q", stream.ProtocolWebSocket, handlers[1].Protocol)
	}
	if handlers[1].Path != "/ws" {
		t.Fatalf("expected second path %q, got %q", "/ws", handlers[1].Path)
	}

	router := gin.New()
	group.Register(router)

	sseRecorder := httptest.NewRecorder()
	sseReq, _ := http.NewRequest(http.MethodGet, "/events", nil)
	router.ServeHTTP(sseRecorder, sseReq)

	if sseRecorder.Code != http.StatusOK {
		t.Fatalf("expected SSE status 200, got %d", sseRecorder.Code)
	}
	if got := sseRecorder.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("expected SSE content type %q, got %q", "text/event-stream", got)
	}

	wsRecorder := httptest.NewRecorder()
	wsReq, _ := http.NewRequest(http.MethodGet, "/ws", nil)
	router.ServeHTTP(wsRecorder, wsReq)

	if wsRecorder.Code != http.StatusSwitchingProtocols {
		t.Fatalf("expected WebSocket status 101, got %d", wsRecorder.Code)
	}
}

func TestStreamGroup_Bad_DropsInvalidHandlersAndClonesMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	group := stream.NewGroup(
		"invalid",
		stream.Handler{
			Protocol: stream.ProtocolSSE,
			Method:   http.MethodGet,
			Path:     "",
			Handle:   func(*gin.Context) {},
		},
		stream.Handler{
			Protocol: stream.ProtocolWebSocket,
			Method:   http.MethodGet,
			Path:     "/ws",
			Handle:   nil,
		},
		stream.SSE("/events", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		}),
	)

	handlers := group.Handlers()
	if len(handlers) != 1 {
		t.Fatalf("expected 1 valid handler, got %d", len(handlers))
	}

	handlers[0].Path = "/mutated"

	fresh := group.Handlers()
	if len(fresh) != 1 {
		t.Fatalf("expected 1 fresh handler, got %d", len(fresh))
	}
	if fresh[0].Path != "/events" {
		t.Fatalf("expected cloned handler path %q, got %q", "/events", fresh[0].Path)
	}

	router := gin.New()
	group.Register(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/events", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected valid handler to remain registered, got %d", w.Code)
	}
}

func TestStreamGroup_Ugly_NormalisesWhitespaceWrappedMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	group := stream.NewGroup(
		"  ugly  ",
		stream.Handler{
			Protocol: " WS ",
			Method:   " get ",
			Path:     " /tenant/socket/ ",
			Handle: func(c *gin.Context) {
				c.String(http.StatusAccepted, "ok")
			},
		},
	)

	if group.Name() != "ugly" {
		t.Fatalf("expected trimmed name %q, got %q", "ugly", group.Name())
	}

	handlers := group.Handlers()
	if len(handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(handlers))
	}
	if handlers[0].Protocol != stream.ProtocolWebSocket {
		t.Fatalf("expected normalised protocol %q, got %q", stream.ProtocolWebSocket, handlers[0].Protocol)
	}
	if handlers[0].Method != http.MethodGet {
		t.Fatalf("expected normalised method %q, got %q", http.MethodGet, handlers[0].Method)
	}
	if handlers[0].Path != "/tenant/socket" {
		t.Fatalf("expected normalised path %q, got %q", "/tenant/socket", handlers[0].Path)
	}

	router := gin.New()
	group.Register(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/tenant/socket", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected normalised handler status 202, got %d", w.Code)
	}
}

func TestEngineRegisterStreamGroup_Good_MultiTenantRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, err := api.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	engine.RegisterStreamGroup(stream.NewGroup(
		"tenant-a",
		stream.SSE("/tenants/a/events", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/event-stream", []byte("data: tenant-a\n\n"))
		}),
	))
	engine.RegisterStreamGroup(stream.NewGroup(
		"tenant-b",
		stream.SSE("/tenants/b/events", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/event-stream", []byte("data: tenant-b\n\n"))
		}),
	))

	server := httptest.NewServer(engine.Handler())
	defer server.Close()

	for _, tc := range []struct {
		path string
		body string
	}{
		{path: "/tenants/a/events", body: "data: tenant-a\n\n"},
		{path: "/tenants/b/events", body: "data: tenant-b\n\n"},
	} {
		resp, reqErr := http.Get(server.URL + tc.path)
		if reqErr != nil {
			t.Fatalf("request %s failed: %v", tc.path, reqErr)
		}

		func() {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("%s: expected status 200, got %d", tc.path, resp.StatusCode)
			}
			if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
				t.Fatalf("%s: expected content type %q, got %q", tc.path, "text/event-stream", got)
			}

			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				t.Fatalf("%s: read body failed: %v", tc.path, readErr)
			}
			if string(body) != tc.body {
				t.Fatalf("%s: expected body %q, got %q", tc.path, tc.body, string(body))
			}
		}()
	}
}
