package types

// MonthlyStats represents monthly statistics
type MonthlyStats struct {
	Month                 string  `json:"month"`
	TotalReservations     int     `json:"totalReservations"`
	CompletedReservations int     `json:"completedReservations"`
	CancelledReservations int     `json:"cancelledReservations"`
	Revenue               float64 `json:"revenue"`
}

// DetailedMonthlyStats represents detailed monthly statistics
type DetailedMonthlyStats struct {
	MonthlyStats
	PopularTables []PopularTable `json:"popularTables"`
	PeakHours     []PeakHour     `json:"peakHours"`
}

// PopularTable represents a popular table statistic
type PopularTable struct {
	TableNumber string `json:"tableNumber"`
	Count       int    `json:"count"`
}

// PeakHour represents a peak hour statistic
type PeakHour struct {
	Hour  string `json:"hour"`
	Count int    `json:"count"`
}

