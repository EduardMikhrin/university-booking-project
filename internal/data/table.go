package data

import (
	"context"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

// TableQ defines methods for table-related database operations
type TableQ interface {
	// Create creates a new table
	Create(ctx context.Context, table *types.Table) error

	// GetByID retrieves a table by ID
	GetByID(ctx context.Context, id uuid.UUID) (*types.Table, error)

	// GetByNumber retrieves a table by table number
	GetByNumber(ctx context.Context, number string) (*types.Table, error)

	// GetAll retrieves all tables
	GetAll(ctx context.Context) ([]*types.Table, error)

	// GetAvailable retrieves available tables with optional filters
	GetAvailable(ctx context.Context, filters *types.TableAvailabilityFilters) ([]*types.Table, error)

	// UpdateAvailability updates the availability status of a table
	UpdateAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error

	// Update updates a table's information
	Update(ctx context.Context, id uuid.UUID, table *types.Table) error
}
