package postgres

import (
	"github.com/EduardMikhrin/university-booking-project/internal/data"

	"github.com/jmoiron/sqlx"
)

// Master implements the MasterQ interface
type Master struct {
	db *sqlx.DB

	userQ        data.UserQ
	reservationQ data.ReservationQ
	tableQ       data.TableQ
	reportsQ     data.ReportsQ
}

// NewMaster creates a new Master instance
func NewMaster(db *sqlx.DB) data.MasterQ {
	return &Master{
		db: db,
	}
}

// UserQ returns the user query interface
func (m *Master) UserQ() data.UserQ {
	if m.userQ == nil {
		m.userQ = NewUserQ(m.db)
	}
	return m.userQ
}

// ReservationQ returns the reservation query interface
func (m *Master) ReservationQ() data.ReservationQ {
	if m.reservationQ == nil {
		m.reservationQ = NewReservationQ(m.db)
	}
	return m.reservationQ
}

// TableQ returns the table query interface
func (m *Master) TableQ() data.TableQ {
	if m.tableQ == nil {
		m.tableQ = NewTableQ(m.db)
	}
	return m.tableQ
}

// ReportsQ returns the reports query interface
func (m *Master) ReportsQ() data.ReportsQ {
	if m.reportsQ == nil {
		m.reportsQ = NewReportsQ(m.db)
	}
	return m.reportsQ
}
