package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/cache"
	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	userKeyPrefix      = "user:"
	userEmailKeyPrefix = "user:email:"
)

// UserCache implements cache.UserCacheQ interface using Redis
type UserCache struct {
	client *redis.Client
}

// NewUserCache creates a new UserCache instance
func NewUserCache(client *redis.Client) cache.UserCacheQ {
	return &UserCache{client: client}
}

// SetUser caches user data
func (c *UserCache) SetUser(ctx context.Context, userID uuid.UUID, user *types.User, expiration time.Duration) error {
	key := userKeyPrefix + userID.String()
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// GetUser retrieves cached user data
func (c *UserCache) GetUser(ctx context.Context, userID uuid.UUID) (*types.User, error) {
	key := userKeyPrefix + userID.String()
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("user not found in cache")
		}
		return nil, err
	}

	var user types.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser removes user data from cache
func (c *UserCache) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	key := userKeyPrefix + userID.String()
	return c.client.Del(ctx, key).Err()
}

// SetUserByEmail caches user data by email
func (c *UserCache) SetUserByEmail(ctx context.Context, email string, user *types.User, expiration time.Duration) error {
	key := userEmailKeyPrefix + email
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// GetUserByEmail retrieves cached user data by email
func (c *UserCache) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	key := userEmailKeyPrefix + email
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("user not found in cache")
		}
		return nil, err
	}

	var user types.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

