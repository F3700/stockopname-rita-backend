package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type UploadRepository interface {
	Save(ctx context.Context, tx pgx.Tx, upload *model.Upload) error
	FindAllSummary(ctx context.Context, coordinatorId int) ([]*model.UploadSummary, error)
	FindDetail(ctx context.Context, uploadID int) (*model.UploadDetail, error)
}
