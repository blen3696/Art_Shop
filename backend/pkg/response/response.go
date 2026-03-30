package response

import (
	"encoding/json"
	"net/http"
)

// APIResponse is the standardized envelope for all API responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError carries machine-readable and human-readable error information.
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta provides pagination metadata for list endpoints.
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// JSON writes a success response with the given HTTP status code and data payload.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	writeJSON(w, status, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Error writes an error response with the given HTTP status, error code, and
// human-readable message.
func Error(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorWithDetails writes an error response that includes additional detail
// data (e.g., per-field validation errors).
func ErrorWithDetails(w http.ResponseWriter, status int, code string, message string, details interface{}) {
	writeJSON(w, status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// ValidationError writes a 422 Unprocessable Entity response with field-level
// validation errors. The errors map is keyed by field name.
func ValidationError(w http.ResponseWriter, errors map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "One or more fields failed validation",
			Details: errors,
		},
	})
}

// Paginated writes a success response that includes pagination metadata.
func Paginated(w http.ResponseWriter, status int, data interface{}, meta Meta) {
	writeJSON(w, status, APIResponse{
		Success: true,
		Data:    data,
		Meta:    &meta,
	})
}

// Created is a convenience wrapper for 201 Created responses.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 No Content response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// writeJSON marshals the payload and writes it to the response writer.
func writeJSON(w http.ResponseWriter, status int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// If encoding fails we cannot send a JSON error (the header is
		// already written), so we just log via the stdlib.
		http.Error(w, "internal encoding error", http.StatusInternalServerError)
	}
}
