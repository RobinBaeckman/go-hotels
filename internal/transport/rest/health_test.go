package rest

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type stubDB struct{ err error }

func (s stubDB) Ping(_ context.Context) error { return s.err }

func TestGetHealth_OK(t *testing.T) {
	h := NewHandler(nil, stubDB{}, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	h.GetHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
}

func TestGetReady_OK(t *testing.T) {
	h := NewHandler(nil, stubDB{err: nil}, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()

	h.GetReady(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
}

func TestGetReady_Unhealthy(t *testing.T) {
	h := NewHandler(nil, stubDB{err: errors.New("db down")}, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()

	h.GetReady(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rr.Code)
	}
}

func TestGetReady_RespectsTimeout(t *testing.T) {
	// Simulera långsam Ping som ändå returnerar fel => ska bli 503
	h := NewHandler(nil, pingerFunc(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
			return errors.New("too slow")
		}
	}), slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	h.GetReady(rr, req)
	elapsed := time.Since(start)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rr.Code)
	}
	// Borde vara ~2s pga context timeout i handlern (med viss tolerans)
	if elapsed > 2500*time.Millisecond {
		t.Fatalf("expected handler to respect ~2s timeout, took %v", elapsed)
	}
}

type pingerFunc func(ctx context.Context) error

func (f pingerFunc) Ping(ctx context.Context) error { return f(ctx) }
