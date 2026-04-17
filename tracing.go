// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"time"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/gin-gonic/gin"
)

// WithTracing adds OpenTelemetry distributed tracing middleware via otelgin.
// Each incoming request produces a span tagged with the HTTP method, route,
// and status code. Trace context is propagated using W3C traceparent headers.
//
// The serviceName identifies this service in distributed traces. If a
// TracerProvider has not been configured globally via otel.SetTracerProvider,
// the middleware uses the global default (which is a no-op until set).
//
// Typical setup in main:
//
//	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
//	otel.SetTracerProvider(tp)
//	otel.SetTextMapPropagator(propagation.TraceContext{})
//
//	engine, _ := api.New(api.WithTracing("my-service"))
//
// Example:
//
//	api.New(api.WithTracing("my-service"))
func WithTracing(serviceName string, opts ...otelgin.Option) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, otelgin.Middleware(serviceName, opts...))
		e.middlewares = append(e.middlewares, tracingAttributesMiddleware())
	}
}

// NewTracerProvider creates a TracerProvider configured with the given
// SpanExporter and returns it. The caller is responsible for calling
// Shutdown on the returned provider when the application exits.
//
// This is a convenience helper for tests and simple deployments.
// Production setups should build their own TracerProvider with batching,
// resource attributes, and appropriate exporters.
//
// Example:
//
//	tp := api.NewTracerProvider(exporter)
//	_ = tp.Shutdown(context.Background())
func NewTracerProvider(exporter sdktrace.SpanExporter) *sdktrace.TracerProvider {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp
}

// tracingAttributesMiddleware enriches the active span with request size and
// request duration metadata after the handler completes.
//
//	api.New(api.WithTracing("core-api"))
func tracingAttributesMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		span := trace.SpanFromContext(c.Request.Context())
		if span == nil || !span.IsRecording() {
			return
		}

		size := c.Writer.Size()
		attrs := []attribute.KeyValue{
			attribute.Int64("http.server.duration_ms", time.Since(start).Milliseconds()),
		}

		if size >= 0 {
			attrs = append(attrs, attribute.Int("http.response.body.size", size))
		}

		if size := c.Request.ContentLength; size >= 0 {
			attrs = append(attrs, attribute.Int64("http.request.body.size", size))
		}

		span.SetAttributes(attrs...)
	}
}
