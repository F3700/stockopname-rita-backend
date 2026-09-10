package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type InspectorRepository interface {
	FindAll(ctx context.Context) ([]*model.InspectorSummary, error)
	FindByCoorId(ctx context.Context, coordinatorId int) ([]*model.InspectorSummary, error)
	Save(ctx context.Context, tx pgx.Tx, inspector *model.Inspector) error
}
