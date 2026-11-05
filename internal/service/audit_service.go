package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/repository"
)

type AuditService interface {
	FindAll(ctx context.Context, filter repository.AuditFilter, pagination *dto.Pagination) ([]dto.AuditResponse, error)
}
