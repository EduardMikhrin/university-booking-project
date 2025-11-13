-- Create reservations table
CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    guest_name VARCHAR(255) NOT NULL,
    guest_phone VARCHAR(50) NOT NULL,
    guest_email VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    time TIME NOT NULL,
    guests INTEGER NOT NULL CHECK (guests > 0),
    table_number VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),
    special_requests TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on user_id for faster user-specific queries
CREATE INDEX IF NOT EXISTS idx_reservations_user_id ON reservations(user_id);

-- Create index on date for filtering by date
CREATE INDEX IF NOT EXISTS idx_reservations_date ON reservations(date);

-- Create index on status for filtering by status
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);

-- Create index on table_number for checking availability
CREATE INDEX IF NOT EXISTS idx_reservations_table_number ON reservations(table_number);

-- Create composite index on date, time, and table_number for availability checks
CREATE INDEX IF NOT EXISTS idx_reservations_date_time_table ON reservations(date, time, table_number);

-- Create index on guest_email for searching
CREATE INDEX IF NOT EXISTS idx_reservations_guest_email ON reservations(guest_email);

-- Create index on guest_phone for searching
CREATE INDEX IF NOT EXISTS idx_reservations_guest_phone ON reservations(guest_phone);

-- Create index on guest_name for searching (using text pattern)
CREATE INDEX IF NOT EXISTS idx_reservations_guest_name ON reservations(guest_name);


