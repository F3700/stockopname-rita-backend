package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type RackRepository interface {
	FindAll(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackState, error)
	FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error)
	Save(ctx context.Context, tx pgx.Tx, rack *model.Rack) error
	ClaimForUpload(ctx context.Context, tx pgx.Tx, uploadID int, rakIDs []int32, coorID int) (int64, error)
	SetSnapshot(ctx context.Context, tx pgx.Tx, rakID int, snapshot string) error
	FindById(ctx context.Context, tx pgx.Tx, rakID int) (*model.RackState, error)
	FindByIdForUpdate(ctx context.Context, tx pgx.Tx, rakID int) (*model.RackState, error)
	CountByCoordinator(ctx context.Context, tx pgx.Tx, coorID int) (int, error)
	Delete(ctx context.Context, tx pgx.Tx, rakID int) error
}
