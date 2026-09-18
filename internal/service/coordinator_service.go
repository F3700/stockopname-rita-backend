package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type CoordinatorService interface {
	FindByIdSummary(ctx context.Context, id int) (*dto.CoordinatorResponse, error)
	FindByIdDetail(ctx context.Context, id int) (*dto.CoordinatorDetailResponse, error)
	FindByIdReport(ctx context.Context, id int) (*dto.CoordinatorReportResponse, error)
	FindAllSummary(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error)
	Update(ctx context.Context, id int, req *dto.UpdateCoordinatorRequest) error
}
