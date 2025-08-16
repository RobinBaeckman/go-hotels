package rest

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey int

const loggerKey ctxKey = iota

// LoggerMiddleware enriches logs and also tags OTel span/metrics with http.route.
func LoggerMiddleware(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := routePattern(r)

			// just use the existing context
			ctx := r.Context()

			// If a server span exists (from otelhttp), decorate it with http.route
			if span := trace.SpanFromContext(ctx); span != nil {
				span.SetAttributes(attribute.String("http.route", route))
			}

			// Build request logger (includes trace/span if present)
			sc := trace.SpanContextFromContext(ctx)
			l := base.With(
				"http.method", r.Method,
				"http.route", route,
				"http.path", r.URL.Path,
			)
			if sc.HasTraceID() {
				l = l.With("trace_id", sc.TraceID().String())
			}
			if sc.HasSpanID() {
				l = l.With("span_id", sc.SpanID().String())
			}
			ctx = context.WithValue(ctx, loggerKey, l)

			start := time.Now()
			next.ServeHTTP(w, r.WithContext(ctx))
			l.Debug("request completed", "duration_ms", time.Since(start).Milliseconds())
		})
	}
}

func LoggerFrom(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return fallback
}

// routePattern tries to fetch the chi route pattern (e.g. /hotels/{id})
// Fallback to raw path if unknown.
func routePattern(r *http.Request) string {
	if rc := chiRouteContext(r.Context()); rc != nil {
		if pat := rc.RoutePattern(); pat != "" {
			return pat
		}
	}
	return r.URL.Path
}

// tiny wrapper to avoid direct import in this file
type chiRouterContext interface{ RoutePattern() string }

func chiRouteContext(ctx context.Context) chiRouterContext {
	// same shape chi uses for RouteCtxKey
	var routeCtxKey interface{} = struct{ name string }{"RouteContextKey"}
	if v := ctx.Value(routeCtxKey); v != nil {
		if rc, ok := v.(chiRouterContext); ok {
			return rc
		}
	}
	return nil
}
