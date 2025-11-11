package cache

import (
	"context"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
)

// ReportCacheQ defines methods for report/statistics caching
type ReportCacheQ interface {
	// SetMonthlyStatsList caches list of monthly statistics
	SetMonthlyStatsList(ctx context.Context, stats []*types.MonthlyStats, expiration time.Duration) error

	// GetMonthlyStatsList retrieves cached monthly statistics list
	GetMonthlyStatsList(ctx context.Context) ([]*types.MonthlyStats, error)

	// SetDetailedMonthlyStats caches detailed monthly statistics
	SetDetailedMonthlyStats(ctx context.Context, month string, stats *types.DetailedMonthlyStats, expiration time.Duration) error

	// GetDetailedMonthlyStats retrieves cached detailed monthly statistics
	GetDetailedMonthlyStats(ctx context.Context, month string) (*types.DetailedMonthlyStats, error)

	// InvalidateMonthlyStats invalidates monthly statistics cache
	InvalidateMonthlyStats(ctx context.Context, month string) error

	// InvalidateAllStats invalidates all statistics cache
	InvalidateAllStats(ctx context.Context) error
}

