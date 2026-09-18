package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

func TestCoordinatorFindById(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE coor_id").WithArgs(12).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).AddRow(12, "KOR-12", 1, "IN_PROGRESS"),
	)

	repo := NewCoordinatorRepository(mock)
	got, err := repo.FindById(context.Background(), tx, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.CoorID != 12 || got.CoorCode != "KOR-12" || got.CoorSesiID != 1 || got.CoorStatus != "IN_PROGRESS" {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestCoordinatorFindByIdNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE coor_id").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}),
	)

	repo := NewCoordinatorRepository(mock)
	_, err := repo.FindById(context.Background(), tx, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
