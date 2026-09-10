package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"
)

type RackRepository interface {
	FindAll(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]*model.Rack, error)
	FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error)
}
