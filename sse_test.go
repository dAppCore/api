// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── SSE endpoint ────────────────────────────────────────────────────────

func TestWithSSE_Good_EndpointExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(api.WithSSE(broker))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/events")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("expected Content-Type starting with text/event-stream, got %q", ct)
	}
}

func TestWithSSE_Good_ReceivesPublishedEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(api.WithSSE(broker))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/events")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Wait for the client to register before publishing.
	waitForClients(t, broker, 1)

	// Publish an event on the default channel.
	broker.Publish("test", "greeting", map[string]string{"msg": "hello"})

	// Read SSE lines from the response body.
	scanner := bufio.NewScanner(resp.Body)
	var eventLine, dataLine string

	deadline := time.After(3 * time.Second)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for scanner.Scan() {
			line := scanner.Text()
			if after, ok := strings.CutPrefix(line, "event: "); ok {
				eventLine = after
			}
			if after, ok := strings.CutPrefix(line, "data: "); ok {
				dataLine = after
				return
			}
		}
	}()

	select {
	case <-done:
	case <-deadline:
		t.Fatal("timed out waiting for SSE event")
	}

	if eventLine != "greeting" {
		t.Fatalf("expected event=%q, got %q", "greeting", eventLine)
	}
	if !strings.Contains(dataLine, `"msg":"hello"`) {
		t.Fatalf("expected data containing msg:hello, got %q", dataLine)
	}
}

func TestWithSSE_Good_ChannelFiltering(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(api.WithSSE(broker))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	// Subscribe to channel "foo" only.
	resp, err := http.Get(srv.URL + "/events?channel=foo")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Wait for client to register.
	waitForClients(t, broker, 1)

	// Publish to "bar" (should not be received), then to "foo" (should be received).
	broker.Publish("bar", "ignore", "bar-data")
	// Small delay to ensure ordering.
	time.Sleep(50 * time.Millisecond)
	broker.Publish("foo", "match", "foo-data")

	// Read the first event from the stream.
	scanner := bufio.NewScanner(resp.Body)
	var eventLine string

	deadline := time.After(3 * time.Second)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for scanner.Scan() {
			line := scanner.Text()
			if after, ok := strings.CutPrefix(line, "event: "); ok {
				eventLine = after
				// Read past the data and blank line.
				scanner.Scan() // data line
				return
			}
		}
	}()

	select {
	case <-done:
	case <-deadline:
		t.Fatal("timed out waiting for SSE event")
	}

	// The first event received should be "match" (from channel foo), not "ignore" (from bar).
	if eventLine != "match" {
		t.Fatalf("expected event=%q, got %q (channel filtering failed)", "match", eventLine)
	}
}

func TestWithSSE_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(
		api.WithRequestID(),
		api.WithSSE(broker),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/events")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// RequestID middleware should have injected the header.
	reqID := resp.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Fatal("expected X-Request-ID header from RequestID middleware")
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("expected Content-Type starting with text/event-stream, got %q", ct)
	}
}

func TestWithSSE_Good_MultipleClients(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(api.WithSSE(broker))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	// Connect two clients.
	resp1, err := http.Get(srv.URL + "/events")
	if err != nil {
		t.Fatalf("client 1 request failed: %v", err)
	}
	defer resp1.Body.Close()

	resp2, err := http.Get(srv.URL + "/events")
	if err != nil {
		t.Fatalf("client 2 request failed: %v", err)
	}
	defer resp2.Body.Close()

	// Wait for both clients to register.
	waitForClients(t, broker, 2)

	// Publish a single event.
	broker.Publish("broadcast", "ping", "pong")

	// Both clients should receive it.
	var wg sync.WaitGroup
	wg.Add(2)

	readEvent := func(name string, resp *http.Response) {
		defer wg.Done()
		scanner := bufio.NewScanner(resp.Body)
		deadline := time.After(3 * time.Second)
		done := make(chan string, 1)

		go func() {
			for scanner.Scan() {
				line := scanner.Text()
				if after, ok := strings.CutPrefix(line, "event: "); ok {
					done <- after
					return
				}
			}
		}()

		select {
		case evt := <-done:
			if evt != "ping" {
				t.Errorf("%s: expected event=%q, got %q", name, "ping", evt)
			}
		case <-deadline:
			t.Errorf("%s: timed out waiting for SSE event", name)
		}
	}

	go readEvent("client1", resp1)
	go readEvent("client2", resp2)

	wg.Wait()
}

func TestWithSSE_Good_DrainDisconnectsClients(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broker := api.NewSSEBroker()
	e, err := api.New(api.WithSSE(broker))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/events")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	waitForClients(t, broker, 1)

	streamDone := make(chan struct{})
	go func() {
		defer close(streamDone)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
		}
	}()

	drainDone := make(chan struct{})
	go func() {
		broker.Drain()
		close(drainDone)
	}()

	select {
	case <-drainDone:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for SSE drain to complete")
	}

	select {
	case <-streamDone:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for SSE client to disconnect")
	}

	if got := broker.ClientCount(); got != 0 {
		t.Fatalf("expected 0 connected SSE clients after drain, got %d", got)
	}

	_ = resp.Body.Close()
}

// ── No SSE broker ────────────────────────────────────────────────────────

func TestNoSSEBroker_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Without WithSSE, GET /events should return 404.
	e, _ := api.New()

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/events", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /events without broker, got %d", w.Code)
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────

// waitForClients polls the broker until the expected number of clients
// are connected or the timeout expires.
func waitForClients(t *testing.T, broker *api.SSEBroker, want int) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		if broker.ClientCount() >= want {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %d SSE clients (have %d)", want, broker.ClientCount())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
