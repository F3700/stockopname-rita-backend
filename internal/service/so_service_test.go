package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func TestStockOpnameFindAll(t *testing.T) {
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllFunc: func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error) {
			if limit != 10 || offset != 10 {
				t.Errorf("expected limit 10 offset 10, got limit %d offset %d", limit, offset)
			}
			if search != "indomie" {
				t.Errorf("expected search %q, got %q", "indomie", search)
			}
			if sesiId == nil || *sesiId != 2 {
				t.Errorf("expected sesiId 2, got %v", sesiId)
			}
			if coorId == nil || *coorId != 1 {
				t.Errorf("expected coorId 1, got %v", coorId)
			}
			return []*model.StockOpnameSummary{
				{Id: 1, Barcode: "8991234567890", Name: "Indomie", Quantity: 48, RackName: "R1", InspectorCode: "INSP-01", CoordinatorCode: "KOR-01", UpdatedAt: updatedAt},
				{Id: 2, Barcode: "8991234567891", Name: "Soto", Quantity: 36, RackName: "R1", InspectorCode: "INSP-01", CoordinatorCode: "KOR-01", UpdatedAt: updatedAt},
			}, 12, nil
		},
	}, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 2, Limit: 10}
	sesiId, coorId := 2, 1
	responses, err := svc.FindAll(context.Background(), pagination, "indomie", &coorId, &sesiId)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if responses[0].Barcode != "8991234567890" || responses[0].Quantity != 48 {
		t.Errorf("unexpected first response %+v", responses[0])
	}
	if responses[0].UpdatedAt != updatedAt.Format(time.RFC3339) {
		t.Errorf("unexpected updatedAt %q", responses[0].UpdatedAt)
	}
	if pagination.TotalItems != 12 {
		t.Errorf("expected TotalItems 12, got %d", pagination.TotalItems)
	}
	if pagination.TotalPages != 2 {
		t.Errorf("expected TotalPages 2, got %d", pagination.TotalPages)
	}
}

func TestStockOpnameFindAllEmpty(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllFunc: func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error) {
			return nil, 0, nil
		},
	}, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 1, Limit: 10}
	responses, err := svc.FindAll(context.Background(), pagination, "", nil, nil)
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

func TestStockOpnameFindAllRepoError(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllFunc: func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error) {
			return nil, 0, errors.New("boom")
		},
	}, nil, newTestValidator(t))

	_, err := svc.FindAll(context.Background(), &dto.Pagination{Page: 1, Limit: 10}, "", nil, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestStockOpnameCreateValidationFailure(t *testing.T) {
	repoCalled := false
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		saveFunc: func(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
			repoCalled = true
			return nil
		},
	}, nil, newTestValidator(t))

	_, err := svc.Create(context.Background(), dto.CreateStockOpnameRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if repoCalled {
		t.Error("repository must not be called when validation fails")
	}
}

func TestStockOpnameCreateByRackValidationFailure(t *testing.T) {
	repoCalled := false
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		saveFunc: func(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
			repoCalled = true
			return nil
		},
	}, nil, newTestValidator(t))

	err := svc.CreateByRack(context.Background(), dto.CreateStockOpnameByRackRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if repoCalled {
		t.Error("repository must not be called when validation fails")
	}
}
