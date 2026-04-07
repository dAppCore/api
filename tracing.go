// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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
