# What Can Be Cached with Redis in Booking System

## Quick Reference

| Data Type | Priority | TTL | Cache Key Pattern | Use Case |
|-----------|----------|-----|-------------------|----------|
| **JWT Tokens** | 🔴 Highest | 24h | `token:{token}` | Authentication, token validation |
| **Reports/Stats** | 🟠 High | 30m-24h | `reports:monthly:{month}` | Expensive aggregation queries |
| **Table Data** | 🟠 High | 30m-1h | `tables:all`, `table:{id}` | Frequently accessed, rarely changes |
| **User Data** | 🟡 Medium | 30m-1h | `user:{id}`, `user:email:{email}` | Profile lookups, login |
| **Reservations** | 🟡 Medium | 5-15m | `reservation:{id}`, `reservations:user:{id}` | User reservations, quick lookups |
| **Rate Limiting** | ⚪ Optional | 1m-15m | `ratelimit:{user}:{endpoint}` | Security, abuse prevention |

## Detailed Breakdown

### 1. JWT Tokens 🔴
**Why cache:** Every API request needs token validation
- Token → User ID mapping
- Token blacklist for logout
- **Impact:** Reduces database queries on every request

### 2. Reports/Statistics 🟠
**Why cache:** Expensive database aggregation queries
- Monthly statistics list
- Detailed monthly stats (popular tables, peak hours, revenue)
- **Impact:** Seconds → milliseconds for report generation

### 3. Table Data 🟠
**Why cache:** Tables rarely change, frequently accessed
- Complete table list
- Table availability for date/time
- **Impact:** Fast availability checks without DB queries

### 4. User Data 🟡
**Why cache:** Frequently accessed user profiles
- User by ID (for `/auth/me`)
- User by email (for login)
- **Impact:** Faster authentication and profile retrieval

### 5. Reservation Data 🟡
**Why cache:** User frequently views their reservations
- User's reservation list
- Individual reservation lookups
- **Impact:** Faster reservation listing (but needs frequent invalidation)

### 6. Rate Limiting ⚪
**Why cache:** Security and abuse prevention
- API request counts
- Login attempt tracking
- **Impact:** Prevents abuse, DDoS protection

## Cache Hit Rate Expectations

- **JWT Tokens:** 95%+ (every request uses tokens)
- **Reports:** 80%+ (reports viewed multiple times)
- **Tables:** 90%+ (tables rarely change)
- **Users:** 70%+ (frequent profile access)
- **Reservations:** 60%+ (changes frequently)
- **Rate Limits:** 100% (always checked)

## Memory Estimation

Assuming:
- 1000 users
- 100 tables
- 10,000 reservations/month
- 1000 active tokens

**Estimated Redis Memory:**
- Tokens: ~2MB
- Tables: ~100KB
- Users: ~500KB
- Reservations: ~5MB (with TTL)
- Reports: ~1MB
- **Total: ~8-10MB** (very manageable)

## Implementation Priority

1. **Start with JWT Tokens** - Highest impact, easiest to implement
2. **Add Table Caching** - High impact, low complexity
3. **Add Report Caching** - High impact, significant performance gain
4. **Add User Caching** - Medium impact, improves UX
5. **Add Reservation Caching** - Medium impact, needs careful invalidation
6. **Add Rate Limiting** - Optional, security feature

