# Database Migrations

This directory contains SQL migrations for the booking system using go-migrate.

## Migration Files

### 000001_create_users_table
Creates the `users` table for authentication and user management.
- Fields: id, email, password, name, phone, role, created_at
- Indexes: email, role

### 000002_create_tables_table
Creates the `tables` table for restaurant table management.
- Fields: id, number, capacity, is_available, location, created_at, updated_at
- Indexes: number, is_available, location, capacity

### 000003_create_reservations_table
Creates the `reservations` table for booking management.
- Fields: id, user_id, guest_name, guest_phone, guest_email, date, time, guests, table_number, status, special_requests, created_at, updated_at
- Foreign Key: user_id → users(id)
- Indexes: user_id, date, status, table_number, date+time+table_number (composite), guest_email, guest_phone, guest_name

### 000004_add_foreign_keys_and_triggers
Adds foreign key constraint and automatic timestamp updates.
- Foreign Key: table_number → tables(number)
- Triggers: Automatically update `updated_at` timestamp on table updates

## Usage

### Run migrations up:
```bash
migrate -path ./migrations -database "postgres://user:password@localhost/dbname?sslmode=disable" up
```

### Run migrations down:
```bash
migrate -path ./migrations -database "postgres://user:password@localhost/dbname?sslmode=disable" down
```

### Create a new migration:
```bash
migrate create -ext sql -dir ./migrations -seq migration_name
```

## Database Requirements

- PostgreSQL (uses PostgreSQL-specific features like `gen_random_uuid()`, triggers, and functions)
- For other databases, you may need to modify:
  - UUID generation (use database-specific functions)
  - Trigger syntax
  - Function syntax

## Notes

- All timestamps use `TIMESTAMP WITH TIME ZONE`
- UUIDs are used as primary keys
- Foreign key constraints ensure referential integrity
- Indexes are created for common query patterns (filtering, searching, availability checks)
- The `updated_at` field is automatically maintained by database triggers


