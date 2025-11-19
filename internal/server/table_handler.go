package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

type UpdateTableAvailabilityRequest struct {
	IsAvailable bool `json:"isAvailable"`
}

// @Summary Get all tables
// @Description Get list of all tables
// @Tags Tables
// @Security BearerAuth
// @Produce json
// @Success 200 {array} types.Table
// @Failure 500 {object} ErrorResponse
// @Router /tables [get]
func (s *Server) handleGetTables(w http.ResponseWriter, r *http.Request) {
	tables, err := s.db.TableQ().GetAll(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get tables")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}
	writeJSONResponse(w, http.StatusOK, tables)
}

// @Summary Get table by ID
// @Description Get a specific table by ID
// @Tags Tables
// @Security BearerAuth
// @Produce json
// @Param id path string true "Table ID"
// @Success 200 {object} types.Table
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tables/{id} [get]
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

// @Summary Get available tables
// @Description Get tables available for specified date/time/guests
// @Tags Tables
// @Security BearerAuth
// @Produce json
// @Param date query string false "Date (YYYY-MM-DD)"
// @Param time query string false "Time (HH:mm)"
// @Param guests query int false "Number of guests"
// @Success 200 {array} types.Table
// @Failure 500 {object} ErrorResponse
// @Router /tables/available [get]
func (s *Server) handleGetAvailableTables(w http.ResponseWriter, r *http.Request) {
	filters := &types.TableAvailabilityFilters{}

	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			filters.Date = &date
		}
	}
	if timeStr := r.URL.Query().Get("time"); timeStr != "" {
		filters.Time = &timeStr
	}
	if guestsStr := r.URL.Query().Get("guests"); guestsStr != "" {
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

// @Summary Update table availability
// @Description Update availability for a specific table
// @Tags Tables
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Table ID"
// @Param body body UpdateTableAvailabilityRequest true "Availability payload"
// @Success 200 {object} types.Table
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tables/{id}/availability [patch]
func (s *Server) handleUpdateTableAvailability(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateTableAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.WithError(err).Debug("failed to decode request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := s.db.TableQ().UpdateAvailability(r.Context(), tableID, req.IsAvailable); err != nil {
		s.log.WithError(err).Error("failed to update table availability")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	table, err = s.db.TableQ().GetByID(r.Context(), tableID)
	if err != nil {
		s.log.WithError(err).Error("failed to get updated table")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if err := s.cache.TableCache().InvalidateTableCache(r.Context()); err != nil {
		s.log.WithError(err).Warn("failed to invalidate table cache")
	}

	writeJSONResponse(w, http.StatusOK, table)
}
