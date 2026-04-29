// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"context"
	"dappco.re/go/api/internal/stdcompat/errors"
	"dappco.re/go/api/internal/stdcompat/strings"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	api "dappco.re/go/api"
)

// setupTracing creates an in-memory span exporter, wires it into a
// synchronous TracerProvider, and installs it as the global provider.
// The returned cleanup function restores the previous global state.
func setupTracing(t *testing.T) (*tracetest.InMemoryExporter, func()) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	cleanup := func() {
		_ = tp.Shutdown(context.Background())
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	}

	return exporter, cleanup
}

// hasAttribute returns true if the span stub's attributes contain a
// matching key (and optionally value).
func hasAttribute(attrs []attribute.KeyValue, key attribute.Key) (attribute.KeyValue, bool) {
	for _, a := range attrs {
		if a.Key == key {
			return a, true
		}
	}
	return attribute.KeyValue{}, false
}

type tracingTestExporter struct {
	exports int
}

func (e *tracingTestExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.exports += len(spans)
	return nil
}

func (e *tracingTestExporter) Shutdown(context.Context) error { return nil }

type failingTracingTestExporter struct {
	exports int
}

func (e *failingTracingTestExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.exports += len(spans)
	return errors.New("tracing exporter failed")
}

func (e *failingTracingTestExporter) Shutdown(context.Context) error { return nil }

type traceBodyGroup struct{}

func (g *traceBodyGroup) Name() string     { return "trace-body" }
func (g *traceBodyGroup) BasePath() string { return "/trace" }
func (g *traceBodyGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/echo", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
}

type traceNoopGroup struct {
	sawRecording bool
}

func (g *traceNoopGroup) Name() string     { return "trace-noop" }
func (g *traceNoopGroup) BasePath() string { return "/trace" }
func (g *traceNoopGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/noop", func(c *gin.Context) {
		g.sawRecording = trace.SpanFromContext(c.Request.Context()).IsRecording()
		c.Status(http.StatusOK)
	})
}

type traceEmptyGroup struct{}

func (g *traceEmptyGroup) Name() string     { return "trace-empty" }
func (g *traceEmptyGroup) BasePath() string { return "/trace" }
func (g *traceEmptyGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/empty", func(c *gin.Context) {
	})
}

// ── WithTracing ─────────────────────────────────────────────────────────

func TestWithTracing_Good_CreatesSpan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exporter, cleanup := setupTracing(t)
	defer cleanup()

	e, _ := api.New(api.WithTracing("test-service"))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Fatal("expected at least one span, got none")
	}

	// The span name should contain the route.
	span := spans[0]
	if span.Name == "" {
		t.Fatal("expected span to have a name")
	}
}

func TestWithTracing_Good_SpanHasHTTPAttributes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exporter, cleanup := setupTracing(t)
	defer cleanup()

	e, _ := api.New(api.WithTracing("test-service"))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Fatal("expected at least one span")
	}

	span := spans[0]
	attrs := span.Attributes

	// Check http.request.method attribute.
	if kv, ok := hasAttribute(attrs, attribute.Key("http.request.method")); !ok {
		t.Error("expected span to have http.request.method attribute")
	} else if kv.Value.AsString() != "GET" {
		t.Errorf("expected http.request.method=GET, got %q", kv.Value.AsString())
	}

	// Check http.route attribute.
	if _, ok := hasAttribute(attrs, attribute.Key("http.route")); !ok {
		t.Error("expected span to have http.route attribute")
	}

	// Check http.response.status_code attribute.
	if kv, ok := hasAttribute(attrs, attribute.Key("http.response.status_code")); !ok {
		t.Error("expected span to have http.response.status_code attribute")
	} else if kv.Value.AsInt64() != 200 {
		t.Errorf("expected http.response.status_code=200, got %d", kv.Value.AsInt64())
	}
}

func TestWithTracing_Good_PropagatesTraceContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exporter, cleanup := setupTracing(t)
	defer cleanup()

	e, _ := api.New(api.WithTracing("test-service"))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)

	// Inject a W3C traceparent header to simulate an upstream service.
	// Format: version-traceID-spanID-flags
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Fatal("expected at least one span")
	}

	span := spans[0]

	// The span should have a parent with the trace ID from the traceparent header.
	parentTraceID := span.Parent.TraceID()
	expectedTraceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	if parentTraceID != expectedTraceID {
		t.Errorf("expected parent trace ID %s, got %s", expectedTraceID, parentTraceID)
	}

	// The span should also share the same trace ID (trace propagation).
	spanTraceID := span.SpanContext.TraceID()
	if spanTraceID != expectedTraceID {
		t.Errorf("expected span trace ID %s to match parent %s", spanTraceID, expectedTraceID)
	}

	// The parent span ID should match what was in the traceparent header.
	parentSpanID := span.Parent.SpanID()
	expectedSpanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	if parentSpanID != expectedSpanID {
		t.Errorf("expected parent span ID %s, got %s", expectedSpanID, parentSpanID)
	}
}

func TestWithTracing_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exporter, cleanup := setupTracing(t)
	defer cleanup()

	e, _ := api.New(
		api.WithTracing("test-service"),
		api.WithRequestID(),
	)
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Tracing should produce spans.
	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Fatal("expected at least one span from WithTracing")
	}

	// WithRequestID should set the X-Request-ID header.
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}

func TestWithTracing_Good_ServiceNameInSpan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exporter, cleanup := setupTracing(t)
	defer cleanup()

	const serviceName = "my-awesome-api"
	e, _ := api.New(api.WithTracing(serviceName))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Fatal("expected at least one span")
	}

	span := spans[0]

	// otelgin uses the serviceName as the server.address attribute.
	if kv, ok := hasAttribute(span.Attributes, attribute.Key("server.address")); !ok {
		t.Error("expected span to have server.address attribute for service name")
	} else if kv.Value.AsString() != serviceName {
		t.Errorf("expected server.address=%q, got %q", serviceName, kv.Value.AsString())
	}
}

func TestTracing_NewTracerProvider_Good_ExportsSpansAndSetsGlobals(t *testing.T) {
	exporter := &tracingTestExporter{}
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()

	tp := api.NewTracerProvider(exporter)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})

	if got := otel.GetTracerProvider(); got != tp {
		t.Fatalf("expected global tracer provider to be replaced, got %T", got)
	}
	if _, ok := otel.GetTextMapPropagator().(propagation.TraceContext); !ok {
		t.Fatalf("expected TraceContext propagator, got %T", otel.GetTextMapPropagator())
	}

	tracer := otel.Tracer("tracing-test")
	_, span := tracer.Start(context.Background(), "exported-span")
	span.End()

	if got := exporter.exports; got != 1 {
		t.Fatalf("expected exporter to receive one span, got %d", got)
	}
}

func TestTracing_NewTracerProvider_Bad_AllowsNilExporter(t *testing.T) {
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()

	tp := api.NewTracerProvider(nil)
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})

	if got := otel.GetTracerProvider(); got != tp {
		t.Fatalf("expected global tracer provider to be replaced, got %T", got)
	}

	tracer := otel.Tracer("tracing-test")
	_, span := tracer.Start(context.Background(), "nil-exporter-span")
	span.End()
}

func TestTracing_NewTracerProvider_Ugly_HandlesExporterErrors(t *testing.T) {
	exporter := &failingTracingTestExporter{}
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()

	tp := api.NewTracerProvider(exporter)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})

	tracer := otel.Tracer("tracing-test")
	_, span := tracer.Start(context.Background(), "failing-exporter-span")
	span.End()

	if got := exporter.exports; got != 1 {
		t.Fatalf("expected failing exporter to be invoked once, got %d", got)
	}
}

func TestTracing_WithTracing_Good_AttachesDurationAndSizeAttributes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exporter, cleanup := setupTracing(t)
	defer cleanup()

	e, _ := api.New(api.WithTracing("trace-service"))
	e.Register(&traceBodyGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/trace/echo", strings.NewReader("abc"))
	req.Header.Set("Content-Type", "text/plain")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Fatal("expected at least one span")
	}

	span := spans[0]
	if kv, ok := hasAttribute(span.Attributes, attribute.Key("http.request.body.size")); !ok {
		t.Fatal("expected span to record request body size")
	} else if kv.Value.AsInt64() != 3 {
		t.Fatalf("expected request body size 3, got %d", kv.Value.AsInt64())
	}
	if kv, ok := hasAttribute(span.Attributes, attribute.Key("http.response.body.size")); !ok {
		t.Fatal("expected span to record response body size")
	} else if kv.Value.AsInt64() != 4 {
		t.Fatalf("expected response body size 4, got %d", kv.Value.AsInt64())
	}
	if _, ok := hasAttribute(span.Attributes, attribute.Key("http.server.duration_ms")); !ok {
		t.Fatal("expected span to record server duration")
	}
}

func TestTracing_WithTracing_Bad_SkipsAttributesWhenSpanIsNotRecording(t *testing.T) {
	gin.SetMode(gin.TestMode)

	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})

	e, _ := api.New(api.WithTracing("noop-service"))
	group := &traceNoopGroup{}
	e.Register(group)

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/trace/noop", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if group.sawRecording {
		t.Fatal("expected no-op tracer provider to expose a non-recording span")
	}
}

func TestTracing_WithTracing_Ugly_OmitsResponseSizeForEmptyResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exporter, cleanup := setupTracing(t)
	defer cleanup()

	e, _ := api.New(api.WithTracing("trace-service"))
	e.Register(&traceEmptyGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/trace/empty", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Fatal("expected at least one span")
	}

	span := spans[0]
	if _, ok := hasAttribute(span.Attributes, attribute.Key("http.response.body.size")); ok {
		t.Fatal("expected empty responses to skip the response body size attribute")
	}
	if _, ok := hasAttribute(span.Attributes, attribute.Key("http.server.duration_ms")); !ok {
		t.Fatal("expected span to record server duration")
	}
}
