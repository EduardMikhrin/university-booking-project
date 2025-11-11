package cache

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TokenCacheQ defines methods for JWT token caching
type TokenCacheQ interface {
	// SetToken stores a JWT token with user ID and expiration
	SetToken(ctx context.Context, token string, userID uuid.UUID, expiration time.Duration) error

	// GetUserIDByToken retrieves user ID by token
	GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error)

	// DeleteToken removes a token from cache (logout/blacklist)
	DeleteToken(ctx context.Context, token string) error

	// TokenExists checks if token exists and is valid
	TokenExists(ctx context.Context, token string) (bool, error)

	// SetTokenBlacklist adds token to blacklist (for logout)
	SetTokenBlacklist(ctx context.Context, token string, expiration time.Duration) error

	// IsTokenBlacklisted checks if token is blacklisted
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

