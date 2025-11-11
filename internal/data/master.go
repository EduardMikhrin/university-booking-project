package data

// MasterQ is the master query interface that combines all query interfaces
// It provides access to all database operations through a single interface
type MasterQ interface {
	// UserQ returns the user query interface
	UserQ() UserQ

	// ReservationQ returns the reservation query interface
	ReservationQ() ReservationQ

	// TableQ returns the table query interface
	TableQ() TableQ

	// ReportsQ returns the reports query interface
	ReportsQ() ReportsQ
}
