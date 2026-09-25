package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func soSummaryModel() *model.StockOpnameSummary {
	return &model.StockOpnameSummary{
		Id:              7,
		Barcode:         "8991001010016",
		Name:            "Sari Roti",
		Quantity:        10,
		RackName:        "R1",
		InspectorCode:   "INSP-01",
		CoordinatorCode: "KOR-01",
		UpdatedAt:       time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC),
	}
}

func TestStockOpnameFindAll(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllFunc: func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error) {
			if limit != 10 || offset != 0 || search != "roti" {
				t.Errorf("unexpected args limit %d offset %d search %q", limit, offset, search)
			}
			return []*model.StockOpnameSummary{soSummaryModel()}, 1, nil
		},
	}, &fakeProductRepository{}, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 1, Limit: 10}
	responses, err := svc.FindAll(context.Background(), pagination, "roti", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 || responses[0].Barcode != "8991001010016" || responses[0].Name != "Sari Roti" {
		t.Errorf("unexpected responses %+v", responses)
	}
	if pagination.TotalItems != 1 || pagination.TotalPages != 1 {
		t.Errorf("unexpected pagination %+v", pagination)
	}
}

func TestStockOpnameFindAllRepoError(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllFunc: func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error) {
			return nil, 0, errors.New("boom")
		},
	}, &fakeProductRepository{}, nil, newTestValidator(t))

	_, err := svc.FindAll(context.Background(), &dto.Pagination{Page: 1, Limit: 10}, "", nil, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestStockOpnameFindAllForExport(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllForExportFunc: func(ctx context.Context, sesiId int) ([]*model.StockOpnameExport, error) {
			if sesiId != 3 {
				t.Errorf("expected sesi 3, got %d", sesiId)
			}
			return []*model.StockOpnameExport{
				{Id: 7, Barcode: "8991001010016", Name: "Sari Roti", BuyPrice: 14500, SellPrice: 17200, Quantity: 10, RackName: "R1", InspectorCode: "INSP-01", CoordinatorCode: "KOR-01", UpdatedAt: time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)},
			}, nil
		},
	}, &fakeProductRepository{}, nil, newTestValidator(t))

	responses, err := svc.FindAllForExport(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 || responses[0].BuyPrice != 14500 || responses[0].SellPrice != 17200 {
		t.Errorf("unexpected responses %+v", responses)
	}
}
