package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"
)

type InspectorRepository interface {
	FindAll(ctx context.Context) ([]*model.InspectorSummary, error)
	FindByCoorId(ctx context.Context, coordinatorId int) ([]*model.InspectorSummary, error)
}
