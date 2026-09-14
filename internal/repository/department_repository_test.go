package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

func TestDepartmentDelete(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM department").WithArgs(1).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := NewDepartmentRepositoryImpl(mock)
	if err := repo.Delete(context.Background(), tx, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDepartmentDeleteNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM department").WithArgs(99).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewDepartmentRepositoryImpl(mock)
	err := repo.Delete(context.Background(), tx, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestDepartmentFindAll(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM department").WillReturnRows(
		pgxmock.NewRows([]string{"department_id", "department_code", "department_name", "department_desc"}).
			AddRow(1, "D01", "Gudang", "desc"),
	)

	repo := NewDepartmentRepositoryImpl(mock)
	results, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].DepartmentCode != "D01" {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestDepartmentFindAllInPageSearch(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("department_code ILIKE").WithArgs("%gud%", 10, 5).WillReturnRows(
		pgxmock.NewRows([]string{"department_id", "department_code", "department_name", "department_desc"}).
			AddRow(1, "D01", "Gudang", "desc"),
	)
	mock.ExpectQuery("SELECT COUNT").WithArgs("%gud%").WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(1),
	)

	repo := NewDepartmentRepositoryImpl(mock)
	results, total, err := repo.FindAllInPageSearch(context.Background(), 10, 5, "gud")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(results) != 1 {
		t.Errorf("unexpected results %v total %d", results, total)
	}
}

func TestDepartmentFindById(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE department_id").WithArgs(2).WillReturnRows(
		pgxmock.NewRows([]string{"department_id", "department_code", "department_name", "department_desc"}).
			AddRow(2, "D02", "Kasir", "desc"),
	)

	repo := NewDepartmentRepositoryImpl(mock)
	got, err := repo.FindById(context.Background(), tx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.DepartmentID != 2 || got.DepartmentName != "Kasir" {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestDepartmentFindByIdNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE department_id").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"department_id", "department_code", "department_name", "department_desc"}),
	)

	repo := NewDepartmentRepositoryImpl(mock)
	_, err := repo.FindById(context.Background(), tx, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestDepartmentSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO department").WithArgs("D09", "Baru", "desc").WillReturnRows(
		pgxmock.NewRows([]string{"department_id"}).AddRow(9),
	)

	repo := NewDepartmentRepositoryImpl(mock)
	department := &model.Department{DepartmentCode: "D09", DepartmentName: "Baru", DepartmentDesc: "desc"}
	if err := repo.Save(context.Background(), tx, department); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if department.DepartmentID != 9 {
		t.Errorf("expected id 9, got %d", department.DepartmentID)
	}
}

func TestDepartmentSaveConflict(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO department").WithArgs("D09", "Baru", "desc").WillReturnError(
		&pgconn.PgError{Code: "23505", ConstraintName: "department_department_code_key"},
	)

	repo := NewDepartmentRepositoryImpl(mock)
	err := repo.Save(context.Background(), tx, &model.Department{DepartmentCode: "D09", DepartmentName: "Baru", DepartmentDesc: "desc"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestDepartmentUpdate(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("UPDATE department").WithArgs("D09", "Baru", "desc", 9).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := NewDepartmentRepositoryImpl(mock)
	if err := repo.Update(context.Background(), tx, &model.Department{DepartmentID: 9, DepartmentCode: "D09", DepartmentName: "Baru", DepartmentDesc: "desc"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
