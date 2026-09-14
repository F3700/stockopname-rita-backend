package helper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
