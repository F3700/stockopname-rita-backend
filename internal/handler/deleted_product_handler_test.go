package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"stockopname-rita-backend/internal/dto"
)

func TestDeleteDeletedProducts(t *testing.T) {
	h := NewDeletedProductHandler(&fakeDeletedProductService{
		deleteFunc: func(ctx context.Context) error { return nil },
	})

	req := newJSONRequest(t, http.MethodDelete, "/deleted-products", "")
	rec := httptest.NewRecorder()
	h.DeleteDeletedProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Deleted products deleted successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestDeleteDeletedProductsServiceError(t *testing.T) {
	h := NewDeletedProductHandler(&fakeDeletedProductService{
		deleteFunc: func(ctx context.Context) error { return errors.New("boom") },
	})

	req := newJSONRequest(t, http.MethodDelete, "/deleted-products", "")
	rec := httptest.NewRecorder()
	h.DeleteDeletedProducts(rec, req, nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestGetDeletedProducts(t *testing.T) {
	h := NewDeletedProductHandler(&fakeDeletedProductService{
		findAllFunc: func(ctx context.Context, date *time.Time) ([]dto.DeletedProductResponse, error) {
			if date != nil {
				t.Errorf("expected nil date, got %v", date)
			}
			return []dto.DeletedProductResponse{{ProductID: 9}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/deleted-products", "")
	rec := httptest.NewRecorder()
	h.GetDeletedProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Deleted products fetched successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetDeletedProductsWithDate(t *testing.T) {
	h := NewDeletedProductHandler(&fakeDeletedProductService{
		findAllFunc: func(ctx context.Context, date *time.Time) ([]dto.DeletedProductResponse, error) {
			if date == nil {
				t.Fatal("expected non-nil date")
			}
			return []dto.DeletedProductResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/deleted-products?updated_after=2026-09-10T08:00:00%2B07:00", "")
	rec := httptest.NewRecorder()
	h.GetDeletedProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetDeletedProductsInvalidDate(t *testing.T) {
	called := false
	h := NewDeletedProductHandler(&fakeDeletedProductService{
		findAllFunc: func(ctx context.Context, date *time.Time) ([]dto.DeletedProductResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/deleted-products?updated_after=nope", "")
	rec := httptest.NewRecorder()
	h.GetDeletedProducts(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid date")
	}
}
