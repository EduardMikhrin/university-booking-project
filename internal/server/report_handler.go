package server

import (
	"net/http"
)

// handleGetMonthlyReports handles GET /reports/monthly
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
func (s *Server) handleGetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	month := r.PathValue("month")

	// Validate month format (YYYY-MM)
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

