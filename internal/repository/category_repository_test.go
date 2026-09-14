package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

func TestCategoryDelete(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM category").WithArgs(1).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := NewCategoryRepositoryImpl(mock)
	if err := repo.Delete(context.Background(), tx, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCategoryDeleteNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("DELETE FROM category").WithArgs(99).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewCategoryRepositoryImpl(mock)
	err := repo.Delete(context.Background(), tx, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCategoryFindAll(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM category").WillReturnRows(
		pgxmock.NewRows([]string{"category_id", "category_name", "category_desc"}).
			AddRow(1, "Makanan", "desc").
			AddRow(2, "Minuman", "desc"),
	)

	repo := NewCategoryRepositoryImpl(mock)
	results, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 || results[0].CategoryName != "Makanan" {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestCategoryFindAllQueryError(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM category").WillReturnError(errors.New("boom"))

	repo := NewCategoryRepositoryImpl(mock)
	_, err := repo.FindAll(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCategoryFindAllInPageSearch(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("WHERE category_name ILIKE").WithArgs("%mak%", 10, 0).WillReturnRows(
		pgxmock.NewRows([]string{"category_id", "category_name", "category_desc"}).
			AddRow(1, "Makanan", "desc"),
	)
	mock.ExpectQuery("SELECT COUNT").WithArgs("%mak%").WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(1),
	)

	repo := NewCategoryRepositoryImpl(mock)
	results, total, err := repo.FindAllInPageSearch(context.Background(), 10, 0, "mak")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(results) != 1 || results[0].CategoryID != 1 {
		t.Errorf("unexpected results %v total %d", results, total)
	}
}

func TestCategoryFindById(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE category_id").WithArgs(2).WillReturnRows(
		pgxmock.NewRows([]string{"category_id", "category_name", "category_desc"}).
			AddRow(2, "Minuman", "desc"),
	)

	repo := NewCategoryRepositoryImpl(mock)
	got, err := repo.FindById(context.Background(), tx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.CategoryID != 2 || got.CategoryName != "Minuman" {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestCategoryFindByIdNotFound(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("WHERE category_id").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"category_id", "category_name", "category_desc"}),
	)

	repo := NewCategoryRepositoryImpl(mock)
	_, err := repo.FindById(context.Background(), tx, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCategorySave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO category").WithArgs("Snack", "desc").WillReturnRows(
		pgxmock.NewRows([]string{"category_id"}).AddRow(5),
	)

	repo := NewCategoryRepositoryImpl(mock)
	category := &model.Category{CategoryName: "Snack", CategoryDescription: "desc"}
	if err := repo.Save(context.Background(), tx, category); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if category.CategoryID != 5 {
		t.Errorf("expected id 5, got %d", category.CategoryID)
	}
}

func TestCategorySaveConflict(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO category").WithArgs("Snack", "desc").WillReturnError(
		&pgconn.PgError{Code: "23505", ConstraintName: "category_category_name_key"},
	)

	repo := NewCategoryRepositoryImpl(mock)
	err := repo.Save(context.Background(), tx, &model.Category{CategoryName: "Snack", CategoryDescription: "desc"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestCategoryUpdate(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectExec("UPDATE category").WithArgs("Snack", "desc", 5).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := NewCategoryRepositoryImpl(mock)
	if err := repo.Update(context.Background(), tx, &model.Category{CategoryID: 5, CategoryName: "Snack", CategoryDescription: "desc"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
