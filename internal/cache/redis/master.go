package redis

import (
	"github.com/EduardMikhrin/university-booking-project/internal/cache"
	"github.com/redis/go-redis/v9"
)

// Master implements the CacheQ interface using Redis
type Master struct {
	client *redis.Client

	tokenCache       cache.TokenCacheQ
	userCache        cache.UserCacheQ
	tableCache       cache.TableCacheQ
	reservationCache cache.ReservationCacheQ
	reportCache      cache.ReportCacheQ
}

// NewMaster creates a new Master cache instance
func NewMaster(client *redis.Client) cache.CacheQ {
	return &Master{
		client: client,
	}
}

// TokenCache returns the token cache interface
func (m *Master) TokenCache() cache.TokenCacheQ {
	if m.tokenCache == nil {
		m.tokenCache = NewTokenCache(m.client)
	}
	return m.tokenCache
}

// UserCache returns the user cache interface
func (m *Master) UserCache() cache.UserCacheQ {
	if m.userCache == nil {
		m.userCache = NewUserCache(m.client)
	}
	return m.userCache
}

// TableCache returns the table cache interface
func (m *Master) TableCache() cache.TableCacheQ {
	if m.tableCache == nil {
		m.tableCache = NewTableCache(m.client)
	}
	return m.tableCache
}

// ReservationCache returns the reservation cache interface
func (m *Master) ReservationCache() cache.ReservationCacheQ {
	if m.reservationCache == nil {
		m.reservationCache = NewReservationCache(m.client)
	}
	return m.reservationCache
}

// ReportCache returns the report cache interface
func (m *Master) ReportCache() cache.ReportCacheQ {
	if m.reportCache == nil {
		m.reportCache = NewReportCache(m.client)
	}
	return m.reportCache
}

