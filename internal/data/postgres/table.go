package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/data"
	"github.com/EduardMikhrin/university-booking-project/internal/types"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TableQ implements data.TableQ interface
type TableQ struct {
	db *sqlx.DB
}

// NewTableQ creates a new TableQ instance
func NewTableQ(db *sqlx.DB) data.TableQ {
	return &TableQ{db: db}
}

// Create creates a new table
func (q *TableQ) Create(ctx context.Context, table *types.Table) error {
	query := `
		INSERT INTO tables (id, number, capacity, is_available, location, created_at, updated_at)
		VALUES (:id, :number, :capacity, :is_available, :location, :created_at, :updated_at)
	`

	if table.ID == uuid.Nil {
		table.ID = uuid.New()
	}

	if table.CreatedAt.IsZero() {
		table.CreatedAt = time.Now()
	}

	if table.UpdatedAt.IsZero() {
		table.UpdatedAt = time.Now()
	}

	_, err := q.db.NamedExecContext(ctx, query, table)
	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a table by ID
func (q *TableQ) GetByID(ctx context.Context, id uuid.UUID) (*types.Table, error) {
	query := `
		SELECT id, number, capacity, is_available, location, created_at, updated_at
		FROM tables
		WHERE id = $1
	`

	var table types.Table
	err := q.db.GetContext(ctx, &table, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("table not found")
		}
		return nil, err
	}

	return &table, nil
}

// GetByNumber retrieves a table by table number
func (q *TableQ) GetByNumber(ctx context.Context, number string) (*types.Table, error) {
	query := `
		SELECT id, number, capacity, is_available, location, created_at, updated_at
		FROM tables
		WHERE number = $1
	`

	var table types.Table
	err := q.db.GetContext(ctx, &table, query, number)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("table not found")
		}
		return nil, err
	}

	return &table, nil
}

// GetAll retrieves all tables
func (q *TableQ) GetAll(ctx context.Context) ([]*types.Table, error) {
	query := `
		SELECT id, number, capacity, is_available, location, created_at, updated_at
		FROM tables
		ORDER BY number
	`

	var tables []*types.Table
	err := q.db.SelectContext(ctx, &tables, query)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

// GetAvailable retrieves available tables with optional filters
func (q *TableQ) GetAvailable(ctx context.Context, filters *types.TableAvailabilityFilters) ([]*types.Table, error) {
	query := `
		SELECT DISTINCT t.id, t.number, t.capacity, t.is_available, t.location, t.created_at, t.updated_at
		FROM tables t
		WHERE t.is_available = true
	`

	args := []interface{}{}
	argPos := 1

	// Filter by minimum capacity if provided
	if filters != nil && filters.Guests != nil {
		query += fmt.Sprintf(" AND t.capacity >= $%d", argPos)
		args = append(args, *filters.Guests)
		argPos++
	}

	// Filter by date and time if provided (check for conflicting reservations)
	if filters != nil && filters.Date != nil && filters.Time != nil {
		query += fmt.Sprintf(`
			AND t.number NOT IN (
				SELECT r.table_number
				FROM reservations r
				WHERE r.table_number = t.number
				  AND r.date = $%d::date
				  AND r.time = $%d::time
				  AND r.status IN ('pending', 'confirmed')
			)
		`, argPos, argPos+1)
		args = append(args, filters.Date.Format("2006-01-02"), *filters.Time)
		argPos += 2
	} else if filters != nil && filters.Date != nil {
		// Only date filter - exclude tables with any reservation on that date
		query += fmt.Sprintf(`
			AND t.number NOT IN (
				SELECT r.table_number
				FROM reservations r
				WHERE r.table_number = t.number
				  AND r.date = $%d::date
				  AND r.status IN ('pending', 'confirmed')
			)
		`, argPos)
		args = append(args, filters.Date.Format("2006-01-02"))
		argPos++
	}

	query += " ORDER BY t.number"

	var tables []*types.Table
	err := q.db.SelectContext(ctx, &tables, query, args...)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

// UpdateAvailability updates the availability status of a table
func (q *TableQ) UpdateAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error {
	query := `
		UPDATE tables
		SET is_available = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := q.db.ExecContext(ctx, query, isAvailable, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("table not found")
	}

	return nil
}

// Update updates a table's information
func (q *TableQ) Update(ctx context.Context, id uuid.UUID, table *types.Table) error {
	query := `
		UPDATE tables
		SET number = :number, capacity = :capacity, is_available = :is_available,
		    location = :location, updated_at = NOW()
		WHERE id = :id
	`

	table.ID = id
	result, err := q.db.NamedExecContext(ctx, query, table)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("table not found")
	}

	return nil
}
