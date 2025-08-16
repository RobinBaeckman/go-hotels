package hotel_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/robinbaeckman/go-hotels/api"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
	"github.com/robinbaeckman/go-hotels/internal/pkg/utils"
)

type mockStore struct {
	RegisterFunc   func(ctx context.Context, h *hotel.Hotel) error
	ListByCityFunc func(ctx context.Context, city string) ([]hotel.Hotel, error)
}

func (m *mockStore) Register(ctx context.Context, h *hotel.Hotel) error {
	return m.RegisterFunc(ctx, h)
}

func (m *mockStore) ListByCity(ctx context.Context, city string) ([]hotel.Hotel, error) {
	return m.ListByCityFunc(ctx, city)
}

func TestRegisterHotel(t *testing.T) {
	tests := []struct {
		name      string
		input     hotel.RegisterHotelInput
		wantErr   bool
		setupMock func() *mockStore
	}{
		{
			name: "success",
			input: hotel.RegisterHotelInput{
				Name:          "My Hotel",
				City:          "Tokyo",
				Stars:         5,
				PricePerNight: 1500.0,
				Amenities:     []string{"WiFi"},
			},
			setupMock: func() *mockStore {
				return &mockStore{
					RegisterFunc: func(ctx context.Context, h *hotel.Hotel) error {
						return nil
					},
				}
			},
		},
		{
			name: "missing name",
			input: hotel.RegisterHotelInput{
				Name:          " ",
				City:          "Tokyo",
				Stars:         3,
				PricePerNight: 100,
				Amenities:     []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid stars",
			input: hotel.RegisterHotelInput{
				Name:          "Hotel",
				City:          "Tokyo",
				Stars:         10,
				PricePerNight: 100,
				Amenities:     []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid price",
			input: hotel.RegisterHotelInput{
				Name:          "Hotel",
				City:          "Tokyo",
				Stars:         3,
				PricePerNight: 0,
				Amenities:     []string{},
			},
			wantErr: true,
		},
		{
			name: "store failure",
			input: hotel.RegisterHotelInput{
				Name:          "Hotel",
				City:          "Tokyo",
				Stars:         3,
				PricePerNight: 100,
				Amenities:     []string{},
			},
			wantErr: true,
			setupMock: func() *mockStore {
				return &mockStore{
					RegisterFunc: func(ctx context.Context, h *hotel.Hotel) error {
						return errors.New("fail")
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var store hotel.Store
			if tt.setupMock != nil {
				store = tt.setupMock()
			} else {
				store = &mockStore{}
			}

			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))
			svc := hotel.NewService(store, logger)
			got, err := svc.RegisterHotel(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}

			if err == nil && got.Name != tt.input.Name {
				t.Errorf("name mismatch: got %q, want %q", got.Name, tt.input.Name)
			}
		})
	}
}

func TestSearchHotels(t *testing.T) {
	mock := &mockStore{
		ListByCityFunc: func(ctx context.Context, city string) ([]hotel.Hotel, error) {
			return []hotel.Hotel{
				{
					ID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Name:          "A",
					City:          city,
					Stars:         3,
					PricePerNight: 100,
					Amenities:     []string{"WiFi"},
				},
			}, nil
		},
	}
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	svc := hotel.NewService(mock, logger)
	got, err := svc.SearchHotels(context.Background(), "Osaka")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []hotel.Hotel{
		{
			ID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Name:          "A",
			City:          "Osaka",
			Stars:         3,
			PricePerNight: 100,
			Amenities:     []string{"WiFi"},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("SearchHotels mismatch (-want +got):\n%s", diff)
	}
}

func TestToOpenAPI(t *testing.T) {
	id := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	h := hotel.Hotel{
		ID:            id,
		Name:          "Test",
		City:          "Tokyo",
		Stars:         4,
		PricePerNight: 120,
		Amenities:     []string{"Gym", "Spa"},
	}

	got := hotel.ToOpenAPI(h)
	want := api.Hotel{
		Id:            &id,
		Name:          utils.String("Test"),
		City:          utils.String("Tokyo"),
		Stars:         utils.Int(4),
		PricePerNight: utils.Float32(120),
		Amenities:     utils.Strings([]string{"Gym", "Spa"}),
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ToOpenAPI mismatch (-want +got):\n%s", diff)
	}
}

func TestToOpenAPIList(t *testing.T) {
	id := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	list := []hotel.Hotel{
		{
			ID:            id,
			Name:          "Hotel A",
			City:          "City A",
			Stars:         5,
			PricePerNight: 200,
			Amenities:     []string{"Breakfast"},
		},
	}

	got := hotel.ToOpenAPIList(list)
	want := []api.Hotel{
		{
			Id:            &id,
			Name:          utils.String("Hotel A"),
			City:          utils.String("City A"),
			Stars:         utils.Int(5),
			PricePerNight: utils.Float32(200),
			Amenities:     utils.Strings([]string{"Breakfast"}),
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ToOpenAPIList mismatch (-want +got):\n%s", diff)
	}
}
