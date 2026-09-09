package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type SesiRepository interface {
	Save(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error
	Update(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error
}
