package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type ProductService interface {
	Create(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductCreateResponse, error)
	Update()
	Delete()
	FindAll()
}
