package service

import (
	"context"
	"errors"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestRackFindAllSummary(t *testing.T) {
	svc := NewRackService(&fakeRackRepository{
		findAllFunc: func(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackState, error) {
			if coordinatorId == nil || *coordinatorId != 1 {
				t.Errorf("expected coordinatorId 1, got %v", coordinatorId)
			}
			if sessionId != nil {
				t.Errorf("expected nil sessionId, got %v", sessionId)
			}
			return []*model.RackState{
				{RakID: 1, RakName: "R1", CoorID: 1},
				{RakID: 2, RakName: "R2", CoorID: 1},
			}, nil
		},
	}, nil, nil, nil, nil)

	coordinatorId := 1
	responses, err := svc.FindAllSummary(context.Background(), &coordinatorId, nil)
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
		findAllFunc: func(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackState, error) {
			return nil, errors.New("boom")
		},
	}, nil, nil, nil, nil)

	_, err := svc.FindAllSummary(context.Background(), nil, nil)
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
	}, nil, nil, nil, nil)

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
	}, nil, nil, nil, nil)

	_, err := svc.FindProgress(context.Background(), nil, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
