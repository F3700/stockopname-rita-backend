package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/dto"
)

func TestGetRackProgress(t *testing.T) {
	h := NewRackHandler(&fakeRackService{
		findProgressFunc: func(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error) {
			if sessionId == nil || *sessionId != 2 {
				t.Errorf("expected sessionId 2, got %v", sessionId)
			}
			if coordinatorId != nil {
				t.Errorf("expected nil coordinatorId, got %d", *coordinatorId)
			}
			return []dto.RackProgressResponse{{RackAssigned: 5, RackCompleted: 3}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/racks/progress?sessionId=2", "")
	rec := httptest.NewRecorder()
	h.GetRackProgress(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Rack progress retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetRackProgressInvalidSession(t *testing.T) {
	called := false
	h := NewRackHandler(&fakeRackService{
		findProgressFunc: func(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/racks/progress?sessionId=bad", "")
	rec := httptest.NewRecorder()
	h.GetRackProgress(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid sessionId")
	}
}

func TestGetRackProgressInvalidCoordinator(t *testing.T) {
	called := false
	h := NewRackHandler(&fakeRackService{
		findProgressFunc: func(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/racks/progress?coordinatorId=bad", "")
	rec := httptest.NewRecorder()
	h.GetRackProgress(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid coordinatorId")
	}
}

func TestGetRackProgressServiceError(t *testing.T) {
	h := NewRackHandler(&fakeRackService{
		findProgressFunc: func(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error) {
			return nil, errors.New("boom")
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/racks/progress", "")
	rec := httptest.NewRecorder()
	h.GetRackProgress(rec, req, nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestGetRacks(t *testing.T) {
	h := NewRackHandler(&fakeRackService{
		findAllSummaryFunc: func(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error) {
			if inspectorId == nil || *inspectorId != 1 {
				t.Errorf("expected inspectorId 1, got %v", inspectorId)
			}
			if coordinatorId == nil || *coordinatorId != 2 {
				t.Errorf("expected coordinatorId 2, got %v", coordinatorId)
			}
			if sessionId == nil || *sessionId != 3 {
				t.Errorf("expected sessionId 3, got %v", sessionId)
			}
			return []dto.RackResponse{{RackID: 1, RackName: "R1"}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/racks?inspectorId=1&coordinatorId=2&sessionId=3", "")
	rec := httptest.NewRecorder()
	h.GetRacks(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Racks retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetRacksInvalidInspector(t *testing.T) {
	called := false
	h := NewRackHandler(&fakeRackService{
		findAllSummaryFunc: func(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/racks?inspectorId=bad", "")
	rec := httptest.NewRecorder()
	h.GetRacks(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid inspectorId")
	}
}

func TestGetRacksInvalidCoordinator(t *testing.T) {
	called := false
	h := NewRackHandler(&fakeRackService{
		findAllSummaryFunc: func(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/racks?coordinatorId=bad", "")
	rec := httptest.NewRecorder()
	h.GetRacks(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid coordinatorId")
	}
}

func TestGetRacksInvalidSession(t *testing.T) {
	called := false
	h := NewRackHandler(&fakeRackService{
		findAllSummaryFunc: func(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/racks?sessionId=bad", "")
	rec := httptest.NewRecorder()
	h.GetRacks(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid sessionId")
	}
}

func TestCreateRack(t *testing.T) {
	h := NewRackHandler(&fakeRackService{
		createFunc: func(ctx context.Context, request dto.CreateRackRequest) (dto.RackResponse, error) {
			if request.RackName != "R9" || request.InspectorID != 3 {
				t.Errorf("unexpected request %+v", request)
			}
			return dto.RackResponse{RackID: 9, RackName: "R9"}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/racks", `{"inspector_id":3,"rak_name":"R9"}`)
	rec := httptest.NewRecorder()
	h.CreateRack(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Rack created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateRackInvalidBody(t *testing.T) {
	called := false
	h := NewRackHandler(&fakeRackService{
		createFunc: func(ctx context.Context, request dto.CreateRackRequest) (dto.RackResponse, error) {
			called = true
			return dto.RackResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/racks", `{"rak_name":`)
	rec := httptest.NewRecorder()
	h.CreateRack(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}
