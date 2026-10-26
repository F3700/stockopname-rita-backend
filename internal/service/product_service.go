package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type ProductService interface {
	Create(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error)
	Update(ctx context.Context, id int, sesiId int, req dto.ProductUpdateRequest) (dto.ProductResponse, error)
	Delete(ctx context.Context, id int, sesiId int) error
	Clear(ctx context.Context, sesiId int) error
	FindAll(ctx context.Context, sesiId int, pagination *dto.Pagination, search string) ([]dto.ProductResponse, error)
}
