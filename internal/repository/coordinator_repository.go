package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type CoordinatorRepository interface {
	Save(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error
	FindAllSummary(ctx context.Context, tx pgx.Tx) ([]*model.CoordinatorSummary, error)
	FindBySesiIdSummary(ctx context.Context, tx pgx.Tx, sesiId int) ([]*model.CoordinatorSummary, error)
	FindByIdSummary(ctx context.Context, tx pgx.Tx, id int) (*model.CoordinatorSummary, error)
	Update(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error
}
