package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but can't change response at this point
		// This should rarely happen
	}
}

// writeErrorResponse writes an error JSON response
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string, details map[string]string) {
	response := ErrorResponse{
		Error: message,
	}
	if details != nil {
		response.Details = details
	}
	writeJSONResponse(w, statusCode, response)
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

