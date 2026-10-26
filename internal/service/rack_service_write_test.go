package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

func newRackServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) RackService {
	t.Helper()
	return NewRackService(
		repository.NewRackRepository(mock),
		repository.NewCoordinatorRepository(mock),
		repository.NewSesiRepository(mock),
		repository.NewStockOpnameRepository(mock),
		mock,
	)
}

func rackStateRows(rakID int, uploaded bool) *pgxmock.Rows {
	var uploadedAt any
	if uploaded {
		uploadedAt = time.Now()
	}
	return pgxmock.NewRows([]string{"RakID", "RakName", "CoorID", "UploadedAt"}).
		AddRow(rakID, "Rak 1", 1, uploadedAt)
}

func coordinatorRows(coorID int, sesiID int) *pgxmock.Rows {
	return pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).
		AddRow(coorID, "KOR-01", sesiID, "IN_PROGRESS")
}

func sesiStatusRows(sesiID int, status string) *pgxmock.Rows {
	started := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"sesi_id", "sesi_location", "sesi_code", "sesi_status", "sesi_startedat", "sesi_endedat"}).
		AddRow(sesiID, "Gudang A", "SESI-01", status, started, nil)
}

func TestRackDeleteTx(t *testing.T) {
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
	mock.ExpectQuery("FROM rak").WithArgs(9).WillReturnRows(rackStateRows(9, false))
	mock.ExpectQuery("FROM coordinator").WithArgs(1).WillReturnRows(coordinatorRows(1, 7))
	mock.ExpectQuery("FOR SHARE").WithArgs(7).WillReturnRows(sesiStatusRows(7, "IN_PROGRESS"))
	mock.ExpectQuery("FROM rak").WithArgs(9).WillReturnRows(rackStateRows(9, false))
	mock.ExpectQuery("FROM stock_opname").WithArgs(9).WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(0),
	)
	mock.ExpectQuery("SELECT coor_id FROM coordinator").WithArgs(1).WillReturnRows(
		pgxmock.NewRows([]string{"coor_id"}).AddRow(1),
	)
	mock.ExpectQuery("FROM rak WHERE rak_coor_id").WithArgs(1).WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(2),
	)
	mock.ExpectExec("DELETE FROM rak").WithArgs(9).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectExec("recalc_coordinator_status").WithArgs(1).WillReturnResult(pgxmock.NewResult("SELECT", 0))
	mock.ExpectCommit()

	svc := newRackServiceWithMock(mock, t)
	if err := svc.DeleteRack(context.Background(), 9); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRackDeleteSessionFinal(t *testing.T) {
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
	mock.ExpectQuery("FROM rak").WithArgs(9).WillReturnRows(rackStateRows(9, false))
	mock.ExpectQuery("FROM coordinator").WithArgs(1).WillReturnRows(coordinatorRows(1, 7))
	mock.ExpectQuery("FOR SHARE").WithArgs(7).WillReturnRows(sesiStatusRows(7, "COMPLETED"))
	mock.ExpectRollback()

	svc := newRackServiceWithMock(mock, t)
	err = svc.DeleteRack(context.Background(), 9)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestRackDeleteAlreadyUploaded(t *testing.T) {
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
	mock.ExpectQuery("FROM rak").WithArgs(9).WillReturnRows(rackStateRows(9, true))
	mock.ExpectQuery("FROM coordinator").WithArgs(1).WillReturnRows(coordinatorRows(1, 7))
	mock.ExpectQuery("FOR SHARE").WithArgs(7).WillReturnRows(sesiStatusRows(7, "IN_PROGRESS"))
	mock.ExpectQuery("FROM rak").WithArgs(9).WillReturnRows(rackStateRows(9, true))
	mock.ExpectRollback()

	svc := newRackServiceWithMock(mock, t)
	err = svc.DeleteRack(context.Background(), 9)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestRackDeleteLastRack(t *testing.T) {
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
	mock.ExpectQuery("FROM rak").WithArgs(9).WillReturnRows(rackStateRows(9, false))
	mock.ExpectQuery("FROM coordinator").WithArgs(1).WillReturnRows(coordinatorRows(1, 7))
	mock.ExpectQuery("FOR SHARE").WithArgs(7).WillReturnRows(sesiStatusRows(7, "IN_PROGRESS"))
	mock.ExpectQuery("FROM rak").WithArgs(9).WillReturnRows(rackStateRows(9, false))
	mock.ExpectQuery("FROM stock_opname").WithArgs(9).WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(0),
	)
	mock.ExpectQuery("SELECT coor_id FROM coordinator").WithArgs(1).WillReturnRows(
		pgxmock.NewRows([]string{"coor_id"}).AddRow(1),
	)
	mock.ExpectQuery("FROM rak WHERE rak_coor_id").WithArgs(1).WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(1),
	)
	mock.ExpectRollback()

	svc := newRackServiceWithMock(mock, t)
	err = svc.DeleteRack(context.Background(), 9)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}
