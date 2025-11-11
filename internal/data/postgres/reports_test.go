package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReportsTestDB(t *testing.T) (*ReportsQ, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sqlxDB := sqlx.NewDb(db, "postgres")
	reportsQ := NewReportsQ(sqlxDB).(*ReportsQ)

	teardown := func() {
		db.Close()
	}

	return reportsQ, mock, teardown
}

func TestReportsQ_GetMonthlyStatsList(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name: "successful get monthly stats list",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"month", "total_reservations", "completed_reservations", "cancelled_reservations", "revenue"}).
					AddRow("2025-12", 10, 8, 1, 400.0).
					AddRow("2025-11", 15, 12, 2, 600.0)
				mock.ExpectQuery(`SELECT.*FROM reservations.*GROUP BY.*ORDER BY month DESC`).
					WillReturnRows(rows)
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "empty result",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"month", "total_reservations", "completed_reservations", "cancelled_reservations", "revenue"})
				mock.ExpectQuery(`SELECT.*FROM reservations.*GROUP BY.*ORDER BY month DESC`).
					WillReturnRows(rows)
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "database error",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT.*FROM reservations.*GROUP BY.*ORDER BY month DESC`).
					WillReturnError(sql.ErrConnDone)
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportsQ, mock, teardown := setupReportsTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			got, err := reportsQ.GetMonthlyStatsList(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)
				assert.Len(t, got, tt.want)
				if tt.want > 0 {
					assert.Equal(t, "2025-12", got[0].Month)
					assert.Equal(t, 10, got[0].TotalReservations)
					assert.Equal(t, 8, got[0].CompletedReservations)
					assert.Equal(t, 1, got[0].CancelledReservations)
					assert.Equal(t, 400.0, got[0].Revenue)
				}
			}

			// Note: Complex query matching might not work perfectly with sqlmock
			// The important part is that we test the logic
		})
	}
}

func TestReportsQ_GetDetailedMonthlyStats(t *testing.T) {
	tests := []struct {
		name    string
		month   string
		mock    func(mock sqlmock.Sqlmock)
		want    *types.DetailedMonthlyStats
		wantErr bool
		errMsg  string
	}{
		{
			name:  "successful get detailed monthly stats",
			month: "2025-12",
			mock: func(mock sqlmock.Sqlmock) {
				// Mock stats query
				statsRows := sqlmock.NewRows([]string{"month", "total_reservations", "completed_reservations", "cancelled_reservations", "revenue"}).
					AddRow("2025-12", 10, 8, 1, 400.0)
				mock.ExpectQuery(`SELECT.*FROM reservations WHERE date >= \$1::date AND date <= \$2::date.*GROUP BY`).
					WithArgs("2025-12-01", "2025-12-31").
					WillReturnRows(statsRows)

				// Mock popular tables query
				popularTablesRows := sqlmock.NewRows([]string{"table_number", "count"}).
					AddRow("T1", 5).
					AddRow("T2", 3)
				mock.ExpectQuery(`SELECT table_number, COUNT.*FROM reservations WHERE date >= \$1::date AND date <= \$2::date AND status = 'completed'.*GROUP BY table_number.*ORDER BY count DESC.*LIMIT 10`).
					WithArgs("2025-12-01", "2025-12-31").
					WillReturnRows(popularTablesRows)

				// Mock peak hours query
				peakHoursRows := sqlmock.NewRows([]string{"hour", "count"}).
					AddRow("19:00", 4).
					AddRow("20:00", 3)
				mock.ExpectQuery(`SELECT time AS hour, COUNT.*FROM reservations WHERE date >= \$1::date AND date <= \$2::date AND status = 'completed'.*GROUP BY time.*ORDER BY count DESC.*LIMIT 10`).
					WithArgs("2025-12-01", "2025-12-31").
					WillReturnRows(peakHoursRows)
			},
			want: &types.DetailedMonthlyStats{
				MonthlyStats: types.MonthlyStats{
					Month:                 "2025-12",
					TotalReservations:     10,
					CompletedReservations: 8,
					CancelledReservations: 1,
					Revenue:               400.0,
				},
				PopularTables: []types.PopularTable{
					{TableNumber: "T1", Count: 5},
					{TableNumber: "T2", Count: 3},
				},
				PeakHours: []types.PeakHour{
					{Hour: "19:00", Count: 4},
					{Hour: "20:00", Count: 3},
				},
			},
			wantErr: false,
		},
		{
			name:  "invalid month format",
			month: "invalid",
			mock:  func(mock sqlmock.Sqlmock) {},
			want:  nil,
			wantErr: true,
			errMsg: "invalid month format, expected YYYY-MM",
		},
		{
			name:  "month not found",
			month: "2025-12",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT.*FROM reservations WHERE date >= \$1::date AND date <= \$2::date.*GROUP BY`).
					WithArgs("2025-12-01", "2025-12-31").
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "statistics for this month not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportsQ, mock, teardown := setupReportsTestDB(t)
			defer teardown()

			tt.mock(mock)

			ctx := context.Background()
			got, err := reportsQ.GetDetailedMonthlyStats(ctx, tt.month)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Month, got.Month)
				assert.Equal(t, tt.want.TotalReservations, got.TotalReservations)
				assert.Len(t, got.PopularTables, len(tt.want.PopularTables))
				assert.Len(t, got.PeakHours, len(tt.want.PeakHours))
			}

			// Note: Complex query matching might not work perfectly with sqlmock
		})
	}
}

