package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type CategoryService interface {
	Create(ctx context.Context, req dto.CategoryCreateRequest) (dto.CategoryResponse, error)
	FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.CategoryResponse, error)
	Update(ctx context.Context, id int, req dto.CategoryUpdateRequest) (dto.CategoryResponse, error)
	Delete(ctx context.Context, id int) error
}
