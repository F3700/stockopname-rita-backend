package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stockopname-rita-backend/internal/dto"
)

func TestGetDeletedProducts(t *testing.T) {
	h := NewDeletedProductHandler(&fakeDeletedProductService{
		findAllFunc: func(ctx context.Context, date *time.Time) ([]dto.DeletedProductResponse, error) {
			if date != nil {
				t.Errorf("expected nil date, got %v", date)
			}
			return []dto.DeletedProductResponse{{ProductPLU: "100251"}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/deleted/products", "")
	rec := httptest.NewRecorder()
	h.GetDeletedProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "100251") {
		t.Errorf("expected plu in body, got %s", rec.Body.String())
	}
}

func TestDeleteDeletedProducts(t *testing.T) {
	called := false
	h := NewDeletedProductHandler(&fakeDeletedProductService{
		deleteFunc: func(ctx context.Context) error {
			called = true
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/deleted/products", "")
	rec := httptest.NewRecorder()
	h.DeleteDeletedProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !called {
		t.Error("service Delete must be called")
	}
}
