package store

import (
	"context"
	"fmt"

	"github.com/robinbaeckman/go-hotels/internal/hotel"
	pg "github.com/robinbaeckman/go-hotels/internal/store/postgres"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresStore) Register(ctx context.Context, h *hotel.Hotel) error {
	// Convert float to pgtype.Numeric
	var price pgtype.Numeric
	if err := price.Scan(fmt.Sprintf("%f", h.PricePerNight)); err != nil {
		return fmt.Errorf("failed to convert price: %w", err)
	}

	_, err := r.q.CreateHotel(ctx, pg.CreateHotelParams{
		Name:          h.Name,
		City:          h.City,
		Stars:         int32(h.Stars),
		PricePerNight: price,
		Amenities:     h.Amenities,
	})
	return err
}

func (r *memoryStore) Register(ctx context.Context, h *hotel.Hotel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hotels = append(r.hotels, *h)
	return nil
}
