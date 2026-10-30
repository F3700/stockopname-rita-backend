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
		CoordinatorCode: "KOR-01",
		UpdatedAt:       time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC),
	}
}

func TestStockOpnameFindAll(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllFunc: func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int, rakId *int) ([]*model.StockOpnameSummary, int, error) {
			if limit != 10 || offset != 0 || search != "roti" {
				t.Errorf("unexpected args limit %d offset %d search %q", limit, offset, search)
			}
			return []*model.StockOpnameSummary{soSummaryModel()}, 1, nil
		},
	}, &fakeProductRepository{}, &fakeSesiRepository{}, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 1, Limit: 10}
	responses, err := svc.FindAll(context.Background(), pagination, "roti", nil, nil, nil)
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
		findAllFunc: func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int, rakId *int) ([]*model.StockOpnameSummary, int, error) {
			return nil, 0, errors.New("boom")
		},
	}, &fakeProductRepository{}, &fakeSesiRepository{}, nil, newTestValidator(t))

	_, err := svc.FindAll(context.Background(), &dto.Pagination{Page: 1, Limit: 10}, "", nil, nil, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestStockOpnameFindAllForExport(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllForExportFunc: func(ctx context.Context, sesiId *int, coorId *int) ([]*model.StockOpnameExport, error) {
			if sesiId == nil || *sesiId != 3 {
				t.Errorf("expected sesi 3, got %v", sesiId)
			}
			if coorId != nil {
				t.Errorf("expected nil coorId, got %d", *coorId)
			}
			return []*model.StockOpnameExport{
				{Id: 7, Barcode: "8991001010016", Name: "Sari Roti", BuyPrice: 14500, SellPrice: 17200, Quantity: 10, RackName: "R1", CoordinatorCode: "KOR-01", UpdatedAt: time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)},
			}, nil
		},
	}, &fakeProductRepository{}, &fakeSesiRepository{}, nil, newTestValidator(t))

	sesiId := 3
	responses, err := svc.FindAllForExport(context.Background(), &sesiId, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 || responses[0].BuyPrice != 14500 || responses[0].SellPrice != 17200 {
		t.Errorf("unexpected responses %+v", responses)
	}
}

func TestStockOpnameQuantityValidation(t *testing.T) {
	validate := newTestValidator(t)

	tests := []struct {
		name    string
		request any
		wantErr bool
	}{
		{"zero create", dto.CreateStockOpnameRequest{Quantity: qtyPtr(0), Barcode: "x", RakID: 1}, false},
		{"missing create", dto.CreateStockOpnameRequest{Barcode: "x", RakID: 1}, true},
		{"negative create", dto.CreateStockOpnameRequest{Quantity: qtyPtr(-1), Barcode: "x", RakID: 1}, true},
		{"zero update", dto.UpdateStockOpnameRequest{Quantity: qtyPtr(0)}, false},
		{"missing update", dto.UpdateStockOpnameRequest{}, true},
		{"negative update", dto.UpdateStockOpnameRequest{Quantity: qtyPtr(-1)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
