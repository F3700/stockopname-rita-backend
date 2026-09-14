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

func newDepartmentServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) DepartmentService {
	t.Helper()
	return NewDepartmentService(repository.NewDepartmentRepositoryImpl(mock), mock, newTestValidator(t))
}

func TestDepartmentCreateTx(t *testing.T) {
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
	mock.ExpectQuery("INSERT INTO department").WithArgs("D09", "Baru", "desc").WillReturnRows(
		pgxmock.NewRows([]string{"department_id"}).AddRow(9),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newDepartmentServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.DepartmentCreateRequest{Code: "D09", Name: "Baru", Description: "desc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 9 || res.Code != "D09" || res.Name != "Baru" {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestDepartmentDeleteTx(t *testing.T) {
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
	mock.ExpectExec("DELETE FROM department").WithArgs(1).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newDepartmentServiceWithMock(mock, t)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDepartmentDeleteTxNotFound(t *testing.T) {
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
	mock.ExpectExec("DELETE FROM department").WithArgs(99).WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectRollback()

	svc := newDepartmentServiceWithMock(mock, t)
	err = svc.Delete(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestDepartmentUpdateTx(t *testing.T) {
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
	mock.ExpectQuery("WHERE department_id").WithArgs(9).WillReturnRows(
		pgxmock.NewRows([]string{"department_id", "department_code", "department_name", "department_desc"}).
			AddRow(9, "D09", "Lama", "desc"),
	)
	mock.ExpectExec("UPDATE department").WithArgs("D09", "Baru", "desc", 9).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newDepartmentServiceWithMock(mock, t)
	name := "Baru"
	res, err := svc.Update(context.Background(), 9, dto.DepartmentUpdateRequest{Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 9 || res.Name != "Baru" {
		t.Errorf("unexpected response %+v", res)
	}
}
