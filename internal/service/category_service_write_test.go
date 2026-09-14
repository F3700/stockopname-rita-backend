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

func newCategoryServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) CategoryService {
	t.Helper()
	return NewCategoryService(repository.NewCategoryRepositoryImpl(mock), mock, newTestValidator(t))
}

func TestCategoryCreateTx(t *testing.T) {
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
	mock.ExpectQuery("INSERT INTO category").WithArgs("Snack", "desc").WillReturnRows(
		pgxmock.NewRows([]string{"category_id"}).AddRow(5),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newCategoryServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.CategoryCreateRequest{CategoryName: "Snack", CategoryDescription: "desc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.CategoryID != 5 || res.CategoryName != "Snack" {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestCategoryDeleteTx(t *testing.T) {
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
	mock.ExpectExec("DELETE FROM category").WithArgs(1).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newCategoryServiceWithMock(mock, t)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCategoryDeleteTxNotFound(t *testing.T) {
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
	mock.ExpectExec("DELETE FROM category").WithArgs(99).WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectRollback()

	svc := newCategoryServiceWithMock(mock, t)
	err = svc.Delete(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCategoryUpdateTx(t *testing.T) {
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
	mock.ExpectQuery("WHERE category_id").WithArgs(5).WillReturnRows(
		pgxmock.NewRows([]string{"category_id", "category_name", "category_desc"}).AddRow(5, "Old", "old desc"),
	)
	mock.ExpectExec("UPDATE category").WithArgs("New", "old desc", 5).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newCategoryServiceWithMock(mock, t)
	name := "New"
	res, err := svc.Update(context.Background(), 5, dto.CategoryUpdateRequest{CategoryName: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.CategoryID != 5 || res.CategoryName != "New" {
		t.Errorf("unexpected response %+v", res)
	}
}
