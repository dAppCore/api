// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"io"
	"net/http"
	"strconv"
	"sync"

	core "dappco.re/go/core"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
)

const (
	// BrotliBestSpeed is the lowest (fastest) Brotli compression level.
	BrotliBestSpeed = brotli.BestSpeed

	// BrotliBestCompression is the highest (smallest output) Brotli level.
	BrotliBestCompression = brotli.BestCompression

	// BrotliDefaultCompression is the default Brotli compression level.
	BrotliDefaultCompression = brotli.DefaultCompression
)

// brotliHandler manages a pool of brotli writers for reuse across requests.
type brotliHandler struct {
	pool  sync.Pool
	level int
}

// newBrotliHandler creates a handler that pools brotli writers at the given level.
func newBrotliHandler(level int) *brotliHandler {
	if level < BrotliBestSpeed || level > BrotliBestCompression {
		level = BrotliDefaultCompression
	}
	return &brotliHandler{
		level: level,
		pool: sync.Pool{
			New: func() any {
				return brotli.NewWriterLevel(io.Discard, level)
			},
		},
	}
}

// Handle is the Gin middleware function that compresses responses with Brotli.
func (h *brotliHandler) Handle(c *gin.Context) {
	if !core.Contains(c.Request.Header.Get("Accept-Encoding"), "br") {
		c.Next()
		return
	}

	w := h.pool.Get().(*brotli.Writer)
	w.Reset(c.Writer)

	c.Header("Content-Encoding", "br")
	c.Writer.Header().Add("Vary", "Accept-Encoding")

	bw := &brotliWriter{ResponseWriter: c.Writer, writer: w}
	c.Writer = bw

	defer func() {
		if bw.status >= http.StatusBadRequest {
			bw.Header().Del("Content-Encoding")
			bw.Header().Del("Vary")
			w.Reset(io.Discard)
		} else if c.Writer.Size() < 0 {
			w.Reset(io.Discard)
		}
		_ = w.Close()
		if c.Writer.Size() > -1 {
			c.Header("Content-Length", strconv.Itoa(c.Writer.Size()))
		}
		h.pool.Put(w)
	}()

	c.Next()
}

// brotliWriter wraps gin.ResponseWriter to intercept writes through brotli.
type brotliWriter struct {
	gin.ResponseWriter
	writer        *brotli.Writer
	statusWritten bool
	status        int
}

func (b *brotliWriter) Write(data []byte) (int, error) {
	b.Header().Del("Content-Length")

	if !b.statusWritten {
		b.status = b.ResponseWriter.Status()
	}

	if b.status >= http.StatusBadRequest {
		b.Header().Del("Content-Encoding")
		b.Header().Del("Vary")
		return b.ResponseWriter.Write(data)
	}

	return b.writer.Write(data)
}

func (b *brotliWriter) WriteString(s string) (int, error) {
	return b.Write([]byte(s))
}

func (b *brotliWriter) WriteHeader(code int) {
	b.status = code
	b.statusWritten = true
	b.Header().Del("Content-Length")
	b.ResponseWriter.WriteHeader(code)
}

func (b *brotliWriter) Flush() {
	_ = b.writer.Flush()
	b.ResponseWriter.Flush()
}
