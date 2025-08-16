package rest

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggerMiddleware_SetsLoggerAndPassesThrough(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug, // ensure our .Info is emitted
	}))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The middleware should have put a logger in the context.
		l := LoggerFrom(r.Context(), nil)
		if l == nil {
			t.Fatalf("expected logger in context")
		}

		// Write a line via the injected logger to prove propagation works.
		l.Info("handler saw logger")
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/m", nil)
	rr := httptest.NewRecorder()

	LoggerMiddleware(base)(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rr.Code)
	}

	out := buf.String()
	if !strings.Contains(out, "handler saw logger") {
		t.Fatalf("expected log line from handler via context logger, got: %q", out)
	}
}

func TestLoggerFrom_NilWhenAbsent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/m", nil)

	// Pass explicit fallback, should be returned if none in context
	fallback := slog.Default()
	got := LoggerFrom(req.Context(), fallback)

	if got != fallback {
		t.Fatalf("expected fallback logger, got %+v", got)
	}
}
