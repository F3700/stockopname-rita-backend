package service

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func factorPtr(value int) *int { return &value }

func soAuditDataRowsWithFactor(id, rakID, coorID int, factor int) *pgxmock.Rows {
	return pgxmock.NewRows([]string{"stock_opname_id", "quantity", "plu", "name", "department_code", "barcode", "buy", "sell", "qty_factor", "rak_id", "rak_name", "coor_id", "coor_code", "coor_status", "rak_upload_id", "upload_createdat"}).
		AddRow(id, 10.0, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, factor, rakID, "R1", coorID, "KOR-01", "IN_PROGRESS", nil, nil)
}

func TestSetQtyFactorItemsTx(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "COMPLETED"))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(3, "items", nil, pgxmock.AnyArg(), nil, nil, nil).WillReturnRows(soAuditDataRowsWithFactor(4, 2, 4, 1))
	mock.ExpectExec("UPDATE stock_opname").WithArgs(-1, 4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec("INSERT INTO stock_opname_audit").WithArgs(
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	result, err := svc.SetQtyFactor(context.Background(), 3, dto.SetQtyFactorRequest{
		Scope:     "items",
		ResultIDs: []int{4},
		QtyFactor: factorPtr(-1),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UpdatedCount != 1 || result.BatchID == "" {
		t.Errorf("unexpected result %+v", result)
	}
}

func TestSetQtyFactorNoOp(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(3, "items", nil, pgxmock.AnyArg(), nil, nil, nil).WillReturnRows(soAuditDataRowsWithFactor(4, 2, 4, 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	result, err := svc.SetQtyFactor(context.Background(), 3, dto.SetQtyFactorRequest{
		Scope:     "items",
		ResultIDs: []int{4},
		QtyFactor: factorPtr(1),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UpdatedCount != 0 || result.BatchID != "" {
		t.Errorf("unexpected result %+v", result)
	}
}

func TestSetQtyFactorFilterTx(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("SELECT COUNT").WithArgs(3, "filter", nil, nil, "%roti%", nil, nil).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(3, "filter", nil, nil, "%roti%", nil, nil).WillReturnRows(soAuditDataRowsWithFactor(4, 2, 4, 1))
	mock.ExpectExec("UPDATE stock_opname").WithArgs(0, 4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec("INSERT INTO stock_opname_audit").WithArgs(
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	result, err := svc.SetQtyFactor(context.Background(), 3, dto.SetQtyFactorRequest{
		Scope:     "filter",
		Filter:    &dto.QtyFactorFilterRequest{Search: "roti"},
		QtyFactor: factorPtr(0),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UpdatedCount != 1 {
		t.Errorf("unexpected result %+v", result)
	}
}

func TestSetQtyFactorRejectsScopeTooLarge(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("SELECT COUNT").WithArgs(3, "session", nil, nil, nil, nil, nil).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1001))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.SetQtyFactor(context.Background(), 3, dto.SetQtyFactorRequest{
		Scope:     "session",
		QtyFactor: factorPtr(-1),
	})
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestSetQtyFactorItemsStrict(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(3, "items", nil, pgxmock.AnyArg(), nil, nil, nil).WillReturnRows(soAuditDataRowsWithFactor(4, 2, 4, 1))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.SetQtyFactor(context.Background(), 3, dto.SetQtyFactorRequest{
		Scope:     "items",
		ResultIDs: []int{4, 5},
		QtyFactor: factorPtr(-1),
	})
	if err == nil {
		t.Fatal("expected not found, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSetQtyFactorRejectsEmptyFilter(t *testing.T) {
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
	_, err = svc.SetQtyFactor(context.Background(), 3, dto.SetQtyFactorRequest{
		Scope:     "filter",
		Filter:    &dto.QtyFactorFilterRequest{},
		QtyFactor: factorPtr(-1),
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var validation *model.ValidationError
	if !errors.As(err, &validation) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestFindSellSummaryTx(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectQuery("FROM sesi").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("GROUP BY s.sesi_id").WithArgs(3).WillReturnRows(
		pgxmock.NewRows([]string{"key", "label", "item_count", "total_qty", "total_sell_original", "total_sell_effective"}).
			AddRow("3", "SESI-01", 2, 15.5, 250000.0, 200000.0),
	)

	svc := newStockOpnameServiceWithMock(mock, t)
	result, err := svc.FindSellSummary(context.Background(), 3, "session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GroupBy != "session" || len(result.Groups) != 1 || result.Groups[0].TotalSellEffective != 200000 {
		t.Errorf("unexpected result %+v", result)
	}
}
