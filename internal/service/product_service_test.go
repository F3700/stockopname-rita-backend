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
		ProductSesiID:         3,
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

func fakeSesiForProduct() *fakeSesiRepository {
	return &fakeSesiRepository{
		findByIdFn: func(ctx context.Context, id int) (*model.Sesi, error) {
			return &model.Sesi{SesiID: id, SesiStatus: "IN_PROGRESS"}, nil
		},
	}
}

func TestProductFindAll(t *testing.T) {
	svc := NewProductService(&fakeProductRepository{
		findAllInPageSearchFn: func(ctx context.Context, sesiId int, limit int, offset int, search string) ([]*model.Product, int, error) {
			if sesiId != 3 || limit != 10 || offset != 20 {
				t.Errorf("expected session 3 limit 10 offset 20, got %d %d %d", sesiId, limit, offset)
			}
			if search != "roti" {
				t.Errorf("expected search %q, got %q", "roti", search)
			}
			return []*model.Product{productModel()}, 21, nil
		},
	}, nil, fakeSesiForProduct(), nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 3, Limit: 10}
	responses, err := svc.FindAll(context.Background(), 3, pagination, "roti")
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
		findAllInPageSearchFn: func(ctx context.Context, sesiId int, limit int, offset int, search string) ([]*model.Product, int, error) {
			return nil, 0, errors.New("boom")
		},
	}, nil, fakeSesiForProduct(), nil, newTestValidator(t))

	_, err := svc.FindAll(context.Background(), 3, &dto.Pagination{Page: 1, Limit: 10}, "")
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
