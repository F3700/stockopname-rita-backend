package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type CategoryRepository interface {
	Save(ctx context.Context, tx pgx.Tx, category *model.Category) error
	FindAll(ctx context.Context) ([]*model.Category, error)
	FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Category, int, error)
	FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Category, error)
	Update(ctx context.Context, tx pgx.Tx, category *model.Category) error
	Delete(ctx context.Context, tx pgx.Tx, id int) error
}
