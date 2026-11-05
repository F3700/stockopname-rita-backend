package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type AuditFilter struct {
	SesiID        *int
	CoordinatorID *int
	ResultID      *int
	Limit         int
	Offset        int
}

type AuditRepository interface {
	Save(ctx context.Context, tx pgx.Tx, audit *model.StockOpnameAudit) error
	FindAll(ctx context.Context, filter AuditFilter) ([]*model.StockOpnameAudit, int, error)
}
