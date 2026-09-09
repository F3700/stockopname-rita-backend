package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type SesiRepository interface {
	FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Sesi, error)
	FindAll(ctx context.Context, tx pgx.Tx) ([]*model.Sesi, error)
	Save(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error
	Update(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error
	Delete(ctx context.Context, tx pgx.Tx, id int) error
}
