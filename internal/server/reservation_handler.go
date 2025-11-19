package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
)

type CreateReservationRequest struct {
	GuestName       string  `json:"guestName"`
	GuestPhone      string  `json:"guestPhone"`
	GuestEmail      string  `json:"guestEmail"`
	Date            string  `json:"date"`
	Time            string  `json:"time"`
	Guests          int     `json:"guests"`
	TableNumber     string  `json:"tableNumber"`
	SpecialRequests *string `json:"specialRequests,omitempty"`
}

type UpdateReservationRequest struct {
	GuestName       *string `json:"guestName,omitempty"`
	GuestPhone      *string `json:"guestPhone,omitempty"`
	GuestEmail      *string `json:"guestEmail,omitempty"`
	Date            *string `json:"date,omitempty"`
	Time            *string `json:"time,omitempty"`
	Guests          *int    `json:"guests,omitempty"`
	TableNumber     *string `json:"tableNumber,omitempty"`
	SpecialRequests *string `json:"specialRequests,omitempty"`
}

type UpdateReservationStatusRequest struct {
	Status string `json:"status"`
}

type DeleteResponse struct {
	Message string `json:"message"`
}

// @Summary Get reservations
// @Description Get reservations for current user (admin – all reservations)
// @Tags Reservations
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param date query string false "Filter by date (YYYY-MM-DD)"
// @Param search query string false "Search"
// @Success 200 {array} types.Reservation
// @Failure 500 {object} ErrorResponse
// @Router /reservations [get]
func (s *Server) handleGetReservations(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

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

// @Summary Get reservation by ID
// @Description Get single reservation (only owner or admin)
// @Tags Reservations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Reservation ID"
// @Success 200 {object} types.Reservation
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/{id} [get]
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

	if user.Role != adminRole && reservation.UserID != user.ID {
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, reservation)
}

// @Summary Get reservations by user
// @Description Admin may fetch any user; user may fetch only their own
// @Tags Reservations
// @Security BearerAuth
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {array} types.Reservation
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/user/{userId} [get]
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

// @Summary Create reservation
// @Description Create reservation for authenticated user
// @Tags Reservations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param reservation body CreateReservationRequest true "Reservation payload"
// @Success 201 {object} types.Reservation
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations [post]
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
		validationErrors["date"] = "Invalid date format"
	}
	if req.Time == "" {
		validationErrors["time"] = "Time is required"
	} else if _, err := time.Parse("15:04", req.Time); err != nil {
		validationErrors["time"] = "Invalid time format"
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

	date, _ := time.Parse("2006-01-02", req.Date)

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

	if err := s.cache.ReservationCache().InvalidateUserReservations(r.Context(), user.ID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate reservation cache")
	}

	writeJSONResponse(w, http.StatusCreated, reservation)
}

// @Summary Update reservation
// @Description Update reservation fields (owner or admin)
// @Tags Reservations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Reservation ID"
// @Param body body UpdateReservationRequest true "Payload"
// @Success 200 {object} types.Reservation
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/{id} [patch]
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
			validationErrors["date"] = "Invalid date format"
		} else {
			reservation.Date = date
			hasUpdates = true
		}
	}
	if req.Time != nil {
		if _, err := time.Parse("15:04", *req.Time); err != nil {
			validationErrors["time"] = "Invalid time format"
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

	if err := s.cache.ReservationCache().DeleteReservation(r.Context(), reservationID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate reservation cache")
	}
	if err := s.cache.ReservationCache().InvalidateUserReservations(r.Context(), reservation.UserID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate user reservations cache")
	}

	writeJSONResponse(w, http.StatusOK, reservation)
}

// @Summary Update reservation status
// @Description Update reservation status (pending, confirmed, cancelled, completed)
// @Tags Reservations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Reservation ID"
// @Param body body UpdateReservationStatusRequest true "Status payload"
// @Success 200 {object} types.Reservation
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/{id}/status [patch]
func (s *Server) handleUpdateReservationStatus(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateReservationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.WithError(err).Debug("failed to decode request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"cancelled": true,
		"completed": true,
	}
	if !validStatuses[req.Status] {
		writeErrorResponse(w, http.StatusBadRequest, "Validation error", map[string]string{
			"status": "Invalid status",
		})
		return
	}

	if err := s.db.ReservationQ().UpdateStatus(r.Context(), reservationID, req.Status); err != nil {
		s.log.WithError(err).Error("failed to update reservation status")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	reservation, err = s.db.ReservationQ().GetByID(r.Context(), reservationID)
	if err != nil {
		s.log.WithError(err).Error("failed to get updated reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if err := s.cache.ReservationCache().DeleteReservation(r.Context(), reservationID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate reservation cache")
	}
	if err := s.cache.ReservationCache().InvalidateUserReservations(r.Context(), reservation.UserID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate user reservations cache")
	}

	writeJSONResponse(w, http.StatusOK, reservation)
}

// @Summary Delete reservation
// @Description Delete reservation (owner or admin)
// @Tags Reservations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Reservation ID"
// @Success 200 {object} DeleteResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/{id} [delete]
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

	if user.Role != adminRole && reservation.UserID != user.ID {
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	if err := s.db.ReservationQ().Delete(r.Context(), reservationID); err != nil {
		s.log.WithError(err).Error("failed to delete reservation")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if err := s.cache.ReservationCache().DeleteReservation(r.Context(), reservationID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate reservation cache")
	}
	if err := s.cache.ReservationCache().InvalidateUserReservations(r.Context(), reservation.UserID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate user reservations cache")
	}

	writeJSONResponse(w, http.StatusOK, DeleteResponse{
		Message: "Reservation deleted successfully",
	})
}
