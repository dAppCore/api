// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// SSEBroker manages Server-Sent Events connections and broadcasts events
// to subscribed clients. Clients connect via a GET endpoint and receive
// a streaming text/event-stream response. Each client may optionally
// subscribe to a specific channel via the ?channel= query parameter.
type SSEBroker struct {
	mu      sync.RWMutex
	clients map[*sseClient]struct{}
}

// sseClient represents a single connected SSE consumer.
type sseClient struct {
	channel string
	events  chan sseEvent
	done    chan struct{}
}

// sseEvent is an internal representation of a single SSE message.
type sseEvent struct {
	Event string
	Data  string
}

// NewSSEBroker creates a ready-to-use SSE broker.
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		clients: make(map[*sseClient]struct{}),
	}
}

// Publish sends an event to all clients subscribed to the given channel.
// Clients subscribed to an empty channel (no ?channel= param) receive
// events on every channel. The data value is JSON-encoded before sending.
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
		b.mu.Unlock()

		defer func() {
			close(client.done)
			b.mu.Lock()
			delete(b.clients, client)
			b.mu.Unlock()
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
			case evt := <-client.events:
				_, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", evt.Event, evt.Data)
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
func (b *SSEBroker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// Drain closes all connected clients by writing an empty response.
// Useful for graceful shutdown.
func (b *SSEBroker) Drain() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for client := range b.clients {
		select {
		case <-client.done:
		default:
			// Write EOF to trigger client disconnect via their event loop.
			close(client.events)
		}
	}
}
