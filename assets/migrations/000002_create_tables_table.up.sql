-- Create tables table
CREATE TABLE IF NOT EXISTS tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number VARCHAR(50) NOT NULL UNIQUE,
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    is_available BOOLEAN NOT NULL DEFAULT true,
    location VARCHAR(20) NOT NULL CHECK (location IN ('main', 'terrace', 'private')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on table number for faster lookups
CREATE INDEX IF NOT EXISTS idx_tables_number ON tables(number);

-- Create index on availability for filtering
CREATE INDEX IF NOT EXISTS idx_tables_is_available ON tables(is_available);

-- Create index on location for filtering
CREATE INDEX IF NOT EXISTS idx_tables_location ON tables(location);

-- Create index on capacity for filtering
CREATE INDEX IF NOT EXISTS idx_tables_capacity ON tables(capacity);


