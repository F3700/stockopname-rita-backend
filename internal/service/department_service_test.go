package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func TestDepartmentFindAll(t *testing.T) {
	svc := NewDepartmentService(&fakeDepartmentRepository{
		findAllInPageSearchFn: func(ctx context.Context, limit int, offset int, search string) ([]*model.Department, int, error) {
			if limit != 10 || offset != 0 {
				t.Errorf("expected limit 10 offset 0, got limit %d offset %d", limit, offset)
			}
			if search != "gudang" {
				t.Errorf("expected search %q, got %q", "gudang", search)
			}
			return []*model.Department{
				{DepartmentID: 1, DepartmentCode: "D01", DepartmentName: "Gudang", DepartmentDesc: "desc"},
			}, 1, nil
		},
	}, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 1, Limit: 10}
	responses, err := svc.FindAll(context.Background(), pagination, "gudang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	got := responses[0]
	if got.ID != 1 || got.Code != "D01" || got.Name != "Gudang" || got.Description != "desc" {
		t.Errorf("unexpected response %+v", got)
	}
	if pagination.TotalItems != 1 || pagination.TotalPages != 1 {
		t.Errorf("unexpected pagination %+v", pagination)
	}
}

func TestDepartmentFindAllRepoError(t *testing.T) {
	svc := NewDepartmentService(&fakeDepartmentRepository{
		findAllInPageSearchFn: func(ctx context.Context, limit int, offset int, search string) ([]*model.Department, int, error) {
			return nil, 0, errors.New("boom")
		},
	}, nil, newTestValidator(t))

	_, err := svc.FindAll(context.Background(), &dto.Pagination{Page: 1, Limit: 10}, "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestDepartmentCreateValidationFailure(t *testing.T) {
	repoCalled := false
	svc := NewDepartmentService(&fakeDepartmentRepository{
		saveFunc: func(ctx context.Context, tx pgx.Tx, department *model.Department) error {
			repoCalled = true
			return nil
		},
	}, nil, newTestValidator(t))

	_, err := svc.Create(context.Background(), dto.DepartmentCreateRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if repoCalled {
		t.Error("repository must not be called when validation fails")
	}
}
