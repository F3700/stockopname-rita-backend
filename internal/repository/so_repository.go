package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type StockOpnameRepository interface {
	Create(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	// SaveMany upserts a whole upload batch in one statement. Callers must not
	// repeat a (rack, plu) pair: Postgres rejects a second conflict on it.
	SaveMany(ctx context.Context, tx pgx.Tx, stockOpnames []*model.StockOpname) error
	FindById(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameSummary, error)
	FindByIdForUpdate(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameAuditData, error)
	FindAll(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int, rakId *int) ([]*model.StockOpnameSummary, int, error)
	Update(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	Delete(ctx context.Context, tx pgx.Tx, id int) error
	FindAllForExport(ctx context.Context, sesiId *int, coorId *int) ([]*model.StockOpnameExport, error)
	FindSesiIdByRak(ctx context.Context, tx pgx.Tx, rakID int) (int, error)
	FindSesiIdByStockOpnameID(ctx context.Context, tx pgx.Tx, stockOpnameID int) (int, error)
	FindRackContextForUpdate(ctx context.Context, tx pgx.Tx, rakID int) (*model.StockOpnameRackContext, error)
	FindForMove(ctx context.Context, tx pgx.Tx, rakID int, ids []int) ([]*model.StockOpnameAuditData, error)
	FindByRakAndPLUForUpdate(ctx context.Context, tx pgx.Tx, rakID int, plu string) (*model.StockOpnameAuditData, error)
	UpsertForMove(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) (*model.StockOpnameAuditData, error)
	CountBySesi(ctx context.Context, tx pgx.Tx, sesiId int) (int, error)
	CountByRak(ctx context.Context, tx pgx.Tx, rakID int) (int, error)
	ReopenUploadIfEmpty(ctx context.Context, tx pgx.Tx, rakID int) (bool, error)
	CountForQtyFactorUpdate(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) (int, error)
	FindForQtyFactorUpdate(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) ([]*model.StockOpnameAuditData, error)
	UpdateQtyFactor(ctx context.Context, tx pgx.Tx, id int, factor int) error
	FindSellSummary(ctx context.Context, sessionID int, groupBy string) ([]*model.SellSummaryGroup, error)
}
