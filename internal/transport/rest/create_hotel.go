package rest

import (
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/robinbaeckman/go-hotels/api"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
)

func (h *Handler) CreateHotel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := LoggerFrom(ctx, h.log)

	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.SetAttributes(attribute.String("http.payload.type", "HotelInput"))
	}

	// Cap body size (e.g. 1MB)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input api.HotelInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.kind", "decode"))
		}
		log.Warn("invalid JSON", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	hInput := hotel.RegisterHotelInput{
		Name:          input.Name,
		City:          input.City,
		Stars:         input.Stars,
		PricePerNight: float64(input.PricePerNight),
		Amenities:     input.Amenities,
	}

	hDomain, err := h.svc.RegisterHotel(ctx, hInput)
	if err != nil {
		if span != nil {
			span.RecordError(err)
		}
		log.Error("create hotel failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "could not create hotel")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(hotel.ToOpenAPI(*hDomain)); err != nil {
		if span != nil {
			span.RecordError(err)
		}
		log.Error("encode response failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "could not create hotel")
		return
	}

	log.Info("hotel created", "name", hInput.Name, "city", hInput.City, "stars", hInput.Stars)
}
