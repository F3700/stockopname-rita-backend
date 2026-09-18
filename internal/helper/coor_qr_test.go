package helper

import (
	"errors"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestFormatCoordinatorQR(t *testing.T) {
	if got := FormatCoordinatorQR(12); got != "RITA-COOR-12" {
		t.Errorf("expected RITA-COOR-12, got %q", got)
	}
}

func TestParseCoordinatorQR(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantID  int
		wantErr bool
	}{
		{"valid", "RITA-COOR-12", 12, false},
		{"valid lowercase", "rita-coor-7", 7, false},
		{"valid with spaces", "  RITA-COOR-3  ", 3, false},
		{"empty", "", 0, true},
		{"blank", "   ", 0, true},
		{"wrong prefix", "SESI-01", 0, true},
		{"missing id", "RITA-COOR-", 0, true},
		{"non numeric", "RITA-COOR-abc", 0, true},
		{"zero", "RITA-COOR-0", 0, true},
		{"negative", "RITA-COOR--5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCoordinatorQR(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var validation *model.ValidationError
				if !errors.As(err, &validation) {
					t.Errorf("expected ValidationError, got %T: %v", err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantID {
				t.Errorf("expected %d, got %d", tt.wantID, got)
			}
		})
	}
}

func TestParseFormatRoundTrip(t *testing.T) {
	got, err := ParseCoordinatorQR(FormatCoordinatorQR(99))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 99 {
		t.Errorf("expected 99, got %d", got)
	}
}
