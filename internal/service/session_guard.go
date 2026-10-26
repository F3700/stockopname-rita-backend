package service

import "stockopname-rita-backend/internal/model"

// ensureSessionWritable rejects writes on final sessions (D3).
func ensureSessionWritable(sesi *model.Sesi) error {
	if sesi.SesiStatus == "COMPLETED" || sesi.SesiStatus == "CANCELLED" {
		return &model.ConflictError{Resource: "Session", Detail: "Session is final"}
	}
	return nil
}
