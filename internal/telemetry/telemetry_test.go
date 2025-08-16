package telemetry_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/robinbaeckman/go-hotels/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"go.opentelemetry.io/otel/attribute"
)

func TestWithCtx_NoSpan_NoFieldsAdded(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	log := telemetry.WithCtx(context.Background(), logger)
	log.Info("hello")

	got := buf.String()
	if contains(got, "trace_id") || contains(got, "span_id") {
		t.Fatalf("expected no trace/span fields, got: %s", got)
	}
}

func TestWithCtx_WithSpan_AddsTraceAndSpanIDs(t *testing.T) {
	// Setup a test tracer provider with an in-memory recorder
	rec := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(rec))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev); _ = tp.Shutdown(context.Background()) })

	tr := telemetry.TracerProvider("test/logger")
	ctx, span := tr.Start(context.Background(), "test-span")
	defer span.End()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	log := telemetry.WithCtx(ctx, logger)
	log.Info("hello")

	got := buf.String()
	if !contains(got, "trace_id=") {
		t.Fatalf("expected trace_id in log, got: %s", got)
	}
	if !contains(got, "span_id=") {
		t.Fatalf("expected span_id in log, got: %s", got)
	}
}

// tiny helper (avoids pulling in testify for en enkel contains)
func contains(s, sub string) bool { return bytes.Contains([]byte(s), []byte(sub)) }

func TestStartSpan_RecordsNameAndAttributes(t *testing.T) {
	// In-memory recorder
	rec := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(rec))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev); _ = tp.Shutdown(context.Background()) })

	tr := telemetry.TracerProvider("test/tracer")

	ctx := context.Background()
	attr := attribute.String("key", "value")
	ctx, span := telemetry.StartSpan(ctx, tr, "my-op", attr)
	span.End()

	// Assert recorded spans
	spans := rec.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	s := spans[0]
	if s.Name() != "my-op" {
		t.Fatalf("expected span name 'my-op', got %q", s.Name())
	}
	// verify attribute
	attrs := s.Attributes()
	found := false
	for _, a := range attrs {
		if a.Key == "key" && a.Value.AsString() == "value" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected attribute key=value, got %+v", attrs)
	}

	_ = ctx // (ctx används inte vidare i detta test, men verifierar att API:t returnerar nytt ctx)
}

func TestTracerProvider_ReturnsUsableTracer(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(rec))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev); _ = tp.Shutdown(context.Background()) })

	tr := telemetry.TracerProvider("test/usable")
	ctx, span := tr.Start(context.Background(), "op")
	span.End()

	if got := len(rec.Ended()); got != 1 {
		t.Fatalf("expected 1 span recorded, got %d", got)
	}
	_ = ctx
}
