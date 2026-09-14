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

func TestUpdateCoordinator(t *testing.T) {
	h := NewCoordinatorHandler(&fakeCoordinatorService{
		updateFunc: func(ctx context.Context, id int, req *dto.UpdateCoordinatorRequest) error {
			if id != 2 {
				t.Errorf("expected id 2, got %d", id)
			}
			if req.Status != "COMPLETED" {
				t.Errorf("expected status COMPLETED, got %q", req.Status)
			}
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/coordinators/2", `{"status":"COMPLETED"}`)
	rec := httptest.NewRecorder()
	h.UpdateCoordinator(rec, req, paramsWithID("2"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Coordinator updated successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestUpdateCoordinatorInvalidID(t *testing.T) {
	called := false
	h := NewCoordinatorHandler(&fakeCoordinatorService{
		updateFunc: func(ctx context.Context, id int, req *dto.UpdateCoordinatorRequest) error {
			called = true
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/coordinators/x", `{"status":"COMPLETED"}`)
	rec := httptest.NewRecorder()
	h.UpdateCoordinator(rec, req, paramsWithID("x"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestUpdateCoordinatorServiceError(t *testing.T) {
	h := NewCoordinatorHandler(&fakeCoordinatorService{
		updateFunc: func(ctx context.Context, id int, req *dto.UpdateCoordinatorRequest) error {
			return &model.NotFoundError{Resource: "Coordinator", ID: id}
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/coordinators/99", `{"status":"COMPLETED"}`)
	rec := httptest.NewRecorder()
	h.UpdateCoordinator(rec, req, paramsWithID("99"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetCoordinatorById(t *testing.T) {
	h := NewCoordinatorHandler(&fakeCoordinatorService{
		findByIdSummaryFunc: func(ctx context.Context, id int) (*dto.CoordinatorResponse, error) {
			if id != 5 {
				t.Errorf("expected id 5, got %d", id)
			}
			return &dto.CoordinatorResponse{ID: 5, Code: "KOR-01"}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/coordinators/5", "")
	rec := httptest.NewRecorder()
	h.GetCoordinatorById(rec, req, paramsWithID("5"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Coordinator retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetCoordinatorByIdServiceError(t *testing.T) {
	h := NewCoordinatorHandler(&fakeCoordinatorService{
		findByIdSummaryFunc: func(ctx context.Context, id int) (*dto.CoordinatorResponse, error) {
			return nil, errors.New("boom")
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/coordinators/5", "")
	rec := httptest.NewRecorder()
	h.GetCoordinatorById(rec, req, paramsWithID("5"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestGetCoordinatorsWithoutFilter(t *testing.T) {
	h := NewCoordinatorHandler(&fakeCoordinatorService{
		findAllSummaryFunc: func(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error) {
			if sesiId != nil {
				t.Errorf("expected nil sesiId, got %d", *sesiId)
			}
			return []*dto.CoordinatorResponse{{ID: 1, Code: "KOR-01"}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/coordinators", "")
	rec := httptest.NewRecorder()
	h.GetCoordinators(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Coordinators retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetCoordinatorsWithSessionFilter(t *testing.T) {
	h := NewCoordinatorHandler(&fakeCoordinatorService{
		findAllSummaryFunc: func(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error) {
			if sesiId == nil || *sesiId != 4 {
				t.Errorf("expected sesiId 4, got %v", sesiId)
			}
			return []*dto.CoordinatorResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/coordinators?sessionId=4", "")
	rec := httptest.NewRecorder()
	h.GetCoordinators(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetCoordinatorsInvalidSessionFilter(t *testing.T) {
	called := false
	h := NewCoordinatorHandler(&fakeCoordinatorService{
		findAllSummaryFunc: func(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/coordinators?sessionId=abc", "")
	rec := httptest.NewRecorder()
	h.GetCoordinators(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid sessionId")
	}
}
