// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"io" // Note: AX-6 - brotli writer pooling needs io.Discard as the reset sink; no core primitive.
	"net/http"
	"strconv"
	"sync" // AX-6-exception: core has no Pool wrapper; brotli writers are pooled per compression level.

	core "dappco.re/go"

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
	if !acceptsBrotli(c.Request.Header.Get("Accept-Encoding")) {
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
		bw.release(&h.pool)
	}()

	c.Next()
}

func acceptsBrotli(acceptEncoding string) bool {
	found := false
	for _, part := range core.Split(acceptEncoding, ",") {
		token := core.Trim(part)
		params := ""
		if i := indexByte(token, ';'); i >= 0 {
			params = token[i+1:]
			token = core.Trim(token[:i])
		}
		if core.Lower(token) != "br" {
			continue
		}
		if hasZeroQValue(params) {
			return false
		}
		found = true
	}
	return found
}

func hasZeroQValue(params string) bool {
	for _, part := range core.Split(params, ";") {
		name, value, ok := cutString(core.Trim(part), "=")
		if !ok || core.Lower(core.Trim(name)) != "q" {
			continue
		}

		q, err := strconv.ParseFloat(core.Trim(value), 64)
		return err == nil && q == 0
	}
	return false
}

// brotliWriter wraps gin.ResponseWriter to intercept writes through brotli.
type brotliWriter struct {
	gin.ResponseWriter
	mu            core.Mutex
	writer        *brotli.Writer
	released      bool
	statusWritten bool
	status        int
}

func (b *brotliWriter) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.released {
		return len(data), nil
	}

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
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.released {
		return
	}

	b.status = code
	b.statusWritten = true
	b.Header().Del("Content-Length")
	if code >= http.StatusBadRequest {
		b.Header().Del("Content-Encoding")
		b.Header().Del("Vary")
	}
	b.ResponseWriter.WriteHeader(code)
}

func (b *brotliWriter) WriteHeaderNow() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.released {
		return
	}

	if !b.statusWritten {
		b.status = b.ResponseWriter.Status()
		b.statusWritten = true
	}
	b.Header().Del("Content-Length")
	if b.status >= http.StatusBadRequest {
		b.Header().Del("Content-Encoding")
		b.Header().Del("Vary")
	}
	b.ResponseWriter.WriteHeaderNow()
}

func (b *brotliWriter) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.released {
		return
	}

	if err := b.writer.Flush(); err != nil {
		return
	}
	b.ResponseWriter.Flush()
}

func (b *brotliWriter) release(pool *sync.Pool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.released {
		return
	}
	b.released = true

	if b.status >= http.StatusBadRequest {
		b.Header().Del("Content-Encoding")
		b.Header().Del("Vary")
		b.writer.Reset(io.Discard)
	} else if b.ResponseWriter.Size() < 0 {
		b.writer.Reset(io.Discard)
	}
	if err := b.writer.Close(); err != nil {
		b.Header().Del("Content-Length")
	}
	if b.ResponseWriter.Size() > -1 {
		b.Header().Set("Content-Length", core.Sprintf("%d", b.ResponseWriter.Size()))
	}
	b.writer.Reset(io.Discard)
	pool.Put(b.writer)
	b.writer = nil
}
