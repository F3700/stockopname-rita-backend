package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type StockOpnameRepository interface {
	Save(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	FindById(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameSummary, error)
	FindAll(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error)
	Update(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	Delete(ctx context.Context, tx pgx.Tx, id int) error
	FindAllForExport(ctx context.Context, sesiId int) ([]*model.StockOpnameExport, error)
}
