-- Drop reservations table and indexes
DROP INDEX IF EXISTS idx_reservations_guest_name;
DROP INDEX IF EXISTS idx_reservations_guest_phone;
DROP INDEX IF EXISTS idx_reservations_guest_email;
DROP INDEX IF EXISTS idx_reservations_date_time_table;
DROP INDEX IF EXISTS idx_reservations_table_number;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_date;
DROP INDEX IF EXISTS idx_reservations_user_id;
DROP TABLE IF EXISTS reservations;


