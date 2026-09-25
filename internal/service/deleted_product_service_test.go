package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"stockopname-rita-backend/internal/model"
)

func TestDeletedProductDelete(t *testing.T) {
	svc := NewDeletedProductService(&fakeDeletedProductRepository{
		deleteFunc: func(ctx context.Context) error { return nil },
	})

	if err := svc.Delete(context.Background()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeletedProductDeleteError(t *testing.T) {
	svc := NewDeletedProductService(&fakeDeletedProductRepository{
		deleteFunc: func(ctx context.Context) error { return errors.New("boom") },
	})

	if err := svc.Delete(context.Background()); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestDeletedProductFindAllWithoutDate(t *testing.T) {
	afterCalled := false
	svc := NewDeletedProductService(&fakeDeletedProductRepository{
		findAllFunc: func(ctx context.Context) ([]*model.DeletedProduct, error) {
			return []*model.DeletedProduct{{ProductPLU: "100251"}, {ProductPLU: "100252"}}, nil
		},
		findAllUpdatedAfterFunc: func(ctx context.Context, date time.Time) ([]*model.DeletedProduct, error) {
			afterCalled = true
			return nil, nil
		},
	})

	responses, err := svc.FindAll(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if afterCalled {
		t.Error("FindAllUpdatedAfter must not be called without date")
	}
	if len(responses) != 2 || responses[0].ProductPLU != "100251" || responses[1].ProductPLU != "100252" {
		t.Errorf("unexpected responses %+v", responses)
	}
}

func TestDeletedProductFindAllWithDate(t *testing.T) {
	allCalled := false
	date := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	svc := NewDeletedProductService(&fakeDeletedProductRepository{
		findAllFunc: func(ctx context.Context) ([]*model.DeletedProduct, error) {
			allCalled = true
			return nil, nil
		},
		findAllUpdatedAfterFunc: func(ctx context.Context, got time.Time) ([]*model.DeletedProduct, error) {
			if !got.Equal(date) {
				t.Errorf("expected date %v, got %v", date, got)
			}
			return []*model.DeletedProduct{{ProductPLU: "100251"}}, nil
		},
	})

	responses, err := svc.FindAll(context.Background(), &date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allCalled {
		t.Error("FindAll must not be called with date")
	}
	if len(responses) != 1 || responses[0].ProductPLU != "100251" {
		t.Errorf("unexpected responses %+v", responses)
	}
}

func TestDeletedProductFindAllError(t *testing.T) {
	svc := NewDeletedProductService(&fakeDeletedProductRepository{
		findAllFunc: func(ctx context.Context) ([]*model.DeletedProduct, error) {
			return nil, errors.New("boom")
		},
	})

	_, err := svc.FindAll(context.Background(), nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
