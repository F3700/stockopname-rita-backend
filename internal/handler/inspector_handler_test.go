package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/dto"
)

func TestGetInspectorsWithoutFilter(t *testing.T) {
	h := NewInspectorHandler(&fakeInspectorService{
		findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
			if coorId != nil {
				t.Errorf("expected nil coorId, got %d", *coorId)
			}
			return []dto.InspectorResponse{{ID: 1, Code: "INSP-01"}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/inspectors", "")
	rec := httptest.NewRecorder()
	h.GetInspectors(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Inspectors retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetInspectorsWithFilter(t *testing.T) {
	h := NewInspectorHandler(&fakeInspectorService{
		findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
			if coorId == nil || *coorId != 6 {
				t.Errorf("expected coorId 6, got %v", coorId)
			}
			return []dto.InspectorResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/inspectors?coordinatorId=6", "")
	rec := httptest.NewRecorder()
	h.GetInspectors(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetInspectorsInvalidFilter(t *testing.T) {
	called := false
	h := NewInspectorHandler(&fakeInspectorService{
		findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/inspectors?coordinatorId=oops", "")
	rec := httptest.NewRecorder()
	h.GetInspectors(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid coordinatorId")
	}
}

func TestGetInspectorsServiceError(t *testing.T) {
	h := NewInspectorHandler(&fakeInspectorService{
		findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
			return nil, errors.New("boom")
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/inspectors", "")
	rec := httptest.NewRecorder()
	h.GetInspectors(rec, req, nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestCreateInspector(t *testing.T) {
	h := NewInspectorHandler(&fakeInspectorService{
		createFunc: func(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error) {
			if request.InspectorCode != "INSP-02" {
				t.Errorf("unexpected request %+v", request)
			}
			return dto.InspectorJoinResponse{InspectorID: 2}, nil
		},
	})

	body := `{"sesi_code":"SESI-01","coor_code":"KOR-01","inspector_code":"INSP-02","rak":["R1"]}`
	req := newJSONRequest(t, http.MethodPost, "/inspectors", body)
	rec := httptest.NewRecorder()
	h.CreateInspector(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Inspector created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateInspectorInvalidBody(t *testing.T) {
	called := false
	h := NewInspectorHandler(&fakeInspectorService{
		createFunc: func(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error) {
			called = true
			return dto.InspectorJoinResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/inspectors", `{"sesi_code":`)
	rec := httptest.NewRecorder()
	h.CreateInspector(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}
