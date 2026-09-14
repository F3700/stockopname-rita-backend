package service

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

func newCoordinatorServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) CoordinatorService {
	t.Helper()
	return NewCoordinatorService(repository.NewCoordinatorRepository(mock), mock)
}

func TestCoordinatorUpdateTx(t *testing.T) {
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
	mock.ExpectExec("UPDATE coordinator").WithArgs("COMPLETED", 2).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newCoordinatorServiceWithMock(mock, t)
	if err := svc.Update(context.Background(), 2, &dto.UpdateCoordinatorRequest{Status: "COMPLETED"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCoordinatorUpdateTxNotFound(t *testing.T) {
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
	mock.ExpectExec("UPDATE coordinator").WithArgs("COMPLETED", 99).WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	mock.ExpectRollback()

	svc := newCoordinatorServiceWithMock(mock, t)
	err = svc.Update(context.Background(), 99, &dto.UpdateCoordinatorRequest{Status: "COMPLETED"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
