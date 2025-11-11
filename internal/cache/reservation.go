package cache

import (
	"context"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

// ReservationCacheQ defines methods for reservation data caching
type ReservationCacheQ interface {
	// SetReservation caches a single reservation
	SetReservation(ctx context.Context, reservationID uuid.UUID, reservation *types.Reservation, expiration time.Duration) error

	// GetReservation retrieves cached reservation
	GetReservation(ctx context.Context, reservationID uuid.UUID) (*types.Reservation, error)

	// SetUserReservations caches reservations for a specific user
	SetUserReservations(ctx context.Context, userID uuid.UUID, reservations []*types.Reservation, expiration time.Duration) error

	// GetUserReservations retrieves cached user reservations
	GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*types.Reservation, error)

	// SetReservationList caches filtered reservation list
	SetReservationList(ctx context.Context, key string, reservations []*types.Reservation, expiration time.Duration) error

	// GetReservationList retrieves cached reservation list
	GetReservationList(ctx context.Context, key string) ([]*types.Reservation, error)

	// DeleteReservation removes reservation from cache
	DeleteReservation(ctx context.Context, reservationID uuid.UUID) error

	// InvalidateUserReservations invalidates cache for user's reservations
	InvalidateUserReservations(ctx context.Context, userID uuid.UUID) error
}

