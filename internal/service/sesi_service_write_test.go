package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

func newSesiServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) SesiService {
	t.Helper()
	return NewSesiService(repository.NewSesiRepository(mock), repository.NewCoordinatorRepository(mock), mock, newTestValidator(t))
}

func TestSesiCreateTx(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	started := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO sesi").WithArgs("Gudang A", "SESI-01", "IN_PROGRESS").WillReturnRows(
		pgxmock.NewRows([]string{"sesi_id", "sesi_startedat", "sesi_endedat"}).
			AddRow(1, started, pgtype.Timestamptz{Valid: false}),
	)
	mock.ExpectExec("INSERT INTO coordinator").WithArgs("KOR-01", 1, "IN_PROGRESS").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO coordinator").WithArgs("KOR-02", 1, "IN_PROGRESS").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.CreateSesiRequest{Code: "SESI-01", Location: "Gudang A", Coordinator: []string{"KOR-01", "KOR-02"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 || res.Code != "SESI-01" || res.Status != "IN_PROGRESS" {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestSesiCreateTxCoordinatorError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	started := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO sesi").WithArgs("Gudang A", "SESI-01", "IN_PROGRESS").WillReturnRows(
		pgxmock.NewRows([]string{"sesi_id", "sesi_startedat", "sesi_endedat"}).
			AddRow(1, started, pgtype.Timestamptz{Valid: false}),
	)
	mock.ExpectExec("INSERT INTO coordinator").WithArgs("KOR-01", 1, "IN_PROGRESS").WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateSesiRequest{Code: "SESI-01", Location: "Gudang A", Coordinator: []string{"KOR-01"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSesiDeleteTx(t *testing.T) {
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
	mock.ExpectExec("DELETE FROM sesi").WithArgs(1).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSesiDeleteTxNotFound(t *testing.T) {
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
	mock.ExpectExec("DELETE FROM sesi").WithArgs(99).WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	err = svc.Delete(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSesiUpdateTx(t *testing.T) {
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
	mock.ExpectQuery("SELECT EXISTS").WithArgs(1).WillReturnRows(
		pgxmock.NewRows([]string{"exists"}).AddRow(true),
	)
	mock.ExpectExec("update_sesi_status").WithArgs(1, "COMPLETED").WillReturnResult(pgxmock.NewResult("SELECT", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	if err := svc.Update(context.Background(), 1, dto.UpdateSesiRequest{Status: "COMPLETED"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
