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
	// Every row must carry SoUploadID; the upload owns what it writes.
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
	// AdoptUnownedRows gives every row without an owner in the claimed racks to
	// one upload. Called right after the rack claim, before the batch is saved.
	AdoptUnownedRows(ctx context.Context, tx pgx.Tx, uploadID int, rakIDs []int32) (int64, error)
	// HasOwnedRows reports whether the upload still owns a row in the rack.
	HasOwnedRows(ctx context.Context, tx pgx.Tx, rakID int, uploadID int) (bool, error)
	// CountByUpload returns the number of rows and distinct racks owned by one
	// upload. Used for the upload response, so counts match the reports.
	CountByUpload(ctx context.Context, tx pgx.Tx, uploadID int) (items int, racks int, err error)
	// ClearUploadPointerIfNoOwnedRows releases the rack host once it no longer
	// owns any row in the rack. The snapshot is intentionally left untouched.
	ClearUploadPointerIfNoOwnedRows(ctx context.Context, tx pgx.Tx, rakID int) (bool, error)
	CountForQtyFactorUpdate(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) (int, error)
	FindForQtyFactorUpdate(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) ([]*model.StockOpnameAuditData, error)
	UpdateQtyFactor(ctx context.Context, tx pgx.Tx, id int, factor int) error
	FindSellSummary(ctx context.Context, sessionID int, groupBy string) ([]*model.SellSummaryGroup, error)
}
