package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

func newProductServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) ProductService {
	t.Helper()
	return NewProductService(repository.NewProductRepository(mock), repository.NewDeletedProductRepository(mock), mock, newTestValidator(t))
}

func productDetailRows() *pgxmock.Rows {
	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"product_id", "product_barcode", "product_name", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat", "product_category_id", "product_department_id", "category_name", "department_code"}).
		AddRow(11, "8991234567890", "Indomie", 2500.0, 3500.0, ts, ts, 1, 1, "Makanan", "D01")
}

func TestProductCreateTx(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO product").WithArgs("8991234567890", "Indomie", 2500.0, 3500.0, 1, 1).WillReturnRows(
		pgxmock.NewRows([]string{"product_id"}).AddRow(11),
	)
	mock.ExpectQuery("WHERE p.product_id").WithArgs(11).WillReturnRows(productDetailRows())
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newProductServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.ProductCreateRequest{Barcode: "8991234567890", Name: "Indomie", BuyPrice: 2500, SellPrice: 3500, CategoryID: 1, DepartmentID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 11 || res.Barcode != "8991234567890" || res.CategoryName != "Makanan" {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestProductDeleteTx(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM product").WithArgs(11).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectExec("INSERT INTO deleted_product").WithArgs(11).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newProductServiceWithMock(mock, t)
	if err := svc.Delete(context.Background(), 11); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductDeleteTxNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM product").WithArgs(99).WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectRollback()

	svc := newProductServiceWithMock(mock, t)
	err = svc.Delete(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestProductUpdateTx(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectBegin()
	mock.ExpectQuery("WHERE p.product_id").WithArgs(11).WillReturnRows(productDetailRows())
	mock.ExpectExec("UPDATE product").WithArgs("8991234567890", "Indomie Goreng", 2500.0, 3500.0, 1, 1, 11).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery("WHERE p.product_id").WithArgs(11).WillReturnRows(productDetailRows())
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newProductServiceWithMock(mock, t)
	name := "Indomie Goreng"
	res, err := svc.Update(context.Background(), 11, dto.ProductUpdateRequest{Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 11 {
		t.Errorf("unexpected response %+v", res)
	}
}
