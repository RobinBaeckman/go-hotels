package hotel

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

func (s *service) RegisterHotel(ctx context.Context, input RegisterHotelInput) (*Hotel, error) {
	// Business layer validation
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("hotel name is required")
	}
	if input.Stars < 1 || input.Stars > 5 {
		return nil, errors.New("stars must be between 1 and 5")
	}
	if input.PricePerNight <= 0 {
		return nil, errors.New("price must be greater than 0")
	}

	hotel := &Hotel{
		ID:            uuid.New(),
		Name:          input.Name,
		City:          input.City,
		Stars:         input.Stars,
		PricePerNight: input.PricePerNight,
		Amenities:     input.Amenities,
	}

	if err := s.store.Register(ctx, hotel); err != nil {
		return nil, err
	}
	return hotel, nil
}
