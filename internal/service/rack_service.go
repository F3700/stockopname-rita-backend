package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type RackService interface {
	FindAllSummary(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error)
	FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error)
}
