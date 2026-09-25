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

func productModel() *model.Product {
	return &model.Product{
		ProductID:             1,
		ProductPLU:            "100251",
		ProductName:           "Sari Roti",
		ProductDepartmentCode: "1138",
		ProductBuyPrice:       14500,
		ProductSellPrice:      17200,
		ProductCreatedat:      time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC),
		ProductUpdatedat:      time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC),
		ProductBarcodes:       []string{"1002515550011", "8991001010016"},
	}
}

func TestProductFindAll(t *testing.T) {
	svc := NewProductService(&fakeProductRepository{
		findAllInPageSearchFn: func(ctx context.Context, limit int, offset int, search string) ([]*model.Product, int, error) {
			if limit != 10 || offset != 20 {
				t.Errorf("expected limit 10 offset 20, got limit %d offset %d", limit, offset)
			}
			if search != "roti" {
				t.Errorf("expected search %q, got %q", "roti", search)
			}
			return []*model.Product{productModel()}, 21, nil
		},
	}, nil, nil, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 3, Limit: 10}
	responses, err := svc.FindAll(context.Background(), pagination, "roti")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	got := responses[0]
	if got.Id != 1 || got.PLU != "100251" || got.Name != "Sari Roti" {
		t.Errorf("unexpected response %+v", got)
	}
	if got.Barcode != "1002515550011" || len(got.Barcodes) != 2 {
		t.Errorf("unexpected barcodes %+v", got)
	}
	if got.BuyPrice != 14500 || got.SellPrice != 17200 {
		t.Errorf("unexpected prices %+v", got)
	}
	if got.DepartmentCode != "1138" {
		t.Errorf("unexpected department %+v", got)
	}
	if got.DateCreated != "2026-09-10T08:00:00Z" || got.DateUpdated != "2026-09-11T08:00:00Z" {
		t.Errorf("unexpected dates %+v", got)
	}
	if pagination.TotalItems != 21 {
		t.Errorf("expected TotalItems 21, got %d", pagination.TotalItems)
	}
	if pagination.TotalPages != 3 {
		t.Errorf("expected TotalPages 3, got %d", pagination.TotalPages)
	}
}

func TestProductFindAllRepoError(t *testing.T) {
	svc := NewProductService(&fakeProductRepository{
		findAllInPageSearchFn: func(ctx context.Context, limit int, offset int, search string) ([]*model.Product, int, error) {
			return nil, 0, errors.New("boom")
		},
	}, nil, nil, nil, newTestValidator(t))

	_, err := svc.FindAll(context.Background(), &dto.Pagination{Page: 1, Limit: 10}, "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestProductFindAllUpdatedAfter(t *testing.T) {
	date := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	svc := NewProductService(&fakeProductRepository{
		findAllUpdatedAfterFunc: func(ctx context.Context, got time.Time, limit int, offset int) ([]*model.Product, int, error) {
			if !got.Equal(date) {
				t.Errorf("expected date %v, got %v", date, got)
			}
			return []*model.Product{productModel()}, 1, nil
		},
	}, nil, nil, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 1, Limit: 10}
	responses, err := svc.FindAllUpdatedAfter(context.Background(), pagination, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 || responses[0].Id != 1 {
		t.Errorf("unexpected responses %+v", responses)
	}
	if pagination.TotalItems != 1 || pagination.TotalPages != 1 {
		t.Errorf("unexpected pagination %+v", pagination)
	}
}

func TestProductFindAllLastSession(t *testing.T) {
	sessionDate := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	svc := NewProductService(&fakeProductRepository{
		findAllLastSessionFunc: func(ctx context.Context, limit int) ([]*model.ProductLastSession, error) {
			if limit != 20 {
				t.Errorf("expected limit 20, got %d", limit)
			}
			return []*model.ProductLastSession{
				{Id: 1, Barcode: "1002515550011", Name: "Sari Roti", LastSessionDate: &sessionDate, LastSessionCode: "SESI-01"},
				{Id: 2, Barcode: "", Name: "Krat", LastSessionDate: nil, LastSessionCode: ""},
			}, nil
		},
	}, nil, nil, nil, newTestValidator(t))

	responses, err := svc.FindAllLastSession(context.Background(), 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if responses[0].LastSessionDate != "2026-09-10T08:00:00Z" || responses[0].LastSessionCode != "SESI-01" {
		t.Errorf("unexpected first response %+v", responses[0])
	}
	if responses[1].LastSessionDate != "" || responses[1].Barcode != "" {
		t.Errorf("expected empty date/barcode for nil session, got %+v", responses[1])
	}
}

func TestProductFindAllLastSessionError(t *testing.T) {
	svc := NewProductService(&fakeProductRepository{
		findAllLastSessionFunc: func(ctx context.Context, limit int) ([]*model.ProductLastSession, error) {
			return nil, errors.New("boom")
		},
	}, nil, nil, nil, newTestValidator(t))

	_, err := svc.FindAllLastSession(context.Background(), 20)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestProductCreateValidationFailure(t *testing.T) {
	repoCalled := false
	svc := NewProductService(&fakeProductRepository{
		saveFunc: func(ctx context.Context, tx pgx.Tx, product *model.Product) error {
			repoCalled = true
			return nil
		},
	}, nil, nil, nil, newTestValidator(t))

	_, err := svc.Create(context.Background(), dto.ProductCreateRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if repoCalled {
		t.Error("repository must not be called when validation fails")
	}
}
