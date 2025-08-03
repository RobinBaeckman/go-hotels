package hotel

import (
	"context"
)

type StoreCreator interface {
	Register(ctx context.Context, h *Hotel) error
}

type StoreLister interface {
	ListByCity(ctx context.Context, city string) ([]Hotel, error)
}

type Store interface {
	Register(ctx context.Context, h *Hotel) error
	ListByCity(ctx context.Context, city string) ([]Hotel, error)
}

type Service interface {
	RegisterHotel(ctx context.Context, input RegisterHotelInput) (*Hotel, error)
	SearchHotels(ctx context.Context, city string) ([]Hotel, error)
}

type service struct {
	store Store
}

func NewService(store Store) Service {
	return &service{store: store}
}
