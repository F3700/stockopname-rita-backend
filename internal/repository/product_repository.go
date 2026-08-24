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
	FindAll(ctx context.Context) ([]*model.Product, error)
	FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error)
}
