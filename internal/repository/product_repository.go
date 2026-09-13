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
	FindAll(ctx context.Context) ([]*model.Product, error)
	FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Product, int, error)
	FindAllUpdatedAfter(ctx context.Context, date time.Time, limit int, offset int) ([]*model.Product, int, error)
	FindAllLastSession(ctx context.Context, limit int) ([]*model.ProductLastSession, error)
	FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error)
}
