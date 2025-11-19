package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/distributed_lab/logan/v3"
)

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty"`
	Phone *string `json:"phone,omitempty"`
	Email *string `json:"email,omitempty"`
}

// @Summary Get user by ID
// @Description Get user profile by ID (only self or admin)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} types.User
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userIDStr).Debug("invalid user ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	authenticatedUser, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get authenticated user")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if authenticatedUser.ID != userID && authenticatedUser.Role != adminRole {
		s.log.WithFields(logan.F{
			"authenticated_user_id": authenticatedUser.ID,
			"requested_user_id":     userID,
		}).Debug("unauthorized access attempt")
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	user, err := s.db.UserQ().GetByID(r.Context(), userID)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("failed to get user from database")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if user == nil {
		s.log.WithField("user_id", userID).Debug("user not found")
		writeErrorResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, user)
}

// @Summary Update user
// @Description Update user profile (only self or admin)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body UpdateUserRequest true "User update payload"
// @Success 200 {object} types.User
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [patch]
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userIDStr).Debug("invalid user ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	authenticatedUser, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get authenticated user")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if authenticatedUser.ID != userID && authenticatedUser.Role != adminRole {
		s.log.WithFields(logan.F{
			"authenticated_user_id": authenticatedUser.ID,
			"requested_user_id":     userID,
		}).Debug("unauthorized update attempt")
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	user, err := s.db.UserQ().GetByID(r.Context(), userID)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("failed to get user from database")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if user == nil {
		s.log.WithField("user_id", userID).Debug("user not found")
		writeErrorResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	var updateReq UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		s.log.WithError(err).Debug("failed to decode request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	validationErrors := make(map[string]string)
	hasUpdates := false

	if updateReq.Name != nil {
		name := strings.TrimSpace(*updateReq.Name)
		if name == "" {
			validationErrors["name"] = "Name cannot be empty"
		} else {
			user.Name = name
			hasUpdates = true
		}
	}

	if updateReq.Phone != nil {
		phone := strings.TrimSpace(*updateReq.Phone)
		user.Phone = &phone
		hasUpdates = true
	}

	if updateReq.Email != nil {
		email := strings.TrimSpace(*updateReq.Email)
		if email == "" {
			validationErrors["email"] = "Email cannot be empty"
		} else if !isValidEmail(email) {
			validationErrors["email"] = "Invalid email format"
		} else if email != user.Email {
			existingUser, err := s.db.UserQ().GetByEmail(r.Context(), email)
			if err != nil {
				s.log.WithError(err).Error("failed to check email existence")
				writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
				return
			}
			if existingUser != nil && existingUser.ID != userID {
				validationErrors["email"] = "Email already exists"
			} else {
				user.Email = email
				hasUpdates = true
			}
		}
	}

	if len(validationErrors) > 0 {
		writeErrorResponse(w, http.StatusBadRequest, "Validation error", validationErrors)
		return
	}

	if !hasUpdates {
		writeJSONResponse(w, http.StatusOK, user)
		return
	}

	if err := s.db.UserQ().Update(r.Context(), userID, user); err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("failed to update user")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if err := s.cache.UserCache().DeleteUser(r.Context(), userID); err != nil {
		s.log.WithError(err).WithField("user_id", userID).Warn("failed to invalidate user cache")
	}

	writeJSONResponse(w, http.StatusOK, user)
}
