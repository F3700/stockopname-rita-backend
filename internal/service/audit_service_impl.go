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

func (a *AuditServiceImpl) FindBatches(ctx context.Context, filter repository.AuditFilter, pagination *dto.Pagination) ([]dto.AuditBatchSummaryResponse, error) {
	filter.Limit = pagination.Limit
	filter.Offset = (pagination.Page - 1) * pagination.Limit

	batches, total, err := a.AuditRepository.FindBatches(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.AuditBatchSummaryResponse, 0, len(batches))
	for _, batch := range batches {
		responses = append(responses, dto.AuditBatchSummaryResponse{
			BatchID:   batch.BatchID,
			Action:    batch.Action,
			Actor:     batch.Actor,
			Note:      batch.Note,
			ItemCount: batch.ItemCount,
			CreatedAt: batch.CreatedAt.Format(time.RFC3339),
		})
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}

func (a *AuditServiceImpl) FindTimeline(ctx context.Context, filter repository.AuditFilter, pagination *dto.Pagination) ([]dto.AuditTimelineItemResponse, error) {
	filter.Limit = pagination.Limit
	filter.Offset = (pagination.Page - 1) * pagination.Limit

	items, total, err := a.AuditRepository.FindTimeline(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.AuditTimelineItemResponse, 0, len(items))
	for _, item := range items {
		response := dto.AuditTimelineItemResponse{
			Kind:      item.Kind,
			BatchID:   item.BatchID,
			Action:    item.Action,
			Actor:     item.Actor,
			Note:      item.Note,
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
		}
		if item.Kind == "batch" {
			response.ItemCount = item.ItemCount
			response.FactorBefore = item.FactorBefore
			response.FactorAfter = item.FactorAfter
			response.SourceRakName = item.SourceRakName
			response.TargetRakName = item.TargetRakName
			response.MergedCount = item.MergedCount
		} else {
			response.ID = item.AuditID
			before := json.RawMessage(item.BeforeData)
			if len(before) == 0 {
				before = json.RawMessage("null")
			}
			after := json.RawMessage(item.AfterData)
			if len(after) == 0 {
				after = json.RawMessage("null")
			}
			response.Before = before
			response.After = after
		}
		responses = append(responses, response)
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}
