package service

import (
	"math"
	"stockopname-rita-backend/internal/model"
)

// validateQuantity rejects negative values and more than 2 decimals.
func validateQuantity(q float64) error {
	if math.IsNaN(q) || math.IsInf(q, 0) || q < 0 {
		return &model.ValidationError{Detail: "quantity must be a non-negative number"}
	}
	if math.Abs(q*100-math.Round(q*100)) > 1e-9 {
		return &model.ValidationError{Detail: "quantity supports at most 2 decimals"}
	}
	return nil
}
