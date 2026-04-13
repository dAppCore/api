// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

// defaultSSEPath is the URL path where the SSE endpoint is mounted.
const defaultSSEPath = "/events"

// SSEBroker manages Server-Sent Events connections and broadcasts events
// to subscribed clients. Clients connect via a GET endpoint and receive
// a streaming text/event-stream response. Each client may optionally
// subscribe to a specific channel via the ?channel= query parameter.
//
// Example:
//
//	broker := api.NewSSEBroker()
//	engine.GET("/events", broker.Handler())
type SSEBroker struct {
	mu      sync.RWMutex
	wg      sync.WaitGroup
	clients map[*sseClient]struct{}
}

// sseClient represents a single connected SSE consumer.
type sseClient struct {
	channel    string
	events     chan sseEvent
	done       chan struct{}
	doneOnce   sync.Once
	eventsOnce sync.Once
}

// sseEvent is an internal representation of a single SSE message.
type sseEvent struct {
	Event string
	Data  string
}

// NewSSEBroker creates a ready-to-use SSE broker.
//
// Example:
//
//	broker := api.NewSSEBroker()
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		clients: make(map[*sseClient]struct{}),
	}
}

// normaliseSSEPath coerces custom SSE paths into a stable form.
// The path always begins with a single slash and never ends with one.
func normaliseSSEPath(path string) string {
	path = core.Trim(path)
	if path == "" {
		return defaultSSEPath
	}

	path = "/" + strings.Trim(path, "/")
	if path == "/" {
		return defaultSSEPath
	}

	return path
}

// resolveSSEPath returns the configured SSE path or the default path when
// no override has been provided.
func resolveSSEPath(path string) string {
	if core.Trim(path) == "" {
		return defaultSSEPath
	}
	return normaliseSSEPath(path)
}

// Publish sends an event to all clients subscribed to the given channel.
// Clients subscribed to an empty channel (no ?channel= param) receive
// events on every channel. The data value is JSON-encoded before sending.
//
// Example:
//
//	broker.Publish("system", "ready", map[string]any{"status": "ok"})
func (b *SSEBroker) Publish(channel, event string, data any) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return
	}

	msg := sseEvent{
		Event: event,
		Data:  string(encoded),
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for client := range b.clients {
		// Send to clients on the matching channel, or clients with no channel filter.
		if client.channel == "" || client.channel == channel {
			select {
			case <-client.done:
				continue
			default:
			}
			select {
			case client.events <- msg:
			case <-client.done:
			default:
				// Drop event if client buffer is full.
			}
		}
	}
}

// Handler returns a Gin handler for the SSE endpoint. Clients connect with
// a GET request and receive events as text/event-stream. An optional
// ?channel=<name> query parameter subscribes the client to a specific channel.
//
// Example:
//
//	engine.GET("/events", broker.Handler())
func (b *SSEBroker) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Query("channel")

		client := &sseClient{
			channel: channel,
			events:  make(chan sseEvent, 64),
			done:    make(chan struct{}),
		}

		b.mu.Lock()
		b.clients[client] = struct{}{}
		b.wg.Add(1)
		b.mu.Unlock()

		defer func() {
			b.mu.Lock()
			client.signalDone()
			delete(b.clients, client)
			b.mu.Unlock()
			b.wg.Done()
		}()

		// Set SSE headers.
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("X-Accel-Buffering", "no")
		c.Status(http.StatusOK)
		c.Writer.Flush()

		// Stream events until client disconnects.
		ctx := c.Request.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case <-client.done:
				return
			default:
			}

			select {
			case <-ctx.Done():
				return
			case <-client.done:
				return
			case evt, ok := <-client.events:
				if !ok {
					return
				}
				_, err := c.Writer.Write([]byte(core.Sprintf("event: %s\ndata: %s\n\n", evt.Event, evt.Data)))
				if err != nil {
					return
				}
				// Flush to ensure the event is sent immediately.
				if f, ok := c.Writer.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
	}
}

// ClientCount returns the number of currently connected SSE clients.
//
// Example:
//
//	n := broker.ClientCount()
func (b *SSEBroker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// Drain signals all connected clients to disconnect and waits for their
// handler goroutines to exit. Useful for graceful shutdown.
//
// Example:
//
//	broker.Drain()
func (b *SSEBroker) Drain() {
	b.mu.Lock()
	for client := range b.clients {
		client.signalDone()
		client.closeEvents()
	}
	b.mu.Unlock()

	b.wg.Wait()
}

// signalDone closes the client done channel once.
func (c *sseClient) signalDone() {
	c.doneOnce.Do(func() {
		close(c.done)
	})
}

// closeEvents closes the client event channel once.
func (c *sseClient) closeEvents() {
	c.eventsOnce.Do(func() {
		close(c.events)
	})
}
