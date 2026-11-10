package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLimitBodyRejectsOversizedContentLength(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(strings.Repeat("a", 100)))
	rec := httptest.NewRecorder()

	LimitBody(next, 10).ServeHTTP(rec, req)

	if called {
		t.Fatal("expected next handler not to be called")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}

	var res struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if res.Message == "" {
		t.Error("expected a message in the response body")
	}
}

func TestLimitBodyAllowsSmallBody(t *testing.T) {
	var read string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected read error: %v", err)
		}
		read = string(data)
	})

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader("hello"))
	rec := httptest.NewRecorder()

	LimitBody(next, 1024).ServeHTTP(rec, req)

	if read != "hello" {
		t.Errorf("expected body %q, got %q", "hello", read)
	}
}

func TestLimitBodyCapsChunkedBody(t *testing.T) {
	var readErr error
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	})

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(strings.Repeat("a", 100)))
	req.ContentLength = -1 // simulate a chunked upload without Content-Length
	rec := httptest.NewRecorder()

	LimitBody(next, 10).ServeHTTP(rec, req)

	if readErr == nil {
		t.Fatal("expected reading an oversized chunked body to fail")
	}
}

func TestLimitBodyDisabled(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(strings.Repeat("a", 100)))
	rec := httptest.NewRecorder()

	LimitBody(next, 0).ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called when the limit is disabled")
	}
}
