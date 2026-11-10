package middleware

import (
	"encoding/json"
	"net/http"
)

// errorBody mirrors the API response envelope (dto.Response) so a rejection
// produced by middleware looks like every other error to the frontend.
type errorBody struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func writeJsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{Message: message})
}
