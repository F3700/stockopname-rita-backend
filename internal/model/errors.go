package model

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %d not found", e.Resource, e.ID)
}

type ConflictError struct {
	Resource string
	Detail   string
}

func (e *ConflictError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s conflict: %s", e.Resource, e.Detail)
	}
	return fmt.Sprintf("%s already exists", e.Resource)
}

type ForeignKeyError struct {
	Resource string
	Detail   string
}

func (e *ForeignKeyError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s foreign key violation: %s", e.Resource, e.Detail)
	}
	return fmt.Sprintf("%s is referenced by other records", e.Resource)
}

type ValidationError struct {
	Detail string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s", e.Detail)
}

var constraintMessages = map[string]string{
	"department_department_code_key":    "Department code already exists",
	"category_category_name_key":        "Category name already exists",
	"product_product_barcode_key":       "Product barcode already exists",
	"sesi_sesi_code_key":                "Session code already exists",
	"uq_coordinator_sesi_code":          "Coordinator code already exists in this session",
	"uq_inspector_coor_code":            "Inspector code already exists for this coordinator",
	"uq_rak_inspector_name":             "Rack name already exists for this inspector",
	"uq_stock_opname_rak_product":       "Product already recorded for this rack",
	"fk_product_category_id":            "Category not found",
	"fk_product_department_id":          "Department not found",
	"fk_coordinator_sesi":               "Session not found",
	"fk_inspector_coordinator":          "Coordinator not found",
	"fk_rak_inspector":                  "Inspector not found",
	"fk_stock_opname_product":           "Product not found",
	"fk_stock_opname_rak":               "Rack not found",
}

// MapPgError converts a PostgreSQL error into a typed application error.
// Returns nil if err is not a PgError.
func MapPgError(resource string, err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case "23505": // unique_violation
		if msg, ok := constraintMessages[pgErr.ConstraintName]; ok {
			return &ConflictError{Resource: resource, Detail: msg}
		}
		return &ConflictError{Resource: resource, Detail: "unique constraint violation"}
	case "23503": // foreign_key_violation
		if msg, ok := constraintMessages[pgErr.ConstraintName]; ok {
			return &ForeignKeyError{Resource: resource, Detail: msg}
		}
		return &ForeignKeyError{Resource: resource, Detail: "referenced record not found"}
	case "23514": // check_violation
		return &ValidationError{Detail: pgErr.ConstraintName}
	}

	return err
}
