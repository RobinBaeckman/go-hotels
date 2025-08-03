package rest

import (
	"encoding/json"
	"net/http"

	"github.com/robinbaeckman/go-hotels/api"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
)

func (h *Handler) ListHotels(w http.ResponseWriter, r *http.Request, params api.ListHotelsParams) {
	ctx := r.Context()

	hs, err := h.svc.SearchHotels(ctx, params.City)
	if err != nil {
		http.Error(w, "failed to list hotels: "+err.Error(), http.StatusInternalServerError)
		return
	}

	apiHotels := hotel.ToOpenAPIList(hs)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(apiHotels); err != nil {
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}
