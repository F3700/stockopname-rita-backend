package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func TestCreateDepartment(t *testing.T) {
	h := NewDepartmentHandler(&fakeDepartmentService{
		createFunc: func(ctx context.Context, req dto.DepartmentCreateRequest) (dto.DepartmentResponse, error) {
			if req.Code != "D01" || req.Name != "Gudang" {
				t.Errorf("unexpected request %+v", req)
			}
			return dto.DepartmentResponse{ID: 1, Code: "D01", Name: "Gudang"}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/departments", `{"code":"D01","name":"Gudang"}`)
	rec := httptest.NewRecorder()
	h.CreateDepartment(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Department created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateDepartmentInvalidBody(t *testing.T) {
	called := false
	h := NewDepartmentHandler(&fakeDepartmentService{
		createFunc: func(ctx context.Context, req dto.DepartmentCreateRequest) (dto.DepartmentResponse, error) {
			called = true
			return dto.DepartmentResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/departments", `{"code":`)
	rec := httptest.NewRecorder()
	h.CreateDepartment(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}

func TestDeleteDepartment(t *testing.T) {
	h := NewDepartmentHandler(&fakeDepartmentService{
		deleteFunc: func(ctx context.Context, id int) error {
			if id != 2 {
				t.Errorf("expected id 2, got %d", id)
			}
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/departments/2", "")
	rec := httptest.NewRecorder()
	h.DeleteDepartment(rec, req, paramsWithID("2"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Department deleted successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestDeleteDepartmentInvalidID(t *testing.T) {
	called := false
	h := NewDepartmentHandler(&fakeDepartmentService{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/departments/NaN", "")
	rec := httptest.NewRecorder()
	h.DeleteDepartment(rec, req, paramsWithID("NaN"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestGetDepartments(t *testing.T) {
	h := NewDepartmentHandler(&fakeDepartmentService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.DepartmentResponse, error) {
			pagination.TotalItems = 2
			pagination.TotalPages = 1
			return []dto.DepartmentResponse{
				{ID: 1, Code: "D01"},
				{ID: 2, Code: "D02"},
			}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/departments?page=1&limit=10", "")
	rec := httptest.NewRecorder()
	h.GetDepartments(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Departments retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetDepartmentsMissingLimit(t *testing.T) {
	called := false
	h := NewDepartmentHandler(&fakeDepartmentService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.DepartmentResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/departments?page=1", "")
	rec := httptest.NewRecorder()
	h.GetDepartments(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called when limit is missing")
	}
}

func TestUpdateDepartment(t *testing.T) {
	h := NewDepartmentHandler(&fakeDepartmentService{
		updateFunc: func(ctx context.Context, id int, req dto.DepartmentUpdateRequest) (dto.DepartmentResponse, error) {
			if id != 4 {
				t.Errorf("expected id 4, got %d", id)
			}
			return dto.DepartmentResponse{ID: 4, Code: "D04"}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/departments/4", `{"code":"D04"}`)
	rec := httptest.NewRecorder()
	h.UpdateDepartment(rec, req, paramsWithID("4"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Department updated successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestUpdateDepartmentNotFound(t *testing.T) {
	h := NewDepartmentHandler(&fakeDepartmentService{
		updateFunc: func(ctx context.Context, id int, req dto.DepartmentUpdateRequest) (dto.DepartmentResponse, error) {
			return dto.DepartmentResponse{}, &model.NotFoundError{Resource: "Department", ID: id}
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/departments/77", `{"code":"D77"}`)
	rec := httptest.NewRecorder()
	h.UpdateDepartment(rec, req, paramsWithID("77"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdateDepartmentUnexpectedError(t *testing.T) {
	h := NewDepartmentHandler(&fakeDepartmentService{
		updateFunc: func(ctx context.Context, id int, req dto.DepartmentUpdateRequest) (dto.DepartmentResponse, error) {
			return dto.DepartmentResponse{}, errors.New("boom")
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/departments/4", `{"code":"D04"}`)
	rec := httptest.NewRecorder()
	h.UpdateDepartment(rec, req, paramsWithID("4"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
