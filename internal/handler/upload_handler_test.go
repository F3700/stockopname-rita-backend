package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/dto"
)

func TestCreateUpload(t *testing.T) {
	h := NewUploadHandler(&fakeUploadService{
		createFunc: func(ctx context.Context, req dto.CreateUploadRequest) (dto.UploadResponse, error) {
			if req.CoordinatorID != 7 || len(req.Racks) != 2 {
				t.Errorf("unexpected request %+v", req)
			}
			return dto.UploadResponse{UploadID: 3, RackCount: 2, ItemCount: 2}, nil
		},
	})

	body := `{"coordinator_id":7,"racks":[{"rak_id":31,"so_products":[{"quantity":5,"barcode":"111"}]},{"rak_id":32,"so_products":[{"quantity":3,"barcode":"222"}]}]}`
	req := newJSONRequest(t, http.MethodPost, "/uploads", body)
	rec := httptest.NewRecorder()
	h.CreateUpload(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Upload created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateUploadInvalidBody(t *testing.T) {
	called := false
	h := NewUploadHandler(&fakeUploadService{
		createFunc: func(ctx context.Context, req dto.CreateUploadRequest) (dto.UploadResponse, error) {
			called = true
			return dto.UploadResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/uploads", `{"coordinator_id":`)
	rec := httptest.NewRecorder()
	h.CreateUpload(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}

func TestGetUploads(t *testing.T) {
	h := NewUploadHandler(&fakeUploadService{
		findAllSummaryFunc: func(ctx context.Context, coordinatorId int) ([]dto.UploadResponse, error) {
			if coordinatorId != 7 {
				t.Errorf("expected coordinatorId 7, got %d", coordinatorId)
			}
			return []dto.UploadResponse{{UploadID: 3, RackCount: 2, ItemCount: 10}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/uploads?coordinatorId=7", "")
	rec := httptest.NewRecorder()
	h.GetUploads(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Uploads retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetUploadsMissingCoordinator(t *testing.T) {
	called := false
	h := NewUploadHandler(&fakeUploadService{
		findAllSummaryFunc: func(ctx context.Context, coordinatorId int) ([]dto.UploadResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/uploads", "")
	rec := httptest.NewRecorder()
	h.GetUploads(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called without coordinatorId")
	}
}
