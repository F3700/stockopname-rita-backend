package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func TestCreateCategory(t *testing.T) {
	h := NewCategoryHandler(&fakeCategoryService{
		createFunc: func(ctx context.Context, req dto.CategoryCreateRequest) (dto.CategoryResponse, error) {
			if req.CategoryName != "Makanan" {
				t.Errorf("expected name %q, got %q", "Makanan", req.CategoryName)
			}
			return dto.CategoryResponse{CategoryID: 1, CategoryName: "Makanan"}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/categories", `{"name":"Makanan"}`)
	rec := httptest.NewRecorder()
	h.CreateCategory(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Category created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateCategoryInvalidBody(t *testing.T) {
	called := false
	h := NewCategoryHandler(&fakeCategoryService{
		createFunc: func(ctx context.Context, req dto.CategoryCreateRequest) (dto.CategoryResponse, error) {
			called = true
			return dto.CategoryResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/categories", `{"name":`)
	rec := httptest.NewRecorder()
	h.CreateCategory(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}

func TestCreateCategoryServiceError(t *testing.T) {
	h := NewCategoryHandler(&fakeCategoryService{
		createFunc: func(ctx context.Context, req dto.CategoryCreateRequest) (dto.CategoryResponse, error) {
			return dto.CategoryResponse{}, &model.ConflictError{Resource: "Category"}
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/categories", `{"name":"Makanan"}`)
	rec := httptest.NewRecorder()
	h.CreateCategory(rec, req, nil)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestDeleteCategory(t *testing.T) {
	h := NewCategoryHandler(&fakeCategoryService{
		deleteFunc: func(ctx context.Context, id int) error {
			if id != 7 {
				t.Errorf("expected id 7, got %d", id)
			}
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/categories/7", "")
	rec := httptest.NewRecorder()
	h.DeleteCategory(rec, req, paramsWithID("7"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Category deleted successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestDeleteCategoryInvalidID(t *testing.T) {
	called := false
	h := NewCategoryHandler(&fakeCategoryService{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/categories/abc", "")
	rec := httptest.NewRecorder()
	h.DeleteCategory(rec, req, paramsWithID("abc"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestDeleteCategoryNotFound(t *testing.T) {
	h := NewCategoryHandler(&fakeCategoryService{
		deleteFunc: func(ctx context.Context, id int) error {
			return &model.NotFoundError{Resource: "Category", ID: id}
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/categories/99", "")
	rec := httptest.NewRecorder()
	h.DeleteCategory(rec, req, paramsWithID("99"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetCategories(t *testing.T) {
	h := NewCategoryHandler(&fakeCategoryService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.CategoryResponse, error) {
			if search != "mak" {
				t.Errorf("expected search %q, got %q", "mak", search)
			}
			pagination.TotalItems = 1
			pagination.TotalPages = 1
			return []dto.CategoryResponse{{CategoryID: 1, CategoryName: "Makanan"}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/categories?page=1&limit=10&search=mak", "")
	rec := httptest.NewRecorder()
	h.GetCategories(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Categories retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetCategoriesMissingPage(t *testing.T) {
	called := false
	h := NewCategoryHandler(&fakeCategoryService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.CategoryResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/categories?limit=10", "")
	rec := httptest.NewRecorder()
	h.GetCategories(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called when page is missing")
	}
}

func TestUpdateCategory(t *testing.T) {
	h := NewCategoryHandler(&fakeCategoryService{
		updateFunc: func(ctx context.Context, id int, req dto.CategoryUpdateRequest) (dto.CategoryResponse, error) {
			if id != 3 {
				t.Errorf("expected id 3, got %d", id)
			}
			return dto.CategoryResponse{CategoryID: 3, CategoryName: *req.CategoryName}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/categories/3", `{"name":"Minuman"}`)
	rec := httptest.NewRecorder()
	h.UpdateCategory(rec, req, paramsWithID("3"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Category updated successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestUpdateCategoryServiceError(t *testing.T) {
	h := NewCategoryHandler(&fakeCategoryService{
		updateFunc: func(ctx context.Context, id int, req dto.CategoryUpdateRequest) (dto.CategoryResponse, error) {
			return dto.CategoryResponse{}, errors.New("boom")
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/categories/3", `{"name":"Minuman"}`)
	rec := httptest.NewRecorder()
	h.UpdateCategory(rec, req, paramsWithID("3"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
