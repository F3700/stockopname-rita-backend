package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"time"
)

type DeletedProductService interface {
	FindAll(ctx context.Context, date *time.Time) ([]dto.DeletedProductResponse, error)
	Delete(ctx context.Context) error
}
