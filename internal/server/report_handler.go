package server

import (
	"net/http"
)

// handleGetMonthlyReports handles GET /reports/monthly
// @Summary Get monthly statistics list
// @Description Returns aggregated statistics for all months
// @Tags Reports
// @Produce json
// @Success 200 {array} types.MonthlyStats
// @Failure 500 {object} ErrorResponse "Server error"
// @Router /reports/monthly [get]
func (s *Server) handleGetMonthlyReports(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.ReportsQ().GetMonthlyStatsList(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get monthly reports")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, stats)
}

// handleGetMonthlyReport handles GET /reports/monthly/{month}
// @Summary Get detailed monthly report
// @Description Returns detailed statistics for a specific month (YYYY-MM)
// @Tags Reports
// @Produce json
// @Param month path string true "Month in format YYYY-MM"
// @Success 200 {object} types.DetailedMonthlyStats
// @Failure 400 {object} ErrorResponse "Invalid month format"
// @Failure 404 {object} ErrorResponse "Statistics not found"
// @Failure 500 {object} ErrorResponse "Server error"
// @Router /reports/monthly/{month} [get]
func (s *Server) handleGetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	month := r.PathValue("month")

	if len(month) != 7 || month[4] != '-' {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid month format (expected YYYY-MM)", nil)
		return
	}

	stats, err := s.db.ReportsQ().GetDetailedMonthlyStats(r.Context(), month)
	if err != nil {
		s.log.WithError(err).Error("failed to get monthly report")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if stats == nil {
		writeErrorResponse(w, http.StatusNotFound, "Statistics for this month not found", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, stats)
}
