package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

func newStockOpnameServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) StockOpnameService {
	t.Helper()
	return NewStockOpnameService(
		repository.NewStockOpnameRepository(mock),
		repository.NewProductRepository(mock),
		repository.NewSesiRepository(mock),
		repository.NewAuditRepository(mock),
		repository.NewCoordinatorRepository(mock),
		mock,
		newTestValidator(t),
	)
}

func soResolveProductRows() *pgxmock.Rows {
	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"product_id", "product_sesi_id", "product_plu", "product_name", "product_department_code", "product_buyprice", "product_sellprice", "product_createdat", "product_updatedat"}).
		AddRow(11, 3, "100251", "Sari Roti", "1138", 14500.0, 17200.0, ts, ts)
}

func soSesiIdByRakRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"coor_sesi_id"}).AddRow(3)
}

func soSesiIdByResultRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"coor_sesi_id"}).AddRow(3)
}

func soResolveBarcodeRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"barcode_product_id", "barcode_code"}).
		AddRow(11, "1002515550011").
		AddRow(11, "8991001010016")
}

func soSummaryRows() *pgxmock.Rows {
	return soSummaryRowsWithQuantity(10)
}

func soSummaryRowsWithQuantity(quantity float64) *pgxmock.Rows {
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"id", "barcode", "productPLU", "name", "quantity", "rackId", "rackName", "coordinatorId", "coordinatorCode", "sell_factor", "updatedAt"}).
		AddRow(7, "8991001010016", "100251", "Sari Roti", quantity, 2, "R1", 4, "KOR-01", 1, updatedAt)
}

func soRackContextRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"rak_id", "rak_name", "coor_id", "coor_code", "coor_status", "coor_sesi_id", "rak_upload_id", "upload_createdat"}).
		AddRow(2, "R1", 4, "KOR-01", "IN_PROGRESS", 3, nil, nil)
}

func soAuditDataRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"stock_opname_id", "quantity", "plu", "name", "department_code", "barcode", "buy", "sell", "sell_factor", "rak_id", "rak_name", "coor_id", "coor_code", "coor_status", "rak_upload_id", "upload_createdat"}).
		AddRow(4, 10.0, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, 1, 2, "R1", 4, "KOR-01", "IN_PROGRESS", nil, nil)
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
	mock.ExpectQuery("FROM rak r").WithArgs(2).WillReturnRows(soSesiIdByRakRows())
	mock.ExpectQuery("FOR SHARE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(3).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(2).WillReturnRows(soRackContextRows())
	mock.ExpectQuery("JOIN barcode").WithArgs(3, "8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10.0, 2, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, 1).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(7),
	)
	mock.ExpectExec("INSERT INTO stock_opname_audit").WithArgs(
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("UPDATE coordinator").WithArgs(4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(7).WillReturnRows(soSummaryRows())
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: qtyPtr(10), Barcode: "8991001010016", RakID: 2})
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
	mock.ExpectQuery("FROM rak r").WithArgs(2).WillReturnRows(soSesiIdByRakRows())
	mock.ExpectQuery("FOR SHARE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(3).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(2).WillReturnRows(soRackContextRows())
	mock.ExpectQuery("WHERE p.product_sesi_id").WithArgs(3, "100251").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	// PLU-based input falls back to the primary barcode for the snapshot.
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10.0, 2, "100251", "Sari Roti", "1138", "1002515550011", 14500.0, 17200.0, 1).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(7),
	)
	mock.ExpectExec("INSERT INTO stock_opname_audit").WithArgs(
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("UPDATE coordinator").WithArgs(4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(7).WillReturnRows(soSummaryRows())
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: qtyPtr(10), PLU: "100251", RakID: 2})
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
	_, err = svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: qtyPtr(10), RakID: 2})
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
	mock.ExpectQuery("FROM rak r").WithArgs(2).WillReturnRows(soSesiIdByRakRows())
	mock.ExpectQuery("FOR SHARE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(3).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(2).WillReturnRows(soRackContextRows())
	mock.ExpectQuery("JOIN barcode").WithArgs(3, "8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10.0, 2, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, 1).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: qtyPtr(10), Barcode: "8991001010016", RakID: 2})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStockOpnameCreateDuplicateRejected(t *testing.T) {
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
	mock.ExpectQuery("FROM rak r").WithArgs(2).WillReturnRows(soSesiIdByRakRows())
	mock.ExpectQuery("FOR SHARE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(3).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(2).WillReturnRows(soRackContextRows())
	mock.ExpectQuery("JOIN barcode").WithArgs(3, "8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10.0, 2, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, 1).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}),
	)
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: qtyPtr(10), Barcode: "8991001010016", RakID: 2})
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
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
	mock.ExpectQuery("FROM stock_opname so").WithArgs(4).WillReturnRows(soSesiIdByResultRows())
	mock.ExpectQuery("FOR SHARE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(4).WillReturnRows(soAuditDataRows())
	mock.ExpectExec("DELETE FROM stock_opname").WithArgs(4).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectExec("INSERT INTO stock_opname_audit").WithArgs(
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("UPDATE coordinator").WithArgs(4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	if err := svc.Delete(context.Background(), 4, "tester"); err != nil {
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
	mock.ExpectQuery("FROM stock_opname so").WithArgs(404).WillReturnRows(
		pgxmock.NewRows([]string{"coor_sesi_id"}),
	)
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	err = svc.Delete(context.Background(), 404, "tester")
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
	mock.ExpectQuery("FROM stock_opname so").WithArgs(4).WillReturnRows(soSesiIdByResultRows())
	mock.ExpectQuery("FOR SHARE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(4).WillReturnRows(soAuditDataRows())
	mock.ExpectExec("UPDATE stock_opname").WithArgs(25.0, 4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec("INSERT INTO stock_opname_audit").WithArgs(
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("UPDATE coordinator").WithArgs(4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(4).WillReturnRows(soSummaryRowsWithQuantity(25))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	res, err := svc.Update(context.Background(), 4, dto.UpdateStockOpnameRequest{Quantity: qtyPtr(25)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 7 || res.Quantity != 25 {
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

func TestStockOpnameUpdateRejectsInvalidQuantity(t *testing.T) {
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
	_, err = svc.Update(context.Background(), 4, dto.UpdateStockOpnameRequest{Quantity: qtyPtr(-1)})
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		t.Fatalf("expected validation error, got %T: %v", err, err)
	}
}

func TestStockOpnameCreateNoProducts(t *testing.T) {
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
	mock.ExpectQuery("FROM rak r").WithArgs(2).WillReturnRows(soSesiIdByRakRows())
	mock.ExpectQuery("FOR SHARE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(3).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: qtyPtr(10), Barcode: "8991001010016", RakID: 2})
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestStockOpnameCreateFinalSession(t *testing.T) {
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
	mock.ExpectQuery("FROM rak r").WithArgs(2).WillReturnRows(soSesiIdByRakRows())
	mock.ExpectQuery("FOR SHARE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "COMPLETED"))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: qtyPtr(10), Barcode: "8991001010016", RakID: 2})
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestStockOpnameUpdateFinalSession(t *testing.T) {
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
	mock.ExpectQuery("FROM stock_opname so").WithArgs(4).WillReturnRows(soSesiIdByResultRows())
	mock.ExpectQuery("FOR SHARE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "CANCELLED"))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.Update(context.Background(), 4, dto.UpdateStockOpnameRequest{Quantity: qtyPtr(25)})
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}
