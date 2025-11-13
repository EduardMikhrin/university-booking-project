package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/distributed_lab/logan/v3"
)

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty"`
	Phone *string `json:"phone,omitempty"`
	Email *string `json:"email,omitempty"`
}

// handleGetUser handles GET /users/{id}
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path parameter
	userIDStr := r.PathValue("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userIDStr).Debug("invalid user ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	// Get authenticated user from context
	authenticatedUser, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get authenticated user")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Check authorization: users can only view their own profile unless they are admin
	if authenticatedUser.ID != userID && authenticatedUser.Role != adminRole {
		s.log.WithFields(logan.F{
			"authenticated_user_id": authenticatedUser.ID,
			"requested_user_id":     userID,
		}).Debug("unauthorized access attempt")
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	// Get user from database
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

	// Return user (password is already excluded via json:"-" tag)
	writeJSONResponse(w, http.StatusOK, user)
}

// handleUpdateUser handles PATCH /users/{id}
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path parameter
	userIDStr := r.PathValue("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userIDStr).Debug("invalid user ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	// Get authenticated user from context
	authenticatedUser, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get authenticated user")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Check authorization: users can only update their own profile unless they are admin
	if authenticatedUser.ID != userID && authenticatedUser.Role != adminRole {
		s.log.WithFields(logan.F{
			"authenticated_user_id": authenticatedUser.ID,
			"requested_user_id":     userID,
		}).Debug("unauthorized update attempt")
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	// Get existing user
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

	// Parse request body
	var updateReq UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		s.log.WithError(err).Debug("failed to decode request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Validate and update fields
	validationErrors := make(map[string]string)
	hasUpdates := false

	// Update name if provided
	if updateReq.Name != nil {
		name := strings.TrimSpace(*updateReq.Name)
		if name == "" {
			validationErrors["name"] = "Name cannot be empty"
		} else {
			user.Name = name
			hasUpdates = true
		}
	}

	// Update phone if provided
	if updateReq.Phone != nil {
		phone := strings.TrimSpace(*updateReq.Phone)
		user.Phone = &phone
		hasUpdates = true
	}

	// Update email if provided
	if updateReq.Email != nil {
		email := strings.TrimSpace(*updateReq.Email)
		if email == "" {
			validationErrors["email"] = "Email cannot be empty"
		} else if !isValidEmail(email) {
			validationErrors["email"] = "Invalid email format"
		} else if email != user.Email {
			// Check if email already exists
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

	// Return validation errors if any
	if len(validationErrors) > 0 {
		writeErrorResponse(w, http.StatusBadRequest, "Validation error", validationErrors)
		return
	}

	// Check if there are any updates
	if !hasUpdates {
		// No updates provided, return current user
		writeJSONResponse(w, http.StatusOK, user)
		return
	}

	// Update user in database
	if err := s.db.UserQ().Update(r.Context(), userID, user); err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("failed to update user")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	// Invalidate user cache
	if err := s.cache.UserCache().DeleteUser(r.Context(), userID); err != nil {
		s.log.WithError(err).WithField("user_id", userID).Warn("failed to invalidate user cache")
	}
	// Note: Email cache will expire naturally, or we could add DeleteUserByEmail to cache interface if needed

	// Return updated user
	writeJSONResponse(w, http.StatusOK, user)
}
