package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/EduardMikhrin/university-booking-project/internal/data"
	"github.com/EduardMikhrin/university-booking-project/internal/types"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserQ implements data.UserQ interface
type UserQ struct {
	db *sqlx.DB
}

// NewUserQ creates a new UserQ instance
func NewUserQ(db *sqlx.DB) data.UserQ {
	return &UserQ{db: db}
}

// Create creates a new user
func (q *UserQ) Create(ctx context.Context, user *types.User) error {
	query := `
		INSERT INTO users (id, email, password, name, phone, photo, role, created_at)
		VALUES (:id, :email, :password, :name, :phone, :photo, :role, :created_at)
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Set default photo if not provided
	if user.Photo == nil || *user.Photo == "" {
		defaultPhoto := types.DefaultUserPhoto
		user.Photo = &defaultPhoto
	}

	_, err := q.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a user by ID
func (q *UserQ) GetByID(ctx context.Context, id uuid.UUID) (*types.User, error) {
	query := `
		SELECT id, email, password, name, phone, photo, role, created_at
		FROM users
		WHERE id = $1
	`

	var user types.User
	err := q.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Set default photo if not set
	if user.Photo == nil || *user.Photo == "" {
		defaultPhoto := types.DefaultUserPhoto
		user.Photo = &defaultPhoto
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (q *UserQ) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	query := `
		SELECT id, email, password, name, phone, photo, role, created_at
		FROM users
		WHERE email = $1
	`

	var user types.User
	err := q.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Set default photo if not set
	if user.Photo == nil || *user.Photo == "" {
		defaultPhoto := types.DefaultUserPhoto
		user.Photo = &defaultPhoto
	}

	return &user, nil
}

// Update updates a user's information
func (q *UserQ) Update(ctx context.Context, id uuid.UUID, user *types.User) error {
	query := `
		UPDATE users
		SET email = :email, name = :name, phone = :phone, photo = :photo
		WHERE id = :id
	`

	user.ID = id
	result, err := q.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
