// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio"
	core "dappco.re/go"
	"net"
	"net/http"
	"testing"
)

type responseMetaWriterStub struct {
	header            http.Header
	status            int
	body              responseMetaBodyBuffer
	wroteHeader       bool
	writeHeaderNowHit bool
	flushed           bool
	hijacked          bool
}

func newResponseMetaWriterStub() *responseMetaWriterStub {
	return &responseMetaWriterStub{
		header: make(http.Header),
		status: http.StatusOK,
		body:   core.NewBuffer(),
	}
}

func (w *responseMetaWriterStub) Header() http.Header {
	return w.header
}

func (w *responseMetaWriterStub) WriteHeader(code int) {
	w.status = code
	w.wroteHeader = true
}

func (w *responseMetaWriterStub) WriteHeaderNow() {
	w.writeHeaderNowHit = true
	w.wroteHeader = true
}

func (w *responseMetaWriterStub) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(data)
}

func (w *responseMetaWriterStub) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *responseMetaWriterStub) Status() int {
	if w.wroteHeader {
		return w.status
	}
	return http.StatusOK
}

func (w *responseMetaWriterStub) Size() int {
	return w.body.Len()
}

func (w *responseMetaWriterStub) Written() bool {
	return w.wroteHeader
}

func (w *responseMetaWriterStub) Flush() {
	w.flushed = true
}

func (w *responseMetaWriterStub) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return nil, nil, core.NewError("hijack not supported")
}

func (w *responseMetaWriterStub) CloseNotify() <-chan bool {
	return make(chan bool)
}

func (w *responseMetaWriterStub) Pusher() http.Pusher {
	return nil
}

func TestResponseMetaRecorder_Good_BuffersAndCommits(t *testing.T) {
	base := newResponseMetaWriterStub()
	base.Header().Set("X-Preexisting", "keep")

	rec := newResponseMetaRecorder(base)
	base.Header().Set("X-Preexisting", "mutated")

	if got := rec.Header().Get("X-Preexisting"); got != "keep" {
		t.Fatalf("expected header snapshot to be isolated, got %q", got)
	}

	rec.Header().Set("Content-Type", "application/json")
	rec.WriteHeader(http.StatusCreated)
	rec.WriteHeaderNow()
	if !rec.Written() {
		t.Fatal("expected recorder to report written after WriteHeaderNow")
	}
	if base.writeHeaderNowHit {
		t.Fatal("expected WriteHeaderNow to stay buffered before passthrough")
	}

	if _, err := rec.WriteString(`{"success":true,"data":"hello"}`); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	if rec.Size() != len(`{"success":true,"data":"hello"}`) {
		t.Fatalf("expected buffered size to match body length, got %d", rec.Size())
	}
	if rec.Status() != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Status())
	}

	rec.Flush()
	if !base.flushed {
		t.Fatal("expected underlying writer to be flushed")
	}
	if got := base.body.String(); got != `{"success":true,"data":"hello"}` {
		t.Fatalf("expected buffered body to be committed, got %q", got)
	}
	if got := base.Header().Get("X-Preexisting"); got != "keep" {
		t.Fatalf("expected original header value to be restored, got %q", got)
	}

	if _, err := rec.WriteString("!"); err != nil {
		t.Fatalf("WriteString after passthrough failed: %v", err)
	}
	if got := base.body.String(); got != `{"success":true,"data":"hello"}!` {
		t.Fatalf("expected passthrough writes to reach base writer, got %q", got)
	}
	if rec.Size() != base.Size() {
		t.Fatalf("expected Size to delegate after passthrough, got %d vs %d", rec.Size(), base.Size())
	}
}

func TestResponseMetaRecorder_Bad_RejectsNonJSONPayloads(t *testing.T) {
	meta := &Meta{RequestID: "req-123", Duration: "1ms"}

	if got := shouldAttachResponseMeta("text/plain", []byte(`{"success":true}`)); got {
		t.Fatal("expected text/plain to be rejected")
	}
	if got := shouldAttachResponseMeta("application/json", []byte(`[]`)); got {
		t.Fatal("expected array body to be rejected")
	}

	body := []byte(`{"success":true,"meta":{"page":2,"per_page":10,"total":100}}`)
	updated := refreshResponseMetaBody(body, meta)
	if coreBytesEqual(updated, body) {
		t.Fatal("expected metadata body to be updated")
	}

	var refreshed map[string]any
	if err := coreJSONUnmarshal(updated, &refreshed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	metaObj, ok := refreshed["meta"].(map[string]any)
	if !ok {
		t.Fatal("expected meta object in refreshed body")
	}
	if metaObj["request_id"] != "req-123" || metaObj["duration"] != "1ms" {
		t.Fatalf("expected request metadata to be injected, got %+v", metaObj)
	}
	if metaObj["page"] != float64(2) || metaObj["per_page"] != float64(10) || metaObj["total"] != float64(100) {
		t.Fatalf("expected pagination metadata to be preserved, got %+v", metaObj)
	}
}

func TestResponseMetaRecorder_Ugly_HandlesMalformedBodiesAndHijack(t *testing.T) {
	base := newResponseMetaWriterStub()
	rec := newResponseMetaRecorder(base)
	rec.Header().Set("Content-Type", "application/json")

	if got := refreshResponseMetaBody([]byte(`not-json`), &Meta{RequestID: "x"}); string(got) != "not-json" {
		t.Fatalf("expected malformed JSON to be returned unchanged, got %q", got)
	}
	if got := refreshResponseMetaBody([]byte(`{"success":false}`), nil); string(got) != `{"success":false}` {
		t.Fatalf("expected nil meta to leave body unchanged, got %q", got)
	}

	// Missing seam: gin.ResponseWriter already requires http.Hijacker, so the
	// recorder's non-Hijacker fallback branch is not reachable in a unit test.
	if _, _, err := rec.Hijack(); err == nil {
		t.Fatal("expected hijack to fail on the stub writer")
	}
	if !base.hijacked {
		t.Fatal("expected underlying Hijack method to be invoked")
	}
	if !rec.Written() {
		t.Fatal("expected recorder to report written after hijack")
	}
}
