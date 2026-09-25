package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestDeletedProductSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("INSERT INTO deleted_product").WithArgs("100251").WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := NewDeletedProductRepository(mock)
	if err := repo.Save(context.Background(), tx, "100251"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeletedProductFindAll(t *testing.T) {
	mock := mustMockPool(t)
	deletedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM deleted_product").WillReturnRows(
		pgxmock.NewRows([]string{"product_plu", "deleted_at"}).
			AddRow("100251", deletedAt).
			AddRow("100252", deletedAt),
	)

	repo := NewDeletedProductRepository(mock)
	results, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 || results[0].ProductPLU != "100251" {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestDeletedProductFindAllUpdatedAfter(t *testing.T) {
	mock := mustMockPool(t)
	date := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("WHERE deleted_at >").WithArgs(date).WillReturnRows(
		pgxmock.NewRows([]string{"product_plu", "deleted_at"}).AddRow("100251", deletedAt),
	)

	repo := NewDeletedProductRepository(mock)
	results, err := repo.FindAllUpdatedAfter(context.Background(), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].ProductPLU != "100251" {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestDeletedProductDelete(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectExec("DELETE FROM deleted_product").WillReturnResult(pgxmock.NewResult("DELETE", 2))

	repo := NewDeletedProductRepository(mock)
	if err := repo.Delete(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
