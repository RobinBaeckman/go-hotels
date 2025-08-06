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
		writeJSONError(w, http.StatusInternalServerError, "could not create hotel")
		return
	}

	apiHotels := hotel.ToOpenAPIList(hs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiHotels); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not create hotel")
	}
}
