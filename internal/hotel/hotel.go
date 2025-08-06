package hotel

import (
	"github.com/google/uuid"
	"github.com/robinbaeckman/go-hotels/api"
	"github.com/robinbaeckman/go-hotels/internal/pkg/utils"
)

type Hotel struct {
	ID            uuid.UUID
	Name          string
	City          string
	Stars         int
	PricePerNight float64
	Amenities     []string
}

type RegisterHotelInput struct {
	ID            string
	Name          string
	City          string
	Stars         int
	PricePerNight float64
	Amenities     []string
}

func FromInput(input api.HotelInput) *Hotel {
	return &Hotel{
		ID:            uuid.New(),
		Name:          input.Name,
		City:          input.City,
		Stars:         input.Stars,
		PricePerNight: float64(input.PricePerNight),
		Amenities:     input.Amenities,
	}
}

func ToOpenAPI(h Hotel) api.Hotel {
	return api.Hotel{
		Id:            utils.UUIDToOAPIPtr(h.ID),
		Name:          utils.String(h.Name),
		City:          utils.String(h.City),
		Stars:         utils.Int(h.Stars),
		PricePerNight: utils.Float32(float32(h.PricePerNight)),
		Amenities:     utils.Strings(h.Amenities),
	}
}

func ToOpenAPIList(list []Hotel) []api.Hotel {
	res := make([]api.Hotel, len(list))
	for i, h := range list {
		res[i] = ToOpenAPI(h)
	}
	return res
}
