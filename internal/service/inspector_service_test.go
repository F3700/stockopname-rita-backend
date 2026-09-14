package service

import (
	"context"
	"errors"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestInspectorFindAllSummaryWithoutFilter(t *testing.T) {
	byCoorCalled := false
	svc := NewInspectorService(&fakeInspectorRepository{
		findAllFunc: func(ctx context.Context) ([]*model.InspectorSummary, error) {
			return []*model.InspectorSummary{
				{InspectorID: 1, InspectorCode: "INSP-01", RackAssigned: 3, RackCompleted: 2, TotalItems: 40},
			}, nil
		},
		findByCoorIdFunc: func(ctx context.Context, coordinatorId int) ([]*model.InspectorSummary, error) {
			byCoorCalled = true
			return nil, nil
		},
	}, nil, nil, nil)

	responses, err := svc.FindAllSummary(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if byCoorCalled {
		t.Error("FindByCoorId must not be called without filter")
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	got := responses[0]
	if got.ID != 1 || got.Code != "INSP-01" || got.RackAssigned != 3 || got.RackCompleted != 2 || got.TotalItems != 40 {
		t.Errorf("unexpected response %+v", got)
	}
}

func TestInspectorFindAllSummaryWithFilter(t *testing.T) {
	allCalled := false
	svc := NewInspectorService(&fakeInspectorRepository{
		findAllFunc: func(ctx context.Context) ([]*model.InspectorSummary, error) {
			allCalled = true
			return nil, nil
		},
		findByCoorIdFunc: func(ctx context.Context, coordinatorId int) ([]*model.InspectorSummary, error) {
			if coordinatorId != 5 {
				t.Errorf("expected coordinatorId 5, got %d", coordinatorId)
			}
			return []*model.InspectorSummary{}, nil
		},
	}, nil, nil, nil)

	coorId := 5
	responses, err := svc.FindAllSummary(context.Background(), &coorId)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allCalled {
		t.Error("FindAll must not be called with filter")
	}
	if len(responses) != 0 {
		t.Errorf("expected 0 responses, got %d", len(responses))
	}
}

func TestInspectorFindAllSummaryError(t *testing.T) {
	svc := NewInspectorService(&fakeInspectorRepository{
		findAllFunc: func(ctx context.Context) ([]*model.InspectorSummary, error) {
			return nil, errors.New("boom")
		},
	}, nil, nil, nil)

	_, err := svc.FindAllSummary(context.Background(), nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestInspectorFindAllSummaryEmpty(t *testing.T) {
	svc := NewInspectorService(&fakeInspectorRepository{
		findAllFunc: func(ctx context.Context) ([]*model.InspectorSummary, error) {
			return nil, nil
		},
	}, nil, nil, nil)

	responses, err := svc.FindAllSummary(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 0 {
		t.Errorf("expected 0 responses, got %d", len(responses))
	}
}
