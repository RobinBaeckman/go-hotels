package telemetry

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	runtimeinstr "go.opentelemetry.io/contrib/instrumentation/runtime"
)

type Config struct {
	Endpoint        string
	Insecure        bool
	ServiceName     string
	ServiceVersion  string
	Environment     string
	MetricsInterval time.Duration
}

type Shutdown func(context.Context) error

func Setup(ctx context.Context, c Config) (Shutdown, error) {
	if c.Endpoint == "" {
		if ep := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); ep != "" {
			c.Endpoint = ep
		} else {
			c.Endpoint = "alloy:4317"
		}
	}
	if c.MetricsInterval == 0 {
		c.MetricsInterval = 10 * time.Second
	}

	// Resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(c.ServiceName),
			semconv.ServiceVersionKey.String(c.ServiceVersion),
			attribute.String("deployment.environment", c.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// ── Traces ────────────────────────────────────────────────────────────────
	topts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(c.Endpoint)}
	if c.Insecure {
		topts = append(topts, otlptracegrpc.WithInsecure())
	}
	traceExp, err := otlptracegrpc.New(ctx, topts...)
	if err != nil {
		return nil, err
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExp),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// ── Metrics (v1.37 views + explicit buckets incl. 0.3s) ─────────────────
	mopts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(c.Endpoint)}
	if c.Insecure {
		mopts = append(mopts, otlpmetricgrpc.WithInsecure())
	}
	metricExp, err := otlpmetricgrpc.New(ctx, mopts...)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, err
	}

	reader := sdkmetric.NewPeriodicReader(
		metricExp,
		sdkmetric.WithInterval(c.MetricsInterval),
	)

	// Buckets in seconds (Prometheus will expose _bucket/_sum/_count).
	// Includes 0.3s (300ms) for your latency SLI.
	buckets := []float64{
		0.005, 0.010, 0.025, 0.050, 0.075, 0.100,
		0.250, 0.300, 0.500, 0.750, 1.0, 2.5, 5.0, 7.5, 10.0,
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),

		// Your custom HTTP histogram
		sdkmetric.WithView(sdkmetric.NewView(
			sdkmetric.Instrument{
				Name: "go_hotels_http_request_duration_seconds",
			},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: buckets,
				},
			},
		)),

		// Override default OTel HTTP server histogram to share the same buckets
		sdkmetric.WithView(sdkmetric.NewView(
			sdkmetric.Instrument{
				Name: "http.server.request.duration",
			},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: buckets,
				},
			},
		)),
	)

	otel.SetMeterProvider(mp)

	// Runtime metrics
	_ = runtimeinstr.Start(runtimeinstr.WithMeterProvider(mp))

	// Propagation
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		),
	)

	return func(ctx context.Context) error {
		var first error
		if err := mp.Shutdown(ctx); err != nil && first == nil {
			first = err
		}
		if err := tp.Shutdown(ctx); err != nil && first == nil {
			first = err
		}
		return first
	}, nil
}
