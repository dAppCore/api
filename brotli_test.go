// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// ── WithBrotli ────────────────────────────────────────────────────────

func TestWithBrotli_Good_CompressesResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBrotli())
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "br" {
		t.Fatalf("expected Content-Encoding=%q, got %q", "br", ce)
	}
}

func TestWithBrotli_Good_NoCompressionWithoutAcceptHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBrotli())
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	// Deliberately not setting Accept-Encoding header.
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce == "br" {
		t.Fatal("expected no br Content-Encoding when client does not request it")
	}
}

func TestWithBrotli_Good_DefaultLevel(t *testing.T) {
	// Calling WithBrotli() with no arguments should use default compression
	// and not panic.
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBrotli())
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "br" {
		t.Fatalf("expected Content-Encoding=%q with default level, got %q", "br", ce)
	}
}

func TestWithBrotli_Good_CustomLevel(t *testing.T) {
	// WithBrotli(BrotliBestSpeed) should work without panicking and still compress.
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithBrotli(api.BrotliBestSpeed))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ce := w.Header().Get("Content-Encoding")
	if ce != "br" {
		t.Fatalf("expected Content-Encoding=%q with BestSpeed, got %q", "br", ce)
	}
}

func TestWithBrotli_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithBrotli(),
		api.WithRequestID(),
	)
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	req.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Both brotli compression and request ID should be present.
	ce := w.Header().Get("Content-Encoding")
	if ce != "br" {
		t.Fatalf("expected Content-Encoding=%q, got %q", "br", ce)
	}

	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}

func TestWithBrotli_Good_DropsLateWritesAfterHandlerReturn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	oldProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldProcs)

	group := newBrotliLateWriteGroup()
	lateWriterLaunched := false
	t.Cleanup(func() {
		if !lateWriterLaunched {
			return
		}
		group.startLateWrites()
		group.stopLateWrites()
		select {
		case <-group.done:
		case <-time.After(time.Second):
			t.Error("late writer goroutine did not stop")
		}
	})

	e, _ := api.New(api.WithBrotli(api.BrotliBestSpeed))
	e.Register(group)
	h := e.Handler()

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/brotli-late/leaky", nil)
	req1.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w1, req1)

	select {
	case <-group.ready:
		lateWriterLaunched = true
	case <-time.After(time.Second):
		t.Fatal("late writer goroutine did not start")
	}
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first request status 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/brotli-late/target", nil)
	req2.Header.Set("Accept-Encoding", "br")
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected second request status 200, got %d", w2.Code)
	}
	if ce := w2.Header().Get("Content-Encoding"); ce != "br" {
		t.Fatalf("expected second response Content-Encoding=%q, got %q", "br", ce)
	}

	decoded := decodeBrotliResponse(t, w2)
	if string(decoded) != brotliLateTargetBody {
		t.Fatalf("expected second response body %q, got %q", brotliLateTargetBody, string(decoded))
	}
}

const brotliLateTargetBody = "second request body"

type brotliLateWriteGroup struct {
	ready     chan struct{}
	start     chan struct{}
	attempted chan struct{}
	stop      chan struct{}
	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
}

func newBrotliLateWriteGroup() *brotliLateWriteGroup {
	return &brotliLateWriteGroup{
		ready:     make(chan struct{}),
		start:     make(chan struct{}),
		attempted: make(chan struct{}),
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
	}
}

func (g *brotliLateWriteGroup) Name() string     { return "brotli-late" }
func (g *brotliLateWriteGroup) BasePath() string { return "/brotli-late" }

func (g *brotliLateWriteGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/leaky", func(c *gin.Context) {
		writer := c.Writer
		go func() {
			defer close(g.done)
			close(g.ready)
			<-g.start

			payload := bytes.Repeat([]byte("late write from first request;"), 64)
			attempted := false
			for {
				select {
				case <-g.stop:
					return
				default:
				}

				_, _ = writer.Write(payload)
				if !attempted {
					close(g.attempted)
					attempted = true
				}
				runtime.Gosched()
			}
		}()

		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("first request body"))
	})

	rg.GET("/target", func(c *gin.Context) {
		g.startLateWrites()
		select {
		case <-g.attempted:
		case <-time.After(time.Second):
			c.Data(http.StatusInternalServerError, "text/plain; charset=utf-8", []byte("late writer did not run"))
			return
		}

		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(brotliLateTargetBody))
	})
}

func (g *brotliLateWriteGroup) startLateWrites() {
	g.startOnce.Do(func() {
		close(g.start)
	})
}

func (g *brotliLateWriteGroup) stopLateWrites() {
	g.stopOnce.Do(func() {
		close(g.stop)
	})
}

func decodeBrotliResponse(t *testing.T, w *httptest.ResponseRecorder) []byte {
	t.Helper()

	decoded, err := io.ReadAll(brotli.NewReader(bytes.NewReader(w.Body.Bytes())))
	if err != nil {
		t.Fatalf("failed to decode brotli response: %v", err)
	}
	return decoded
}
