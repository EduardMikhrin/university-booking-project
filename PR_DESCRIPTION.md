# Database and Cache Implementation

## Overview
This PR implements the complete database layer with PostgreSQL and Redis caching mechanism for the booking system. It includes database migrations, data access interfaces, PostgreSQL implementations, and Redis cache layer for improved performance.

## 🗄️ Database Layer

### Migrations
- **000001_create_users_table**: Users table with authentication fields
- **000002_create_tables_table**: Restaurant tables management
- **000003_create_reservations_table**: Booking reservations with foreign keys
- **000004_add_foreign_keys_and_triggers**: Foreign key constraints and auto-update triggers
- **000005_add_photo_to_users**: User profile photo support with default avatar

### Data Access Layer
- **Interfaces** (`internal/data/`):
  - `UserQ`: User CRUD operations
  - `ReservationQ`: Reservation management with filtering
  - `TableQ`: Table availability and management
  - `ReportsQ`: Monthly statistics and analytics
  - `MasterQ`: Master interface combining all query interfaces

- **PostgreSQL Implementation** (`internal/data/postgres/`):
  - Full implementation of all data interfaces
  - Parameterized queries for SQL injection prevention
  - Context support for cancellation/timeouts
  - Proper error handling with meaningful messages
  - Date/time handling with PostgreSQL types

### Type Definitions
- **Models** (`internal/types/models.go`):
  - `User`: User entity with photo support
  - `Reservation`: Booking entity with status tracking
  - `Table`: Restaurant table entity
  - Filter structs for querying

- **Constants** (`internal/types/constants.go`):
  - `DefaultUserPhoto`: Default avatar URL for users

## 🚀 Cache Layer (Redis)

### Cache Interfaces (`internal/cache/`)
- **TokenCacheQ**: JWT token management and blacklisting
- **UserCacheQ**: User data caching
- **TableCacheQ**: Table data and availability caching
- **ReservationCacheQ**: Reservation data caching
- **ReportCacheQ**: Statistics caching
- **CacheQ**: Master cache interface

### Redis Implementation (`internal/cache/redis/`)
- Complete Redis implementation for all cache interfaces
- JSON serialization for complex objects
- TTL support for automatic expiration
- Pattern-based cache invalidation
- Error handling with proper fallbacks

### Cache Strategy
- **JWT Tokens**: Token → User ID mapping with blacklist support
- **User Data**: Profile caching with email lookup
- **Table Data**: All tables and availability caching
- **Reservations**: User reservations and filtered lists
- **Reports**: Monthly statistics caching (expensive queries)

## 📦 Dependencies Added
- `github.com/jmoiron/sqlx v1.4.0` - PostgreSQL database driver
- `github.com/redis/go-redis/v9 v9.5.1` - Redis client
- `github.com/google/uuid v1.6.0` - UUID generation

## ✨ Features

### Database Features
- ✅ UUID primary keys for all entities
- ✅ Foreign key constraints for data integrity
- ✅ Automatic `updated_at` timestamp triggers
- ✅ Comprehensive indexes for performance
- ✅ Table availability checking with date/time validation
- ✅ User profile photo support with default avatar
- ✅ Reservation filtering (status, date, search)
- ✅ Monthly statistics and analytics

### Cache Features
- ✅ JWT token caching and blacklisting
- ✅ User profile caching
- ✅ Table availability caching
- ✅ Reservation list caching
- ✅ Report statistics caching
- ✅ Cache invalidation strategies
- ✅ TTL-based expiration

## 🏗️ Architecture

### Data Layer Structure
```
internal/
├── data/              # Data access interfaces
│   ├── master.go      # MasterQ interface
│   ├── user.go        # UserQ interface
│   ├── reservation.go # ReservationQ interface
│   ├── table.go       # TableQ interface
│   └── reports.go     # ReportsQ interface
├── data/postgres/     # PostgreSQL implementations
│   ├── master.go      # Master implementation
│   ├── user.go        # User operations
│   ├── reservation.go # Reservation operations
│   ├── table.go       # Table operations
│   └── reports.go     # Report operations
└── types/             # Shared types
    ├── models.go      # Entity models
    ├── reports.go     # Report types
    └── constants.go   # Constants
```

### Cache Layer Structure
```
internal/
├── cache/             # Cache interfaces
│   ├── cache.go       # CacheQ interface
│   ├── token.go       # TokenCacheQ interface
│   ├── user.go        # UserCacheQ interface
│   ├── table.go       # TableCacheQ interface
│   ├── reservation.go # ReservationCacheQ interface
│   └── report.go      # ReportCacheQ interface
└── cache/redis/       # Redis implementations
    ├── master.go      # Master cache
    ├── token.go       # Token cache
    ├── user.go        # User cache
    ├── table.go       # Table cache
    ├── reservation.go # Reservation cache
    └── report.go      # Report cache
```

## 🔧 Database Schema

### Users Table
- `id` (UUID, PK)
- `email` (VARCHAR, UNIQUE)
- `password` (VARCHAR)
- `name` (VARCHAR)
- `phone` (VARCHAR, nullable)
- `photo` (VARCHAR, nullable, default: Flaticon avatar)
- `role` (VARCHAR: 'admin' | 'user')
- `created_at` (TIMESTAMP)

### Tables Table
- `id` (UUID, PK)
- `number` (VARCHAR, UNIQUE)
- `capacity` (INTEGER)
- `is_available` (BOOLEAN)
- `location` (VARCHAR: 'main' | 'terrace' | 'private')
- `created_at`, `updated_at` (TIMESTAMP)

### Reservations Table
- `id` (UUID, PK)
- `user_id` (UUID, FK → users)
- `guest_name`, `guest_phone`, `guest_email` (VARCHAR)
- `date` (DATE)
- `time` (TIME)
- `guests` (INTEGER)
- `table_number` (VARCHAR, FK → tables.number)
- `status` (VARCHAR: 'pending' | 'confirmed' | 'cancelled' | 'completed')
- `special_requests` (TEXT, nullable)
- `created_at`, `updated_at` (TIMESTAMP)

## 📝 Migration Instructions

1. **Run migrations**:
   ```bash
   migrate -path ./migrations -database "postgres://user:password@localhost/dbname?sslmode=disable" up
   ```

2. **Initialize Redis client**:
   ```go
   import "github.com/redis/go-redis/v9"
   
   client := redis.NewClient(&redis.Options{
       Addr: "localhost:6379",
   })
   ```

3. **Initialize data layer**:
   ```go
   master := postgres.NewMaster(db)
   cache := redis.NewMaster(redisClient)
   ```

## 🧪 Testing Considerations

- All database operations use parameterized queries
- Context support for request cancellation
- Proper error handling and validation
- Cache fallback to database on cache miss
- Foreign key constraints ensure data integrity

## 📚 Documentation

- Migration files include detailed comments
- Interface methods are fully documented
- Cache strategy documented in `docs/CACHE_STRATEGY.md`
- API endpoints documented in `API_ENDPOINTS.md`

## 🔄 Breaking Changes
None - this is a new implementation.

## ✅ Checklist
- [x] Database migrations created and tested
- [x] All data interfaces implemented
- [x] PostgreSQL implementations complete
- [x] Redis cache implementations complete
- [x] Type definitions moved to `internal/types`
- [x] User photo field added with default
- [x] All imports updated
- [x] Dependencies added to `go.mod`
- [x] Code follows project structure
- [x] Error handling implemented
- [x] Context support added

## 🎯 Next Steps
- [ ] Add unit tests for data layer
- [ ] Add integration tests for cache layer
- [ ] Implement cache warming strategies
- [ ] Add database connection pooling configuration
- [ ] Add Redis connection pooling configuration
- [ ] Implement transaction support for complex operations

