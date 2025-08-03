package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
	"github.com/robinbaeckman/go-hotels/internal/store"
	pg "github.com/robinbaeckman/go-hotels/internal/store/postgres"
)

// fakeQuerier implements pg.Querier for testing
type fakeQuerier struct {
	createCalled   bool
	createParams   *pg.CreateHotelParams
	createFail     bool
	listFail       bool
	listHotelsResp []pg.Hotel
}

func (f *fakeQuerier) CreateHotel(ctx context.Context, args pg.CreateHotelParams) (pg.Hotel, error) {
	f.createCalled = true
	f.createParams = &args
	if f.createFail {
		return pg.Hotel{}, errors.New("create failed")
	}
	return pg.Hotel{}, nil
}

func (f *fakeQuerier) ListHotels(ctx context.Context, city string) ([]pg.Hotel, error) {
	if f.listFail {
		return nil, errors.New("list failed")
	}
	return f.listHotelsResp, nil
}

func TestPostgresStore_Register(t *testing.T) {
	tests := []struct {
		name        string
		createFail  bool
		expectError bool
	}{
		{
			name:        "success",
			createFail:  false,
			expectError: false,
		},
		{
			name:        "error from CreateHotel",
			createFail:  true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeQuerier{createFail: tt.createFail}
			ps := store.NewPostgresStore(fake)

			h := &hotel.Hotel{
				ID:            uuid.New(),
				Name:          "TestHotel",
				City:          "TestCity",
				Stars:         4,
				PricePerNight: 123.45,
				Amenities:     []string{"wifi"},
			}

			err := ps.Register(context.Background(), h)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && (!fake.createCalled || fake.createParams == nil) {
				t.Errorf("expected CreateHotel to be called with params")
			}
		})
	}
}

func TestPostgresStore_ListByCity(t *testing.T) {
	tests := []struct {
		name           string
		listFail       bool
		listHotelsResp []pg.Hotel
		expectError    bool
		expectedCount  int
	}{
		{
			name:     "success with one hotel",
			listFail: false,
			listHotelsResp: func() []pg.Hotel {
				// Skapa ett giltigt pgtype.Numeric via Scan
				var price pgtype.Numeric
				if err := price.Scan("500.0"); err != nil {
					t.Fatalf("failed to scan numeric: %v", err)
				}
				return []pg.Hotel{
					{
						Name:          "HotelOne",
						City:          "Tokyo",
						Stars:         3,
						Amenities:     []string{"wifi"},
						PricePerNight: price,
					},
				}
			}(),
			expectError:   false,
			expectedCount: 1,
		},
		{
			name:           "error from ListHotels",
			listFail:       true,
			listHotelsResp: nil,
			expectError:    true,
			expectedCount:  0,
		},
		{
			name:           "empty list",
			listFail:       false,
			listHotelsResp: []pg.Hotel{},
			expectError:    false,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeQuerier{
				listFail:       tt.listFail,
				listHotelsResp: tt.listHotelsResp,
			}
			ps := store.NewPostgresStore(fake)

			got, err := ps.ListByCity(context.Background(), "Tokyo")

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(got) != tt.expectedCount {
				t.Errorf("expected %d hotels, got %d", tt.expectedCount, len(got))
			}
		})
	}
}
