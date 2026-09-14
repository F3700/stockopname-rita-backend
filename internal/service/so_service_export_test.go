package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"stockopname-rita-backend/internal/model"
)

func TestStockOpnameFindAllForExport(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllForExportFunc: func(ctx context.Context, sesiId int) ([]*model.StockOpnameExport, error) {
			if sesiId != 1 {
				t.Errorf("expected sesiId 1, got %d", sesiId)
			}
			return []*model.StockOpnameExport{
				{Id: 1, Barcode: "8991234567890", Name: "Indomie", BuyPrice: 2500, SellPrice: 3500, Quantity: 48, RackName: "R1", InspectorCode: "INSP-01", CoordinatorCode: "KOR-01", UpdatedAt: time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)},
			}, nil
		},
	}, nil, newTestValidator(t))

	responses, err := svc.FindAllForExport(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	got := responses[0]
	if got.Id != 1 || got.Barcode != "8991234567890" || got.Name != "Indomie" {
		t.Errorf("unexpected response %+v", got)
	}
	if got.BuyPrice != 2500 || got.SellPrice != 3500 {
		t.Errorf("unexpected prices %+v", got)
	}
	if got.Quantity != 48 || got.RackName != "R1" || got.InspectorCode != "INSP-01" || got.CoordinatorCode != "KOR-01" {
		t.Errorf("unexpected response %+v", got)
	}
	if got.UpdatedAt != "2026-09-10T08:00:00Z" {
		t.Errorf("unexpected updatedAt %q", got.UpdatedAt)
	}
}

func TestStockOpnameFindAllForExportEmpty(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllForExportFunc: func(ctx context.Context, sesiId int) ([]*model.StockOpnameExport, error) {
			return nil, nil
		},
	}, nil, newTestValidator(t))

	responses, err := svc.FindAllForExport(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 0 {
		t.Errorf("expected 0 responses, got %d", len(responses))
	}
}

func TestStockOpnameFindAllForExportRepoError(t *testing.T) {
	svc := NewStockOpnameService(&fakeStockOpnameRepository{
		findAllForExportFunc: func(ctx context.Context, sesiId int) ([]*model.StockOpnameExport, error) {
			return nil, errors.New("boom")
		},
	}, nil, newTestValidator(t))

	_, err := svc.FindAllForExport(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
