package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

func newSesiServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) SesiService {
	t.Helper()
	return NewSesiService(
		repository.NewSesiRepository(mock),
		repository.NewCoordinatorRepository(mock),
		repository.NewProductRepository(mock),
		repository.NewStockOpnameRepository(mock),
		repository.NewRackRepository(mock),
		mock,
		newTestValidator(t),
	)
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
	mock.ExpectQuery("INSERT INTO coordinator").WithArgs("KOR-01", 1, "IN_PROGRESS").WillReturnRows(
		pgxmock.NewRows([]string{"coor_id"}).AddRow(1),
	)
	mock.ExpectQuery("INSERT INTO rak").WithArgs("Rak 3", 1).WillReturnRows(
		pgxmock.NewRows([]string{"rak_id"}).AddRow(1),
	)
	mock.ExpectQuery("INSERT INTO rak").WithArgs("Rak 4", 1).WillReturnRows(
		pgxmock.NewRows([]string{"rak_id"}).AddRow(2),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.CreateSesiRequest{
		Code:     "SESI-01",
		Location: "Gudang A",
		Coordinators: []dto.CoordinatorInput{
			{Code: "KOR-01", RackStart: 3, RackEnd: 4},
		},
	})
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
	mock.ExpectQuery("INSERT INTO coordinator").WithArgs("KOR-01", 1, "IN_PROGRESS").WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateSesiRequest{
		Code:         "SESI-01",
		Location:     "Gudang A",
		Coordinators: []dto.CoordinatorInput{{Code: "KOR-01", RackStart: 1, RackEnd: 1}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSesiCreateRejectsOverlappingRanges(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	svc := newSesiServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateSesiRequest{
		Code:     "SESI-01",
		Location: "Gudang A",
		Coordinators: []dto.CoordinatorInput{
			{Code: "KOR-01", RackStart: 1, RackEnd: 5},
			{Code: "KOR-02", RackStart: 4, RackEnd: 8},
		},
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var validation *model.ValidationError
	if !errors.As(err, &validation) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSesiCreateRejectsOversizedRange(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	svc := newSesiServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateSesiRequest{
		Code:     "SESI-01",
		Location: "Gudang A",
		Coordinators: []dto.CoordinatorInput{
			{Code: "KOR-01", RackStart: 1, RackEnd: 201},
		},
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var validation *model.ValidationError
	if !errors.As(err, &validation) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSesiCreateRejectsInvertedRange(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	svc := newSesiServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), dto.CreateSesiRequest{
		Code:     "SESI-01",
		Location: "Gudang A",
		Coordinators: []dto.CoordinatorInput{
			{Code: "KOR-01", RackStart: 5, RackEnd: 3},
		},
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		t.Errorf("expected validator error, got %T: %v", err, err)
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("FROM stock_opname so").WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("DELETE FROM sesi").WithArgs(1).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSesiDeleteTxFinalSession(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "COMPLETED"))
	mock.ExpectExec("DELETE FROM sesi").WithArgs(1).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSesiDeleteTxActiveWithData(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	err = svc.Delete(context.Background(), 1)
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"sesi_id", "sesi_location", "sesi_code", "sesi_status", "sesi_startedat", "sesi_endedat"}),
	)
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
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

func TestSesiUpdateTxCompletionGuard(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	err = svc.Update(context.Background(), 1, dto.UpdateSesiRequest{Status: "COMPLETED"})
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestSesiUpdateTxFinalSession(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "COMPLETED"))
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	err = svc.Update(context.Background(), 1, dto.UpdateSesiRequest{Status: "IN_PROGRESS"})
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestSesiDeleteTxActiveWithResults(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("FROM stock_opname so").WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	svc := newSesiServiceWithMock(mock, t)
	err = svc.Delete(context.Background(), 1)
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}
