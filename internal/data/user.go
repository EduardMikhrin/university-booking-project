package data

import (
	"context"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

// UserQ defines methods for user-related database operations
type UserQ interface {
	// Create creates a new user
	Create(ctx context.Context, user *types.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*types.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*types.User, error)

	// Update updates a user's information
	Update(ctx context.Context, id uuid.UUID, user *types.User) error
}
