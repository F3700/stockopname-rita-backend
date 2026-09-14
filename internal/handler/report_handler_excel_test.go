package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestSessionExcel(t *testing.T) {
	h := NewReportHandler(&fakeReportService{
		sessionExcelFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			return []byte("PK-fake"), "SESI-01.xlsx", nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/1/excel", "")
	rec := httptest.NewRecorder()
	h.SessionExcel(rec, req, paramsWithID("1"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("unexpected Content-Type %q", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="SESI-01.xlsx"` {
		t.Errorf("unexpected Content-Disposition %q", got)
	}
	if rec.Body.String() != "PK-fake" {
		t.Errorf("unexpected body %q", rec.Body.String())
	}
}

func TestSessionExcelInvalidID(t *testing.T) {
	called := false
	h := NewReportHandler(&fakeReportService{
		sessionExcelFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			called = true
			return nil, "", nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/nope/excel", "")
	rec := httptest.NewRecorder()
	h.SessionExcel(rec, req, paramsWithID("nope"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestSessionExcelNotFound(t *testing.T) {
	h := NewReportHandler(&fakeReportService{
		sessionExcelFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			return nil, "", &model.NotFoundError{Resource: "Session", ID: id}
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/99/excel", "")
	rec := httptest.NewRecorder()
	h.SessionExcel(rec, req, paramsWithID("99"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
