-- name: CreateHotel :one
INSERT INTO hotels (name, city, stars, price_per_night, amenities)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListHotels :many
SELECT * FROM hotels WHERE city = $1;
