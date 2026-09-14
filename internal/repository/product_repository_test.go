package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

var productColumns = []string{"product_id", "product_barcode", "product_name", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat", "category_name", "department_code"}

func productRow() []any {
	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return []any{1, "8991234567890", "Indomie", 2500.0, 3500.0, ts, ts, "Makanan", "D01"}
}

func TestProductDelete(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM product").WithArgs(1).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := NewProductRepository(mock)
	if err := repo.Delete(context.Background(), tx, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductDeleteNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM product").WithArgs(99).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewProductRepository(mock)
	err := repo.Delete(context.Background(), tx, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestProductFindAll(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM product p").WillReturnRows(
		pgxmock.NewRows(productColumns).AddRow(productRow()...),
	)

	repo := NewProductRepository(mock)
	results, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].ProductBarcode != "8991234567890" {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestProductFindAllUpdatedAfter(t *testing.T) {
	mock := mustMockPool(t)
	date := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("product_updatedat >").WithArgs(pgxmock.AnyArg(), 10, 0).WillReturnRows(
		pgxmock.NewRows(productColumns).AddRow(productRow()...),
	)
	mock.ExpectQuery("SELECT COUNT").WithArgs(pgxmock.AnyArg()).WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(1),
	)

	repo := NewProductRepository(mock)
	results, total, err := repo.FindAllUpdatedAfter(context.Background(), date, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(results) != 1 {
		t.Errorf("unexpected results %v total %d", results, total)
	}
}

func TestProductFindAllInPageSearch(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("product_barcode ILIKE").WithArgs("%mie%", 10, 0).WillReturnRows(
		pgxmock.NewRows(productColumns).AddRow(productRow()...),
	)
	mock.ExpectQuery("SELECT COUNT").WithArgs("%mie%").WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(1),
	)

	repo := NewProductRepository(mock)
	results, total, err := repo.FindAllInPageSearch(context.Background(), 10, 0, "mie")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(results) != 1 || results[0].ProductName != "Indomie" {
		t.Errorf("unexpected results %v total %d", results, total)
	}
}

func TestProductSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO product").WithArgs("8991234567890", "Indomie", 2500.0, 3500.0, 1, 1).WillReturnRows(
		pgxmock.NewRows([]string{"product_id"}).AddRow(11),
	)

	repo := NewProductRepository(mock)
	product := &model.Product{ProductBarcode: "8991234567890", ProductName: "Indomie", ProductBuyPrice: 2500, ProductSellPrice: 3500, ProductCategoryID: 1, ProductDepartmentID: 1}
	if err := repo.Save(context.Background(), tx, product); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product.ProductID != 11 {
		t.Errorf("expected id 11, got %d", product.ProductID)
	}
}

func TestProductSaveConflict(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO product").WithArgs("8991234567890", "Indomie", 2500.0, 3500.0, 1, 1).WillReturnError(
		&pgconn.PgError{Code: "23505", ConstraintName: "product_product_barcode_key"},
	)

	repo := NewProductRepository(mock)
	err := repo.Save(context.Background(), tx, &model.Product{ProductBarcode: "8991234567890", ProductName: "Indomie", ProductBuyPrice: 2500, ProductSellPrice: 3500, ProductCategoryID: 1, ProductDepartmentID: 1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestProductUpdate(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("UPDATE product").WithArgs("8991234567890", "Indomie", 2500.0, 3500.0, 1, 1, 11).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := NewProductRepository(mock)
	product := &model.Product{ProductID: 11, ProductBarcode: "8991234567890", ProductName: "Indomie", ProductBuyPrice: 2500, ProductSellPrice: 3500, ProductCategoryID: 1, ProductDepartmentID: 1}
	if err := repo.Update(context.Background(), tx, product); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductFindById(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("WHERE p.product_id").WithArgs(11).WillReturnRows(
		pgxmock.NewRows([]string{"product_id", "product_barcode", "product_name", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat", "product_category_id", "product_department_id", "category_name", "department_code"}).
			AddRow(11, "8991234567890", "Indomie", 2500.0, 3500.0, ts, ts, 1, 1, "Makanan", "D01"),
	)

	repo := NewProductRepository(mock)
	got, err := repo.FindById(context.Background(), tx, 11)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ProductID != 11 || got.ProductCategoryID != 1 || got.ProductDepartmentCode != "D01" {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestProductFindByIdNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE p.product_id").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"product_id", "product_barcode", "product_name", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat", "product_category_id", "product_department_id", "category_name", "department_code"}),
	)

	repo := NewProductRepository(mock)
	_, err := repo.FindById(context.Background(), tx, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestProductFindAllLastSession(t *testing.T) {
	mock := mustMockPool(t)
	sessionDate := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("DISTINCT ON").WithArgs(20).WillReturnRows(
		pgxmock.NewRows([]string{"product_id", "product_barcode", "product_name", "lastSessionDate", "lastSessionCode"}).
			AddRow(1, "8991234567890", "Indomie", &sessionDate, "SESI-01").
			AddRow(2, "8991234567891", "Soto", nil, ""),
	)

	repo := NewProductRepository(mock)
	results, err := repo.FindAllLastSession(context.Background(), 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].LastSessionDate == nil || results[0].LastSessionCode != "SESI-01" {
		t.Errorf("unexpected first result %+v", results[0])
	}
	if results[1].LastSessionDate != nil {
		t.Errorf("expected nil date, got %v", results[1].LastSessionDate)
	}
}
