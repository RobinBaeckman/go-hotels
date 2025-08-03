package hotel

import "context"

func (s *service) SearchHotels(ctx context.Context, city string) ([]Hotel, error) {
	return s.store.ListByCity(ctx, city)
}
