-- Add foreign key constraint from reservations.table_number to tables.number
ALTER TABLE reservations 
ADD CONSTRAINT fk_reservations_table_number 
FOREIGN KEY (table_number) REFERENCES tables(number) ON DELETE RESTRICT;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at for tables table
CREATE TRIGGER update_tables_updated_at 
BEFORE UPDATE ON tables
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create trigger to automatically update updated_at for reservations table
CREATE TRIGGER update_reservations_updated_at 
BEFORE UPDATE ON reservations
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


