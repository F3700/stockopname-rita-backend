package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/repository"
	"stockopname-rita-backend/internal/service"
)

func TestSetQtyFactorHandler(t *testing.T) {
	fake := &fakeStockOpnameService{
		setQtyFactorFunc: func(_ context.Context, sessionID int, req dto.SetQtyFactorRequest) (dto.SetQtyFactorResponse, error) {
			if sessionID != 3 || req.Scope != "session" || req.QtyFactor == nil || *req.QtyFactor != -1 {
				t.Fatalf("unexpected request session=%d req=%+v", sessionID, req)
			}
			return dto.SetQtyFactorResponse{UpdatedCount: 2, BatchID: "11111111-1111-4111-8111-111111111111"}, nil
		},
	}
	handler := NewStockOpnameHandler(fake)

	req := newJSONRequest(t, http.MethodPatch, "/stockopname/sessions/3/qty-factor", `{"scope":"session","qty_factor":-1}`)
	req.Header.Set("X-Actor", "Niko")
	rec := httptest.NewRecorder()
	handler.SetQtyFactor(rec, req, paramsWithID("3"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var response struct {
		Data dto.SetQtyFactorResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if response.Data.UpdatedCount != 2 {
		t.Errorf("unexpected response %+v", response.Data)
	}
}

func TestGetSellSummaryHandler(t *testing.T) {
	fake := &fakeStockOpnameService{
		findSellSummaryFunc: func(_ context.Context, sessionID int, groupBy string) (dto.SellSummaryResponse, error) {
			if sessionID != 3 || groupBy != "coordinator" {
				t.Fatalf("unexpected session=%d groupBy=%s", sessionID, groupBy)
			}
			return dto.SellSummaryResponse{
				GroupBy: "coordinator",
				Groups: []dto.SellSummaryGroup{{
					Key: "5", Label: "Eko", ItemCount: 2, TotalQty: 3,
					TotalSellOriginal: 1000, TotalSellEffective: 500,
				}},
			}, nil
		},
	}
	handler := NewStockOpnameHandler(fake)

	req := httptest.NewRequest(http.MethodGet, "/stockopname/sessions/3/sell-summary?groupBy=coordinator", nil)
	rec := httptest.NewRecorder()
	handler.GetSellSummary(rec, req, paramsWithID("3"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var response struct {
		Data dto.SellSummaryResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if len(response.Data.Groups) != 1 || response.Data.Groups[0].Label != "Eko" {
		t.Errorf("unexpected response %+v", response.Data)
	}
}

func TestGetAuditBatchesHandler(t *testing.T) {
	fake := &fakeAuditService{
		findBatchesFunc: func(_ context.Context, filter repository.AuditFilter, pagination *dto.Pagination) ([]dto.AuditBatchSummaryResponse, error) {
			if filter.SesiID == nil || *filter.SesiID != 3 || pagination.Page != 1 || pagination.Limit != 10 {
				t.Fatalf("unexpected filter=%+v pagination=%+v", filter, pagination)
			}
			return []dto.AuditBatchSummaryResponse{{
				BatchID: "11111111-1111-4111-8111-111111111111", Action: "UPDATE",
				Actor: "Niko", ItemCount: 2, CreatedAt: "2026-11-08T10:00:00Z",
			}}, nil
		},
	}
	handler := NewAuditHandler(fake)

	req := httptest.NewRequest(http.MethodGet, "/stockopname/audits/batches?sessionId=3&page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	handler.GetAuditBatches(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var response struct {
		Data []dto.AuditBatchSummaryResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if len(response.Data) != 1 || response.Data[0].ItemCount != 2 {
		t.Errorf("unexpected response %+v", response.Data)
	}
}

var _ service.StockOpnameService = (*fakeStockOpnameService)(nil)
var _ service.AuditService = (*fakeAuditService)(nil)
