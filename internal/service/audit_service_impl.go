package service

import (
	"context"
	"encoding/json"
	"math"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/repository"
	"time"
)

type AuditServiceImpl struct {
	AuditRepository repository.AuditRepository
}

func NewAuditService(auditRepository repository.AuditRepository) AuditService {
	return &AuditServiceImpl{AuditRepository: auditRepository}
}

func (a *AuditServiceImpl) FindAll(ctx context.Context, filter repository.AuditFilter, pagination *dto.Pagination) ([]dto.AuditResponse, error) {
	filter.Limit = pagination.Limit
	filter.Offset = (pagination.Page - 1) * pagination.Limit

	audits, total, err := a.AuditRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.AuditResponse, 0, len(audits))
	for _, audit := range audits {
		before := json.RawMessage(audit.BeforeData)
		if len(before) == 0 {
			before = json.RawMessage("null")
		}
		after := json.RawMessage(audit.AfterData)
		if len(after) == 0 {
			after = json.RawMessage("null")
		}
		responses = append(responses, dto.AuditResponse{
			ID:          audit.AuditID,
			BatchID:     audit.BatchID,
			Action:      audit.Action,
			SesiID:      audit.SesiID,
			SourceRakID: audit.SourceRakID,
			TargetRakID: audit.TargetRakID,
			Actor:       audit.Actor,
			Note:        audit.Note,
			Before:      before,
			After:       after,
			CreatedAt:   audit.CreatedAt.Format(time.RFC3339),
		})
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}
