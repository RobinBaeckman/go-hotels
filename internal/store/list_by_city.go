package store

import (
	"context"

	"github.com/robinbaeckman/go-hotels/internal/hotel"
	"github.com/robinbaeckman/go-hotels/internal/pkg/utils"
)

func (r *PostgresStore) ListByCity(ctx context.Context, city string) ([]hotel.Hotel, error) {
	hotels, err := r.q.ListHotels(ctx, city)
	if err != nil {
		return nil, err
	}

	result := make([]hotel.Hotel, len(hotels))
	for i, h := range hotels {
		var price float64
		if h.PricePerNight.Valid {
			f, _ := h.PricePerNight.Float64Value()
			price = f.Float64
		}

		result[i] = hotel.Hotel{
			ID:            utils.ToUUID(h.ID),
			Name:          h.Name,
			City:          h.City,
			Stars:         int(h.Stars),
			PricePerNight: price,
			Amenities:     h.Amenities,
		}
	}
	return result, nil
}

func (r *memoryStore) ListByCity(ctx context.Context, city string) ([]hotel.Hotel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []hotel.Hotel
	for _, h := range r.hotels {
		if h.City == city {
			result = append(result, h)
		}
	}
	return result, nil
}
