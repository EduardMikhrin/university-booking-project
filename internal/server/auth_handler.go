package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the request body for login
// @Description Login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents the request body for registration
// @Description Registration request body
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

// AuthResponse represents the response for login and register
type AuthResponse struct {
	User  *types.User `json:"user"`
	Token string      `json:"token"`
}

// LogoutResponse represents the response for logout
type LogoutResponse struct {
	Message string `json:"message"`
}

// handleLogin handles POST /auth/login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid email or password"
// @Failure 500 {object} ErrorResponse "Server error"
// @Router /auth/login [post]
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.WithError(err).Debug("failed to decode login request")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Email and password are required", nil)
		return
	}

	user, err := s.db.UserQ().GetByEmail(r.Context(), req.Email)
	if err != nil {
		s.log.WithError(err).Error("failed to get user by email")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		s.log.WithError(err).Error("failed to generate token")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if err := s.cache.TokenCache().SetToken(r.Context(), token, user.ID, s.jwtConfig.AccessTokenLifetime); err != nil {
		s.log.WithError(err).Warn("failed to cache token")
	}

	response := AuthResponse{
		User:  user,
		Token: token,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleRegister handles POST /auth/register
// @Summary User registration
// @Description Create a new user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Server error"
// @Router /auth/register [post]
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.WithError(err).Debug("failed to decode register request")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	validationErrors := make(map[string]string)
	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)
	req.Phone = strings.TrimSpace(req.Phone)

	if req.Email == "" {
		validationErrors["email"] = "Email is required"
	} else if !isValidEmail(req.Email) {
		validationErrors["email"] = "Invalid email format"
	}

	if req.Password == "" {
		validationErrors["password"] = "Password is required"
	} else if len(req.Password) < 6 {
		validationErrors["password"] = "Password must be at least 6 characters"
	}

	if req.Name == "" {
		validationErrors["name"] = "Name is required"
	}

	if len(validationErrors) > 0 {
		writeErrorResponse(w, http.StatusBadRequest, "Validation error", validationErrors)
		return
	}

	existingUser, err := s.db.UserQ().GetByEmail(r.Context(), req.Email)
	if err != nil {
		s.log.WithError(err).Error("failed to check email existence")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}
	if existingUser != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Validation error", map[string]string{
			"email": "Email already exists",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.WithError(err).Error("failed to hash password")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	user := &types.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Name:      req.Name,
		Phone:     &req.Phone,
		Role:      "user",
		CreatedAt: time.Now(),
	}

	if err := s.db.UserQ().Create(r.Context(), user); err != nil {
		s.log.WithError(err).Error("failed to create user")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		s.log.WithError(err).Error("failed to generate token")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if err := s.cache.TokenCache().SetToken(r.Context(), token, user.ID, s.jwtConfig.AccessTokenLifetime); err != nil {
		s.log.WithError(err).Warn("failed to cache token")
	}

	response := AuthResponse{
		User:  user,
		Token: token,
	}
	writeJSONResponse(w, http.StatusCreated, response)
}

// handleGetMe handles GET /auth/me
// @Summary Get current user
// @Description Get authenticated user from JWT token
// @Tags Auth
// @Produce json
// @Success 200 {object} types.User
// @Failure 500 {object} ErrorResponse "Server error"
// @Router /auth/me [get]
func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	writeJSONResponse(w, http.StatusOK, user)
}

// handleLogout handles POST /auth/logout
// @Summary Logout user
// @Description Invalidate JWT token and remove from cache
// @Tags Auth
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} LogoutResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Server error"
// @Router /auth/logout [post]
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token, err := extractToken(r)
	if err != nil {
		s.log.WithError(err).Debug("failed to extract token")
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := GetUserFromContext(r)
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}

	if err := s.cache.TokenCache().DeleteToken(r.Context(), token); err != nil {
		s.log.WithError(err).WithField("user_id", user.ID).Warn("failed to delete token from cache")
	}

	if err := s.cache.TokenCache().SetTokenBlacklist(r.Context(), token, s.jwtConfig.AccessTokenLifetime); err != nil {
		s.log.WithError(err).WithField("user_id", user.ID).Warn("failed to blacklist token")
	}

	response := LogoutResponse{
		Message: "Logged out successfully",
	}
	writeJSONResponse(w, http.StatusOK, response)
}

func (s *Server) generateToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		Issuer:    s.jwtConfig.Issuer,
		Audience:  []string{s.jwtConfig.Audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtConfig.AccessTokenLifetime)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.SecretKey))
}
