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

func TestDeletedProductDelete(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectExec("DELETE FROM deleted_product").WillReturnResult(pgxmock.NewResult("DELETE", 3))

	repo := NewDeletedProductRepository(mock)
	if err := repo.Delete(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeletedProductDeleteError(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectExec("DELETE FROM deleted_product").WillReturnError(errors.New("boom"))

	repo := NewDeletedProductRepository(mock)
	if err := repo.Delete(context.Background()); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestDeletedProductFindAll(t *testing.T) {
	mock := mustMockPool(t)
	deletedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM deleted_product").WillReturnRows(
		pgxmock.NewRows([]string{"product_id", "deleted_at"}).
			AddRow(1, deletedAt).
			AddRow(2, deletedAt),
	)

	repo := NewDeletedProductRepository(mock)
	results, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 || results[0].ProductID != 1 {
		t.Errorf("unexpected results %+v", results)
	}
	if !results[0].DeletedAt.Equal(deletedAt) {
		t.Errorf("unexpected deleted_at %v", results[0].DeletedAt)
	}
}

func TestDeletedProductFindAllUpdatedAfter(t *testing.T) {
	mock := mustMockPool(t)
	date := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("deleted_at >").WithArgs(pgxmock.AnyArg()).WillReturnRows(
		pgxmock.NewRows([]string{"product_id", "deleted_at"}).
			AddRow(7, date),
	)

	repo := NewDeletedProductRepository(mock)
	results, err := repo.FindAllUpdatedAfter(context.Background(), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].ProductID != 7 {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestDeletedProductSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("INSERT INTO deleted_product").WithArgs(4).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := NewDeletedProductRepository(mock)
	if err := repo.Save(context.Background(), tx, 4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeletedProductSaveError(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("INSERT INTO deleted_product").WithArgs(4).WillReturnError(
		&pgconn.PgError{Code: "23503", ConstraintName: "fk_deleted_product_product"},
	)

	repo := NewDeletedProductRepository(mock)
	err := repo.Save(context.Background(), tx, 4)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var fkErr *model.ForeignKeyError
	if !errors.As(err, &fkErr) {
		t.Errorf("expected ForeignKeyError, got %T: %v", err, err)
	}
}
