package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type CoordinatorService interface {
	FindByIdSummary(ctx context.Context, id int) (*dto.CoordinatorResponse, error)
	FindAllSummary(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error)
	Cancel(ctx context.Context, id int) error
	Complete(ctx context.Context, id int) error
}
