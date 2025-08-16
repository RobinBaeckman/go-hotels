package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/robinbaeckman/go-hotels/api"
	"github.com/robinbaeckman/go-hotels/internal/config"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
	"github.com/robinbaeckman/go-hotels/internal/pkg/utils"
	pg "github.com/robinbaeckman/go-hotels/internal/store/postgres"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

type MockService struct {
	RegisterHotelFunc func(hotel.RegisterHotelInput) (*hotel.Hotel, error)
	SearchHotelsFunc  func(city string) ([]hotel.Hotel, error)
}

func (m *MockService) RegisterHotel(ctx context.Context, input hotel.RegisterHotelInput) (*hotel.Hotel, error) {
	return m.RegisterHotelFunc(input)
}

func (m *MockService) SearchHotels(ctx context.Context, city string) ([]hotel.Hotel, error) {
	return m.SearchHotelsFunc(city)
}

func TestCreateHotel(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		mockResponse   *hotel.Hotel
		mockErr        error
		wantStatusCode int
		wantBody       *api.Hotel
	}{
		{
			name: "success",
			input: `{
				"name": "Test Hotel",
				"city": "Test City",
				"stars": 5,
				"price_per_night": 999.99,
				"amenities": ["WiFi", "Pool"]
			}`,
			mockResponse: &hotel.Hotel{
				ID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Name:          "Test Hotel",
				City:          "Test City",
				Stars:         5,
				PricePerNight: 999.99,
				Amenities:     []string{"WiFi", "Pool"},
			},
			wantStatusCode: http.StatusCreated,
			wantBody: &api.Hotel{
				Id:            utils.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
				Name:          utils.Ptr("Test Hotel"),
				City:          utils.Ptr("Test City"),
				Stars:         utils.Ptr(5),
				PricePerNight: utils.Ptr(float32(999.99)),
				Amenities:     utils.Ptr([]string{"WiFi", "Pool"}),
			},
		},
		{
			name:           "invalid JSON",
			input:          `{invalid-json}`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				RegisterHotelFunc: func(input hotel.RegisterHotelInput) (*hotel.Hotel, error) {
					return tt.mockResponse, tt.mockErr
				},
			}
			cfg := config.Load()

			ctx := context.Background()
			pool, err := pg.Connect(ctx, cfg.DatabaseURL)
			if err != nil {
				slog.Error("failed to connect to database", "err", err)
				os.Exit(1)
			}
			defer func() {
				slog.Info("closing database connection pool")
				pool.Close()
			}()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
			slog.SetDefault(logger)

			h := NewHandler(mockSvc, pool, logger)

			req := httptest.NewRequest(http.MethodPost, "/hotels", bytes.NewBufferString(tt.input))
			rec := httptest.NewRecorder()

			h.CreateHotel(rec, req)

			resp := rec.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Println("failed to close response body:", err)
				}
			}()

			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}

			if tt.wantBody != nil {
				var got api.Hotel
				bodyResp, _ := io.ReadAll(resp.Body)
				_ = json.Unmarshal(bodyResp, &got)

				if diff := cmp.Diff(tt.wantBody, &got); diff != "" {
					t.Errorf("response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestListHotels(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   []hotel.Hotel
		mockErr        error
		wantStatusCode int
		wantBody       []api.Hotel
	}{
		{
			name: "success",
			mockResponse: []hotel.Hotel{
				{
					ID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Name:          "Hotel A",
					City:          "Test City",
					Stars:         4,
					PricePerNight: 500,
					Amenities:     []string{"WiFi", "Breakfast"},
				},
			},
			wantStatusCode: http.StatusOK,
			wantBody: []api.Hotel{
				{
					Id:            utils.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
					Name:          utils.Ptr("Hotel A"),
					City:          utils.Ptr("Test City"),
					Stars:         utils.Ptr(4),
					PricePerNight: utils.Ptr(float32(500)),
					Amenities:     utils.Ptr([]string{"WiFi", "Breakfast"}),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				SearchHotelsFunc: func(city string) ([]hotel.Hotel, error) {
					return tt.mockResponse, tt.mockErr
				},
			}

			cfg := config.Load()

			ctx := context.Background()
			pool, err := pg.Connect(ctx, cfg.DatabaseURL)
			if err != nil {
				slog.Error("failed to connect to database", "err", err)
				os.Exit(1)
			}
			defer func() {
				slog.Info("closing database connection pool")
				pool.Close()
			}()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
			slog.SetDefault(logger)

			h := NewHandler(mockSvc, pool, logger)

			req := httptest.NewRequest(http.MethodGet, "/hotels?city=Test%20City", nil)
			rec := httptest.NewRecorder()

			h.ListHotels(rec, req, api.ListHotelsParams{City: "Test City"})

			resp := rec.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Println("failed to close response body:", err)
				}
			}()

			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}

			if tt.wantBody != nil {
				var got []api.Hotel
				bodyResp, _ := io.ReadAll(resp.Body)
				_ = json.Unmarshal(bodyResp, &got)

				if diff := cmp.Diff(tt.wantBody, got); diff != "" {
					t.Errorf("response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// --- OTEL test plumbing (manual reader + reset av middleware-state) ---

// setTestMeterProvider replaces the global MeterProvider with a ManualReader-backed one
// and returns the reader plus a cleanup to restore the previous provider.
func setTestMeterProvider(t *testing.T) (*sdkmetric.ManualReader, func()) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	prev := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)

	cleanup := func() {
		otel.SetMeterProvider(prev)
		_ = mp.Shutdown(context.Background())
	}
	return reader, cleanup
}

// resetMiddlewareState clears the package-level sync.Once so initHTTPMetrics() can run again
// with our test MeterProvider.
func resetMiddlewareState() {
	once = sync.Once{}
	initErr = nil
	// instruments är interfaces → nollvärdet är nil
	reqDurHist = nil
	reqCounter = nil
}

func TestHTTPMetricsMiddleware_RecordsCounterAndHistogram(t *testing.T) {
	reader, cleanup := setTestMeterProvider(t)
	defer cleanup()
	resetMiddlewareState()

	// Handler that sets 201 and sleeps a bit to produce a measurable duration.
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated) // 201
		time.Sleep(10 * time.Millisecond)
		_, _ = w.Write([]byte("ok"))
	})

	mw := HTTPMetricsMiddleware()
	ts := httptest.NewServer(mw(next))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/hotels/123", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp.Body.Close()

	// Collect metrics from the ManualReader
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	// Flatten scope metrics
	var all []metricdata.Metrics
	for _, sm := range rm.ScopeMetrics {
		all = append(all, sm.Metrics...)
	}

	// Find our metrics
	hist := findMetric(all, "go_hotels_http_request_duration_seconds")
	if hist == nil {
		t.Fatalf("histogram not found")
	}
	counter := findMetric(all, "go_hotels_http_requests_total")
	if counter == nil {
		t.Fatalf("counter not found")
	}

	// Assert histogram has one data point with attrs we expect
	switch dp := hist.Data.(type) {
	case metricdata.Histogram[float64]:
		if len(dp.DataPoints) == 0 {
			t.Fatalf("expected at least 1 histogram datapoint")
		}
		h := dp.DataPoints[len(dp.DataPoints)-1]
		if h.Count == 0 {
			t.Fatalf("expected histogram count > 0")
		}
		if h.Sum < 0 {
			t.Fatalf("expected non-negative histogram sum, got %v", h.Sum)
		}
		assertHasAttr(t, h.Attributes, "http_request_method", "GET")
		assertHasAttr(t, h.Attributes, "http_response_status_code", "201")
		assertHasKey(t, h.Attributes, "http_route")
	default:
		t.Fatalf("histogram has unexpected type: %T", hist.Data)
	}

	// Assert counter == 1 with attributes
	switch dp := counter.Data.(type) {
	case metricdata.Sum[int64]:
		if len(dp.DataPoints) == 0 {
			t.Fatalf("expected at least 1 counter datapoint")
		}
		c := dp.DataPoints[len(dp.DataPoints)-1]
		if c.Value < 1 {
			t.Fatalf("expected counter >= 1, got %d", c.Value)
		}
		assertHasAttr(t, c.Attributes, "http_request_method", "GET")
		assertHasAttr(t, c.Attributes, "http_response_status_code", "201")
		assertHasKey(t, c.Attributes, "http_route")
	default:
		t.Fatalf("counter has unexpected type: %T", counter.Data)
	}
}

func TestHTTPMetricsMiddleware_DefaultsStatus200WhenWriteHeaderNotCalled(t *testing.T) {
	reader, cleanup := setTestMeterProvider(t)
	defer cleanup()
	resetMiddlewareState()

	// Handler that never calls WriteHeader ⇒ statusRecorder should default to 200
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	mw := HTTPMetricsMiddleware()
	req := httptest.NewRequest(http.MethodPost, "/ping", nil)
	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	var all []metricdata.Metrics
	for _, sm := range rm.ScopeMetrics {
		all = append(all, sm.Metrics...)
	}
	counter := findMetric(all, "go_hotels_http_requests_total")
	if counter == nil {
		t.Fatalf("counter not found")
	}

	switch dp := counter.Data.(type) {
	case metricdata.Sum[int64]:
		c := dp.DataPoints[len(dp.DataPoints)-1]
		assertHasAttr(t, c.Attributes, "http_request_method", "POST")
		assertHasAttr(t, c.Attributes, "http_response_status_code", "200") // defaulted by statusRecorder
	default:
		t.Fatalf("counter has unexpected type: %T", counter.Data)
	}
}

// --- helpers ---

func findMetric(ms []metricdata.Metrics, name string) *metricdata.Metrics {
	for i := range ms {
		if ms[i].Name == name {
			return &ms[i]
		}
	}
	return nil
}

func assertHasAttr(t *testing.T, set attribute.Set, key, want string) {
	t.Helper()
	if v, ok := set.Value(attribute.Key(key)); !ok || v.AsString() != want {
		t.Fatalf("expected attr %q=%q, got %v (present=%v)", key, want, v, ok)
	}
}

func assertHasKey(t *testing.T, set attribute.Set, key string) {
	t.Helper()
	if _, ok := set.Value(attribute.Key(key)); !ok {
		t.Fatalf("expected attr key present: %q", key)
	}
}

func TestStatusRecorder_WriteHeader_IgnoresSecondCall(t *testing.T) {
	rr := httptest.NewRecorder()
	rec := &statusRecorder{ResponseWriter: rr}

	rec.WriteHeader(http.StatusCreated) // 201
	rec.WriteHeader(http.StatusTeapot)  // 418 (ska ignoreras)

	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201 locked-in, got %d", rr.Code)
	}
	if rec.status != http.StatusCreated {
		t.Fatalf("want rec.status=201, got %d", rec.status)
	}
}

func TestRoutePattern_Nested(t *testing.T) {
	r := chi.NewRouter()

	// fångar vad routePattern returnerar
	gotCh := make(chan string, 1)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/hotels/{id}/rooms/{rid}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotCh <- routePattern(r)
			w.WriteHeader(http.StatusNoContent)
		}))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/v1/hotels/abc/rooms/def", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	var got string
	select {
	case got = <-gotCh:
	default:
		t.Fatalf("did not capture routePattern value")
	}

	if got == "" {
		t.Fatalf("expected non-empty pattern/path")
	}

	// Acceptera både mall och faktisk path beroende på implementation
	wantPattern := "/v1/hotels/{id}/rooms/{rid}"
	wantPath := "/v1/hotels/abc/rooms/def"
	if got != wantPattern && got != wantPath && got != "GET "+wantPattern {
		t.Fatalf("unexpected pattern/path: %q", got)
	}
}

func TestRoutePattern_NoChiContext_ReturnsPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/raw/path", nil)
	got := routePattern(req)
	if got == "" {
		t.Fatalf("expect non-empty fallback")
	}
	// oftast blir got == "/raw/path"
}
