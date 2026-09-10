package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type InspectorService interface {
	FindAllSummary(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error)
	CreateInspector(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error)
}
