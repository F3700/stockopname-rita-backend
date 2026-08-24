package helper

import (
	"encoding/json"
	"net/http"
)

func ResponseJson(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	return err
}
