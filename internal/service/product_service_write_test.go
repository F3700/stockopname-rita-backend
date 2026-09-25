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
	return NewProductService(
		repository.NewProductRepository(mock),
		repository.NewBarcodeRepository(mock),
		repository.NewDeletedProductRepository(mock),
		mock,
		newTestValidator(t),
	)
}

func productDetailRows() *pgxmock.Rows {
	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"product_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}).
		AddRow(11, "100251", "Sari Roti", "1138", 14500.0, 17200.0, ts, ts)
}

func productBarcodeRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"barcode_product_id", "barcode_code"}).
		AddRow(11, "1002515550011").
		AddRow(11, "8991001010016")
}

func floatPtr(f float64) *float64 { return &f }

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
	mock.ExpectQuery("INSERT INTO product").WithArgs("100251", "Sari Roti", "1138", 14500.0, 17200.0).WillReturnRows(
		pgxmock.NewRows([]string{"product_id"}).AddRow(11),
	)
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "1002515550011").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "8991001010016").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectQuery("WHERE p.product_id").WithArgs(11).WillReturnRows(productDetailRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(productBarcodeRows())
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newProductServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.ProductCreateRequest{
		PLU: "100251", Name: "Sari Roti", DepartmentCode: "1138",
		BuyPrice: floatPtr(14500), SellPrice: floatPtr(17200),
		Barcodes: []string{"1002515550011", "8991001010016"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 11 || res.PLU != "100251" || res.Barcode != "1002515550011" || len(res.Barcodes) != 2 {
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
	mock.ExpectQuery("WHERE p.product_id").WithArgs(11).WillReturnRows(productDetailRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(productBarcodeRows())
	mock.ExpectExec("DELETE FROM product").WithArgs(11).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectExec("INSERT INTO deleted_product").WithArgs("100251").WillReturnResult(pgxmock.NewResult("INSERT", 1))
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
	mock.ExpectQuery("WHERE p.product_id").WithArgs(99).WillReturnRows(pgxmock.NewRows([]string{"product_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}))
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

	newBarcodes := []string{"1002515550011"}
	mock.ExpectBegin()
	mock.ExpectQuery("WHERE p.product_id").WithArgs(11).WillReturnRows(productDetailRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(productBarcodeRows())
	mock.ExpectExec("UPDATE product").WithArgs("100251", "Sari Roti Baru", "1138", 14500.0, 17200.0, 11).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec("DELETE FROM barcode").WithArgs(11).WillReturnResult(pgxmock.NewResult("DELETE", 2))
	mock.ExpectExec("INSERT INTO barcode").WithArgs(11, "1002515550011").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectQuery("WHERE p.product_id").WithArgs(11).WillReturnRows(productDetailRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(
		pgxmock.NewRows([]string{"barcode_product_id", "barcode_code"}).AddRow(11, "1002515550011"),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newProductServiceWithMock(mock, t)
	name := "Sari Roti Baru"
	res, err := svc.Update(context.Background(), 11, dto.ProductUpdateRequest{Name: &name, Barcodes: &newBarcodes})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 11 || len(res.Barcodes) != 1 {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestProductClearTx(t *testing.T) {
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
	mock.ExpectExec("TRUNCATE product, barcode, deleted_product").WillReturnResult(pgxmock.NewResult("TRUNCATE", 0))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newProductServiceWithMock(mock, t)
	if err := svc.Clear(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
