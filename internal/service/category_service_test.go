package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func TestCategoryFindAll(t *testing.T) {
	svc := NewCategoryService(&fakeCategoryRepository{
		findAllInPageSearchFn: func(ctx context.Context, limit int, offset int, search string) ([]*model.Category, int, error) {
			if limit != 10 || offset != 0 {
				t.Errorf("expected limit 10 offset 0, got limit %d offset %d", limit, offset)
			}
			if search != "mak" {
				t.Errorf("expected search %q, got %q", "mak", search)
			}
			return []*model.Category{
				{CategoryID: 1, CategoryName: "Makanan", CategoryDescription: "desc"},
			}, 1, nil
		},
	}, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 1, Limit: 10}
	responses, err := svc.FindAll(context.Background(), pagination, "mak")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	got := responses[0]
	if got.CategoryID != 1 || got.CategoryName != "Makanan" || got.CategoryDescription != "desc" {
		t.Errorf("unexpected response %+v", got)
	}
	if pagination.TotalItems != 1 || pagination.TotalPages != 1 {
		t.Errorf("unexpected pagination %+v", pagination)
	}
}

func TestCategoryFindAllEmpty(t *testing.T) {
	svc := NewCategoryService(&fakeCategoryRepository{
		findAllInPageSearchFn: func(ctx context.Context, limit int, offset int, search string) ([]*model.Category, int, error) {
			return nil, 0, nil
		},
	}, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 1, Limit: 10}
	responses, err := svc.FindAll(context.Background(), pagination, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 0 {
		t.Errorf("expected 0 responses, got %d", len(responses))
	}
	if pagination.TotalItems != 0 || pagination.TotalPages != 0 {
		t.Errorf("unexpected pagination %+v", pagination)
	}
}

func TestCategoryFindAllRepoError(t *testing.T) {
	svc := NewCategoryService(&fakeCategoryRepository{
		findAllInPageSearchFn: func(ctx context.Context, limit int, offset int, search string) ([]*model.Category, int, error) {
			return nil, 0, errors.New("boom")
		},
	}, nil, newTestValidator(t))

	_, err := svc.FindAll(context.Background(), &dto.Pagination{Page: 1, Limit: 10}, "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCategoryCreateValidationFailure(t *testing.T) {
	repoCalled := false
	svc := NewCategoryService(&fakeCategoryRepository{
		saveFunc: func(ctx context.Context, tx pgx.Tx, category *model.Category) error {
			repoCalled = true
			return nil
		},
	}, nil, newTestValidator(t))

	_, err := svc.Create(context.Background(), dto.CategoryCreateRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if repoCalled {
		t.Error("repository must not be called when validation fails")
	}
}
