package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testOrigin = "https://rita.albertt.my.id"

func newNoopHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestCORSEchoesAllowedOrigin(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/items", nil)
	request.Header.Set("Origin", testOrigin)

	recorder := httptest.NewRecorder()
	CORS([]string{testOrigin}, newNoopHandler()).ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Errorf("expected Allow-Origin %q, got %q", testOrigin, got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected Allow-Credentials true, got %q", got)
	}
	if got := recorder.Header().Get("Vary"); got != "Origin" {
		t.Errorf("expected Vary Origin, got %q", got)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("expected next handler status %d, got %d", http.StatusOK, recorder.Code)
	}
}

// A wildcard would be invalid next to credentials, so an unknown origin must
// simply not be echoed back.
func TestCORSRejectsUnknownOrigin(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/items", nil)
	request.Header.Set("Origin", "https://evil.example")

	recorder := httptest.NewRecorder()
	CORS([]string{testOrigin}, newNoopHandler()).ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no Allow-Origin header, got %q", got)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("the request itself must still reach the handler, got %d", recorder.Code)
	}
}

func TestCORSWithoutOrigin(t *testing.T) {
	recorder := httptest.NewRecorder()
	CORS([]string{testOrigin}, newNoopHandler()).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/items", nil))

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no Allow-Origin header, got %q", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })

	request := httptest.NewRequest(http.MethodOptions, "/items", nil)
	request.Header.Set("Origin", testOrigin)

	recorder := httptest.NewRecorder()
	CORS([]string{testOrigin}, next).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if called {
		t.Error("expected next handler not to be called on preflight")
	}
	if got := recorder.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Errorf("expected Max-Age 86400, got %q", got)
	}
}

func TestCORSAllowListIgnoresBlankEntries(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/items", nil)
	request.Header.Set("Origin", "")

	recorder := httptest.NewRecorder()
	CORS([]string{"", "   "}, newNoopHandler()).ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no Allow-Origin header, got %q", got)
	}
	if strings.Contains(recorder.Header().Get("Access-Control-Allow-Headers"), "\n") {
		t.Error("unexpected newline in the allow headers")
	}
}
