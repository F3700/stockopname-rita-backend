package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"

	"stockopname-rita-backend/internal/model"
)

func TestResponseJson(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]interface{}{"message": "ok", "data": 42}

	if err := ResponseJson(rec, http.StatusCreated, payload); err != nil {
		t.Fatalf("ResponseJson returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if got["message"] != "ok" {
		t.Errorf("expected message ok, got %v", got["message"])
	}
}

func TestResponseJsonEncodeFailureDoesNotWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	unencodable := struct {
		Ch chan int
	}{Ch: make(chan int)}

	err := ResponseJson(rec, http.StatusCreated, unencodable)
	if err == nil {
		t.Fatal("expected error for unencodable payload, got nil")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("header must not be written before encoding succeeds, got status %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", rec.Body.String())
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadRequest, "invalid request", errors.New("boom"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var res struct {
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if res.Message != "invalid request" {
		t.Errorf("expected message %q, got %q", "invalid request", res.Message)
	}
}

func TestWriteServiceError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteServiceError(rec, &model.NotFoundError{Resource: "Product", ID: 5})

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	var res struct {
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if res.Message != "Product with ID 5 not found" {
		t.Errorf("expected message %q, got %q", "Product with ID 5 not found", res.Message)
	}
}

func TestMapServiceError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", &model.NotFoundError{Resource: "Product", ID: 1}, http.StatusNotFound},
		{"conflict", &model.ConflictError{Resource: "Product"}, http.StatusConflict},
		{"foreign key", &model.ForeignKeyError{Resource: "Product"}, http.StatusConflict},
		{"validation", &model.ValidationError{Detail: "quantity must be positive"}, http.StatusBadRequest},
		{"unexpected", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _ := mapServiceError(tt.err)
			if status != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, status)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	date, err := ParseDate("2026-09-10T08:00:00+07:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if date == nil {
		t.Fatal("expected non-nil date")
	}
	if got := date.Format(time.RFC3339); got != "2026-09-10T08:00:00+07:00" {
		t.Errorf("expected %q, got %q", "2026-09-10T08:00:00+07:00", got)
	}

	if date, err := ParseDate(""); err != nil || date != nil {
		t.Errorf("expected nil date and nil error for empty input, got date=%v err=%v", date, err)
	}

	if _, err := ParseDate("not-a-date"); err == nil {
		t.Error("expected error for invalid date, got nil")
	}
}

func TestParseIntParam(t *testing.T) {
	params := httprouter.Params{httprouter.Param{Key: "id", Value: "42"}}

	got, err := ParseIntParam(params, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Errorf("expected 42, got %d", got)
	}

	if _, err := ParseIntParam(params, "missing"); err == nil {
		t.Error("expected error for missing param, got nil")
	}

	badParams := httprouter.Params{httprouter.Param{Key: "id", Value: "abc"}}
	if _, err := ParseIntParam(badParams, "id"); err == nil {
		t.Error("expected error for invalid param, got nil")
	}
}

func TestParseIntQueryNotNull(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items?page=3", nil)
	got, err := ParseIntQueryNotNull(req, "page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3 {
		t.Errorf("expected 3, got %d", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/items", nil)
	if _, err := ParseIntQueryNotNull(req, "page"); err == nil {
		t.Error("expected error for missing query param, got nil")
	}

	req = httptest.NewRequest(http.MethodGet, "/items?page=abc", nil)
	if _, err := ParseIntQueryNotNull(req, "page"); err == nil {
		t.Error("expected error for invalid query param, got nil")
	}
}

func TestParseDateNullable(t *testing.T) {
	if got := ParseDateNullable(nil); got != "" {
		t.Errorf("expected empty string for nil, got %q", got)
	}

	date := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	got := ParseDateNullable(&date)
	if !strings.HasPrefix(got, "2026-09-10T08:00:00Z") {
		t.Errorf("unexpected formatted date: %q", got)
	}
}
