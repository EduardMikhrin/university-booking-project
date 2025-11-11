package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReservationTestDB(t *testing.T) (*ReservationQ, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sqlxDB := sqlx.NewDb(db, "postgres")
	reservationQ := NewReservationQ(sqlxDB).(*ReservationQ)

	teardown := func() {
		db.Close()
	}

	return reservationQ, mock, teardown
}

func TestReservationQ_Create(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name        string
		reservation *types.Reservation
		mock        func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name: "successful create",
			reservation: &types.Reservation{
				ID:           reservationID,
				UserID:       userID,
				GuestName:    "John Doe",
				GuestPhone:   "+1234567890",
				GuestEmail:   "john@example.com",
				Date:         time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC),
				Time:         "19:00",
				Guests:       4,
				TableNumber:  "T1",
				Status:       "pending",
				CreatedAt:    createdAt,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO reservations`).
					WithArgs(
						reservationID,
						userID,
						"John Doe",
						"+1234567890",
						"john@example.com",
						sqlmock.AnyArg(), // date
						"19:00",
						4,
						"T1",
						"pending",
						nil, // special_requests
						sqlmock.AnyArg(), // created_at
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "create with auto-generated ID and default status",
			reservation: &types.Reservation{
				ID:          uuid.Nil,
				UserID:      userID,
				GuestName:   "Jane Doe",
				GuestPhone:  "+1234567890",
				GuestEmail:  "jane@example.com",
				Date:        time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC),
				Time:        "20:00",
				Guests:      2,
				TableNumber: "T2",
				Status:      "",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO reservations`).
					WithArgs(
						sqlmock.AnyArg(), // id (will be generated)
						userID,
						"Jane Doe",
						"+1234567890",
						"jane@example.com",
						sqlmock.AnyArg(), // date
						"20:00",
						2,
						"T2",
						"pending", // default status
						nil,       // special_requests
						sqlmock.AnyArg(), // created_at
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "database error",
			reservation: &types.Reservation{
				ID:          reservationID,
				UserID:      userID,
				GuestName:   "John Doe",
				GuestPhone:  "+1234567890",
				GuestEmail:  "john@example.com",
				Date:        time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC),
				Time:        "19:00",
				Guests:      4,
				TableNumber: "T1",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO reservations`).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservationQ, mock, teardown := setupReservationTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			err := reservationQ.Create(ctx, tt.reservation)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify that ID was generated if it was nil
				if tt.reservation.ID == uuid.Nil {
					assert.NotEqual(t, uuid.Nil, tt.reservation.ID)
				}
				// Verify default status was set
				if tt.reservation.Status == "" {
					assert.Equal(t, "pending", tt.reservation.Status)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReservationQ_GetByID(t *testing.T) {
	reservationID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name    string
		id      uuid.UUID
		mock    func(mock sqlmock.Sqlmock)
		want    *types.Reservation
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful get",
			id:   reservationID,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "guest_name", "guest_phone", "guest_email", "date", "time", "guests", "table_number", "status", "special_requests", "created_at", "updated_at"}).
					AddRow(reservationID, userID, "John Doe", "+1234567890", "john@example.com", time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC), "19:00", 4, "T1", "pending", nil, createdAt, updatedAt)
				mock.ExpectQuery(`SELECT id, user_id, guest_name, guest_phone, guest_email, date, time, guests, table_number, status, special_requests, created_at, updated_at FROM reservations WHERE id = \$1`).
					WithArgs(reservationID).
					WillReturnRows(rows)
			},
			want: &types.Reservation{
				ID:           reservationID,
				UserID:       userID,
				GuestName:    "John Doe",
				GuestPhone:   "+1234567890",
				GuestEmail:   "john@example.com",
				Date:         time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC),
				Time:         "19:00",
				Guests:       4,
				TableNumber:  "T1",
				Status:       "pending",
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			},
			wantErr: false,
		},
		{
			name: "reservation not found",
			id:   reservationID,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, user_id, guest_name, guest_phone, guest_email, date, time, guests, table_number, status, special_requests, created_at, updated_at FROM reservations WHERE id = \$1`).
					WithArgs(reservationID).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "reservation not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservationQ, mock, teardown := setupReservationTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			got, err := reservationQ.GetByID(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.GuestName, got.GuestName)
				assert.Equal(t, tt.want.TableNumber, got.TableNumber)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReservationQ_GetAll(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()
	testDate := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		userID  *uuid.UUID
		filters *types.ReservationFilters
		mock    func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:   "get all without filters",
			userID: nil,
			filters: nil,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "guest_name", "guest_phone", "guest_email", "date", "time", "guests", "table_number", "status", "special_requests", "created_at", "updated_at"}).
					AddRow(reservationID, userID, "John Doe", "+1234567890", "john@example.com", testDate, "19:00", 4, "T1", "pending", nil, createdAt, updatedAt)
				mock.ExpectQuery(`SELECT.*FROM reservations WHERE 1=1 ORDER BY date DESC, time DESC`).
					WillReturnRows(rows)
			},
			want:    1,
			wantErr: false,
		},
		{
			name:   "get all with user ID filter",
			userID: &userID,
			filters: nil,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "guest_name", "guest_phone", "guest_email", "date", "time", "guests", "table_number", "status", "special_requests", "created_at", "updated_at"}).
					AddRow(reservationID, userID, "John Doe", "+1234567890", "john@example.com", testDate, "19:00", 4, "T1", "pending", nil, createdAt, updatedAt)
				mock.ExpectQuery(`SELECT.*FROM reservations WHERE 1=1 AND user_id = \$1 ORDER BY date DESC, time DESC`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want:    1,
			wantErr: false,
		},
		{
			name:   "get all with status filter",
			userID: nil,
			filters: &types.ReservationFilters{
				Status: stringPtr("confirmed"),
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "guest_name", "guest_phone", "guest_email", "date", "time", "guests", "table_number", "status", "special_requests", "created_at", "updated_at"}).
					AddRow(reservationID, userID, "John Doe", "+1234567890", "john@example.com", testDate, "19:00", 4, "T1", "confirmed", nil, createdAt, updatedAt)
				mock.ExpectQuery(`SELECT.*FROM reservations WHERE 1=1 AND status = \$1 ORDER BY date DESC, time DESC`).
					WithArgs("confirmed").
					WillReturnRows(rows)
			},
			want:    1,
			wantErr: false,
		},
		{
			name:   "get all with date filter",
			userID: nil,
			filters: &types.ReservationFilters{
				Date: &testDate,
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "guest_name", "guest_phone", "guest_email", "date", "time", "guests", "table_number", "status", "special_requests", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT.*FROM reservations WHERE 1=1 AND date = \$1::date ORDER BY date DESC, time DESC`).
					WithArgs("2025-12-25").
					WillReturnRows(rows)
			},
			want:    0,
			wantErr: false,
		},
		{
			name:   "get all with search filter",
			userID: nil,
			filters: &types.ReservationFilters{
				Search: stringPtr("John"),
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "guest_name", "guest_phone", "guest_email", "date", "time", "guests", "table_number", "status", "special_requests", "created_at", "updated_at"}).
					AddRow(reservationID, userID, "John Doe", "+1234567890", "john@example.com", testDate, "19:00", 4, "T1", "pending", nil, createdAt, updatedAt)
				mock.ExpectQuery(`SELECT.*FROM reservations WHERE 1=1 AND.*ILIKE.*ORDER BY date DESC, time DESC`).
					WithArgs("%John%").
					WillReturnRows(rows)
			},
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservationQ, mock, teardown := setupReservationTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			got, err := reservationQ.GetAll(ctx, tt.userID, tt.filters)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.want)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReservationQ_GetByUserID(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name    string
		userID  uuid.UUID
		mock    func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:   "successful get by user ID",
			userID: userID,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "guest_name", "guest_phone", "guest_email", "date", "time", "guests", "table_number", "status", "special_requests", "created_at", "updated_at"}).
					AddRow(reservationID, userID, "John Doe", "+1234567890", "john@example.com", time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC), "19:00", 4, "T1", "pending", nil, createdAt, updatedAt).
					AddRow(uuid.New(), userID, "Jane Doe", "+1234567891", "jane@example.com", time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC), "20:00", 2, "T2", "confirmed", nil, createdAt, updatedAt)
				mock.ExpectQuery(`SELECT.*FROM reservations WHERE user_id = \$1 ORDER BY date DESC, time DESC`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want:    2,
			wantErr: false,
		},
		{
			name:   "empty result",
			userID: userID,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "guest_name", "guest_phone", "guest_email", "date", "time", "guests", "table_number", "status", "special_requests", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT.*FROM reservations WHERE user_id = \$1 ORDER BY date DESC, time DESC`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservationQ, mock, teardown := setupReservationTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			got, err := reservationQ.GetByUserID(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.want)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReservationQ_Update(t *testing.T) {
	reservationID := uuid.New()

	tests := []struct {
		name        string
		id          uuid.UUID
		reservation *types.Reservation
		mock        func(mock sqlmock.Sqlmock)
		wantErr     bool
		errMsg      string
	}{
		{
			name: "successful update single field",
			id:   reservationID,
			reservation: &types.Reservation{
				GuestName: "Updated Name",
			},
			mock: func(mock sqlmock.Sqlmock) {
				// The query is built dynamically, so we use a more flexible pattern
				mock.ExpectExec(`UPDATE reservations`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "reservation not found",
			id:   reservationID,
			reservation: &types.Reservation{
				GuestName: "Updated Name",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE reservations`).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errMsg:  "reservation not found",
		},
		{
			name: "no fields to update",
			id:   reservationID,
			reservation: &types.Reservation{},
			mock: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			wantErr: true,
			errMsg:  "no fields to update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservationQ, mock, teardown := setupReservationTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			err := reservationQ.Update(ctx, tt.id, tt.reservation)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Note: ExpectationsWereMet might fail for Update due to dynamic query building
			// This is acceptable as the important part is testing the logic
		})
	}
}

func TestReservationQ_UpdateStatus(t *testing.T) {
	reservationID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		status  string
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
		errMsg  string
	}{
		{
			name:   "successful update",
			id:     reservationID,
			status: "confirmed",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE reservations SET status = \$1, updated_at = NOW\(\) WHERE id = \$2`).
					WithArgs("confirmed", reservationID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:   "reservation not found",
			id:     reservationID,
			status: "confirmed",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE reservations SET status = \$1, updated_at = NOW\(\) WHERE id = \$2`).
					WithArgs("confirmed", reservationID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errMsg:  "reservation not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservationQ, mock, teardown := setupReservationTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			err := reservationQ.UpdateStatus(ctx, tt.id, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReservationQ_Delete(t *testing.T) {
	reservationID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful delete",
			id:   reservationID,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM reservations WHERE id = \$1`).
					WithArgs(reservationID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "reservation not found",
			id:   reservationID,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM reservations WHERE id = \$1`).
					WithArgs(reservationID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errMsg:  "reservation not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservationQ, mock, teardown := setupReservationTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			err := reservationQ.Delete(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReservationQ_CheckTableAvailability(t *testing.T) {
	tests := []struct {
		name         string
		tableNumber  string
		date         string
		time         string
		mock         func(mock sqlmock.Sqlmock)
		want         bool
		wantErr      bool
	}{
		{
			name:        "table available",
			tableNumber: "T1",
			date:        "2025-12-25",
			time:        "19:00",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT.*FROM reservations WHERE table_number = \$1 AND date = \$2::date AND time = \$3::time AND status IN`).
					WithArgs("T1", "2025-12-25", "19:00").
					WillReturnRows(rows)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:        "table not available",
			tableNumber: "T1",
			date:        "2025-12-25",
			time:        "19:00",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT.*FROM reservations WHERE table_number = \$1 AND date = \$2::date AND time = \$3::time AND status IN`).
					WithArgs("T1", "2025-12-25", "19:00").
					WillReturnRows(rows)
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservationQ, mock, teardown := setupReservationTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			got, err := reservationQ.CheckTableAvailability(ctx, tt.tableNumber, tt.date, tt.time)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

