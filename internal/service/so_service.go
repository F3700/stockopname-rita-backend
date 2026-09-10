package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type StockOpnameService interface {
	Create(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error)
	Update(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error)
	Delete(ctx context.Context, id int) error
	FindById(ctx context.Context, id int) (dto.StockOpnameResponse, error)
	FindAll(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error)
	CreateByRack(ctx context.Context, req dto.CreateStockOpnameByRackRequest) error
}
