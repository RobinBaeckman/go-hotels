package store

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/robinbaeckman/go-hotels/internal/hotel"
	pg "github.com/robinbaeckman/go-hotels/internal/store/postgres"
)

// safeIntToInt32 ensures the value is within int32 bounds before casting.
func safeIntToInt32(v int) (int32, error) {
	if v > math.MaxInt32 || v < math.MinInt32 {
		return 0, fmt.Errorf("int value %d overflows int32", v)
	}
	return int32(v), nil
}

func (r *PostgresStore) Register(ctx context.Context, h *hotel.Hotel) error {
	// basic validation only; tracing handled in domain + pgxotel
	if h.Stars < 0 || h.Stars > 5 {
		return fmt.Errorf("invalid star rating: %d", h.Stars)
	}

	stars, err := safeIntToInt32(h.Stars)
	if err != nil {
		return err
	}

	var price pgtype.Numeric
	if err = price.Scan(fmt.Sprintf("%f", h.PricePerNight)); err != nil {
		return fmt.Errorf("failed to convert price: %w", err)
	}

	_, err = r.q.CreateHotel(ctx, pg.CreateHotelParams{
		Name:          h.Name,
		City:          h.City,
		Stars:         stars,
		PricePerNight: price,
		Amenities:     h.Amenities,
	})
	return err
}

func (r *memoryStore) Register(ctx context.Context, h *hotel.Hotel) error {
	// Validate stars to keep behavior consistent with PostgresStore
	if h.Stars < 0 || h.Stars > 5 {
		return fmt.Errorf("invalid star rating: %d", h.Stars)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.hotels = append(r.hotels, *h)
	return nil
}
