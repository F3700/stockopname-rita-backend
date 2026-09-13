package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"time"
)

type ProductService interface {
	Create(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error)
	Update(ctx context.Context, id int, req dto.ProductUpdateRequest) (dto.ProductResponse, error)
	Delete(ctx context.Context, id int) error
	FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.ProductResponse, error)
	FindAllUpdatedAfter(ctx context.Context, pagination *dto.Pagination, date time.Time) ([]dto.ProductResponse, error)
	FindAllLastSession(ctx context.Context, limit int) ([]dto.ProductLastSessionResponse, error)
}
