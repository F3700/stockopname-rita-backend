package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

func newTestStockOpname() *model.StockOpname {
	return &model.StockOpname{
		StockOpnameQuantity: 10,
		StockOpnameRakID:    2,
		SoProductPLU:        "100251",
		SoProductName:       "Sari Roti",
		SoBarcode:           "8991001010016",
		SoBuyPrice:          14500,
		SoSellPrice:         17200,
	}
}

var soSummaryColumns = []string{"id", "barcode", "name", "quantity", "rackName", "inspectorCode", "coordinatorCode", "updatedAt"}

func soTestSummaryRow() *pgxmock.Rows {
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows(soSummaryColumns).
		AddRow(7, "8991001010016", "Sari Roti", 10, "R1", "INSP-01", "KOR-01", updatedAt)
}

func TestStockOpnameFindAll(t *testing.T) {
	mock := mustMockPool(t)
	sesiId, coorId := 3, 5
	mock.ExpectQuery("FROM stock_opname so").WithArgs(&coorId, &sesiId, "%roti%", 10, 0).WillReturnRows(soTestSummaryRow())
	mock.ExpectQuery("COUNT").WithArgs(&coorId, &sesiId, "%roti%").WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	repo := NewStockOpnameRepository(mock)
	results, total, err := repo.FindAll(context.Background(), 10, 0, "roti", &sesiId, &coorId)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(results) != 1 || results[0].Barcode != "8991001010016" || results[0].Name != "Sari Roti" {
		t.Errorf("unexpected results %+v total %d", results, total)
	}
}

func TestStockOpnameFindById(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(7).WillReturnRows(soTestSummaryRow())

	repo := NewStockOpnameRepository(mock)
	result, err := repo.FindById(context.Background(), tx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Id != 7 || result.Barcode != "8991001010016" {
		t.Errorf("unexpected result %+v", result)
	}
}

func TestStockOpnameSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("ON CONFLICT \\(stock_opname_rak_id, so_product_plu\\) DO UPDATE").
		WithArgs(10, 2, "100251", "Sari Roti", "8991001010016", 14500.0, 17200.0).
		WillReturnRows(pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(7))

	repo := NewStockOpnameRepository(mock)
	so := newTestStockOpname()
	if err := repo.Save(context.Background(), tx, so); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if so.StockOpnameID != 7 {
		t.Errorf("expected id 7, got %d", so.StockOpnameID)
	}
}

// Re-uploading an edited rack must upsert the existing (rak_id, plu) row
// instead of failing with a unique-constraint conflict.
func TestStockOpnameSaveUpsertOnReupload(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("ON CONFLICT \\(stock_opname_rak_id, so_product_plu\\) DO UPDATE").
		WithArgs(25, 2, "100251", "Sari Roti", "8991001010016", 14500.0, 17200.0).
		WillReturnRows(pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(7))

	repo := NewStockOpnameRepository(mock)
	so := newTestStockOpname()
	so.StockOpnameQuantity = 25
	if err := repo.Save(context.Background(), tx, so); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if so.StockOpnameID != 7 {
		t.Errorf("expected existing id 7 after upsert, got %d", so.StockOpnameID)
	}
}

func TestStockOpnameFindAllForExport(t *testing.T) {
	mock := mustMockPool(t)
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM stock_opname so").WithArgs(3).WillReturnRows(
		pgxmock.NewRows([]string{"id", "barcode", "name", "buyPrice", "sellPrice", "quantity", "rackName", "inspectorCode", "coordinatorCode", "updatedAt"}).
			AddRow(7, "8991001010016", "Sari Roti", 14500.0, 17200.0, 10, "R1", "INSP-01", "KOR-01", updatedAt),
	)

	repo := NewStockOpnameRepository(mock)
	exports, err := repo.FindAllForExport(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exports) != 1 || exports[0].BuyPrice != 14500 || exports[0].SellPrice != 17200 {
		t.Errorf("unexpected exports %+v", exports)
	}
}
