package service

import (
	"context"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestCoordinatorFindByIdDetail(t *testing.T) {
	svc := NewCoordinatorService(&fakeCoordinatorRepository{
		findByIdDetailFunc: func(ctx context.Context, id int) (*model.CoordinatorDetail, error) {
			if id != 5 {
				t.Errorf("expected id 5, got %d", id)
			}
			return &model.CoordinatorDetail{
				CoordinatorSummary: model.CoordinatorSummary{ID: 5, Code: "KOR-05", Inspector: 1, RackAssigned: 2, RackCompleted: 1, Status: "IN_PROGRESS"},
				SessionCode:        "SESI-01",
				SessionLocation:    "Gudang A",
			}, nil
		},
	}, nil)

	res, err := svc.FindByIdDetail(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 5 || res.SessionCode != "SESI-01" || res.SessionLocation != "Gudang A" {
		t.Errorf("unexpected response %+v", res)
	}
}
