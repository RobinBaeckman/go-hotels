package rest

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	once       sync.Once
	reqDurHist metric.Float64Histogram
	reqCounter metric.Int64Counter
	initErr    error
)

func initHTTPMetrics() error {
	m := otel.GetMeterProvider().Meter("go-hotels/http")

	// Histogram i sekunder, Prom-exporter gör _bucket/_sum/_count
	h, err := m.Float64Histogram(
		"go_hotels_http_request_duration_seconds",
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	c, err := m.Int64Counter("go_hotels_http_requests_total")
	if err != nil {
		return err
	}
	reqDurHist = h
	reqCounter = c
	return nil
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(code int) {
	if w.status != 0 { // redan satt
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// HTTPMetricsMiddleware – mäter duration + räknar requests med labels {http_route, method, status}
func HTTPMetricsMiddleware() func(http.Handler) http.Handler {
	once.Do(func() { initErr = initHTTPMetrics() })

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if initErr != nil {
				// Kör vidare utan metrics om init mot förmodan fallit
				next.ServeHTTP(w, r)
				return
			}

			route := routePattern(r)
			rec := &statusRecorder{ResponseWriter: w}

			start := time.Now()
			next.ServeHTTP(rec, r)
			elapsed := time.Since(start).Seconds()

			code := rec.status
			if code == 0 {
				code = http.StatusOK
			}

			attrs := []attribute.KeyValue{
				attribute.String("http_route", route),
				attribute.String("http_request_method", r.Method),
				attribute.String("http_response_status_code", strconv.Itoa(code)),
			}

			reqDurHist.Record(r.Context(), elapsed, metric.WithAttributes(attrs...))
			reqCounter.Add(r.Context(), 1, metric.WithAttributes(attrs...))
		})
	}
}
