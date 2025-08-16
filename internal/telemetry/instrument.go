package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Tracer = trace.Tracer

// TracerProvider returns a service-wide tracer. Call once in main after Setup.
func TracerProvider(name string) Tracer { return otel.Tracer(name) }

// WithCtx enriches a slog.Logger with trace/span IDs from ctx.
func WithCtx(ctx context.Context, base *slog.Logger) *slog.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return base
	}
	sc := span.SpanContext()
	return base.With(
		"trace_id", sc.TraceID().String(),
		"span_id", sc.SpanID().String(),
	)
}

// StartSpan is a tiny wrapper to cut noise in handlers/services.
func StartSpan(ctx context.Context, tracer Tracer, name string, kv ...attribute.KeyValue) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, trace.WithAttributes(kv...))
}
