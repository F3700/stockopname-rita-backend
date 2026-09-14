package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type DepartmentService interface {
	Delete(ctx context.Context, id int) error
	FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.DepartmentResponse, error)
	Create(ctx context.Context, req dto.DepartmentCreateRequest) (dto.DepartmentResponse, error)
	Update(ctx context.Context, id int, req dto.DepartmentUpdateRequest) (dto.DepartmentResponse, error)
}
