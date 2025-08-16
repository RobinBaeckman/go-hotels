package hotel

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/robinbaeckman/go-hotels/internal/telemetry"
)

func (s *service) RegisterHotel(ctx context.Context, input RegisterHotelInput) (*Hotel, error) {
	// tracer from global provider (set in telemetry.Setup)
	tr := otel.Tracer("go-hotels/domain")
	ctx, span := tr.Start(ctx, "HotelService.RegisterHotel",
		trace.WithAttributes(
			attribute.String("hotel.city", input.City),
			attribute.Int("hotel.stars", input.Stars),
		),
	)
	defer span.End()

	// logger enriched with trace/span IDs from ctx
	log := telemetry.WithCtx(ctx, s.log)
	if log == nil {
		log = slog.Default()
	}

	// validation
	if strings.TrimSpace(input.Name) == "" {
		err := errors.New("hotel name is required")
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation")
		log.Warn("validation failed", "err", err)
		return nil, err
	}
	if input.Stars < 1 || input.Stars > 5 {
		err := errors.New("stars must be between 1 och 5")
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation")
		log.Warn("validation failed", "err", err)
		return nil, err
	}
	if input.PricePerNight <= 0 {
		err := errors.New("price must be greater than 0")
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation")
		log.Warn("validation failed", "err", err)
		return nil, err
	}

	h := &Hotel{
		ID:            uuid.New(),
		Name:          input.Name,
		City:          input.City,
		Stars:         input.Stars,
		PricePerNight: input.PricePerNight,
		Amenities:     input.Amenities,
	}
	span.SetAttributes(attribute.String("hotel.id", h.ID.String()))

	// store layer (pgxotel will trace SQL)
	if err := s.store.Register(ctx, h); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "store.register")
		log.Error("store.Register failed", slog.String("err", err.Error()))
		return nil, err
	}

	log.Info("hotel registered", "id", h.ID.String(), "city", h.City, "stars", h.Stars)
	return h, nil
}
