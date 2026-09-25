package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

var productTestColumns = []string{"product_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}

func productTestRow() *pgxmock.Rows {
	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows(productTestColumns).
		AddRow(11, "100251", "Sari Roti", "1138", 14500.0, 17200.0, ts, ts)
}

func barcodeAttachRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"barcode_product_id", "barcode_code"}).
		AddRow(11, "1002515550011").
		AddRow(11, "8991001010016")
}

func TestProductSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO product").
		WithArgs("100251", "Sari Roti", "1138", 14500.0, 17200.0).
		WillReturnRows(pgxmock.NewRows([]string{"product_id"}).AddRow(11))

	repo := NewProductRepository(mock)
	product := &model.Product{ProductPLU: "100251", ProductName: "Sari Roti", ProductDepartmentCode: "1138", ProductBuyPrice: 14500, ProductSellPrice: 17200}
	if err := repo.Save(context.Background(), tx, product); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product.ProductID != 11 {
		t.Errorf("expected product id 11, got %d", product.ProductID)
	}
}

func TestProductFindById(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE p.product_id").WithArgs(11).WillReturnRows(productTestRow())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(barcodeAttachRows())

	repo := NewProductRepository(mock)
	product, err := repo.FindById(context.Background(), tx, 11)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product.ProductPLU != "100251" || product.ProductName != "Sari Roti" || product.ProductDepartmentCode != "1138" {
		t.Errorf("unexpected product %+v", product)
	}
	if len(product.ProductBarcodes) != 2 || product.PrimaryBarcode() != "1002515550011" {
		t.Errorf("unexpected barcodes %+v", product.ProductBarcodes)
	}
}

func TestProductFindByIdNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE p.product_id").WithArgs(99).WillReturnRows(pgxmock.NewRows(productTestColumns))

	repo := NewProductRepository(mock)
	_, err := repo.FindById(context.Background(), tx, 99)
	if _, ok := err.(*model.NotFoundError); !ok {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestProductFindByPLU(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE p.product_plu").WithArgs("100251").WillReturnRows(productTestRow())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(barcodeAttachRows())

	repo := NewProductRepository(mock)
	product, err := repo.FindByPLU(context.Background(), tx, "100251")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product.ProductID != 11 || len(product.ProductBarcodes) != 2 {
		t.Errorf("unexpected product %+v", product)
	}
}

func TestProductFindByBarcode(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("JOIN barcode").WithArgs("8991001010016").WillReturnRows(productTestRow())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(barcodeAttachRows())

	repo := NewProductRepository(mock)
	product, err := repo.FindByBarcode(context.Background(), tx, "8991001010016")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product.ProductPLU != "100251" {
		t.Errorf("unexpected product %+v", product)
	}
}

func TestProductFindAllForImport(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("FROM product p").WillReturnRows(productTestRow())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(barcodeAttachRows())

	repo := NewProductRepository(mock)
	products, err := repo.FindAllForImport(context.Background(), tx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 1 || len(products[0].ProductBarcodes) != 2 {
		t.Errorf("unexpected products %+v", products)
	}
}

func TestProductFindAllInPageSearch(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM product p").WithArgs("%roti%", 10, 0).WillReturnRows(productTestRow())
	mock.ExpectQuery("COUNT").WithArgs("%roti%").WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(barcodeAttachRows())

	repo := NewProductRepository(mock)
	results, total, err := repo.FindAllInPageSearch(context.Background(), 10, 0, "roti")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(results) != 1 || results[0].ProductPLU != "100251" {
		t.Errorf("unexpected results %+v total %d", results, total)
	}
}

func TestProductFindAllUpdatedAfter(t *testing.T) {
	mock := mustMockPool(t)
	date := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("product_updatedat >").WithArgs(date, 10, 0).WillReturnRows(productTestRow())
	mock.ExpectQuery("COUNT").WithArgs(date).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(barcodeAttachRows())

	repo := NewProductRepository(mock)
	results, total, err := repo.FindAllUpdatedAfter(context.Background(), date, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(results) != 1 || results[0].PrimaryBarcode() != "1002515550011" {
		t.Errorf("unexpected results %+v total %d", results, total)
	}
}

func TestProductFindAllLastSession(t *testing.T) {
	mock := mustMockPool(t)
	sessionDate := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("DISTINCT ON").WithArgs(20).WillReturnRows(
		pgxmock.NewRows([]string{"product_id", "barcode", "name", "lastSessionDate", "lastSessionCode"}).
			AddRow(11, "1002515550011", "Sari Roti", sessionDate, "SESI-01").
			AddRow(12, nil, "Krat", nil, ""),
	)

	repo := NewProductRepository(mock)
	results, err := repo.FindAllLastSession(context.Background(), 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 || results[0].Barcode != "1002515550011" {
		t.Errorf("unexpected results %+v", results)
	}
	if results[1].Barcode != "" {
		t.Errorf("expected empty barcode for NULL, got %q", results[1].Barcode)
	}
}

func TestProductUpdate(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("UPDATE product").
		WithArgs("100251", "Sari Roti Baru", "1138", 14500.0, 17200.0, 11).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := NewProductRepository(mock)
	product := &model.Product{ProductID: 11, ProductPLU: "100251", ProductName: "Sari Roti Baru", ProductDepartmentCode: "1138", ProductBuyPrice: 14500, ProductSellPrice: 17200}
	if err := repo.Update(context.Background(), tx, product); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductDelete(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM product").WithArgs(11).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := NewProductRepository(mock)
	if err := repo.Delete(context.Background(), tx, 11); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductDeleteNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM product").WithArgs(99).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewProductRepository(mock)
	if err := repo.Delete(context.Background(), tx, 99); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestProductClear(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("TRUNCATE product, barcode, deleted_product").WillReturnResult(pgxmock.NewResult("TRUNCATE", 0))

	repo := NewProductRepository(mock)
	if err := repo.Clear(context.Background(), tx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
