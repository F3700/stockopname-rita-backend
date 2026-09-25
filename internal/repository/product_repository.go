package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProductRepository interface {
	Save(ctx context.Context, tx pgx.Tx, product *model.Product) error
	Update(ctx context.Context, tx pgx.Tx, product *model.Product) error
	Delete(ctx context.Context, tx pgx.Tx, id int) error
	// Clear removes ALL master data (product, barcode, deleted_product).
	// Stock opname history is untouched (snapshot, no FK to product).
	Clear(ctx context.Context, tx pgx.Tx) error
	FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error)
	FindByPLU(ctx context.Context, tx pgx.Tx, plu string) (*model.Product, error)
	FindByBarcode(ctx context.Context, tx pgx.Tx, code string) (*model.Product, error)
	// FindAllForImport loads every product with barcodes for import diffing.
	FindAllForImport(ctx context.Context, tx pgx.Tx) ([]*model.Product, error)
	FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Product, int, error)
	FindAllUpdatedAfter(ctx context.Context, date time.Time, limit int, offset int) ([]*model.Product, int, error)
	FindAllLastSession(ctx context.Context, limit int) ([]*model.ProductLastSession, error)
}
