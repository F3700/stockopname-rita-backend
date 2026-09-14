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

func TestCreateSesi(t *testing.T) {
	h := NewSesiHandler(&fakeSesiService{
		createFunc: func(ctx context.Context, req dto.CreateSesiRequest) (dto.SesiResponse, error) {
			if req.Code != "SESI-01" {
				t.Errorf("unexpected request %+v", req)
			}
			return dto.SesiResponse{ID: 1, Code: "SESI-01"}, nil
		},
	})

	body := `{"sesi_code":"SESI-01","location":"Gudang A","coor_code":["KOR-01"]}`
	req := newJSONRequest(t, http.MethodPost, "/sessions", body)
	rec := httptest.NewRecorder()
	h.CreateSesi(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Sesi created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateSesiInvalidBody(t *testing.T) {
	called := false
	h := NewSesiHandler(&fakeSesiService{
		createFunc: func(ctx context.Context, req dto.CreateSesiRequest) (dto.SesiResponse, error) {
			called = true
			return dto.SesiResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/sessions", `{"sesi_code":`)
	rec := httptest.NewRecorder()
	h.CreateSesi(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}

func TestDeleteSesi(t *testing.T) {
	h := NewSesiHandler(&fakeSesiService{
		deleteFunc: func(ctx context.Context, id int) error {
			if id != 5 {
				t.Errorf("expected id 5, got %d", id)
			}
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/sessions/5", "")
	rec := httptest.NewRecorder()
	h.DeleteSesi(rec, req, paramsWithID("5"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Sesi deleted successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestDeleteSesiInvalidID(t *testing.T) {
	called := false
	h := NewSesiHandler(&fakeSesiService{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/sessions/bad", "")
	rec := httptest.NewRecorder()
	h.DeleteSesi(rec, req, paramsWithID("bad"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestDeleteSesiNotFound(t *testing.T) {
	h := NewSesiHandler(&fakeSesiService{
		deleteFunc: func(ctx context.Context, id int) error {
			return &model.NotFoundError{Resource: "Session", ID: id}
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/sessions/99", "")
	rec := httptest.NewRecorder()
	h.DeleteSesi(rec, req, paramsWithID("99"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetAllSesi(t *testing.T) {
	h := NewSesiHandler(&fakeSesiService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.SesiResponse, error) {
			if pagination.Page != 1 || pagination.Limit != 10 {
				t.Errorf("unexpected pagination %+v", pagination)
			}
			pagination.TotalItems = 1
			pagination.TotalPages = 1
			return []dto.SesiResponse{{ID: 1, Code: "SESI-01"}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions?page=1&limit=10", "")
	rec := httptest.NewRecorder()
	h.GetAllSesi(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Sesi retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetAllSesiMissingLimit(t *testing.T) {
	called := false
	h := NewSesiHandler(&fakeSesiService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.SesiResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions?page=1", "")
	rec := httptest.NewRecorder()
	h.GetAllSesi(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called when limit is missing")
	}
}

func TestGetSesiById(t *testing.T) {
	h := NewSesiHandler(&fakeSesiService{
		findByIdFn: func(ctx context.Context, id int) (dto.SesiResponse, error) {
			if id != 3 {
				t.Errorf("expected id 3, got %d", id)
			}
			return dto.SesiResponse{ID: 3, Code: "SESI-03"}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/3", "")
	rec := httptest.NewRecorder()
	h.GetSesiById(rec, req, paramsWithID("3"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Sesi retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetSesiByIdServiceError(t *testing.T) {
	h := NewSesiHandler(&fakeSesiService{
		findByIdFn: func(ctx context.Context, id int) (dto.SesiResponse, error) {
			return dto.SesiResponse{}, errors.New("boom")
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/3", "")
	rec := httptest.NewRecorder()
	h.GetSesiById(rec, req, paramsWithID("3"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestUpdateSesi(t *testing.T) {
	h := NewSesiHandler(&fakeSesiService{
		updateFunc: func(ctx context.Context, id int, req dto.UpdateSesiRequest) error {
			if id != 2 {
				t.Errorf("expected id 2, got %d", id)
			}
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/sessions/2", `{"status":"COMPLETED"}`)
	rec := httptest.NewRecorder()
	h.UpdateSesi(rec, req, paramsWithID("2"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Sesi updated successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestUpdateSesiInvalidBody(t *testing.T) {
	called := false
	h := NewSesiHandler(&fakeSesiService{
		updateFunc: func(ctx context.Context, id int, req dto.UpdateSesiRequest) error {
			called = true
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/sessions/2", `{"status":`)
	rec := httptest.NewRecorder()
	h.UpdateSesi(rec, req, paramsWithID("2"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}
