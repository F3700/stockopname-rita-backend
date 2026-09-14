package service

import (
	"context"
	"errors"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestRackFindAllSummary(t *testing.T) {
	svc := NewRackService(&fakeRackRepository{
		findAllFunc: func(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]*model.Rack, error) {
			if inspectorId == nil || *inspectorId != 1 {
				t.Errorf("expected inspectorId 1, got %v", inspectorId)
			}
			if coordinatorId != nil || sessionId != nil {
				t.Errorf("expected nil filters, got %v %v", coordinatorId, sessionId)
			}
			return []*model.Rack{
				{RackID: 1, RackName: "R1", InspectorID: 1},
				{RackID: 2, RackName: "R2", InspectorID: 1},
			}, nil
		},
	}, nil)

	inspectorId := 1
	responses, err := svc.FindAllSummary(context.Background(), &inspectorId, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if responses[0].RackID != 1 || responses[0].RackName != "R1" {
		t.Errorf("unexpected first response %+v", responses[0])
	}
}

func TestRackFindAllSummaryError(t *testing.T) {
	svc := NewRackService(&fakeRackRepository{
		findAllFunc: func(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]*model.Rack, error) {
			return nil, errors.New("boom")
		},
	}, nil)

	_, err := svc.FindAllSummary(context.Background(), nil, nil, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestRackFindProgress(t *testing.T) {
	svc := NewRackService(&fakeRackRepository{
		findProgressFunc: func(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error) {
			if coordinatorId == nil || *coordinatorId != 2 {
				t.Errorf("expected coordinatorId 2, got %v", coordinatorId)
			}
			return []*model.RackProgress{{RackAssigned: 5, RackCompleted: 3, TotalItems: 42}}, nil
		},
	}, nil)

	coorId := 2
	responses, err := svc.FindProgress(context.Background(), &coorId, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	got := responses[0]
	if got.RackAssigned != 5 || got.RackCompleted != 3 || got.TotalItems != 42 {
		t.Errorf("unexpected response %+v", got)
	}
}

func TestRackFindProgressError(t *testing.T) {
	svc := NewRackService(&fakeRackRepository{
		findProgressFunc: func(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error) {
			return nil, errors.New("boom")
		},
	}, nil)

	_, err := svc.FindProgress(context.Background(), nil, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
