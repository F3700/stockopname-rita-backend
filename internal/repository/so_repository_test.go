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

func TestStockOpnameFindAll(t *testing.T) {
	mock := mustMockPool(t)

	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	rows := pgxmock.NewRows([]string{"id", "barcode", "name", "quantity", "rackName", "inspectorCode", "coordinatorCode", "updatedAt"}).
		AddRow(1, "8991234567890", "Indomie", 48, "R1", "INSP-01", "KOR-01", updatedAt).
		AddRow(2, "8991234567891", "Soto", 36, "R1", "INSP-01", "KOR-01", updatedAt)
	mock.ExpectQuery("FROM stock_opname").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), "%mie%", 10, 0).WillReturnRows(rows)
	mock.ExpectQuery("SELECT COUNT").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), "%mie%").WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(12),
	)

	repo := NewStockOpnameRepository(mock)
	coorId, sesiId := 1, 2
	results, total, err := repo.FindAll(context.Background(), 10, 0, "mie", &sesiId, &coorId)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 12 {
		t.Errorf("expected total 12, got %d", total)
	}
	if len(results) != 2 || results[0].Barcode != "8991234567890" || results[1].Quantity != 36 {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestStockOpnameFindAllQueryError(t *testing.T) {
	mock := mustMockPool(t)

	mock.ExpectQuery("FROM stock_opname").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), "%%", 10, 0).WillReturnError(errors.New("boom"))

	repo := NewStockOpnameRepository(mock)
	_, _, err := repo.FindAll(context.Background(), 10, 0, "", nil, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestStockOpnameFindById(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)

	updatedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(4).WillReturnRows(
		pgxmock.NewRows([]string{"id", "barcode", "product_name", "quantity", "rak_name", "inspector_code", "coor_code", "updatedAt"}).
			AddRow(4, "8991234567890", "Indomie", 48, "R1", "INSP-01", "KOR-01", updatedAt),
	)

	repo := NewStockOpnameRepository(mock)
	got, err := repo.FindById(context.Background(), tx, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != 4 || got.Barcode != "8991234567890" || got.Quantity != 48 {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestStockOpnameFindByIdNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)

	mock.ExpectQuery("WHERE so.stock_opname_id").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"id", "barcode", "product_name", "quantity", "rak_name", "inspector_code", "coor_code", "updatedAt"}),
	)

	repo := NewStockOpnameRepository(mock)
	_, err := repo.FindById(context.Background(), tx, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestStockOpnameSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)

	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10, 1, 2).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(7),
	)

	repo := NewStockOpnameRepository(mock)
	so := &model.StockOpname{StockOpnameQuantity: 10, StockOpnameProductID: 1, StockOpnameRakID: 2}
	if err := repo.Save(context.Background(), tx, so); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if so.StockOpnameID != 7 {
		t.Errorf("expected id 7, got %d", so.StockOpnameID)
	}
}

func TestStockOpnameSaveConflict(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)

	mock.ExpectQuery("INSERT INTO stock_opname").WithArgs(10, 1, 2).WillReturnError(
		&pgconn.PgError{Code: "23505", ConstraintName: "uq_stock_opname_rak_product"},
	)

	repo := NewStockOpnameRepository(mock)
	err := repo.Save(context.Background(), tx, &model.StockOpname{StockOpnameQuantity: 10, StockOpnameProductID: 1, StockOpnameRakID: 2})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestStockOpnameUpdate(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)

	mock.ExpectExec("UPDATE stock_opname").WithArgs(25, 4).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := NewStockOpnameRepository(mock)
	if err := repo.Update(context.Background(), tx, &model.StockOpname{StockOpnameID: 4, StockOpnameQuantity: 25}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStockOpnameUpdateNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)

	mock.ExpectExec("UPDATE stock_opname").WithArgs(25, 404).WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	repo := NewStockOpnameRepository(mock)
	err := repo.Update(context.Background(), tx, &model.StockOpname{StockOpnameID: 404, StockOpnameQuantity: 25})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestStockOpnameDelete(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)

	mock.ExpectExec("DELETE FROM stock_opname").WithArgs(4).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := NewStockOpnameRepository(mock)
	if err := repo.Delete(context.Background(), tx, 4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStockOpnameDeleteNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)

	mock.ExpectExec("DELETE FROM stock_opname").WithArgs(404).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewStockOpnameRepository(mock)
	err := repo.Delete(context.Background(), tx, 404)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
