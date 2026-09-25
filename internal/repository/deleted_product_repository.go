package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
)

type DeletedProductRepository interface {
	Save(ctx context.Context, tx pgx.Tx, plu string) error
	Delete(ctx context.Context) error
	FindAll(ctx context.Context) ([]*model.DeletedProduct, error)
	FindAllUpdatedAfter(ctx context.Context, date time.Time) ([]*model.DeletedProduct, error)
}
