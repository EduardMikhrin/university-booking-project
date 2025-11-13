package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

// UpdateTableAvailabilityRequest represents the request body for updating table availability
type UpdateTableAvailabilityRequest struct {
	IsAvailable bool `json:"isAvailable"`
}

// handleGetTables handles GET /tables
func (s *Server) handleGetTables(w http.ResponseWriter, r *http.Request) {
	tables, err := s.db.TableQ().GetAll(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get tables")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, tables)
}

// handleGetTable handles GET /tables/{id}
func (s *Server) handleGetTable(w http.ResponseWriter, r *http.Request) {
	tableIDStr := r.PathValue("id")
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		s.log.WithError(err).Debug("invalid table ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid table ID format", nil)
		return
	}

	table, err := s.db.TableQ().GetByID(r.Context(), tableID)
	if err != nil {
		s.log.WithError(err).Error("failed to get table")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if table == nil {
		writeErrorResponse(w, http.StatusNotFound, "Table not found", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, table)
}

// handleGetAvailableTables handles GET /tables/available
func (s *Server) handleGetAvailableTables(w http.ResponseWriter, r *http.Request) {
	filters := &types.TableAvailabilityFilters{}

	// Parse query parameters
	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			filters.Date = &date
		}
	}
	if timeStr := r.URL.Query().Get("time"); timeStr != "" {
		filters.Time = &timeStr
	}
	if guestsStr := r.URL.Query().Get("guests"); guestsStr != "" {
		// Parse guests as integer
		// Note: This is a simplified parsing, you might want to add proper error handling
		var guests int
		if _, err := fmt.Sscanf(guestsStr, "%d", &guests); err == nil {
			filters.Guests = &guests
		}
	}

	tables, err := s.db.TableQ().GetAvailable(r.Context(), filters)
	if err != nil {
		s.log.WithError(err).Error("failed to get available tables")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, tables)
}

// handleUpdateTableAvailability handles PATCH /tables/{id}/availability
func (s *Server) handleUpdateTableAvailability(w http.ResponseWriter, r *http.Request) {
	tableIDStr := r.PathValue("id")
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		s.log.WithError(err).Debug("invalid table ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid table ID format", nil)
		return
	}

	// Get existing table
	table, err := s.db.TableQ().GetByID(r.Context(), tableID)
	if err != nil {
		s.log.WithError(err).Error("failed to get table")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if table == nil {
		writeErrorResponse(w, http.StatusNotFound, "Table not found", nil)
		return
	}

	var req UpdateTableAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.WithError(err).Debug("failed to decode request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Update availability
	if err := s.db.TableQ().UpdateAvailability(r.Context(), tableID, req.IsAvailable); err != nil {
		s.log.WithError(err).Error("failed to update table availability")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Get updated table
	table, err = s.db.TableQ().GetByID(r.Context(), tableID)
	if err != nil {
		s.log.WithError(err).Error("failed to get updated table")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Invalidate table cache
	if err := s.cache.TableCache().InvalidateTableCache(r.Context()); err != nil {
		s.log.WithError(err).Warn("failed to invalidate table cache")
	}

	writeJSONResponse(w, http.StatusOK, table)
}

