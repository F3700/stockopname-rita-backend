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
	return NewStockOpnameService(repository.NewStockOpnameRepository(mock), mock, newTestValidator(t))
}

func soSummaryRows() *pgxmock.Rows {
	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"id", "barcode", "product_name", "quantity", "rak_name", "inspector_code", "coor_code", "updatedAt"}).
		AddRow(7, "8991234567890", "Indomie", 10, "R1", "INSP-01", "KOR-01", updatedAt)
}

func TestStockOpnameCreateTx(t *testing.T) {
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
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10, 1, 2).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(7),
	)
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(7).WillReturnRows(soSummaryRows())
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: 10, ProductID: 1, RakID: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Id != 7 || res.Barcode != "8991234567890" || res.Quantity != 10 {
		t.Errorf("unexpected response %+v", res)
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
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10, 1, 2).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateStockOpnameRequest{Quantity: 10, ProductID: 1, RakID: 2})
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
		pgxmock.NewRows([]string{"id", "barcode", "product_name", "quantity", "rak_name", "inspector_code", "coor_code", "updatedAt"}).
			AddRow(4, "8991234567890", "Indomie", 25, "R1", "INSP-01", "KOR-01", updatedAt),
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
	if res.Id != 7 || res.Barcode != "8991234567890" {
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
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10, 1, 5).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(20),
	)
	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(5, 2, 5).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(21),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newStockOpnameServiceWithMock(mock, t)
	req := dto.CreateStockOpnameByRackRequest{
		RackID: 5,
		Items: []dto.CreateStockOpnameRequest{
			{Quantity: 10, ProductID: 1, RakID: 5},
			{Quantity: 5, ProductID: 2, RakID: 5},
		},
	}
	if err := svc.CreateByRack(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
