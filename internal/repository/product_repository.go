package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type ProductRepository interface {
	Save(ctx context.Context, tx pgx.Tx, product *model.Product) error
	Update(ctx context.Context, tx pgx.Tx, product *model.Product) error
	Delete(ctx context.Context, tx pgx.Tx, id int) error
	// Clear removes all products of one session. Barcodes cascade.
	// Stock opname history is untouched (snapshot, no FK to product).
	Clear(ctx context.Context, tx pgx.Tx, sesiId int) error
	FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error)
	FindByPLU(ctx context.Context, tx pgx.Tx, sesiId int, plu string) (*model.Product, error)
	FindByBarcode(ctx context.Context, tx pgx.Tx, sesiId int, code string) (*model.Product, error)
	// FindAllForImport loads every product of one session with its barcodes.
	FindAllForImport(ctx context.Context, tx pgx.Tx, sesiId int) ([]*model.Product, error)
	FindAllInPageSearch(ctx context.Context, sesiId int, limit int, offset int, search string) ([]*model.Product, int, error)
	CountBySesi(ctx context.Context, tx pgx.Tx, sesiId int) (int, error)
}
