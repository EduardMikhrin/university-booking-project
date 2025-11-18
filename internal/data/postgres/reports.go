package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/EduardMikhrin/university-booking-project/internal/data"
	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/jmoiron/sqlx"
)

type ReportsQ struct {
	db *sqlx.DB
}

func NewReportsQ(db *sqlx.DB) data.ReportsQ {
	return &ReportsQ{db: db}
}

//
// ────────────────────────────────────────────────────────────────
//   MONTHLY OVERVIEW LIST
// ────────────────────────────────────────────────────────────────
//

func (q *ReportsQ) GetMonthlyStatsList(ctx context.Context) ([]*types.MonthlyStats, error) {
	query := `
		SELECT 
			TO_CHAR(date, 'YYYY-MM') AS month,
			COUNT(*) AS total_reservations,
			COUNT(*) FILTER (WHERE status = 'completed') AS completed_reservations,
			COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_reservations,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) * 50.0, 0) AS revenue
		FROM reservations
		GROUP BY TO_CHAR(date, 'YYYY-MM')
		ORDER BY month DESC
	`

	type result struct {
		Month                 string  `db:"month"`
		TotalReservations     int     `db:"total_reservations"`
		CompletedReservations int     `db:"completed_reservations"`
		CancelledReservations int     `db:"cancelled_reservations"`
		Revenue               float64 `db:"revenue"`
	}

	var results []result
	err := q.db.SelectContext(ctx, &results, query)
	if err != nil {
		return nil, err
	}

	stats := make([]*types.MonthlyStats, len(results))
	for i, r := range results {
		stats[i] = &types.MonthlyStats{
			Month:                 r.Month,
			TotalReservations:     r.TotalReservations,
			CompletedReservations: r.CompletedReservations,
			CancelledReservations: r.CancelledReservations,
			Revenue:               r.Revenue,
		}
	}

	return stats, nil
}

//
// ────────────────────────────────────────────────────────────────
//   MONTHLY DETAILS (POPULAR TABLES + PEAK HOURS)
// ────────────────────────────────────────────────────────────────
//

func (q *ReportsQ) GetDetailedMonthlyStats(ctx context.Context, month string) (*types.DetailedMonthlyStats, error) {
	// Month must be YYYY-MM
	if len(month) != 7 || month[4] != '-' {
		return nil, errors.New("invalid month format (expected YYYY-MM)")
	}

	startDate := month + "-01"

	//
	// ─── BASIC STATS ──────────────────────────────────────────────
	//

	statsQuery := `
        SELECT
            TO_CHAR(date, 'YYYY-MM') AS month,
            COUNT(*) AS total_reservations,
            COUNT(*) FILTER (WHERE status = 'completed') AS completed_reservations,
            COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_reservations,
            COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) * 50.0, 0) AS revenue
        FROM reservations
        WHERE date >= $1::date
          AND date < ($1::date + INTERVAL '1 month')
        GROUP BY TO_CHAR(date, 'YYYY-MM')
    `

	type statsResult struct {
		Month                 string  `db:"month"`
		TotalReservations     int     `db:"total_reservations"`
		CompletedReservations int     `db:"completed_reservations"`
		CancelledReservations int     `db:"cancelled_reservations"`
		Revenue               float64 `db:"revenue"`
	}

	var stats statsResult
	err := q.db.GetContext(ctx, &stats, statsQuery, startDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("statistics for this month not found")
		}
		return nil, err
	}

	//
	// ─── POPULAR TABLES ─────────────────────────────────────────────
	//

	popularTablesQuery := `
        SELECT 
            table_number,
            COUNT(*) AS count
        FROM reservations
        WHERE date >= $1::date
          AND date < ($1::date + INTERVAL '1 month')
          AND status = 'completed'
        GROUP BY table_number
        ORDER BY count DESC
        LIMIT 10
    `

	type popularTableResult struct {
		TableNumber string `db:"table_number"`
		Count       int    `db:"count"`
	}

	var popularTables []popularTableResult
	err = q.db.SelectContext(ctx, &popularTables, popularTablesQuery, startDate)
	if err != nil {
		return nil, err
	}

	//
	// ─── PEAK HOURS — FIXED WITH HH:MI FORMAT ──────────────────────
	//

	peakHoursQuery := `
        SELECT 
            TO_CHAR(time, 'HH24:MI') AS hour,
            COUNT(*) AS count
        FROM reservations
        WHERE date >= $1::date
          AND date < ($1::date + INTERVAL '1 month')
          AND status = 'completed'
        GROUP BY TO_CHAR(time, 'HH24:MI')
        ORDER BY count DESC
        LIMIT 10
    `

	type peakHourResult struct {
		Hour  string `db:"hour"`
		Count int    `db:"count"`
	}

	var peakHours []peakHourResult
	err = q.db.SelectContext(ctx, &peakHours, peakHoursQuery, startDate)
	if err != nil {
		return nil, err
	}

	//
	// ─── BUILD RESPONSE ─────────────────────────────────────────────
	//

	detailedStats := &types.DetailedMonthlyStats{
		MonthlyStats: types.MonthlyStats{
			Month:                 stats.Month,
			TotalReservations:     stats.TotalReservations,
			CompletedReservations: stats.CompletedReservations,
			CancelledReservations: stats.CancelledReservations,
			Revenue:               stats.Revenue,
		},
		PopularTables: make([]types.PopularTable, len(popularTables)),
		PeakHours:     make([]types.PeakHour, len(peakHours)),
	}

	for i, pt := range popularTables {
		detailedStats.PopularTables[i] = types.PopularTable{
			TableNumber: pt.TableNumber,
			Count:       pt.Count,
		}
	}

	for i, ph := range peakHours {
		detailedStats.PeakHours[i] = types.PeakHour{
			Hour:  ph.Hour, // NOW ALWAYS "HH:MM"
			Count: ph.Count,
		}
	}

	return detailedStats, nil
}
