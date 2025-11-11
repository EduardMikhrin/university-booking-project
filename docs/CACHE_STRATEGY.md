# Redis Cache Strategy

This document outlines what data can be cached with Redis in the booking system and the caching strategy for each data type.

## 1. JWT Tokens (High Priority)

### What to Cache:
- **Token → User ID mapping**: Store JWT token as key, user ID as value
- **Token blacklist**: Store invalidated tokens (for logout)
- **Token expiration**: Use Redis TTL matching JWT expiration

### Benefits:
- Fast token validation without database lookup
- Immediate token invalidation on logout
- Reduced database load for authentication

### Cache Keys:
- `token:{token}` → user ID (TTL: token expiration)
- `token:blacklist:{token}` → "1" (TTL: token expiration)

### TTL Strategy:
- Match JWT token expiration time
- Default: 24 hours for access tokens

---

## 2. User Data (Medium Priority)

### What to Cache:
- **User profile data**: User ID → User object
- **User by email**: Email → User object (for login lookup)

### Benefits:
- Faster user profile retrieval
- Reduced database queries for frequent lookups
- Better performance for `/auth/me` endpoint

### Cache Keys:
- `user:{userID}` → User JSON
- `user:email:{email}` → User JSON

### TTL Strategy:
- **User profile**: 1 hour (invalidate on update)
- **User by email**: 30 minutes (used for login)

### Invalidation:
- Invalidate on user profile update
- Invalidate on user deletion

---

## 3. Table Data (High Priority)

### What to Cache:
- **All tables list**: Complete list of tables (rarely changes)
- **Table by ID**: Individual table data
- **Table by number**: Table lookup by table number
- **Available tables**: Available tables for specific date/time/guests combinations

### Benefits:
- Tables rarely change, perfect for caching
- Fast availability checks
- Reduced database load for frequent queries

### Cache Keys:
- `tables:all` → List of tables JSON
- `table:{tableID}` → Table JSON
- `table:number:{number}` → Table JSON
- `tables:available:{date}:{time}:{guests}` → List of available tables

### TTL Strategy:
- **All tables**: 1 hour (invalidate on table update)
- **Individual table**: 30 minutes
- **Available tables**: 5-15 minutes (depends on booking frequency)

### Invalidation:
- Invalidate all table cache on table create/update/delete
- Invalidate available tables cache on reservation create/update/delete

---

## 4. Reservation Data (Medium Priority)

### What to Cache:
- **Reservation by ID**: Individual reservation data
- **User reservations**: All reservations for a specific user
- **Filtered reservation lists**: Cached search/filter results

### Benefits:
- Faster reservation retrieval
- Reduced database load for user's reservation list
- Better performance for admin views

### Cache Keys:
- `reservation:{reservationID}` → Reservation JSON
- `reservations:user:{userID}` → List of user reservations
- `reservations:filter:{hash}` → Filtered reservation list (hash of filter parameters)

### TTL Strategy:
- **Individual reservation**: 15 minutes
- **User reservations**: 5 minutes (frequently updated)
- **Filtered lists**: 2-5 minutes (depends on update frequency)

### Invalidation:
- Invalidate reservation cache on create/update/delete
- Invalidate user reservations cache on any reservation change for that user
- Invalidate filtered lists on any reservation change

---

## 5. Reports/Statistics (High Priority)

### What to Cache:
- **Monthly statistics list**: List of all months with stats
- **Detailed monthly statistics**: Complete stats for a specific month (popular tables, peak hours, revenue)

### Benefits:
- Reports are expensive to calculate (aggregation queries)
- Statistics don't change frequently
- Significant performance improvement for admin dashboard

### Cache Keys:
- `reports:monthly:list` → List of monthly stats
- `reports:monthly:{month}` → Detailed monthly stats (e.g., `reports:monthly:2025-10`)

### TTL Strategy:
- **Monthly list**: 1 hour
- **Detailed monthly stats**: 30 minutes (for current month), 24 hours (for past months)

### Invalidation:
- Invalidate monthly stats when new reservations are completed/cancelled
- Invalidate detailed stats for specific month on reservation status changes
- Consider cache warming for past months (they rarely change)

---

## 6. Rate Limiting (Optional)

### What to Cache:
- **API rate limits**: Request count per user/IP
- **Login attempts**: Failed login attempts tracking

### Benefits:
- Prevent abuse
- DDoS protection
- Account security

### Cache Keys:
- `ratelimit:{userID}:{endpoint}` → Request count
- `login:attempts:{email}` → Attempt count

### TTL Strategy:
- **Rate limits**: Sliding window (1 minute, 1 hour, etc.)
- **Login attempts**: 15 minutes (reset on successful login)

---

## Cache Key Naming Convention

```
{entity}:{identifier}:{optional_subkey}
```

Examples:
- `token:abc123def456`
- `user:550e8400-e29b-41d4-a716-446655440000`
- `tables:available:2025-11-05:19:00:4`
- `reservations:user:550e8400-e29b-41d4-a716-446655440000`
- `reports:monthly:2025-10`

---

## Cache Invalidation Strategy

### On Create:
- Invalidate related lists (e.g., create reservation → invalidate user reservations, available tables)

### On Update:
- Invalidate specific entity cache
- Invalidate related lists
- Invalidate dependent caches (e.g., update reservation → invalidate reports)

### On Delete:
- Delete specific entity cache
- Invalidate related lists
- Invalidate dependent caches

---

## Recommended Cache Priorities

1. **JWT Tokens** (Highest) - Critical for performance and security
2. **Reports/Statistics** (High) - Expensive queries, rarely change
3. **Table Data** (High) - Frequently accessed, rarely changes
4. **User Data** (Medium) - Frequently accessed, moderate change frequency
5. **Reservation Data** (Medium) - Frequently accessed, but changes often
6. **Rate Limiting** (Optional) - Security feature

---

## Implementation Notes

- Use Redis TTL for automatic expiration
- Implement cache-aside pattern (check cache first, then database)
- Use Redis pipelines for batch operations
- Consider Redis transactions for atomic operations
- Monitor cache hit rates and adjust TTLs accordingly
- Implement cache warming for frequently accessed data
- Use Redis pub/sub for cache invalidation across instances (if using multiple servers)

