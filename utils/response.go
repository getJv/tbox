// Package utils provides shared generic and multi-use utility functions.
package utils

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard error payload returned by all API endpoints.
type ErrorResponse struct {
	Message string `json:"message" example:"Error message description"`
}

// RespondWithError writes a JSON error response with the given status code and message.
func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

// RespondWithJSON writes a JSON success response with the given status code and payload.
func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
