CREATE TABLE hotels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    stars INTEGER NOT NULL,
    price_per_night NUMERIC NOT NULL,
    amenities TEXT[]
);
