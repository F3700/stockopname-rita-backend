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

func newStockOpnameServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) StockOpnameService {
	t.Helper()
	return NewStockOpnameService(repository.NewStockOpnameRepository(mock), repository.NewProductRepository(mock), mock, newTestValidator(t))
}

func soResolveProductRows() *pgxmock.Rows {
	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"product_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}).
		AddRow(11, "100251", "Sari Roti", "1138", 14500.0, 17200.0, ts, ts)
}

func soResolveBarcodeRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"barcode_product_id", "barcode_code"}).
		AddRow(11, "1002515550011").
		AddRow(11, "8991001010016")
}

func soSummaryRows() *pgxmock.Rows {
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"id", "barcode", "name", "quantity", "rackName", "inspectorCode", "coordinatorCode", "updatedAt"}).
		AddRow(7, "8991001010016", "Sari Roti", 10, "R1", "INSP-01", "KOR-01", updatedAt)
}

func TestStockOpnameCreateByBarcodeTx(t *testing.T) {
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
	mock.ExpectQuery("JOIN barcode").WithArgs("8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10, 2, "100251", "Sari Roti", "8991001010016", 14500.0, 17200.0).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(7),
	)
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(7).WillReturnRows(soSummaryRows())
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: 10, Barcode: "8991001010016", RakID: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 7 || res.Barcode != "8991001010016" || res.Quantity != 10 {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestStockOpnameCreateByPLUTx(t *testing.T) {
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
	mock.ExpectQuery("WHERE p.product_plu").WithArgs("100251").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	// PLU-based input falls back to the primary barcode for the snapshot.
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10, 2, "100251", "Sari Roti", "1002515550011", 14500.0, 17200.0).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(7),
	)
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(7).WillReturnRows(soSummaryRows())
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: 10, PLU: "100251", RakID: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 7 {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestStockOpnameCreateWithoutIdentifier(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: 10, RakID: 2})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var validation *model.ValidationError
	if !errors.As(err, &validation) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestStockOpnameCreateTxSaveError(t *testing.T) {
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
	mock.ExpectQuery("JOIN barcode").WithArgs("8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10, 2, "100251", "Sari Roti", "8991001010016", 14500.0, 17200.0).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: 10, Barcode: "8991001010016", RakID: 2})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStockOpnameDeleteTx(t *testing.T) {
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
	mock.ExpectExec("DELETE FROM stock_opname").WithArgs(4).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	if err := svc.Delete(context.Background(), 4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStockOpnameDeleteTxNotFound(t *testing.T) {
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
	mock.ExpectExec("DELETE FROM stock_opname").WithArgs(404).WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	err = svc.Delete(context.Background(), 404)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestStockOpnameUpdateTx(t *testing.T) {
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
	mock.ExpectExec("UPDATE stock_opname").WithArgs(25, 4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(4).WillReturnRows(
		pgxmock.NewRows([]string{"id", "barcode", "name", "quantity", "rackName", "inspectorCode", "coordinatorCode", "updatedAt"}).
			AddRow(4, "8991001010016", "Sari Roti", 25, "R1", "INSP-01", "KOR-01", updatedAt),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	res, err := svc.Update(context.Background(), 4, dto.UpdateStockOpnameRequest{Quantity: 25})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 4 || res.Quantity != 25 {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestStockOpnameFindByIdTx(t *testing.T) {
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
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(7).WillReturnRows(soSummaryRows())
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	res, err := svc.FindById(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 7 || res.Barcode != "8991001010016" {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestStockOpnameCreateByRackTx(t *testing.T) {
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
	mock.ExpectQuery("JOIN barcode").WithArgs("8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10, 5, "100251", "Sari Roti", "8991001010016", 14500.0, 17200.0).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(20),
	)
	mock.ExpectQuery("WHERE p.product_plu").WithArgs("100252").WillReturnRows(
		pgxmock.NewRows([]string{"product_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}).
			AddRow(12, "100252", "Ultra Milk", "1139", 18900.0, 21500.0, time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC), time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)),
	)
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(
		pgxmock.NewRows([]string{"barcode_product_id", "barcode_code"}).AddRow(12, "1002525550018"),
	)
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(5, 5, "100252", "Ultra Milk", "1002525550018", 18900.0, 21500.0).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(21),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	req := dto.CreateStockOpnameByRackRequest{
		RackID: 5,
		Items: []dto.CreateStockOpnameRequest{
			{Quantity: 10, Barcode: "8991001010016", RakID: 5},
			{Quantity: 5, PLU: "100252", RakID: 5},
		},
	}
	if err := svc.CreateByRack(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
