package hotel

import "errors"

var (
	ErrInvalidHotelName  = errors.New("hotel name cannot be empty")
	ErrInvalidCity       = errors.New("city name cannot be empty")
	ErrInvalidStarRating = errors.New("star rating must be between 1 and 5")
	ErrInvalidPrice      = errors.New("price must be greater than 0")
	ErrTooManyAmenities  = errors.New("too many amenities (max 20)")
)
