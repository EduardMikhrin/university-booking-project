package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/cache"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	tokenKeyPrefix      = "token:"
	tokenBlacklistPrefix = "token:blacklist:"
)

// TokenCache implements cache.TokenCacheQ interface using Redis
type TokenCache struct {
	client *redis.Client
}

// NewTokenCache creates a new TokenCache instance
func NewTokenCache(client *redis.Client) cache.TokenCacheQ {
	return &TokenCache{client: client}
}

// SetToken stores a JWT token with user ID and expiration
func (c *TokenCache) SetToken(ctx context.Context, token string, userID uuid.UUID, expiration time.Duration) error {
	key := tokenKeyPrefix + token
	return c.client.Set(ctx, key, userID.String(), expiration).Err()
}

// GetUserIDByToken retrieves user ID by token
func (c *TokenCache) GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := tokenKeyPrefix + token
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, errors.New("token not found")
		}
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in cache: %w", err)
	}

	return userID, nil
}

// DeleteToken removes a token from cache (logout/blacklist)
func (c *TokenCache) DeleteToken(ctx context.Context, token string) error {
	key := tokenKeyPrefix + token
	return c.client.Del(ctx, key).Err()
}

// TokenExists checks if token exists and is valid
func (c *TokenCache) TokenExists(ctx context.Context, token string) (bool, error) {
	key := tokenKeyPrefix + token
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SetTokenBlacklist adds token to blacklist (for logout)
func (c *TokenCache) SetTokenBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	key := tokenBlacklistPrefix + token
	return c.client.Set(ctx, key, "1", expiration).Err()
}

// IsTokenBlacklisted checks if token is blacklisted
func (c *TokenCache) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := tokenBlacklistPrefix + token
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

