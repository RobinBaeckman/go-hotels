-- Enable UUID generation if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Ensure the hotels table uses UUID default for id
ALTER TABLE hotels
    ALTER COLUMN id SET DEFAULT uuid_generate_v4();
