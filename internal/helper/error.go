package helper

import (
	"encoding/json"
	"log"
	"net/http"
	"stockopname-rita-backend/internal/dto"
)

func WriteError(writer http.ResponseWriter, status int, message string, err error) {
	log.Println(message, ":\n", err)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	res := dto.Response{
		Message: message,
		Data:    nil,
	}
	_ = json.NewEncoder(writer).Encode(res)
}
