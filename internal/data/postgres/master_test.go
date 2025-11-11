package postgres

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNewMaster(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	master := NewMaster(sqlxDB)

	assert.NotNil(t, master)
	assert.NotNil(t, master.UserQ())
	assert.NotNil(t, master.ReservationQ())
	assert.NotNil(t, master.TableQ())
	assert.NotNil(t, master.ReportsQ())
}

func TestMaster_UserQ(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	master := NewMaster(sqlxDB).(*Master)

	userQ1 := master.UserQ()
	userQ2 := master.UserQ()

	// Should return the same instance (lazy initialization)
	assert.Equal(t, userQ1, userQ2)
}

func TestMaster_ReservationQ(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	master := NewMaster(sqlxDB).(*Master)

	reservationQ1 := master.ReservationQ()
	reservationQ2 := master.ReservationQ()

	// Should return the same instance (lazy initialization)
	assert.Equal(t, reservationQ1, reservationQ2)
}

func TestMaster_TableQ(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	master := NewMaster(sqlxDB).(*Master)

	tableQ1 := master.TableQ()
	tableQ2 := master.TableQ()

	// Should return the same instance (lazy initialization)
	assert.Equal(t, tableQ1, tableQ2)
}

func TestMaster_ReportsQ(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	master := NewMaster(sqlxDB).(*Master)

	reportsQ1 := master.ReportsQ()
	reportsQ2 := master.ReportsQ()

	// Should return the same instance (lazy initialization)
	assert.Equal(t, reportsQ1, reportsQ2)
}

