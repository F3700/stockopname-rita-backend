package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

func TestSesiDelete(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM sesi").WithArgs(1).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := NewSesiRepository(mock)
	if err := repo.Delete(context.Background(), tx, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSesiDeleteNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM sesi").WithArgs(99).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewSesiRepository(mock)
	err := repo.Delete(context.Background(), tx, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSesiDeleteError(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM sesi").WithArgs(1).WillReturnError(errors.New("boom"))

	repo := NewSesiRepository(mock)
	if err := repo.Delete(context.Background(), tx, 1); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestSesiFindAllInPageSearch(t *testing.T) {
	mock := mustMockPool(t)
	started := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	rows := pgxmock.NewRows([]string{"sesi_id", "sesi_location", "sesi_code", "sesi_status", "sesi_startedat", "sesi_endedat"}).
		AddRow(1, "Gudang A", "SESI-01", "IN_PROGRESS", started, pgtype.Timestamptz{Time: started, Valid: true})
	mock.ExpectQuery("FROM sesi").WithArgs(10, 0, "%gudang%").WillReturnRows(rows)
	mock.ExpectQuery("SELECT COUNT").WithArgs("%gudang%").WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(1),
	)

	repo := NewSesiRepository(mock)
	results, total, err := repo.FindAllInPageSearch(context.Background(), 10, 0, "gudang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(results) != 1 || results[0].SesiCode != "SESI-01" {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestSesiFindAllInPageSearchQueryError(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM sesi").WithArgs(10, 0, "%%").WillReturnError(errors.New("boom"))

	repo := NewSesiRepository(mock)
	_, _, err := repo.FindAllInPageSearch(context.Background(), 10, 0, "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestSesiFindById(t *testing.T) {
	mock := mustMockPool(t)
	started := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM sesi").WithArgs(1).WillReturnRows(
		pgxmock.NewRows([]string{"sesi_id", "sesi_location", "sesi_code", "sesi_status", "sesi_startedat", "sesi_endedat"}).
			AddRow(1, "Gudang A", "SESI-01", "COMPLETED", started, pgtype.Timestamptz{Valid: false}),
	)

	repo := NewSesiRepository(mock)
	got, err := repo.FindById(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SesiID != 1 || got.SesiCode != "SESI-01" || got.SesiStatus != "COMPLETED" {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestSesiFindByIdNotFound(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM sesi").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"sesi_id", "sesi_location", "sesi_code", "sesi_status", "sesi_startedat", "sesi_endedat"}),
	)

	repo := NewSesiRepository(mock)
	_, err := repo.FindById(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSesiSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	started := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("INSERT INTO sesi").WithArgs("Gudang A", "SESI-01", "IN_PROGRESS").WillReturnRows(
		pgxmock.NewRows([]string{"sesi_id", "sesi_startedat", "sesi_endedat"}).
			AddRow(3, started, pgtype.Timestamptz{Valid: false}),
	)

	repo := NewSesiRepository(mock)
	sesi := &model.Sesi{SesiLocation: "Gudang A", SesiCode: "SESI-01", SesiStatus: "IN_PROGRESS"}
	if err := repo.Save(context.Background(), tx, sesi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sesi.SesiID != 3 {
		t.Errorf("expected id 3, got %d", sesi.SesiID)
	}
}

func TestSesiSaveConflict(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO sesi").WithArgs("Gudang A", "SESI-01", "IN_PROGRESS").WillReturnError(
		&pgconn.PgError{Code: "23505", ConstraintName: "sesi_sesi_code_key"},
	)

	repo := NewSesiRepository(mock)
	err := repo.Save(context.Background(), tx, &model.Sesi{SesiLocation: "Gudang A", SesiCode: "SESI-01", SesiStatus: "IN_PROGRESS"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestSesiUpdate(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("SELECT EXISTS").WithArgs(1).WillReturnRows(
		pgxmock.NewRows([]string{"exists"}).AddRow(true),
	)
	mock.ExpectExec("update_sesi_status").WithArgs(1, "COMPLETED").WillReturnResult(pgxmock.NewResult("SELECT", 1))

	repo := NewSesiRepository(mock)
	if err := repo.Update(context.Background(), tx, &model.Sesi{SesiID: 1, SesiStatus: "COMPLETED"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSesiUpdateNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("SELECT EXISTS").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"exists"}).AddRow(false),
	)

	repo := NewSesiRepository(mock)
	err := repo.Update(context.Background(), tx, &model.Sesi{SesiID: 99, SesiStatus: "COMPLETED"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
