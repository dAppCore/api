// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
)

// chatRemoteConfig is the remote backend attached to /v1/chat/completions via
// WithChatCompletionsRemote. It reuses the upstream router's balancer/transport.
type chatRemoteConfig struct {
	reg         *UpstreamRegistry
	adapters    map[string]ChatFormatAdapter
	maxAttempts int
	cooldown    time.Duration
	failover    map[int]bool
	transport   http.RoundTripper
	rt          *upstreamTransport // built in finalise
}

func (cfg *chatRemoteConfig) finalise() {
	if cfg.cooldown <= 0 {
		cfg.cooldown = defaultUpstreamCooldown
	}
	if cfg.failover == nil {
		cfg.failover = defaultFailoverStatuses()
	}
	if cfg.transport == nil {
		cfg.transport = http.DefaultTransport.(*http.Transport).Clone()
	}
	balancer := newUpstreamBalancer(cfg.cooldown, time.Now)
	cfg.rt = &upstreamTransport{
		base:        cfg.transport,
		balancer:    balancer,
		maxAttempts: cfg.maxAttempts,
		failover:    cfg.failover,
	}
}

// dispatchRemote proxies a chat request to the resolved remote pool, applying the
// per-model adapter (or verbatim passthrough when adapter == nil).
func (h *chatCompletionsHandler) dispatchRemote(c *gin.Context, req ChatCompletionRequest, raw []byte, pool []Upstream, adapter ChatFormatAdapter) {
	// Stream-capability check BEFORE dispatch (so we can still send an error body).
	if req.Stream && adapter != nil && adapter.Transcoder() == nil {
		writeChatCompletionError(c, http.StatusBadRequest, "invalid_request_error", "stream", "the adapter for this model does not support streaming", "")
		return
	}

	path := defaultChatCompletionsPath
	body := raw
	var hdrs map[string]string
	if adapter != nil {
		b, hh, err := adapter.BuildRequest(req)
		if err != nil {
			writeChatCompletionError(c, http.StatusInternalServerError, "inference_error", "model", err.Error(), "inference_error")
			return
		}
		path, body, hdrs = adapter.UpstreamPath(), b, hh
	}

	outReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		writeChatCompletionError(c, http.StatusInternalServerError, "inference_error", "model", err.Error(), "inference_error")
		return
	}
	bound := body
	outReq.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(bound)), nil }
	outReq.ContentLength = int64(len(bound))
	outReq.Header.Set("Content-Type", "application/json")
	for k, v := range hdrs {
		outReq.Header.Set(k, v)
	}
	ctx := context.WithValue(outReq.Context(), poolCtxKey, pool)
	ctx = context.WithValue(ctx, keyCtxKey, req.Model)
	outReq = outReq.WithContext(ctx)

	resp, err := h.remote.rt.RoundTrip(outReq)
	if err != nil {
		status, code := http.StatusServiceUnavailable, "upstream_unavailable"
		var re *routerError
		if core.As(err, &re) {
			status, code = re.status, re.code
		}
		if status == http.StatusServiceUnavailable {
			c.Header("Retry-After", strconv.Itoa(int(h.remote.cooldown.Seconds())))
		}
		writeChatCompletionError(c, status, "invalid_request_error", "model", "upstream request failed", code)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	h.deliverRemote(c, req, adapter, resp)
}

func (h *chatCompletionsHandler) deliverRemote(c *gin.Context, req ChatCompletionRequest, adapter ChatFormatAdapter, resp *http.Response) {
	// Non-2xx: passthrough copies verbatim; adapter wraps in the OpenAI error shape.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamResponseBytes))
		if adapter == nil {
			c.Header("Content-Type", "application/json")
			c.Status(resp.StatusCode)
			_, _ = c.Writer.Write(body)
			return
		}
		writeChatCompletionError(c, resp.StatusCode, "invalid_request_error", "model", "upstream error: "+string(body), "upstream_error")
		return
	}

	if req.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Status(http.StatusOK)
		flush := c.Writer.Flush
		if adapter == nil {
			copyFlushing(c.Writer, resp.Body, flush)
			return
		}
		meta := ChatStreamMeta{ID: newChatCompletionID(), Model: req.Model, Created: time.Now().Unix()}
		_ = adapter.Transcoder().Transcode(c.Writer, flush, resp.Body, meta)
		return
	}

	// Non-streaming.
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamResponseBytes))
	if adapter == nil {
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(body)
		return
	}
	out, err := adapter.DecodeResponse(req.Model, body)
	if err != nil {
		writeChatCompletionError(c, http.StatusBadGateway, "invalid_request_error", "model", "could not decode upstream response", "invalid_upstream_response")
		return
	}
	c.JSON(http.StatusOK, out)
}

// copyFlushing streams src to dst, flushing after each read so SSE chunks reach
// the client immediately.
func copyFlushing(dst io.Writer, src io.Reader, flush func()) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return
			}
			if flush != nil {
				flush()
			}
		}
		if err != nil {
			return
		}
	}
}
