package data

import (
	"context"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

// ReservationQ defines methods for reservation-related database operations
type ReservationQ interface {
	// Create creates a new reservation
	Create(ctx context.Context, reservation *types.Reservation) error

	// GetByID retrieves a reservation by ID
	GetByID(ctx context.Context, id uuid.UUID) (*types.Reservation, error)

	// GetAll retrieves all reservations with optional filters
	// Admin sees all reservations, users see only their own
	GetAll(ctx context.Context, userID *uuid.UUID, filters *types.ReservationFilters) ([]*types.Reservation, error)

	// GetByUserID retrieves all reservations for a specific user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*types.Reservation, error)

	// Update updates a reservation's information
	Update(ctx context.Context, id uuid.UUID, reservation *types.Reservation) error

	// UpdateStatus updates only the status of a reservation
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error

	// Delete deletes a reservation by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// CheckTableAvailability checks if a table is available at a specific date and time
	CheckTableAvailability(ctx context.Context, tableNumber string, date string, time string) (bool, error)
}
