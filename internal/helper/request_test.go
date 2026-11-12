package helper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stockopname-rita-backend/internal/auth"
)

type readBodyTarget struct {
	Name string `json:"name"`
}

func TestReadFromRequestBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"name":"Rinso"}`))

	var got readBodyTarget
	if err := ReadFromRequestBody(req, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Rinso" {
		t.Errorf("expected name %q, got %q", "Rinso", got.Name)
	}
}

func TestReadFromRequestBodyUnknownField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"name":"Rinso","unknown":1}`))

	var got readBodyTarget
	if err := ReadFromRequestBody(req, &got); err == nil {
		t.Error("expected error for unknown field, got nil")
	}
}

// The session cookie is the authoritative source; a header cannot override it.
func TestActorFromRequestPrefersContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	req.Header.Set("X-Actor", "dari-header")
	req = req.WithContext(auth.WithActor(req.Context(), "Niko"))

	if got := ActorFromRequest(req); got != "Niko" {
		t.Errorf("expected the session actor, got %q", got)
	}
}

func TestActorFromRequestFallsBackToHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	req.Header.Set("X-Actor", "  Mobile  ")

	if got := ActorFromRequest(req); got != "Mobile" {
		t.Errorf("expected the trimmed header, got %q", got)
	}
}

func TestActorFromRequestDefaultsToAdminWeb(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/items", nil)

	if got := ActorFromRequest(req); got != "admin-web" {
		t.Errorf("expected the fallback actor, got %q", got)
	}
}

func TestActorFromRequestTruncatesLongHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	req.Header.Set("X-Actor", strings.Repeat("a", 120))

	if got := []rune(ActorFromRequest(req)); len(got) != 100 {
		t.Errorf("expected 100 runes, got %d", len(got))
	}
}

func TestReadFromRequestBodyInvalid(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty", ``},
		{"malformed", `{"name":`},
		{"wrong type", `{"name":123}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(tt.body))

			var got readBodyTarget
			if err := ReadFromRequestBody(req, &got); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
