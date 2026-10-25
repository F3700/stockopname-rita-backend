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

func newCoordinatorServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) CoordinatorService {
	t.Helper()
	return NewCoordinatorService(repository.NewCoordinatorRepository(mock), repository.NewSesiRepository(mock), mock)
}

func coordinatorRow(id int, sesiId int) *pgxmock.Rows {
	return pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).
		AddRow(id, "KOR-01", sesiId, "IN_PROGRESS")
}

func sesiStatusRow(id int, status string) *pgxmock.Rows {
	started := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"sesi_id", "sesi_location", "sesi_code", "sesi_status", "sesi_startedat", "sesi_endedat"}).
		AddRow(id, "Gudang A", "SESI-01", status, started, nil)
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
	mock.ExpectQuery("FROM coordinator").WithArgs(2).WillReturnRows(coordinatorRow(2, 1))
	mock.ExpectQuery("FOR SHARE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "IN_PROGRESS"))
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
	mock.ExpectQuery("FROM coordinator").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}),
	)
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

func TestCoordinatorUpdateTxFinalSession(t *testing.T) {
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
	mock.ExpectQuery("FROM coordinator").WithArgs(2).WillReturnRows(coordinatorRow(2, 1))
	mock.ExpectQuery("FOR SHARE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "COMPLETED"))
	mock.ExpectRollback()

	svc := newCoordinatorServiceWithMock(mock, t)
	err = svc.Update(context.Background(), 2, &dto.UpdateCoordinatorRequest{Status: "CANCELLED"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}
