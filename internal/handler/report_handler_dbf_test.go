package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestSessionDBF(t *testing.T) {
	h := NewReportHandler(&fakeReportService{
		sessionDBFFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			return []byte{0x30, 'f', 'a', 'k', 'e'}, "SESI-01.dbf", nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/1/dbf", "")
	rec := httptest.NewRecorder()
	h.SessionDBF(rec, req, paramsWithID("1"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/x-dbf" {
		t.Errorf("unexpected Content-Type %q", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="SESI-01.dbf"` {
		t.Errorf("unexpected Content-Disposition %q", got)
	}
	if len(rec.Body.Bytes()) == 0 || rec.Body.Bytes()[0] != 0x30 {
		t.Errorf("unexpected body %v", rec.Body.Bytes())
	}
}

func TestSessionDBFInvalidID(t *testing.T) {
	called := false
	h := NewReportHandler(&fakeReportService{
		sessionDBFFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			called = true
			return nil, "", nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/nope/dbf", "")
	rec := httptest.NewRecorder()
	h.SessionDBF(rec, req, paramsWithID("nope"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestSessionDBFNotFound(t *testing.T) {
	h := NewReportHandler(&fakeReportService{
		sessionDBFFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			return nil, "", &model.NotFoundError{Resource: "Session", ID: id}
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/99/dbf", "")
	rec := httptest.NewRecorder()
	h.SessionDBF(rec, req, paramsWithID("99"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
