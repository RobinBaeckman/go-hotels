package store_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
	"github.com/robinbaeckman/go-hotels/internal/store"
)

func TestMemoryStore_RegisterAndListByCity(t *testing.T) {
	ctx := context.Background()
	mem := store.NewMemoryStore()

	hotels := []hotel.Hotel{
		{
			ID:            uuid.New(),
			Name:          "Hotel One",
			City:          "Tokyo",
			Stars:         4,
			PricePerNight: 10000,
			Amenities:     []string{"WiFi"},
		},
		{
			ID:            uuid.New(),
			Name:          "Hotel Two",
			City:          "Tokyo",
			Stars:         5,
			PricePerNight: 15000,
			Amenities:     []string{"Pool", "WiFi"},
		},
		{
			ID:            uuid.New(),
			Name:          "Hotel Three",
			City:          "Osaka",
			Stars:         3,
			PricePerNight: 8000,
			Amenities:     []string{"Breakfast"},
		},
	}

	// Register all hotels
	for _, h := range hotels {
		hCopy := h // capture range variable
		if err := mem.Register(ctx, &hCopy); err != nil {
			t.Fatalf("failed to register hotel: %v", err)
		}
	}

	t.Run("ListByCity_Tokyo", func(t *testing.T) {
		got, err := mem.ListByCity(ctx, "Tokyo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []hotel.Hotel{hotels[0], hotels[1]}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("ListByCity mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("ListByCity_Osaka", func(t *testing.T) {
		got, err := mem.ListByCity(ctx, "Osaka")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []hotel.Hotel{hotels[2]}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("ListByCity mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("ListByCity_Empty", func(t *testing.T) {
		got, err := mem.ListByCity(ctx, "Sapporo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 0 {
			t.Errorf("expected no results, got %d", len(got))
		}
	})
}
