// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	api "dappco.re/go/core/api"
)

// ── Stub groups ─────────────────────────────────────────────────────────

// wsStubGroup is a basic RouteGroup for WebSocket tests.
type wsStubGroup struct{}

func (s *wsStubGroup) Name() string     { return "wsstub" }
func (s *wsStubGroup) BasePath() string { return "/v1/wsstub" }
func (s *wsStubGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/ping", func(c *gin.Context) {
		c.JSON(200, api.OK("pong"))
	})
}

// wsStubStreamGroup embeds wsStubGroup and implements StreamGroup.
type wsStubStreamGroup struct{ wsStubGroup }

func (s *wsStubStreamGroup) Channels() []string {
	return []string{"wsstub.events", "wsstub.updates"}
}

// ── WebSocket endpoint ──────────────────────────────────────────────────

func TestWSEndpoint_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a WebSocket upgrader that writes "hello" to every connection.
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	wsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}
		defer conn.Close()
		_ = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
	})

	e, err := api.New(api.WithWSHandler(wsHandler))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	// Dial the WebSocket endpoint.
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial WebSocket: %v", err)
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if string(msg) != "hello" {
		t.Fatalf("expected message=%q, got %q", "hello", string(msg))
	}
}

func TestWSEndpoint_Good_CustomPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	wsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}
		defer conn.Close()
		_ = conn.WriteMessage(websocket.TextMessage, []byte("custom"))
	})

	e, err := api.New(api.WithWSPath("/socket"), api.WithWSHandler(wsHandler))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial custom WebSocket: %v", err)
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read custom WebSocket message: %v", err)
	}
	if string(msg) != "custom" {
		t.Fatalf("expected message=%q, got %q", "custom", string(msg))
	}
}

func TestNoWSHandler_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Without WithWSHandler, GET /ws should return 404.
	e, _ := api.New()

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ws", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /ws without handler, got %d", w.Code)
	}
}

// ── Channel listing ─────────────────────────────────────────────────────

func TestChannelListing_Good(t *testing.T) {
	e, _ := api.New()

	// Register a plain RouteGroup (no channels) and a StreamGroup.
	e.Register(&wsStubGroup{})
	e.Register(&wsStubStreamGroup{})

	channels := e.Channels()
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
	if channels[0] != "wsstub.events" {
		t.Fatalf("expected channels[0]=%q, got %q", "wsstub.events", channels[0])
	}
	if channels[1] != "wsstub.updates" {
		t.Fatalf("expected channels[1]=%q, got %q", "wsstub.updates", channels[1])
	}
}
