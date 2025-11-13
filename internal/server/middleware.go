package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"gitlab.com/distributed_lab/logan/v3"
)

const (
	userContextKey = "user"
	adminRole      = "admin"
)

type contextKey string

// GetUserFromContext retrieves the user from the request context
// This can be used by handlers to get the authenticated user
func GetUserFromContext(r *http.Request) (*types.User, error) {
	user, ok := r.Context().Value(contextKey(userContextKey)).(*types.User)
	if !ok || user == nil {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// extractToken extracts the Bearer token from the Authorization header
func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// userMiddleware validates JWT token and loads user into context
func (s *Server) userMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token, err := extractToken(r)
		if err != nil {
			s.log.WithError(err).Debug("failed to extract token")
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Check if token is blacklisted
		isBlacklisted, err := s.cache.TokenCache().IsTokenBlacklisted(r.Context(), token)
		if err != nil {
			s.log.WithError(err).Error("failed to check token blacklist")
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if isBlacklisted {
			s.log.Debug("token is blacklisted")
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Get user ID from token cache
		userID, err := s.cache.TokenCache().GetUserIDByToken(r.Context(), token)
		if err != nil {
			s.log.WithError(err).Debug("failed to get user ID from token")
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Get user from database
		user, err := s.db.UserQ().GetByID(r.Context(), userID)
		if err != nil {
			s.log.WithError(err).Error("failed to get user from database")
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if user == nil {
			s.log.WithField("user_id", userID).Warn("user not found")
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Store user in context
		ctx := context.WithValue(r.Context(), contextKey(userContextKey), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// adminMiddleware validates that the user is an admin
func (s *Server) adminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return s.userMiddleware(func(w http.ResponseWriter, r *http.Request) {
		user, err := GetUserFromContext(r)
		if err != nil {
			s.log.WithError(err).Error("failed to get user from context in admin middleware")
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if user.Role != adminRole {
			s.log.WithFields(logan.F{
				"user_id": user.ID,
				"role":    user.Role,
			}).Debug("non-admin user attempted to access admin endpoint")
			http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

