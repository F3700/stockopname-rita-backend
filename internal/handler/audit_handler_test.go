package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/repository"
)

func TestGetAudits(t *testing.T) {
	handler := NewAuditHandler(&fakeAuditService{
		findAllFunc: func(ctx context.Context, filter repository.AuditFilter, pagination *dto.Pagination) ([]dto.AuditResponse, error) {
			if filter.SesiID == nil || *filter.SesiID != 3 {
				t.Errorf("expected session 3, got %v", filter.SesiID)
			}
			if pagination.Limit != 10 || pagination.Page != 1 {
				t.Errorf("unexpected pagination %+v", pagination)
			}
			return []dto.AuditResponse{{
				ID:        1,
				Action:    "MOVE",
				SesiID:    3,
				Actor:     "Niko",
				Before:    []byte(`{"quantity":5}`),
				After:     []byte(`{"quantity":7}`),
				CreatedAt: "2026-11-05T08:00:00Z",
			}}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/stockopname/audits?page=1&limit=10&sessionId=3", nil)
	rec := httptest.NewRecorder()
	handler.GetAudits(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"action":"MOVE"`) {
		t.Errorf("unexpected body %s", rec.Body.String())
	}
}

func TestGetAuditTimeline(t *testing.T) {
	id := int64(11)
	handler := NewAuditHandler(&fakeAuditService{
		findTimelineFunc: func(ctx context.Context, filter repository.AuditFilter, pagination *dto.Pagination) ([]dto.AuditTimelineItemResponse, error) {
			if filter.SesiID == nil || *filter.SesiID != 3 {
				t.Errorf("expected session 3, got %v", filter.SesiID)
			}
			if len(filter.Actions) != 1 || filter.Actions[0] != "UPDATE" {
				t.Errorf("unexpected actions %v", filter.Actions)
			}
			if pagination.Limit != 10 || pagination.Page != 1 {
				t.Errorf("unexpected pagination %+v", pagination)
			}
			return []dto.AuditTimelineItemResponse{{
				Kind:      "single",
				ID:        &id,
				BatchID:   "11111111-1111-4111-8111-111111111111",
				Action:    "UPDATE",
				Actor:     "Niko",
				CreatedAt: "2026-11-05T08:00:00Z",
				Before:    []byte(`{"quantity":3}`),
				After:     []byte(`{"quantity":3}`),
			}}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/stockopname/audits/timeline?page=1&limit=10&sessionId=3&action=update", nil)
	rec := httptest.NewRecorder()
	handler.GetAuditTimeline(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"kind":"single"`) {
		t.Errorf("unexpected body %s", rec.Body.String())
	}
}

func TestGetAuditTimelineRejectsUnknownAction(t *testing.T) {
	handler := NewAuditHandler(&fakeAuditService{})

	req := httptest.NewRequest(http.MethodGet, "/stockopname/audits/timeline?page=1&limit=10&action=PURGE", nil)
	rec := httptest.NewRecorder()
	handler.GetAuditTimeline(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
