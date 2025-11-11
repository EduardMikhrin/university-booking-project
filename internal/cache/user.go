package cache

import (
	"context"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

// UserCacheQ defines methods for user data caching
type UserCacheQ interface {
	// SetUser caches user data
	SetUser(ctx context.Context, userID uuid.UUID, user *types.User, expiration time.Duration) error

	// GetUser retrieves cached user data
	GetUser(ctx context.Context, userID uuid.UUID) (*types.User, error)

	// DeleteUser removes user data from cache
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	// SetUserByEmail caches user data by email
	SetUserByEmail(ctx context.Context, email string, user *types.User, expiration time.Duration) error

	// GetUserByEmail retrieves cached user data by email
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
}

