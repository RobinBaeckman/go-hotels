package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// This test covers the branch where init failed and middleware should be a no-op passthrough.
func TestHTTPMetricsMiddleware_InitErr_Passthrough(t *testing.T) {
	// Ensure a clean state for once/instruments.
	resetMiddlewareState()

	// Force the initErr path.
	initErr = errors.New("forced init failure")
	t.Cleanup(func() {
		// reset for other tests
		resetMiddlewareState()
		initErr = nil
	})

	// Build middleware
	mw := HTTPMetricsMiddleware()

	// Next handler should run untouched and return 202.
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()

	mw(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("want 202 passthrough, got %d", rr.Code)
	}
}
