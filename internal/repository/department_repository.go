package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type DepartmentRepository interface {
	Save(ctx context.Context, tx pgx.Tx, department *model.Department) error
	Update(ctx context.Context, tx pgx.Tx, department *model.Department) error
	Delete(ctx context.Context, tx pgx.Tx, id int) error
	FindAll(ctx context.Context) ([]*model.Department, error)
	FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Department, int, error)
	FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Department, error)
}
