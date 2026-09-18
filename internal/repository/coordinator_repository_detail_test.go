package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

func TestCoordinatorFindByIdDetail(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("JOIN sesi").WithArgs(5).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "session_code", "session_location", "inspector", "rack_assigned", "rack_completed", "status"}).
			AddRow(5, "KOR-05", "SESI-01", "Gudang A", 1, 2, 1, "IN_PROGRESS"),
	)

	repo := NewCoordinatorRepository(mock)
	got, err := repo.FindByIdDetail(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 5 || got.Code != "KOR-05" || got.SessionCode != "SESI-01" || got.SessionLocation != "Gudang A" || got.Status != "IN_PROGRESS" {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestCoordinatorFindByIdDetailNotFound(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("JOIN sesi").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "session_code", "session_location", "inspector", "rack_assigned", "rack_completed", "status"}),
	)

	repo := NewCoordinatorRepository(mock)
	_, err := repo.FindByIdDetail(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
