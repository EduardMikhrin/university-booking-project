-- Drop triggers
DROP TRIGGER IF EXISTS update_reservations_updated_at ON reservations;
DROP TRIGGER IF EXISTS update_tables_updated_at ON tables;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop foreign key constraint
ALTER TABLE reservations 
DROP CONSTRAINT IF EXISTS fk_reservations_table_number;


