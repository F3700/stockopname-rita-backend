package service

import (
	"errors"
	"testing"

	"stockopname-rita-backend/internal/model"
)

func TestValidateQuantity(t *testing.T) {
	cases := []struct {
		name    string
		qty     float64
		wantErr bool
	}{
		{"zero", 0, false},
		{"integer", 48, false},
		{"one decimal", 2.5, false},
		{"two decimals", 2.55, false},
		{"three decimals", 2.555, true},
		{"negative", -1, true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQuantity(tt.qty)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateQuantity(%v) error = %v, wantErr %v", tt.qty, err, tt.wantErr)
			}
			if err != nil {
				var validation *model.ValidationError
				if !errors.As(err, &validation) {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}
