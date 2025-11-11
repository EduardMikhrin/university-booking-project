package types

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	Name      string    `db:"name" json:"name"`
	Phone     *string   `db:"phone" json:"phone"`
	Photo     *string   `db:"photo" json:"photo"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

// Reservation represents a reservation in the system
type Reservation struct {
	ID              uuid.UUID `db:"id" json:"id"`
	UserID          uuid.UUID `db:"user_id" json:"userId"`
	GuestName       string    `db:"guest_name" json:"guestName"`
	GuestPhone      string    `db:"guest_phone" json:"guestPhone"`
	GuestEmail      string    `db:"guest_email" json:"guestEmail"`
	Date            time.Time `db:"date" json:"date"`
	Time            string    `db:"time" json:"time"`
	Guests          int       `db:"guests" json:"guests"`
	TableNumber     string    `db:"table_number" json:"tableNumber"`
	Status          string    `db:"status" json:"status"`
	SpecialRequests *string   `db:"special_requests" json:"specialRequests,omitempty"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time `db:"updated_at" json:"updatedAt,omitempty"`
}

// Table represents a table in the restaurant
type Table struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Number      string    `db:"number" json:"number"`
	Capacity    int       `db:"capacity" json:"capacity"`
	IsAvailable bool      `db:"is_available" json:"isAvailable"`
	Location    string    `db:"location" json:"location"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt,omitempty"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt,omitempty"`
}

// ReservationFilters represents filters for querying reservations
type ReservationFilters struct {
	Status *string
	Date   *time.Time
	Search *string
}

// TableAvailabilityFilters represents filters for querying available tables
type TableAvailabilityFilters struct {
	Date   *time.Time
	Time   *string
	Guests *int
}

