package service

import (
	"context"
	"errors"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestCoordinatorFindAllSummaryWithoutFilter(t *testing.T) {
	bySesiCalled := false
	svc := NewCoordinatorService(&fakeCoordinatorRepository{
		findAllSummaryFunc: func(ctx context.Context) ([]*model.CoordinatorSummary, error) {
			return []*model.CoordinatorSummary{
				{ID: 1, Code: "KOR-01", Inspector: 2, RackAssigned: 5, RackCompleted: 3, Status: "IN_PROGRESS"},
			}, nil
		},
		findBySesiIdSummaryFn: func(ctx context.Context, sesiId int) ([]*model.CoordinatorSummary, error) {
			bySesiCalled = true
			return nil, nil
		},
	}, nil)

	responses, err := svc.FindAllSummary(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bySesiCalled {
		t.Error("FindBySesiIdSummary must not be called without filter")
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	got := responses[0]
	if got.ID != 1 || got.Code != "KOR-01" || got.Inspector != 2 || got.RackAssigned != 5 || got.RackCompleted != 3 || got.Status != "IN_PROGRESS" {
		t.Errorf("unexpected response %+v", got)
	}
}

func TestCoordinatorFindAllSummaryWithFilter(t *testing.T) {
	allCalled := false
	svc := NewCoordinatorService(&fakeCoordinatorRepository{
		findAllSummaryFunc: func(ctx context.Context) ([]*model.CoordinatorSummary, error) {
			allCalled = true
			return nil, nil
		},
		findBySesiIdSummaryFn: func(ctx context.Context, sesiId int) ([]*model.CoordinatorSummary, error) {
			if sesiId != 4 {
				t.Errorf("expected sesiId 4, got %d", sesiId)
			}
			return []*model.CoordinatorSummary{}, nil
		},
	}, nil)

	sesiId := 4
	responses, err := svc.FindAllSummary(context.Background(), &sesiId)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allCalled {
		t.Error("FindAllSummary must not be called with filter")
	}
	if len(responses) != 0 {
		t.Errorf("expected 0 responses, got %d", len(responses))
	}
}

func TestCoordinatorFindAllSummaryError(t *testing.T) {
	svc := NewCoordinatorService(&fakeCoordinatorRepository{
		findAllSummaryFunc: func(ctx context.Context) ([]*model.CoordinatorSummary, error) {
			return nil, errors.New("boom")
		},
	}, nil)

	_, err := svc.FindAllSummary(context.Background(), nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCoordinatorFindByIdSummary(t *testing.T) {
	svc := NewCoordinatorService(&fakeCoordinatorRepository{
		findByIdSummaryFunc: func(ctx context.Context, id int) (*model.CoordinatorSummary, error) {
			if id != 3 {
				t.Errorf("expected id 3, got %d", id)
			}
			return &model.CoordinatorSummary{ID: 3, Code: "KOR-03", Status: "COMPLETED"}, nil
		},
	}, nil)

	res, err := svc.FindByIdSummary(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 3 || res.Code != "KOR-03" || res.Status != "COMPLETED" {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestCoordinatorFindByIdSummaryNotFound(t *testing.T) {
	svc := NewCoordinatorService(&fakeCoordinatorRepository{
		findByIdSummaryFunc: func(ctx context.Context, id int) (*model.CoordinatorSummary, error) {
			return nil, &model.NotFoundError{Resource: "Coordinator", ID: id}
		},
	}, nil)

	_, err := svc.FindByIdSummary(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError to propagate, got %T: %v", err, err)
	}
}

func TestCoordinatorFindByIdReport(t *testing.T) {
	svc := NewCoordinatorService(&fakeCoordinatorRepository{
		findByIdReportFunc: func(ctx context.Context, id int) (*model.CoordinatorReport, error) {
			if id != 2 {
				t.Errorf("expected id 2, got %d", id)
			}
			return &model.CoordinatorReport{
				CoordinatorSummary: model.CoordinatorSummary{ID: 2, Code: "KOR-02", Inspector: 1, RackAssigned: 4, RackCompleted: 4, Status: "COMPLETED"},
				SessionCode:        "SESI-01",
			}, nil
		},
	}, nil)

	res, err := svc.FindByIdReport(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SessionCode != "SESI-01" {
		t.Errorf("expected session code SESI-01, got %q", res.SessionCode)
	}
	if res.Coordinator.ID != 2 || res.Coordinator.Code != "KOR-02" {
		t.Errorf("unexpected coordinator %+v", res.Coordinator)
	}
}
