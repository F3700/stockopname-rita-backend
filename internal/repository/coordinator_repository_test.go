package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

func TestCoordinatorUpdate(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("UPDATE coordinator").WithArgs("COMPLETED", 2).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := NewCoordinatorRepository(mock)
	if err := repo.Update(context.Background(), tx, &model.Coordinator{CoorID: 2, CoorStatus: "COMPLETED"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCoordinatorUpdateNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("UPDATE coordinator").WithArgs("COMPLETED", 99).WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	repo := NewCoordinatorRepository(mock)
	err := repo.Update(context.Background(), tx, &model.Coordinator{CoorID: 99, CoorStatus: "COMPLETED"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCoordinatorFindAllSummary(t *testing.T) {
	mock := mustMockPool(t)
	rows := pgxmock.NewRows([]string{"id", "code", "inspector", "rack_assigned", "rack_completed", "status"}).
		AddRow(1, "KOR-01", 2, 5, 3, "IN_PROGRESS").
		AddRow(2, "KOR-02", 1, 4, 4, "COMPLETED")
	mock.ExpectQuery("FROM coordinator").WillReturnRows(rows)

	repo := NewCoordinatorRepository(mock)
	results, err := repo.FindAllSummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 || results[0].Code != "KOR-01" || results[1].Status != "COMPLETED" {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestCoordinatorFindAllSummaryError(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM coordinator").WillReturnError(errors.New("boom"))

	repo := NewCoordinatorRepository(mock)
	_, err := repo.FindAllSummary(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCoordinatorFindByIdSummary(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("WHERE c.coor_id").WithArgs(3).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "inspector", "rack_assigned", "rack_completed", "status"}).
			AddRow(3, "KOR-03", 1, 2, 2, "IN_REVIEW"),
	)

	repo := NewCoordinatorRepository(mock)
	got, err := repo.FindByIdSummary(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 3 || got.Code != "KOR-03" || got.Status != "IN_REVIEW" {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestCoordinatorFindByIdSummaryNotFound(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("WHERE c.coor_id").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "inspector", "rack_assigned", "rack_completed", "status"}),
	)

	repo := NewCoordinatorRepository(mock)
	_, err := repo.FindByIdSummary(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCoordinatorFindByIdReport(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("WHERE c.coor_id").WithArgs(2).WillReturnRows(
		pgxmock.NewRows([]string{"coor_id", "coor_code", "coor_status", "sesi_code", "inspector", "rack_assigned", "rack_completed"}).
			AddRow(2, "KOR-02", "COMPLETED", "SESI-01", 1, 4, 4),
	)

	repo := NewCoordinatorRepository(mock)
	got, err := repo.FindByIdReport(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 2 || got.Code != "KOR-02" || got.SessionCode != "SESI-01" || got.RackCompleted != 4 {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestCoordinatorFindByIdReportNotFound(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("WHERE c.coor_id").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"coor_id", "coor_code", "coor_status", "sesi_code", "inspector", "rack_assigned", "rack_completed"}),
	)

	repo := NewCoordinatorRepository(mock)
	_, err := repo.FindByIdReport(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCoordinatorFindBySesiIdSummary(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("c.coor_sesi_id").WithArgs(4).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "inspector", "rack_assigned", "rack_completed", "status"}).
			AddRow(5, "KOR-05", 0, 0, 0, "IN_PROGRESS"),
	)

	repo := NewCoordinatorRepository(mock)
	results, err := repo.FindBySesiIdSummary(context.Background(), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Code != "KOR-05" {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestCoordinatorSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("INSERT INTO coordinator").WithArgs("KOR-09", 1, "IN_PROGRESS").WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := NewCoordinatorRepository(mock)
	if err := repo.Save(context.Background(), tx, &model.Coordinator{CoorCode: "KOR-09", CoorSesiID: 1, CoorStatus: "IN_PROGRESS"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCoordinatorSaveConflict(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("INSERT INTO coordinator").WithArgs("KOR-09", 1, "IN_PROGRESS").WillReturnError(
		&pgconn.PgError{Code: "23505", ConstraintName: "uq_coordinator_sesi_code"},
	)

	repo := NewCoordinatorRepository(mock)
	err := repo.Save(context.Background(), tx, &model.Coordinator{CoorCode: "KOR-09", CoorSesiID: 1, CoorStatus: "IN_PROGRESS"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestCoordinatorFindBySesiAndCoorCode(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("s.sesi_code").WithArgs("SESI-01", "KOR-01").WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).AddRow(1, "KOR-01", 1, "IN_PROGRESS"),
	)

	repo := NewCoordinatorRepository(mock)
	got, err := repo.FindBySesiAndCoorCode(context.Background(), tx, "SESI-01", "KOR-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.CoorID != 1 || got.CoorSesiID != 1 || got.CoorStatus != "IN_PROGRESS" {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestCoordinatorFindBySesiAndCoorCodeNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("s.sesi_code").WithArgs("SESI-01", "NOPE").WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}),
	)

	repo := NewCoordinatorRepository(mock)
	_, err := repo.FindBySesiAndCoorCode(context.Background(), tx, "SESI-01", "NOPE")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
	if notFound.Detail == "" {
		t.Error("expected detail message for code-based lookup")
	}
}
