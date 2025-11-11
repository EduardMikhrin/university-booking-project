package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/data"
	"github.com/EduardMikhrin/university-booking-project/internal/types"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ReservationQ implements data.ReservationQ interface
type ReservationQ struct {
	db *sqlx.DB
}

// NewReservationQ creates a new ReservationQ instance
func NewReservationQ(db *sqlx.DB) data.ReservationQ {
	return &ReservationQ{db: db}
}

// Create creates a new reservation
func (q *ReservationQ) Create(ctx context.Context, reservation *types.Reservation) error {
	query := `
		INSERT INTO reservations (
			id, user_id, guest_name, guest_phone, guest_email,
			date, time, guests, table_number, status, special_requests, created_at
		)
		VALUES (
			:id, :user_id, :guest_name, :guest_phone, :guest_email,
			:date, :time, :guests, :table_number, :status, :special_requests, :created_at
		)
	`

	if reservation.ID == uuid.Nil {
		reservation.ID = uuid.New()
	}

	if reservation.Status == "" {
		reservation.Status = "pending"
	}

	if reservation.CreatedAt.IsZero() {
		reservation.CreatedAt = time.Now()
	}

	_, err := q.db.NamedExecContext(ctx, query, reservation)
	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a reservation by ID
func (q *ReservationQ) GetByID(ctx context.Context, id uuid.UUID) (*types.Reservation, error) {
	query := `
		SELECT id, user_id, guest_name, guest_phone, guest_email,
		       date, time, guests, table_number, status, special_requests,
		       created_at, updated_at
		FROM reservations
		WHERE id = $1
	`

	var reservation types.Reservation
	err := q.db.GetContext(ctx, &reservation, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reservation not found")
		}
		return nil, err
	}

	return &reservation, nil
}

// GetAll retrieves all reservations with optional filters
func (q *ReservationQ) GetAll(ctx context.Context, userID *uuid.UUID, filters *types.ReservationFilters) ([]*types.Reservation, error) {
	query := `
		SELECT id, user_id, guest_name, guest_phone, guest_email,
		       date, time, guests, table_number, status, special_requests,
		       created_at, updated_at
		FROM reservations
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	// Filter by user ID if provided (for regular users)
	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *userID)
		argPos++
	}

	// Apply filters
	if filters != nil {
		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argPos)
			args = append(args, *filters.Status)
			argPos++
		}

		if filters.Date != nil {
			query += fmt.Sprintf(" AND date = $%d::date", argPos)
			args = append(args, filters.Date.Format("2006-01-02"))
			argPos++
		}

		if filters.Search != nil && *filters.Search != "" {
			searchTerm := "%" + *filters.Search + "%"
			query += fmt.Sprintf(" AND (guest_name ILIKE $%d OR guest_phone ILIKE $%d OR guest_email ILIKE $%d)",
				argPos, argPos, argPos)
			args = append(args, searchTerm)
			argPos++
		}
	}

	query += " ORDER BY date DESC, time DESC"

	var reservations []*types.Reservation
	err := q.db.SelectContext(ctx, &reservations, query, args...)
	if err != nil {
		return nil, err
	}

	return reservations, nil
}

// GetByUserID retrieves all reservations for a specific user
func (q *ReservationQ) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*types.Reservation, error) {
	query := `
		SELECT id, user_id, guest_name, guest_phone, guest_email,
		       date, time, guests, table_number, status, special_requests,
		       created_at, updated_at
		FROM reservations
		WHERE user_id = $1
		ORDER BY date DESC, time DESC
	`

	var reservations []*types.Reservation
	err := q.db.SelectContext(ctx, &reservations, query, userID)
	if err != nil {
		return nil, err
	}

	return reservations, nil
}

// Update updates a reservation's information
func (q *ReservationQ) Update(ctx context.Context, id uuid.UUID, reservation *types.Reservation) error {
	setParts := []string{}
	args := []interface{}{}
	argPos := 1

	if reservation.GuestName != "" {
		setParts = append(setParts, fmt.Sprintf("guest_name = $%d", argPos))
		args = append(args, reservation.GuestName)
		argPos++
	}

	if reservation.GuestPhone != "" {
		setParts = append(setParts, fmt.Sprintf("guest_phone = $%d", argPos))
		args = append(args, reservation.GuestPhone)
		argPos++
	}

	if reservation.GuestEmail != "" {
		setParts = append(setParts, fmt.Sprintf("guest_email = $%d", argPos))
		args = append(args, reservation.GuestEmail)
		argPos++
	}

	if !reservation.Date.IsZero() {
		setParts = append(setParts, fmt.Sprintf("date = $%d", argPos))
		args = append(args, reservation.Date)
		argPos++
	}

	if reservation.Time != "" {
		setParts = append(setParts, fmt.Sprintf("time = $%d", argPos))
		args = append(args, reservation.Time)
		argPos++
	}

	if reservation.Guests > 0 {
		setParts = append(setParts, fmt.Sprintf("guests = $%d", argPos))
		args = append(args, reservation.Guests)
		argPos++
	}

	if reservation.TableNumber != "" {
		setParts = append(setParts, fmt.Sprintf("table_number = $%d", argPos))
		args = append(args, reservation.TableNumber)
		argPos++
	}

	if reservation.SpecialRequests != nil {
		setParts = append(setParts, fmt.Sprintf("special_requests = $%d", argPos))
		args = append(args, *reservation.SpecialRequests)
		argPos++
	}

	if len(setParts) == 0 {
		return errors.New("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE reservations
		SET %s, updated_at = NOW()
		WHERE id = $%d
	`, strings.Join(setParts, ", "), argPos)

	args = append(args, id)

	result, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("reservation not found")
	}

	return nil
}

// UpdateStatus updates only the status of a reservation
func (q *ReservationQ) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE reservations
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := q.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("reservation not found")
	}

	return nil
}

// Delete deletes a reservation by ID
func (q *ReservationQ) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reservations WHERE id = $1`

	result, err := q.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("reservation not found")
	}

	return nil
}

// CheckTableAvailability checks if a table is available at a specific date and time
func (q *ReservationQ) CheckTableAvailability(ctx context.Context, tableNumber string, date string, time string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM reservations
		WHERE table_number = $1
		  AND date = $2::date
		  AND time = $3::time
		  AND status IN ('pending', 'confirmed')
	`

	var count int
	err := q.db.GetContext(ctx, &count, query, tableNumber, date, time)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
