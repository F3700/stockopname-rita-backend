package model

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapPgError(t *testing.T) {
	tests := []struct {
		name     string
		pgErr    *pgconn.PgError
		wantType interface{}
	}{
		{
			name:     "unique violation mapped",
			pgErr:    &pgconn.PgError{Code: "23505", ConstraintName: "product_product_plu_key"},
			wantType: &ConflictError{},
		},
		{
			name:     "foreign key violation mapped",
			pgErr:    &pgconn.PgError{Code: "23503", ConstraintName: "fk_barcode_product"},
			wantType: &ForeignKeyError{},
		},
		{
			name:     "check violation mapped",
			pgErr:    &pgconn.PgError{Code: "23514", ConstraintName: "chk_stock_opname_quantity"},
			wantType: &ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MapPgError("Product", tt.pgErr)
			switch tt.wantType.(type) {
			case *ConflictError:
				conflict, ok := err.(*ConflictError)
				if !ok {
					t.Fatalf("expected *ConflictError, got %T", err)
				}
				if conflict.Detail != "Product PLU already exists" {
					t.Errorf("unexpected detail: %q", conflict.Detail)
				}
			case *ForeignKeyError:
				fk, ok := err.(*ForeignKeyError)
				if !ok {
					t.Fatalf("expected *ForeignKeyError, got %T", err)
				}
				if fk.Detail != "Product not found" {
					t.Errorf("unexpected detail: %q", fk.Detail)
				}
			case *ValidationError:
				if _, ok := err.(*ValidationError); !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestMapPgErrorNonPgError(t *testing.T) {
	original := errors.New("boom")
	err := MapPgError("Product", original)
	if err != original {
		t.Errorf("expected original error back, got %v", err)
	}
}

func TestMapPgErrorUnknownConstraint(t *testing.T) {
	err := MapPgError("Product", &pgconn.PgError{Code: "23505", ConstraintName: "some_other_key"})
	conflict, ok := err.(*ConflictError)
	if !ok {
		t.Fatalf("expected *ConflictError, got %T", err)
	}
	if conflict.Detail != "unique constraint violation" {
		t.Errorf("unexpected detail: %q", conflict.Detail)
	}
}

func TestNotFoundErrorMessage(t *testing.T) {
	err := &NotFoundError{Resource: "Session", ID: 3}
	if got := err.Error(); got != "Session with ID 3 not found" {
		t.Errorf("unexpected message: %q", got)
	}

	detailed := &NotFoundError{Resource: "Coordinator", Detail: "coordinator code does not match any coordinator in the session"}
	if got := detailed.Error(); got != detailed.Detail {
		t.Errorf("expected message to use detail, got %q", got)
	}
}

func TestConflictErrorMessage(t *testing.T) {
	withDetail := &ConflictError{Resource: "Product", Detail: "Product PLU already exists"}
	if got := withDetail.Error(); got != "Product conflict: Product PLU already exists" {
		t.Errorf("unexpected message: %q", got)
	}

	plain := &ConflictError{Resource: "Product"}
	if got := plain.Error(); got != "Product already exists" {
		t.Errorf("unexpected message: %q", got)
	}
}

func TestForeignKeyErrorMessage(t *testing.T) {
	withDetail := &ForeignKeyError{Resource: "Barcode", Detail: "Product not found"}
	if got := withDetail.Error(); got != "Barcode foreign key violation: Product not found" {
		t.Errorf("unexpected message: %q", got)
	}

	plain := &ForeignKeyError{Resource: "Barcode"}
	if got := plain.Error(); got != "Barcode is referenced by other records" {
		t.Errorf("unexpected message: %q", got)
	}
}

func TestValidationErrorMessage(t *testing.T) {
	err := &ValidationError{Detail: "quantity must be positive"}
	if got := err.Error(); got != "validation failed: quantity must be positive" {
		t.Errorf("unexpected message: %q", got)
	}
}
