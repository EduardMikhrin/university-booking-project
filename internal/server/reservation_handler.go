package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

// CreateReservationRequest represents the request body for creating a reservation
type CreateReservationRequest struct {
	GuestName       string  `json:"guestName"`
	GuestPhone      string  `json:"guestPhone"`
	GuestEmail      string  `json:"guestEmail"`
	Date            string  `json:"date"` // YYYY-MM-DD
	Time            string  `json:"time"` // HH:mm
	Guests          int     `json:"guests"`
	TableNumber     string  `json:"tableNumber"`
	SpecialRequests *string `json:"specialRequests,omitempty"`
}

// UpdateReservationRequest represents the request body for updating a reservation
type UpdateReservationRequest struct {
	GuestName       *string `json:"guestName,omitempty"`
	GuestPhone      *string `json:"guestPhone,omitempty"`
	GuestEmail      *string `json:"guestEmail,omitempty"`
	Date            *string `json:"date,omitempty"` // YYYY-MM-DD
	Time            *string `json:"time,omitempty"` // HH:mm
	Guests          *int    `json:"guests,omitempty"`
	TableNumber     *string `json:"tableNumber,omitempty"`
	SpecialRequests *string `json:"specialRequests,omitempty"`
}

// UpdateReservationStatusRequest represents the request body for updating reservation status
type UpdateReservationStatusRequest struct {
	Status string `json:"status"`
}

// DeleteResponse represents the response for delete operations
type DeleteResponse struct {
	Message string `json:"message"`
}

// handleGetReservations handles GET /reservations
func (s *Server) handleGetReservations(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Parse query parameters
	filters := &types.ReservationFilters{}
	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
	}
	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			filters.Date = &date
		}
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	// Admin sees all reservations, users see only their own
	var userID *uuid.UUID
	if user.Role != adminRole {
		userID = &user.ID
	}

	reservations, err := s.db.ReservationQ().GetAll(r.Context(), userID, filters)
	if err != nil {
		s.log.WithError(err).Error("failed to get reservations")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, reservations)
}

// handleGetReservation handles GET /reservations/{id}
func (s *Server) handleGetReservation(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	reservationIDStr := r.PathValue("id")
	reservationID, err := uuid.Parse(reservationIDStr)
	if err != nil {
		s.log.WithError(err).Debug("invalid reservation ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid reservation ID format", nil)
		return
	}

	reservation, err := s.db.ReservationQ().GetByID(r.Context(), reservationID)
	if err != nil {
		s.log.WithError(err).Error("failed to get reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if reservation == nil {
		writeErrorResponse(w, http.StatusNotFound, "Reservation not found", nil)
		return
	}

	// Check authorization: users can only view their own reservations unless admin
	if user.Role != adminRole && reservation.UserID != user.ID {
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, reservation)
}

// handleGetUserReservations handles GET /reservations/user/{userId}
func (s *Server) handleGetUserReservations(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	userIDStr := r.PathValue("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.log.WithError(err).Debug("invalid user ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	// Check authorization: users can only view their own reservations unless admin
	if user.Role != adminRole && userID != user.ID {
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	reservations, err := s.db.ReservationQ().GetByUserID(r.Context(), userID)
	if err != nil {
		s.log.WithError(err).Error("failed to get user reservations")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, reservations)
}

// handleCreateReservation handles POST /reservations
func (s *Server) handleCreateReservation(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	var req CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.WithError(err).Debug("failed to decode request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Validate input
	validationErrors := make(map[string]string)
	req.GuestName = strings.TrimSpace(req.GuestName)
	req.GuestPhone = strings.TrimSpace(req.GuestPhone)
	req.GuestEmail = strings.TrimSpace(req.GuestEmail)
	req.TableNumber = strings.TrimSpace(req.TableNumber)

	if req.GuestName == "" {
		validationErrors["guestName"] = "Guest name is required"
	}
	if req.GuestPhone == "" {
		validationErrors["guestPhone"] = "Guest phone is required"
	}
	if req.GuestEmail == "" {
		validationErrors["guestEmail"] = "Guest email is required"
	} else if !isValidEmail(req.GuestEmail) {
		validationErrors["guestEmail"] = "Invalid email format"
	}
	if req.Date == "" {
		validationErrors["date"] = "Date is required"
	} else if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		validationErrors["date"] = "Invalid date format (expected YYYY-MM-DD)"
	}
	if req.Time == "" {
		validationErrors["time"] = "Time is required"
	} else if _, err := time.Parse("15:04", req.Time); err != nil {
		validationErrors["time"] = "Invalid time format (expected HH:mm)"
	}
	if req.Guests <= 0 {
		validationErrors["guests"] = "Number of guests must be greater than 0"
	}
	if req.TableNumber == "" {
		validationErrors["tableNumber"] = "Table number is required"
	}

	if len(validationErrors) > 0 {
		writeErrorResponse(w, http.StatusBadRequest, "Validation error", validationErrors)
		return
	}

	// Parse date
	date, _ := time.Parse("2006-01-02", req.Date)

	// Check table availability
	available, err := s.db.ReservationQ().CheckTableAvailability(r.Context(), req.TableNumber, req.Date, req.Time)
	if err != nil {
		s.log.WithError(err).Error("failed to check table availability")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}
	if !available {
		writeErrorResponse(w, http.StatusBadRequest, "Validation error", map[string]string{
			"tableNumber": "Table not available at this time",
		})
		return
	}

	// Create reservation
	reservation := &types.Reservation{
		ID:              uuid.New(),
		UserID:          user.ID,
		GuestName:       req.GuestName,
		GuestPhone:      req.GuestPhone,
		GuestEmail:      req.GuestEmail,
		Date:            date,
		Time:            req.Time,
		Guests:          req.Guests,
		TableNumber:     req.TableNumber,
		Status:          "pending",
		SpecialRequests: req.SpecialRequests,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.db.ReservationQ().Create(r.Context(), reservation); err != nil {
		s.log.WithError(err).Error("failed to create reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Invalidate reservation cache
	if err := s.cache.ReservationCache().InvalidateUserReservations(r.Context(), user.ID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate reservation cache")
	}

	writeJSONResponse(w, http.StatusCreated, reservation)
}

// handleUpdateReservation handles PATCH /reservations/{id}
func (s *Server) handleUpdateReservation(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	reservationIDStr := r.PathValue("id")
	reservationID, err := uuid.Parse(reservationIDStr)
	if err != nil {
		s.log.WithError(err).Debug("invalid reservation ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid reservation ID format", nil)
		return
	}

	// Get existing reservation
	reservation, err := s.db.ReservationQ().GetByID(r.Context(), reservationID)
	if err != nil {
		s.log.WithError(err).Error("failed to get reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if reservation == nil {
		writeErrorResponse(w, http.StatusNotFound, "Reservation not found", nil)
		return
	}

	// Check authorization: users can only update their own reservations unless admin
	if user.Role != adminRole && reservation.UserID != user.ID {
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	var req UpdateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.WithError(err).Debug("failed to decode request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Update fields
	hasUpdates := false
	validationErrors := make(map[string]string)

	if req.GuestName != nil {
		name := strings.TrimSpace(*req.GuestName)
		if name == "" {
			validationErrors["guestName"] = "Guest name cannot be empty"
		} else {
			reservation.GuestName = name
			hasUpdates = true
		}
	}
	if req.GuestPhone != nil {
		reservation.GuestPhone = strings.TrimSpace(*req.GuestPhone)
		hasUpdates = true
	}
	if req.GuestEmail != nil {
		email := strings.TrimSpace(*req.GuestEmail)
		if email == "" {
			validationErrors["guestEmail"] = "Guest email cannot be empty"
		} else if !isValidEmail(email) {
			validationErrors["guestEmail"] = "Invalid email format"
		} else {
			reservation.GuestEmail = email
			hasUpdates = true
		}
	}
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			validationErrors["date"] = "Invalid date format (expected YYYY-MM-DD)"
		} else {
			reservation.Date = date
			hasUpdates = true
		}
	}
	if req.Time != nil {
		if _, err := time.Parse("15:04", *req.Time); err != nil {
			validationErrors["time"] = "Invalid time format (expected HH:mm)"
		} else {
			reservation.Time = *req.Time
			hasUpdates = true
		}
	}
	if req.Guests != nil {
		if *req.Guests <= 0 {
			validationErrors["guests"] = "Number of guests must be greater than 0"
		} else {
			reservation.Guests = *req.Guests
			hasUpdates = true
		}
	}
	if req.TableNumber != nil {
		reservation.TableNumber = strings.TrimSpace(*req.TableNumber)
		hasUpdates = true
	}
	if req.SpecialRequests != nil {
		reservation.SpecialRequests = req.SpecialRequests
		hasUpdates = true
	}

	if len(validationErrors) > 0 {
		writeErrorResponse(w, http.StatusBadRequest, "Validation error", validationErrors)
		return
	}

	if !hasUpdates {
		writeJSONResponse(w, http.StatusOK, reservation)
		return
	}

	reservation.UpdatedAt = time.Now()

	if err := s.db.ReservationQ().Update(r.Context(), reservationID, reservation); err != nil {
		s.log.WithError(err).Error("failed to update reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Invalidate reservation cache
	if err := s.cache.ReservationCache().DeleteReservation(r.Context(), reservationID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate reservation cache")
	}
	if err := s.cache.ReservationCache().InvalidateUserReservations(r.Context(), reservation.UserID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate user reservations cache")
	}

	writeJSONResponse(w, http.StatusOK, reservation)
}

// handleUpdateReservationStatus handles PATCH /reservations/{id}/status
func (s *Server) handleUpdateReservationStatus(w http.ResponseWriter, r *http.Request) {
	// User is already authenticated via middleware

	reservationIDStr := r.PathValue("id")
	reservationID, err := uuid.Parse(reservationIDStr)
	if err != nil {
		s.log.WithError(err).Debug("invalid reservation ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid reservation ID format", nil)
		return
	}

	// Get existing reservation
	reservation, err := s.db.ReservationQ().GetByID(r.Context(), reservationID)
	if err != nil {
		s.log.WithError(err).Error("failed to get reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if reservation == nil {
		writeErrorResponse(w, http.StatusNotFound, "Reservation not found", nil)
		return
	}

	var req UpdateReservationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.WithError(err).Debug("failed to decode request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"cancelled": true,
		"completed": true,
	}
	if !validStatuses[req.Status] {
		writeErrorResponse(w, http.StatusBadRequest, "Validation error", map[string]string{
			"status": "Invalid status. Must be one of: pending, confirmed, cancelled, completed",
		})
		return
	}

	// Update status
	if err := s.db.ReservationQ().UpdateStatus(r.Context(), reservationID, req.Status); err != nil {
		s.log.WithError(err).Error("failed to update reservation status")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Get updated reservation
	reservation, err = s.db.ReservationQ().GetByID(r.Context(), reservationID)
	if err != nil {
		s.log.WithError(err).Error("failed to get updated reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Invalidate reservation cache
	if err := s.cache.ReservationCache().DeleteReservation(r.Context(), reservationID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate reservation cache")
	}
	if err := s.cache.ReservationCache().InvalidateUserReservations(r.Context(), reservation.UserID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate user reservations cache")
	}

	writeJSONResponse(w, http.StatusOK, reservation)
}

// handleDeleteReservation handles DELETE /reservations/{id}
func (s *Server) handleDeleteReservation(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	reservationIDStr := r.PathValue("id")
	reservationID, err := uuid.Parse(reservationIDStr)
	if err != nil {
		s.log.WithError(err).Debug("invalid reservation ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid reservation ID format", nil)
		return
	}

	// Get existing reservation to check authorization
	reservation, err := s.db.ReservationQ().GetByID(r.Context(), reservationID)
	if err != nil {
		s.log.WithError(err).Error("failed to get reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if reservation == nil {
		writeErrorResponse(w, http.StatusNotFound, "Reservation not found", nil)
		return
	}

	// Check authorization: users can only delete their own reservations unless admin
	if user.Role != adminRole && reservation.UserID != user.ID {
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	// Delete reservation
	if err := s.db.ReservationQ().Delete(r.Context(), reservationID); err != nil {
		s.log.WithError(err).Error("failed to delete reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Invalidate reservation cache
	if err := s.cache.ReservationCache().DeleteReservation(r.Context(), reservationID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate reservation cache")
	}
	if err := s.cache.ReservationCache().InvalidateUserReservations(r.Context(), reservation.UserID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate user reservations cache")
	}

	response := DeleteResponse{
		Message: "Reservation deleted successfully",
	}
	writeJSONResponse(w, http.StatusOK, response)
}

