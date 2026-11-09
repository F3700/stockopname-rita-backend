package service

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func moveRackContextRows(rakID, coorID int) *pgxmock.Rows {
	return pgxmock.NewRows([]string{"rak_id", "rak_name", "coor_id", "coor_code", "coor_status", "coor_sesi_id", "rak_upload_id", "upload_createdat"}).
		AddRow(rakID, "Rak", coorID, "KOR-01", "IN_PROGRESS", 3, nil, nil)
}

func moveAuditDataRows(id, rakID, coorID int, quantity float64) *pgxmock.Rows {
	return moveAuditDataRowsWithFactor(id, rakID, coorID, quantity, 1)
}

func moveAuditDataRowsWithFactor(id, rakID, coorID int, quantity float64, factor int) *pgxmock.Rows {
	return pgxmock.NewRows([]string{"stock_opname_id", "quantity", "plu", "name", "department_code", "barcode", "buy", "sell", "sell_factor", "rak_id", "rak_name", "coor_id", "coor_code", "coor_status", "rak_upload_id", "upload_createdat"}).
		AddRow(id, quantity, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, factor, rakID, "Rak", coorID, "KOR-01", "IN_PROGRESS", nil, nil)
}

func TestStockOpnameMoveMergesAndAudits(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(2).WillReturnRows(moveRackContextRows(2, 4))
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(5).WillReturnRows(moveRackContextRows(5, 4))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(2, pgxmock.AnyArg()).WillReturnRows(moveAuditDataRows(4, 2, 4, 5))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(5, "100251").WillReturnRows(moveAuditDataRows(8, 5, 4, 2))
	mock.ExpectQuery("ON CONFLICT \\(stock_opname_rak_id, so_product_plu\\) DO UPDATE").WithArgs(5.0, 5, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, 1).
		WillReturnRows(pgxmock.NewRows([]string{"stock_opname_id", "stock_opname_quantity", "so_product_plu", "so_product_name", "so_product_dept", "so_barcode", "so_buyprice", "so_sellprice", "so_sell_factor"}).
			AddRow(8, 7.0, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, 1))
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
	ids := []int{4}
	res, err := svc.Move(context.Background(), dto.MoveStockOpnameRequest{SourceRakID: 2, TargetRakID: 5, ResultIDs: &ids, Note: "salah rak", Actor: "Niko"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.MovedCount != 1 || res.MergedCount != 1 {
		t.Errorf("unexpected move response %+v", res)
	}
}

func TestStockOpnameMoveWholeRackCrossCoordinator(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(2).WillReturnRows(moveRackContextRows(2, 4))
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(5).WillReturnRows(moveRackContextRows(5, 6))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(2).WillReturnRows(moveAuditDataRows(4, 2, 4, 5))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(5, "100251").WillReturnRows(pgxmock.NewRows([]string{"stock_opname_id"}))
	mock.ExpectQuery("ON CONFLICT \\(stock_opname_rak_id, so_product_plu\\) DO UPDATE").WithArgs(5.0, 5, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, 1).
		WillReturnRows(pgxmock.NewRows([]string{"stock_opname_id", "stock_opname_quantity", "so_product_plu", "so_product_name", "so_product_dept", "so_barcode", "so_buyprice", "so_sellprice", "so_sell_factor"}).
			AddRow(9, 5.0, "100251", "Sari Roti", "1138", "8991001010016", 14500.0, 17200.0, 1))
	mock.ExpectExec("DELETE FROM stock_opname").WithArgs(4).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectExec("INSERT INTO stock_opname_audit").WithArgs(
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("UPDATE coordinator").WithArgs(4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec("UPDATE coordinator").WithArgs(6).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	res, err := svc.Move(context.Background(), dto.MoveStockOpnameRequest{SourceRakID: 2, TargetRakID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.MovedCount != 1 || res.MergedCount != 0 {
		t.Errorf("unexpected move response %+v", res)
	}
}

func TestStockOpnameMoveRejectsDifferentFactor(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(2).WillReturnRows(moveRackContextRows(2, 4))
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(5).WillReturnRows(moveRackContextRows(5, 4))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(2, pgxmock.AnyArg()).WillReturnRows(moveAuditDataRowsWithFactor(4, 2, 4, 5, 1))
	mock.ExpectQuery("FOR UPDATE OF so").WithArgs(5, "100251").WillReturnRows(moveAuditDataRowsWithFactor(8, 5, 4, 2, 0))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	ids := []int{4}
	_, err = svc.Move(context.Background(), dto.MoveStockOpnameRequest{SourceRakID: 2, TargetRakID: 5, ResultIDs: &ids})
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestStockOpnameMoveRejectsEmptyIDs(t *testing.T) {
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
	ids := []int{}
	_, err = svc.Move(context.Background(), dto.MoveStockOpnameRequest{SourceRakID: 2, TargetRakID: 5, ResultIDs: &ids})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var validation *model.ValidationError
	if !errors.As(err, &validation) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestStockOpnameMoveRejectsFinalCoordinator(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(2).WillReturnRows(
		pgxmock.NewRows([]string{"rak_id", "rak_name", "coor_id", "coor_code", "coor_status", "coor_sesi_id", "rak_upload_id", "upload_createdat"}).
			AddRow(2, "Rak", 4, "KOR-01", "COMPLETED", 3, nil, nil),
	)
	mock.ExpectQuery("FOR UPDATE OF r").WithArgs(5).WillReturnRows(moveRackContextRows(5, 4))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.Move(context.Background(), dto.MoveStockOpnameRequest{SourceRakID: 2, TargetRakID: 5})
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}
