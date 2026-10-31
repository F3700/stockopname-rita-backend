package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

type UploadService interface {
	Create(ctx context.Context, req dto.CreateUploadRequest) (dto.UploadResponse, error)
	FindAllSummary(ctx context.Context, coordinatorId int) ([]dto.UploadResponse, error)
	FindDetail(ctx context.Context, uploadID int) (dto.UploadDetailResponse, error)
	FindSummaryByCoordinator(ctx context.Context, coordinatorId int) ([]*model.UploadReportRow, error)
}
