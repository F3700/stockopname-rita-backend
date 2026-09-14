package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

var exportColumns = []string{"id", "barcode", "name", "buyPrice", "sellPrice", "quantity", "rackName", "inspectorCode", "coordinatorCode", "updatedAt"}

func TestStockOpnameFindAllForExport(t *testing.T) {
	mock := mustMockPool(t)
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("WHERE c.coor_sesi_id").WithArgs(1).WillReturnRows(
		pgxmock.NewRows(exportColumns).
			AddRow(1, "8991234567890", "Indomie", 2500.0, 3500.0, 48, "R1", "INSP-01", "KOR-01", updatedAt).
			AddRow(2, "8991234567891", "Soto", 2500.0, 3500.0, 36, "R1", "INSP-01", "KOR-01", updatedAt),
	)

	repo := NewStockOpnameRepository(mock)
	results, err := repo.FindAllForExport(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	got := results[0]
	if got.Id != 1 || got.Barcode != "8991234567890" || got.Name != "Indomie" {
		t.Errorf("unexpected result %+v", got)
	}
	if got.BuyPrice != 2500.0 || got.SellPrice != 3500.0 {
		t.Errorf("unexpected prices %+v", got)
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Errorf("unexpected updatedAt %v", got.UpdatedAt)
	}
	if got.Quantity != 48 || got.RackName != "R1" || got.InspectorCode != "INSP-01" || got.CoordinatorCode != "KOR-01" {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestStockOpnameFindAllForExportEmpty(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("WHERE c.coor_sesi_id").WithArgs(99).WillReturnRows(
		pgxmock.NewRows(exportColumns),
	)

	repo := NewStockOpnameRepository(mock)
	results, err := repo.FindAllForExport(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestStockOpnameFindAllForExportQueryError(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("WHERE c.coor_sesi_id").WithArgs(1).WillReturnError(errors.New("boom"))

	repo := NewStockOpnameRepository(mock)
	_, err := repo.FindAllForExport(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
