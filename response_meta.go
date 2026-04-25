// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio"
	// Note: AX-6 - streaming decoder UseNumber/extra-token checks have no core primitive.
	"encoding/json"
	// Note: AX-6 - Gin fallback and decoder EOF checks need stdlib sentinels.
	"io"
	"mime"
	"net"
	"net/http" // Note: AX-6 - structural HTTP boundary for Gin response writer contracts; no core primitive.
	"time"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

type responseMetaBodyBuffer interface {
	Write([]byte) (int, error)
	WriteString(string) (int, error)
	Bytes() []byte
	Reset()
}

// responseMetaRecorder buffers JSON responses so request metadata can be
// injected into the standard envelope before the body is written to the client.
type responseMetaRecorder struct {
	gin.ResponseWriter
	headers     http.Header
	body        responseMetaBodyBuffer
	size        int
	status      int
	wroteHeader bool
	committed   bool
	passthrough bool
}

func newResponseMetaRecorder(w gin.ResponseWriter) *responseMetaRecorder {
	headers := make(http.Header)
	for k, vals := range w.Header() {
		headers[k] = append([]string(nil), vals...)
	}

	return &responseMetaRecorder{
		ResponseWriter: w,
		headers:        headers,
		body:           core.NewBuffer(),
		status:         http.StatusOK,
	}
}

func (w *responseMetaRecorder) Header() http.Header {
	return w.headers
}

func (w *responseMetaRecorder) WriteHeader(code int) {
	if w.passthrough {
		w.status = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
		return
	}
	w.status = code
	w.wroteHeader = true
}

func (w *responseMetaRecorder) WriteHeaderNow() {
	if w.passthrough {
		w.wroteHeader = true
		w.ResponseWriter.WriteHeaderNow()
		return
	}
	w.wroteHeader = true
}

func (w *responseMetaRecorder) Write(data []byte) (int, error) {
	if w.passthrough {
		if !w.wroteHeader {
			w.WriteHeader(http.StatusOK)
		}
		return w.ResponseWriter.Write(data)
	}
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.body.Write(data)
	w.size += n
	return n, err
}

func (w *responseMetaRecorder) WriteString(s string) (int, error) {
	if w.passthrough {
		if !w.wroteHeader {
			w.WriteHeader(http.StatusOK)
		}
		return w.ResponseWriter.WriteString(s)
	}
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.body.WriteString(s)
	w.size += n
	return n, err
}

func (w *responseMetaRecorder) Flush() {
	if w.passthrough {
		if f, ok := w.ResponseWriter.(http.Flusher); ok {
			f.Flush()
		}
		return
	}

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	w.commit(true)
	w.passthrough = true

	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseMetaRecorder) Status() int {
	if w.wroteHeader {
		return w.status
	}

	return http.StatusOK
}

func (w *responseMetaRecorder) Size() int {
	if w.passthrough {
		return w.ResponseWriter.Size()
	}
	return w.size
}

func (w *responseMetaRecorder) Written() bool {
	return w.wroteHeader
}

func (w *responseMetaRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.passthrough {
		if h, ok := w.ResponseWriter.(http.Hijacker); ok {
			return h.Hijack()
		}
		return nil, nil, io.ErrClosedPipe
	}

	w.wroteHeader = true
	w.passthrough = true

	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, io.ErrClosedPipe
}

func (w *responseMetaRecorder) commit(writeBody bool) {
	if w.committed {
		return
	}

	for k := range w.ResponseWriter.Header() {
		w.ResponseWriter.Header().Del(k)
	}

	for k, vals := range w.headers {
		for _, v := range vals {
			w.ResponseWriter.Header().Add(k, v)
		}
	}

	w.ResponseWriter.WriteHeader(w.Status())
	if writeBody {
		_, _ = w.ResponseWriter.Write(w.body.Bytes())
		w.body.Reset()
	}
	w.committed = true
}

// responseMetaMiddleware injects request metadata into JSON envelope
// responses before they are written to the client.
func responseMetaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get(requestStartContextKey); !ok {
			c.Set(requestStartContextKey, time.Now())
		}

		recorder := newResponseMetaRecorder(c.Writer)
		c.Writer = recorder

		c.Next()

		if recorder.passthrough {
			return
		}

		body := recorder.body.Bytes()
		if meta := GetRequestMeta(c); meta != nil && shouldAttachResponseMeta(recorder.Header().Get("Content-Type"), body) {
			if refreshed := refreshResponseMetaBody(body, meta); refreshed != nil {
				body = refreshed
			}
		}

		recorder.body.Reset()
		_, _ = recorder.body.Write(body)
		recorder.size = len(body)
		recorder.Header().Set("Content-Length", core.Itoa(len(body)))
		recorder.commit(true)
	}
}

// refreshResponseMetaBody injects request metadata into a cached or buffered
// JSON envelope without disturbing existing pagination metadata.
func refreshResponseMetaBody(body []byte, meta *Meta) []byte {
	if meta == nil {
		return body
	}

	var payload any
	dec := json.NewDecoder(core.NewBuffer(body))
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return body
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return body
	}

	obj, ok := payload.(map[string]any)
	if !ok {
		return body
	}

	if _, ok := obj["success"]; !ok {
		if _, ok := obj["error"]; !ok {
			return body
		}
	}

	current := map[string]any{}
	if existing, ok := obj["meta"].(map[string]any); ok {
		current = existing
	}

	if meta.RequestID != "" {
		current["request_id"] = meta.RequestID
	}
	if meta.Duration != "" {
		current["duration"] = meta.Duration
	}

	obj["meta"] = current

	updated := core.JSONMarshal(obj)
	if !updated.OK {
		return body
	}

	if data, ok := updated.Value.([]byte); ok {
		return data
	}
	return body
}

func shouldAttachResponseMeta(contentType string, body []byte) bool {
	if !isJSONContentType(contentType) {
		return false
	}

	for _, b := range body {
		switch b {
		case ' ', '\t', '\r', '\n':
			continue
		default:
			return b == '{'
		}
	}
	return false
}

func isJSONContentType(contentType string) bool {
	if core.Trim(contentType) == "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = core.Trim(contentType)
	}
	mediaType = core.Lower(mediaType)

	return mediaType == "application/json" ||
		core.HasSuffix(mediaType, "+json") ||
		core.HasSuffix(mediaType, "/json")
}
