package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
)

type StockOpnameService interface {
	Create(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error)
	Update(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error)
	Delete(ctx context.Context, id int, actor string) error
	Move(ctx context.Context, req dto.MoveStockOpnameRequest) (dto.MoveStockOpnameResponse, error)
	FindById(ctx context.Context, id int) (dto.StockOpnameResponse, error)
	FindAll(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int, rakId *int) ([]dto.StockOpnameResponse, error)
	FindAllForExport(ctx context.Context, sesiId *int, coorId *int) ([]dto.StockOpnameExportResponse, error)
	SetSellFactor(ctx context.Context, sessionID int, req dto.SetSellFactorRequest) (dto.SetSellFactorResponse, error)
	FindSellSummary(ctx context.Context, sessionID int, groupBy string) (dto.SellSummaryResponse, error)
}
