package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func TestCreateProduct(t *testing.T) {
	h := NewProductHandler(&fakeProductService{
		createFunc: func(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error) {
			if req.Barcode != "8991234567890" {
				t.Errorf("unexpected request %+v", req)
			}
			return dto.ProductResponse{Id: 1, Barcode: "8991234567890"}, nil
		},
	})

	body := `{"barcode":"8991234567890","name":"Indomie","buy_price":2500,"sell_price":3500,"category_id":1,"department_id":1}`
	req := newJSONRequest(t, http.MethodPost, "/products", body)
	rec := httptest.NewRecorder()
	h.CreateProduct(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Product created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateProductInvalidBody(t *testing.T) {
	called := false
	h := NewProductHandler(&fakeProductService{
		createFunc: func(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error) {
			called = true
			return dto.ProductResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/products", `not-json`)
	rec := httptest.NewRecorder()
	h.CreateProduct(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}

func TestDeleteProduct(t *testing.T) {
	h := NewProductHandler(&fakeProductService{
		deleteFunc: func(ctx context.Context, id int) error {
			if id != 8 {
				t.Errorf("expected id 8, got %d", id)
			}
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/products/8", "")
	rec := httptest.NewRecorder()
	h.DeleteProduct(rec, req, paramsWithID("8"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Product deleted successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestDeleteProductInvalidID(t *testing.T) {
	called := false
	h := NewProductHandler(&fakeProductService{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/products/zz", "")
	rec := httptest.NewRecorder()
	h.DeleteProduct(rec, req, paramsWithID("zz"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestGetProducts(t *testing.T) {
	h := NewProductHandler(&fakeProductService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.ProductResponse, error) {
			if pagination.Page != 2 || pagination.Limit != 5 {
				t.Errorf("unexpected pagination %+v", pagination)
			}
			pagination.TotalItems = 1
			pagination.TotalPages = 1
			return []dto.ProductResponse{{Id: 1, Barcode: "8991234567890"}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/products?page=2&limit=5", "")
	rec := httptest.NewRecorder()
	h.GetProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Products fetched successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetProductsMissingPage(t *testing.T) {
	called := false
	h := NewProductHandler(&fakeProductService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.ProductResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/products?limit=5", "")
	rec := httptest.NewRecorder()
	h.GetProducts(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called when page is missing")
	}
}

func TestGetProductsLastSessionDefaultLimit(t *testing.T) {
	h := NewProductHandler(&fakeProductService{
		findAllLastSessionFn: func(ctx context.Context, limit int) ([]dto.ProductLastSessionResponse, error) {
			if limit != 20 {
				t.Errorf("expected default limit 20, got %d", limit)
			}
			return []dto.ProductLastSessionResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/products/last-session", "")
	rec := httptest.NewRecorder()
	h.GetProductsLastSession(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Products last session retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetProductsLastSessionCustomLimit(t *testing.T) {
	h := NewProductHandler(&fakeProductService{
		findAllLastSessionFn: func(ctx context.Context, limit int) ([]dto.ProductLastSessionResponse, error) {
			if limit != 7 {
				t.Errorf("expected limit 7, got %d", limit)
			}
			return []dto.ProductLastSessionResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/products/last-session?limit=7", "")
	rec := httptest.NewRecorder()
	h.GetProductsLastSession(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetProductsLastSessionInvalidLimit(t *testing.T) {
	called := false
	h := NewProductHandler(&fakeProductService{
		findAllLastSessionFn: func(ctx context.Context, limit int) ([]dto.ProductLastSessionResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/products/last-session?limit=xx", "")
	rec := httptest.NewRecorder()
	h.GetProductsLastSession(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid limit")
	}
}

func TestSyncProducts(t *testing.T) {
	h := NewProductHandler(&fakeProductService{
		findAllUpdatedAfterFn: func(ctx context.Context, pagination *dto.Pagination, date time.Time) ([]dto.ProductResponse, error) {
			return []dto.ProductResponse{{Id: 3}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/products/sync?page=1&limit=10&updated_after=2026-09-10T08:00:00%2B07:00", "")
	rec := httptest.NewRecorder()
	h.SyncProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Products synced successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestSyncProductsMissingDate(t *testing.T) {
	called := false
	h := NewProductHandler(&fakeProductService{
		findAllUpdatedAfterFn: func(ctx context.Context, pagination *dto.Pagination, date time.Time) ([]dto.ProductResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/products/sync?page=1&limit=10", "")
	rec := httptest.NewRecorder()
	h.SyncProducts(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called when updated_after is missing")
	}
}

func TestSyncProductsInvalidDate(t *testing.T) {
	called := false
	h := NewProductHandler(&fakeProductService{
		findAllUpdatedAfterFn: func(ctx context.Context, pagination *dto.Pagination, date time.Time) ([]dto.ProductResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/products/sync?page=1&limit=10&updated_after=kemarin", "")
	rec := httptest.NewRecorder()
	h.SyncProducts(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid date")
	}
}

func TestUpdateProduct(t *testing.T) {
	h := NewProductHandler(&fakeProductService{
		updateFunc: func(ctx context.Context, id int, req dto.ProductUpdateRequest) (dto.ProductResponse, error) {
			if id != 11 {
				t.Errorf("expected id 11, got %d", id)
			}
			return dto.ProductResponse{Id: 11}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/products/11", `{"name":"Teh Botol"}`)
	rec := httptest.NewRecorder()
	h.UpdateProduct(rec, req, paramsWithID("11"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Product updated successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestUpdateProductNotFound(t *testing.T) {
	h := NewProductHandler(&fakeProductService{
		updateFunc: func(ctx context.Context, id int, req dto.ProductUpdateRequest) (dto.ProductResponse, error) {
			return dto.ProductResponse{}, &model.NotFoundError{Resource: "Product", ID: id}
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/products/404", `{"name":"Teh Botol"}`)
	rec := httptest.NewRecorder()
	h.UpdateProduct(rec, req, paramsWithID("404"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdateProductUnexpectedError(t *testing.T) {
	h := NewProductHandler(&fakeProductService{
		updateFunc: func(ctx context.Context, id int, req dto.ProductUpdateRequest) (dto.ProductResponse, error) {
			return dto.ProductResponse{}, errors.New("boom")
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/products/11", `{"name":"Teh Botol"}`)
	rec := httptest.NewRecorder()
	h.UpdateProduct(rec, req, paramsWithID("11"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
