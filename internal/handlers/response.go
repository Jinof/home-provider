package handlers

import (
	"encoding/json"
	"net/http"
)

type ErrorDetails struct {
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Suggestion string                 `json:"suggestion,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, errType, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error": ErrorDetails{
			Type:    errType,
			Message: message,
		},
	})
}

func respondErrorWithDetails(w http.ResponseWriter, status int, errType, message, suggestion string, details map[string]interface{}) {
	respondJSON(w, status, map[string]interface{}{
		"error": ErrorDetails{
			Type:       errType,
			Message:    message,
			Details:    details,
			Suggestion: suggestion,
		},
	})
}
