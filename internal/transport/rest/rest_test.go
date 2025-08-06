package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/robinbaeckman/go-hotels/api"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
	"github.com/robinbaeckman/go-hotels/internal/pkg/utils"
	"github.com/robinbaeckman/go-hotels/internal/transport/rest"
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
			h := rest.NewHandler(mockSvc, nil)

			req := httptest.NewRequest(http.MethodPost, "/hotels", bytes.NewBufferString(tt.input))
			rec := httptest.NewRecorder()

			h.CreateHotel(rec, req)

			resp := rec.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					// valfritt: logga, returnera, ignorera med kommentar
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
			h := rest.NewHandler(mockSvc, nil)

			req := httptest.NewRequest(http.MethodGet, "/hotels?city=Test%20City", nil)
			rec := httptest.NewRecorder()

			h.ListHotels(rec, req, api.ListHotelsParams{City: "Test City"})

			resp := rec.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					// valfritt: logga, returnera, ignorera med kommentar
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
