package rest

import (
	"encoding/json"
	"net/http"

	"github.com/robinbaeckman/go-hotels/api"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
)

func (h *Handler) CreateHotel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input api.HotelInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	hotelInput := hotel.RegisterHotelInput{
		Name:          input.Name,
		City:          input.City,
		Stars:         input.Stars,
		PricePerNight: float64(input.PricePerNight),
		Amenities:     input.Amenities,
	}

	hDomain, err := h.svc.RegisterHotel(ctx, hotelInput)
	if err != nil {
		http.Error(w, "could not create hotel", http.StatusInternalServerError)
		return
	}

	response := hotel.ToOpenAPI(*hDomain)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
