package helper

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"

	"github.com/go-playground/validator/v10"
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

	var fieldErrors validator.ValidationErrors
	if errors.As(err, &fieldErrors) {
		return http.StatusBadRequest, validationErrorMessage(fieldErrors)
	}

	return http.StatusInternalServerError, "Internal server error"
}

func validationErrorMessage(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Invalid request"
	}

	fieldErr := errs[0]
	switch fieldErr.Tag() {
	case "required":
		return fieldErr.Field() + " is required"
	case "gte":
		return fieldErr.Field() + " must be at least " + fieldErr.Param()
	default:
		return fieldErr.Field() + " is invalid"
	}
}
