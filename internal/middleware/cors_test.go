package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()

	CORS(next).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Allow-Origin %q, got %q", "*", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Allow-Methods header to be set")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected Allow-Headers header to be set")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected next handler status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodOptions, "/items", nil)
	rec := httptest.NewRecorder()

	CORS(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if called {
		t.Error("expected next handler not to be called on preflight")
	}
}
