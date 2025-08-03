package rest

import (
	"github.com/robinbaeckman/go-hotels/internal/hotel"
)

type Handler struct {
	svc hotel.Service
}

func NewHandler(svc hotel.Service) *Handler {
	return &Handler{svc: svc}
}
