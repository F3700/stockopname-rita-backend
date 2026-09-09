package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type SesiService interface {
	Create(ctx context.Context, req dto.CreateSesiRequest) (dto.SesiResponse, error)
	Update(ctx context.Context, id int, req dto.UpdateSesiRequest) error
	Delete(ctx context.Context, id int) error
	FindById(ctx context.Context, id int) (dto.SesiResponse, error)
}
