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

func TestCreateStockOpname(t *testing.T) {
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		createFunc: func(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error) {
			if req.Quantity != 10 || req.ProductID != 1 || req.RakID != 2 {
				t.Errorf("unexpected request %+v", req)
			}
			return dto.StockOpnameResponse{Id: 1, Quantity: 10}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/stock-opnames", `{"quantity":10,"product_id":1,"rak_id":2}`)
	rec := httptest.NewRecorder()
	h.CreateStockOpname(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Stock opname created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateStockOpnameInvalidBody(t *testing.T) {
	called := false
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		createFunc: func(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error) {
			called = true
			return dto.StockOpnameResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/stock-opnames", `{"quantity":`)
	rec := httptest.NewRecorder()
	h.CreateStockOpname(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}

func TestDeleteStockOpname(t *testing.T) {
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		deleteFunc: func(ctx context.Context, id int) error {
			if id != 6 {
				t.Errorf("expected id 6, got %d", id)
			}
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/stock-opnames/6", "")
	rec := httptest.NewRecorder()
	h.DeleteStockOpname(rec, req, paramsWithID("6"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Stock opname deleted successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestDeleteStockOpnameInvalidID(t *testing.T) {
	called := false
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodDelete, "/stock-opnames/x", "")
	rec := httptest.NewRecorder()
	h.DeleteStockOpname(rec, req, paramsWithID("x"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestGetAllStockOpname(t *testing.T) {
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
			if coorId == nil || *coorId != 1 {
				t.Errorf("expected coorId 1, got %v", coorId)
			}
			if sesiId == nil || *sesiId != 2 {
				t.Errorf("expected sesiId 2, got %v", sesiId)
			}
			pagination.TotalItems = 1
			pagination.TotalPages = 1
			return []dto.StockOpnameResponse{{Id: 1}}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/stock-opnames?page=1&limit=10&coordinatorId=1&sessionId=2", "")
	rec := httptest.NewRecorder()
	h.GetAllStockOpname(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Stock opname retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetAllStockOpnameNoFilters(t *testing.T) {
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
			if coorId != nil || sesiId != nil {
				t.Errorf("expected nil filters, got coorId=%v sesiId=%v", coorId, sesiId)
			}
			return []dto.StockOpnameResponse{}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/stock-opnames?page=1&limit=10", "")
	rec := httptest.NewRecorder()
	h.GetAllStockOpname(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetAllStockOpnameMissingPage(t *testing.T) {
	called := false
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/stock-opnames?limit=10", "")
	rec := httptest.NewRecorder()
	h.GetAllStockOpname(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called when page is missing")
	}
}

func TestGetAllStockOpnameInvalidSession(t *testing.T) {
	called := false
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/stock-opnames?page=1&limit=10&sessionId=oops", "")
	rec := httptest.NewRecorder()
	h.GetAllStockOpname(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid sessionId")
	}
}

func TestGetAllStockOpnameInvalidCoordinator(t *testing.T) {
	called := false
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
			called = true
			return nil, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/stock-opnames?page=1&limit=10&coordinatorId=oops", "")
	rec := httptest.NewRecorder()
	h.GetAllStockOpname(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid coordinatorId")
	}
}

func TestGetStockOpnameById(t *testing.T) {
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		findByIdFn: func(ctx context.Context, id int) (dto.StockOpnameResponse, error) {
			if id != 4 {
				t.Errorf("expected id 4, got %d", id)
			}
			return dto.StockOpnameResponse{Id: 4}, nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/stock-opnames/4", "")
	rec := httptest.NewRecorder()
	h.GetStockOpnameById(rec, req, paramsWithID("4"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Stock opname retrieved successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestGetStockOpnameByIdNotFound(t *testing.T) {
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		findByIdFn: func(ctx context.Context, id int) (dto.StockOpnameResponse, error) {
			return dto.StockOpnameResponse{}, &model.NotFoundError{Resource: "Stock Opname", ID: id}
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/stock-opnames/404", "")
	rec := httptest.NewRecorder()
	h.GetStockOpnameById(rec, req, paramsWithID("404"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdateStockOpname(t *testing.T) {
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		updateFunc: func(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error) {
			if id != 3 || req.Quantity != 25 {
				t.Errorf("unexpected id/request: %d %+v", id, req)
			}
			return dto.StockOpnameResponse{Id: 3, Quantity: 25}, nil
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/stock-opnames/3", `{"quantity":25}`)
	rec := httptest.NewRecorder()
	h.UpdateStockOpname(rec, req, paramsWithID("3"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Stock opname updated successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestUpdateStockOpnameServiceError(t *testing.T) {
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		updateFunc: func(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error) {
			return dto.StockOpnameResponse{}, errors.New("boom")
		},
	})

	req := newJSONRequest(t, http.MethodPatch, "/stock-opnames/3", `{"quantity":25}`)
	rec := httptest.NewRecorder()
	h.UpdateStockOpname(rec, req, paramsWithID("3"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestCreateStockOpnameByRack(t *testing.T) {
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		createByRackFn: func(ctx context.Context, req dto.CreateStockOpnameByRackRequest) error {
			if req.RackID != 5 || len(req.Items) != 1 {
				t.Errorf("unexpected request %+v", req)
			}
			return nil
		},
	})

	body := `{"rak_id":5,"so_products":[{"quantity":10,"product_id":1,"rak_id":5}]}`
	req := newJSONRequest(t, http.MethodPost, "/stock-opnames/by-rack", body)
	rec := httptest.NewRecorder()
	h.CreateStockOpnameByRack(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Stock opname by rack created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestCreateStockOpnameByRackInvalidBody(t *testing.T) {
	called := false
	h := NewStockOpnameHandler(&fakeStockOpnameService{
		createByRackFn: func(ctx context.Context, req dto.CreateStockOpnameByRackRequest) error {
			called = true
			return nil
		},
	})

	req := newJSONRequest(t, http.MethodPost, "/stock-opnames/by-rack", `{"rak_id":`)
	rec := httptest.NewRecorder()
	h.CreateStockOpnameByRack(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid body")
	}
}
