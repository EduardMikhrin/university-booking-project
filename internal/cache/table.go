package cache

import (
	"context"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

// TableCacheQ defines methods for table data caching
type TableCacheQ interface {
	// SetTable caches a single table
	SetTable(ctx context.Context, tableID uuid.UUID, table *types.Table, expiration time.Duration) error

	// GetTable retrieves cached table data
	GetTable(ctx context.Context, tableID uuid.UUID) (*types.Table, error)

	// SetTableByNumber caches table by table number
	SetTableByNumber(ctx context.Context, number string, table *types.Table, expiration time.Duration) error

	// GetTableByNumber retrieves cached table by number
	GetTableByNumber(ctx context.Context, number string) (*types.Table, error)

	// SetAllTables caches list of all tables
	SetAllTables(ctx context.Context, tables []*types.Table, expiration time.Duration) error

	// GetAllTables retrieves cached list of all tables
	GetAllTables(ctx context.Context) ([]*types.Table, error)

	// SetAvailableTables caches available tables for a specific date/time
	SetAvailableTables(ctx context.Context, date string, time string, guests int, tables []*types.Table, expiration time.Duration) error

	// GetAvailableTables retrieves cached available tables
	GetAvailableTables(ctx context.Context, date string, time string, guests int) ([]*types.Table, error)

	// InvalidateTableCache invalidates all table-related cache
	InvalidateTableCache(ctx context.Context) error
}

