package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type RackRepository interface {
	FindAll(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]*model.Rack, error)
	FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error)
	Save(ctx context.Context, tx pgx.Tx, rack *model.Rack) error
}
