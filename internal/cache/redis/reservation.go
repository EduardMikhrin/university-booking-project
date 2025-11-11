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
	reservationKeyPrefix         = "reservation:"
	userReservationsKeyPrefix    = "reservations:user:"
	reservationListKeyPrefix     = "reservations:list:"
	userReservationsCachePattern = "reservations:user:*"
	reservationListCachePattern  = "reservations:list:*"
)

// ReservationCache implements cache.ReservationCacheQ interface using Redis
type ReservationCache struct {
	client *redis.Client
}

// NewReservationCache creates a new ReservationCache instance
func NewReservationCache(client *redis.Client) cache.ReservationCacheQ {
	return &ReservationCache{client: client}
}

// SetReservation caches a single reservation
func (c *ReservationCache) SetReservation(ctx context.Context, reservationID uuid.UUID, reservation *types.Reservation, expiration time.Duration) error {
	key := reservationKeyPrefix + reservationID.String()
	data, err := json.Marshal(reservation)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// GetReservation retrieves cached reservation
func (c *ReservationCache) GetReservation(ctx context.Context, reservationID uuid.UUID) (*types.Reservation, error) {
	key := reservationKeyPrefix + reservationID.String()
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("reservation not found in cache")
		}
		return nil, err
	}

	var reservation types.Reservation
	if err := json.Unmarshal([]byte(val), &reservation); err != nil {
		return nil, err
	}

	return &reservation, nil
}

// SetUserReservations caches reservations for a specific user
func (c *ReservationCache) SetUserReservations(ctx context.Context, userID uuid.UUID, reservations []*types.Reservation, expiration time.Duration) error {
	key := userReservationsKeyPrefix + userID.String()
	data, err := json.Marshal(reservations)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// GetUserReservations retrieves cached user reservations
func (c *ReservationCache) GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*types.Reservation, error) {
	key := userReservationsKeyPrefix + userID.String()
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("user reservations not found in cache")
		}
		return nil, err
	}

	var reservations []*types.Reservation
	if err := json.Unmarshal([]byte(val), &reservations); err != nil {
		return nil, err
	}

	return reservations, nil
}

// SetReservationList caches filtered reservation list
func (c *ReservationCache) SetReservationList(ctx context.Context, key string, reservations []*types.Reservation, expiration time.Duration) error {
	fullKey := reservationListKeyPrefix + key
	data, err := json.Marshal(reservations)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fullKey, data, expiration).Err()
}

// GetReservationList retrieves cached reservation list
func (c *ReservationCache) GetReservationList(ctx context.Context, key string) ([]*types.Reservation, error) {
	fullKey := reservationListKeyPrefix + key
	val, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("reservation list not found in cache")
		}
		return nil, err
	}

	var reservations []*types.Reservation
	if err := json.Unmarshal([]byte(val), &reservations); err != nil {
		return nil, err
	}

	return reservations, nil
}

// DeleteReservation removes reservation from cache
func (c *ReservationCache) DeleteReservation(ctx context.Context, reservationID uuid.UUID) error {
	key := reservationKeyPrefix + reservationID.String()
	return c.client.Del(ctx, key).Err()
}

// InvalidateUserReservations invalidates cache for user's reservations
func (c *ReservationCache) InvalidateUserReservations(ctx context.Context, userID uuid.UUID) error {
	key := userReservationsKeyPrefix + userID.String()
	return c.client.Del(ctx, key).Err()
}
