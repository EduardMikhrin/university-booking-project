package data

import (
	"context"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
)

// ReportsQ defines methods for reports-related database operations
type ReportsQ interface {
	// GetMonthlyStatsList retrieves a list of all months with available statistics
	GetMonthlyStatsList(ctx context.Context) ([]*types.MonthlyStats, error)

	// GetDetailedMonthlyStats retrieves detailed statistics for a specific month
	GetDetailedMonthlyStats(ctx context.Context, month string) (*types.DetailedMonthlyStats, error)
}
