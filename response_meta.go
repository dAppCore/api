// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// responseMetaRecorder buffers JSON responses so request metadata can be
// injected into the standard envelope before the body is written to the client.
type responseMetaRecorder struct {
	gin.ResponseWriter
	headers     http.Header
	body        bytes.Buffer
	status      int
	wroteHeader bool
}

func newResponseMetaRecorder(w gin.ResponseWriter) *responseMetaRecorder {
	headers := make(http.Header)
	for k, vals := range w.Header() {
		headers[k] = append([]string(nil), vals...)
	}

	return &responseMetaRecorder{
		ResponseWriter: w,
		headers:        headers,
		status:         http.StatusOK,
	}
}

func (w *responseMetaRecorder) Header() http.Header {
	return w.headers
}

func (w *responseMetaRecorder) WriteHeader(code int) {
	w.status = code
	w.wroteHeader = true
}

func (w *responseMetaRecorder) WriteHeaderNow() {
	w.wroteHeader = true
}

func (w *responseMetaRecorder) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(data)
}

func (w *responseMetaRecorder) WriteString(s string) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.WriteString(s)
}

func (w *responseMetaRecorder) Flush() {
}

func (w *responseMetaRecorder) Status() int {
	if w.wroteHeader {
		return w.status
	}

	return http.StatusOK
}

func (w *responseMetaRecorder) Size() int {
	return w.body.Len()
}

func (w *responseMetaRecorder) Written() bool {
	return w.wroteHeader
}

func (w *responseMetaRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, io.ErrClosedPipe
}

func (w *responseMetaRecorder) commit() {
	for k := range w.ResponseWriter.Header() {
		w.ResponseWriter.Header().Del(k)
	}

	for k, vals := range w.headers {
		for _, v := range vals {
			w.ResponseWriter.Header().Add(k, v)
		}
	}

	w.ResponseWriter.WriteHeader(w.Status())
	_, _ = w.ResponseWriter.Write(w.body.Bytes())
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

		body := recorder.body.Bytes()
		if meta := GetRequestMeta(c); meta != nil && shouldAttachResponseMeta(recorder.Header().Get("Content-Type"), body) {
			if refreshed := refreshResponseMetaBody(body, meta); refreshed != nil {
				body = refreshed
			}
		}

		recorder.body.Reset()
		_, _ = recorder.body.Write(body)
		recorder.Header().Set("Content-Length", strconv.Itoa(len(body)))
		recorder.commit()
	}
}

// refreshResponseMetaBody injects request metadata into a cached or buffered
// JSON envelope without disturbing existing pagination metadata.
func refreshResponseMetaBody(body []byte, meta *Meta) []byte {
	if meta == nil {
		return body
	}

	var payload any
	dec := json.NewDecoder(bytes.NewReader(body))
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

	updated, err := json.Marshal(obj)
	if err != nil {
		return body
	}

	return updated
}

func shouldAttachResponseMeta(contentType string, body []byte) bool {
	if !isJSONContentType(contentType) {
		return false
	}

	trimmed := bytes.TrimSpace(body)
	return len(trimmed) > 0 && trimmed[0] == '{'
}

func isJSONContentType(contentType string) bool {
	if strings.TrimSpace(contentType) == "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.TrimSpace(contentType)
	}
	mediaType = strings.ToLower(mediaType)

	return mediaType == "application/json" ||
		strings.HasSuffix(mediaType, "+json") ||
		strings.HasSuffix(mediaType, "/json")
}
