package helper

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
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

func WriteServiceError(writer http.ResponseWriter, err error) {
	status, message := mapServiceError(err)
	log.Printf("[ERROR] %s: %v\n", message, err)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	res := dto.Response{
		Message: message,
		Data:    nil,
	}
	_ = json.NewEncoder(writer).Encode(res)
}

func mapServiceError(err error) (int, string) {
	var notFound *model.NotFoundError
	if errors.As(err, &notFound) {
		return http.StatusNotFound, notFound.Error()
	}

	var conflict *model.ConflictError
	if errors.As(err, &conflict) {
		return http.StatusConflict, conflict.Error()
	}

	var fkErr *model.ForeignKeyError
	if errors.As(err, &fkErr) {
		return http.StatusConflict, fkErr.Error()
	}

	var validation *model.ValidationError
	if errors.As(err, &validation) {
		return http.StatusBadRequest, validation.Error()
	}

	return http.StatusInternalServerError, "Internal server error"
}
