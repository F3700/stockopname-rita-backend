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

func TestCreateInspectorByCoordinatorQR(t *testing.T) {
	h := NewInspectorHandler(&fakeInspectorService{
		findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error) {
			return dto.InspectorJoinResponse{}, nil
		},
		createByCoordinatorQRFunc: func(ctx context.Context, request dto.InspectorJoinByCoordinatorQRRequest) (dto.InspectorJoinResponse, error) {
			if request.CoordinatorQR != "RITA-COOR-12" {
				t.Errorf("unexpected qr %+v", request)
			}
			return dto.InspectorJoinResponse{InspectorID: 9}, nil
		},
	})

	body := `{"coordinator_qr":"RITA-COOR-12","inspector_code":"BUDI","rak":["A-01"]}`
	req := newJSONRequest(t, http.MethodPost, "/stockopname/inspectors/join-by-coordinator-qr", body)
	rec := httptest.NewRecorder()
	h.CreateInspectorByCoordinatorQR(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Inspector created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateInspectorByCoordinatorQRInvalidBody(t *testing.T) {
	called := false
	h := NewInspectorHandler(&fakeInspectorService{
		findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error) {
			return dto.InspectorJoinResponse{}, nil
		},
		createByCoordinatorQRFunc: func(ctx context.Context, request dto.InspectorJoinByCoordinatorQRRequest) (dto.InspectorJoinResponse, error) {
			called = true
			return dto.InspectorJoinResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/stockopname/inspectors/join-by-coordinator-qr", `{"coordinator_qr":`)
	rec := httptest.NewRecorder()
	h.CreateInspectorByCoordinatorQR(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}

func TestCreateInspectorByCoordinatorQRServiceError(t *testing.T) {
	h := NewInspectorHandler(&fakeInspectorService{
		findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error) {
			return dto.InspectorJoinResponse{}, nil
		},
		createByCoordinatorQRFunc: func(ctx context.Context, request dto.InspectorJoinByCoordinatorQRRequest) (dto.InspectorJoinResponse, error) {
			return dto.InspectorJoinResponse{}, &model.ValidationError{Detail: "bad qr"}
		},
	})

	body := `{"coordinator_qr":"BAD","inspector_code":"BUDI","rak":["A-01"]}`
	req := newJSONRequest(t, http.MethodPost, "/stockopname/inspectors/join-by-coordinator-qr", body)
	rec := httptest.NewRecorder()
	h.CreateInspectorByCoordinatorQR(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateInspectorByCoordinatorQRNotFound(t *testing.T) {
	h := NewInspectorHandler(&fakeInspectorService{
		findAllSummaryFunc: func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error) {
			return dto.InspectorJoinResponse{}, nil
		},
		createByCoordinatorQRFunc: func(ctx context.Context, request dto.InspectorJoinByCoordinatorQRRequest) (dto.InspectorJoinResponse, error) {
			return dto.InspectorJoinResponse{}, &model.NotFoundError{Resource: "Coordinator", ID: 99}
		},
	})

	body := `{"coordinator_qr":"RITA-COOR-99","inspector_code":"BUDI","rak":["A-01"]}`
	req := newJSONRequest(t, http.MethodPost, "/stockopname/inspectors/join-by-coordinator-qr", body)
	rec := httptest.NewRecorder()
	h.CreateInspectorByCoordinatorQR(rec, req, nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

var _ = errors.New
