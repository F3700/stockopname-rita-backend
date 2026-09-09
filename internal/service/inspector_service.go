package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type InspectorService interface {
	FindAllSummary(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error)
}
