package repository

import (
	"context"
	"errors"
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
		SoProductDept:       "1138",
		SoBarcode:           "8991001010016",
		SoBuyPrice:          14500,
		SoSellPrice:         17200,
		SoSellFactor:        1,
	}
}

var soSummaryColumns = []string{"id", "barcode", "productPLU", "name", "quantity", "rackId", "rackName", "coordinatorId", "coordinatorCode", "sell_factor", "updatedAt"}

func soTestSummaryRow() *pgxmock.Rows {
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows(soSummaryColumns).
		AddRow(7, "8991001010016", "100251", "Sari Roti", 10, 2, "R1", 4, "KOR-01", 1, updatedAt)
}

func TestStockOpnameFindAll(t *testing.T) {
	mock := mustMockPool(t)
	sesiId, coorId := 3, 5
	mock.ExpectQuery("FROM stock_opname so").WithArgs(&coorId, &sesiId, (*int)(nil), "%roti%", 10, 0).WillReturnRows(soTestSummaryRow())
	mock.ExpectQuery("COUNT").WithArgs(&coorId, &sesiId, (*int)(nil), "%roti%").WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	repo := NewStockOpnameRepository(mock)
	results, total, err := repo.FindAll(context.Background(), 10, 0, "roti", &sesiId, &coorId, nil)
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

func TestStockOpnameSaveMany(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("ON CONFLICT \\(stock_opname_rak_id, so_product_plu\\) DO UPDATE").
		WithArgs(
			[]float64{10, 25},
			[]int32{2, 2},
			[]string{"100251", "100251"},
			[]string{"Sari Roti", "Sari Roti"},
			[]string{"1138", "1138"},
			[]string{"8991001010016", "8991001010016"},
			[]float64{14500, 14500},
			[]float64{17200, 17200},
			[]int16{1, 1},
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))

	repo := NewStockOpnameRepository(mock)
	first := newTestStockOpname()
	second := newTestStockOpname()
	second.StockOpnameQuantity = 25
	if err := repo.SaveMany(context.Background(), tx, []*model.StockOpname{first, second}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// An empty batch must not touch the database at all.
func TestStockOpnameSaveManyEmpty(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)

	repo := NewStockOpnameRepository(mock)
	if err := repo.SaveMany(context.Background(), tx, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStockOpnameFindAllForExport(t *testing.T) {
	mock := mustMockPool(t)
	sesiId := 3
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM stock_opname so").WithArgs(&sesiId, (*int)(nil), 11).WillReturnRows(
		pgxmock.NewRows([]string{"id", "barcode", "ProductPLU", "name", "departmentCode", "buyPrice", "sellPrice", "sell_factor", "per_carton", "quantity", "rackName", "coordinatorCode", "updatedAt"}).
			AddRow(7, "8991001010016", "100251", "Sari Roti", "1138", 14500.0, 17200.0, 1, 12, 10, "R1", "KOR-01", updatedAt),
	)

	repo := NewStockOpnameRepository(mock, WithExportLimit(10))
	exports, err := repo.FindAllForExport(context.Background(), &sesiId, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exports) != 1 || exports[0].BuyPrice != 14500 || exports[0].SellPrice != 17200 || exports[0].ProductPLU != "100251" || exports[0].PerCarton != 12 {
		t.Errorf("unexpected exports %+v", exports)
	}
}

// A result set one row past the limit must be rejected instead of rendered: the
// export is built entirely in memory.
func TestStockOpnameFindAllForExportRejectsOversizedResult(t *testing.T) {
	mock := mustMockPool(t)
	sesiId := 3
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	rows := pgxmock.NewRows([]string{"id", "barcode", "ProductPLU", "name", "departmentCode", "buyPrice", "sellPrice", "sell_factor", "per_carton", "quantity", "rackName", "coordinatorCode", "updatedAt"})
	for i := 0; i < 3; i++ {
		rows.AddRow(i+1, "8991001010016", "100251", "Sari Roti", "1138", 14500.0, 17200.0, 1, 12, 10, "R1", "KOR-01", updatedAt)
	}
	mock.ExpectQuery("FROM stock_opname so").WithArgs(&sesiId, (*int)(nil), 3).WillReturnRows(rows)

	repo := NewStockOpnameRepository(mock, WithExportLimit(2))
	_, err := repo.FindAllForExport(context.Background(), &sesiId, nil)

	var tooLarge *model.TooLargeError
	if !errors.As(err, &tooLarge) {
		t.Fatalf("expected TooLargeError, got %v", err)
	}
	if tooLarge.Limit != 2 {
		t.Errorf("expected the error to carry limit 2, got %d", tooLarge.Limit)
	}
}

// A non-positive limit is the operational escape hatch: no LIMIT clause and no
// truncation check.
func TestStockOpnameFindAllForExportWithoutLimit(t *testing.T) {
	mock := mustMockPool(t)
	sesiId := 3
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM stock_opname so").WithArgs(&sesiId, (*int)(nil)).WillReturnRows(
		pgxmock.NewRows([]string{"id", "barcode", "ProductPLU", "name", "departmentCode", "buyPrice", "sellPrice", "sell_factor", "per_carton", "quantity", "rackName", "coordinatorCode", "updatedAt"}).
			AddRow(7, "8991001010016", "100251", "Sari Roti", "1138", 14500.0, 17200.0, 1, 12, 10, "R1", "KOR-01", updatedAt),
	)

	repo := NewStockOpnameRepository(mock, WithExportLimit(0))
	exports, err := repo.FindAllForExport(context.Background(), &sesiId, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exports) != 1 {
		t.Errorf("expected 1 export, got %d", len(exports))
	}
}

func TestStockOpnameFindSellSummary(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("GROUP BY s.sesi_id").WithArgs(3).WillReturnRows(
		pgxmock.NewRows([]string{"key", "label", "item_count", "total_qty", "total_sell_original", "total_sell_effective"}).
			AddRow("3", "SESI-01", 2, 15.5, 250000.0, 200000.0),
	)

	repo := NewStockOpnameRepository(mock)
	groups, err := repo.FindSellSummary(context.Background(), 3, "session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 || groups[0].Label != "SESI-01" || groups[0].TotalSellEffective != 200000 {
		t.Errorf("unexpected groups %+v", groups)
	}
}
