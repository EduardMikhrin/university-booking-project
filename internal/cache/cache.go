package cache

// CacheQ defines methods for cache-related operations
type CacheQ interface {
	// TokenCache methods for JWT token management
	TokenCache() TokenCacheQ

	// UserCache methods for user data caching
	UserCache() UserCacheQ

	// TableCache methods for table data caching
	TableCache() TableCacheQ

	// ReservationCache methods for reservation data caching
	ReservationCache() ReservationCacheQ

	// ReportCache methods for report/statistics caching
	ReportCache() ReportCacheQ
}
