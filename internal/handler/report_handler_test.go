package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestSessionPDF(t *testing.T) {
	h := NewReportHandler(&fakeReportService{
		sessionPDFFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			return []byte("%PDF-fake"), "sesi-01.pdf", nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/1/pdf", "")
	rec := httptest.NewRecorder()
	h.SessionPDF(rec, req, paramsWithID("1"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %q", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `inline; filename="sesi-01.pdf"` {
		t.Errorf("unexpected Content-Disposition %q", got)
	}
	if rec.Body.String() != "%PDF-fake" {
		t.Errorf("unexpected body %q", rec.Body.String())
	}
}

func TestSessionPDFInvalidID(t *testing.T) {
	called := false
	h := NewReportHandler(&fakeReportService{
		sessionPDFFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			called = true
			return nil, "", nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/nope/pdf", "")
	rec := httptest.NewRecorder()
	h.SessionPDF(rec, req, paramsWithID("nope"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestSessionPDFNotFound(t *testing.T) {
	h := NewReportHandler(&fakeReportService{
		sessionPDFFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			return nil, "", &model.NotFoundError{Resource: "Session", ID: id}
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/sessions/99/pdf", "")
	rec := httptest.NewRecorder()
	h.SessionPDF(rec, req, paramsWithID("99"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestCoordinatorPDF(t *testing.T) {
	h := NewReportHandler(&fakeReportService{
		coordinatorPDFFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			if id != 2 {
				t.Errorf("expected id 2, got %d", id)
			}
			return []byte("%PDF-fake"), "kor-02.pdf", nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/coordinators/2/pdf", "")
	rec := httptest.NewRecorder()
	h.CoordinatorPDF(rec, req, paramsWithID("2"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %q", got)
	}
}

func TestCoordinatorPDFInvalidID(t *testing.T) {
	called := false
	h := NewReportHandler(&fakeReportService{
		coordinatorPDFFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			called = true
			return nil, "", nil
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/coordinators/nope/pdf", "")
	rec := httptest.NewRecorder()
	h.CoordinatorPDF(rec, req, paramsWithID("nope"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service must not be called on invalid id")
	}
}

func TestCoordinatorPDFServiceError(t *testing.T) {
	h := NewReportHandler(&fakeReportService{
		coordinatorPDFFunc: func(ctx context.Context, id int) ([]byte, string, error) {
			return nil, "", errors.New("boom")
		},
	})

	req := newJSONRequest(t, http.MethodGet, "/coordinators/2/pdf", "")
	rec := httptest.NewRecorder()
	h.CoordinatorPDF(rec, req, paramsWithID("2"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
