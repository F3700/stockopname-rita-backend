package service

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

var _ repository.BarcodeRepository = (*fakeBarcodeRepository)(nil)
var _ repository.CoordinatorRepository = (*fakeCoordinatorRepository)(nil)
var _ repository.ProductRepository = (*fakeProductRepository)(nil)
var _ repository.RackRepository = (*fakeRackRepository)(nil)
var _ repository.SesiRepository = (*fakeSesiRepository)(nil)
var _ repository.StockOpnameRepository = (*fakeStockOpnameRepository)(nil)
var _ repository.UploadRepository = (*fakeUploadRepository)(nil)
var _ repository.AuditRepository = (*fakeAuditRepository)(nil)

var _ SesiService = (*fakeSesiService)(nil)
var _ CoordinatorService = (*fakeCoordinatorService)(nil)
var _ RackService = (*fakeRackService)(nil)
var _ UploadService = (*fakeUploadService)(nil)
var _ StockOpnameService = (*fakeStockOpnameService)(nil)

func newTestValidator(t *testing.T) *validator.Validate {
	t.Helper()
	return validator.New()
}

func qtyPtr(v float64) *float64 { return &v }

type fakeBarcodeRepository struct {
	saveBatchFunc func(ctx context.Context, tx pgx.Tx, sesiId int, productID int, codes []string) error
	replaceFunc   func(ctx context.Context, tx pgx.Tx, sesiId int, productID int, codes []string) error
}

func (f *fakeBarcodeRepository) SaveBatch(ctx context.Context, tx pgx.Tx, sesiId int, productID int, codes []string) error {
	if f.saveBatchFunc == nil {
		return nil
	}
	return f.saveBatchFunc(ctx, tx, sesiId, productID, codes)
}

func (f *fakeBarcodeRepository) Replace(ctx context.Context, tx pgx.Tx, sesiId int, productID int, codes []string) error {
	if f.replaceFunc == nil {
		return nil
	}
	return f.replaceFunc(ctx, tx, sesiId, productID, codes)
}

type fakeCoordinatorRepository struct {
	saveFunc               func(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error
	findAllSummaryFunc     func(ctx context.Context) ([]*model.CoordinatorSummary, error)
	findBySesiIdSummaryFn  func(ctx context.Context, sesiId int) ([]*model.CoordinatorSummary, error)
	findByIdSummaryFunc    func(ctx context.Context, id int) (*model.CoordinatorSummary, error)
	findByIdReportFunc     func(ctx context.Context, id int) (*model.CoordinatorReport, error)
	findByIdDetailFunc     func(ctx context.Context, id int) (*model.CoordinatorDetail, error)
	updateFunc             func(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error
	lockByIdFunc           func(ctx context.Context, tx pgx.Tx, coorID int) error
	recalcStatusFunc       func(ctx context.Context, tx pgx.Tx, coorID int) error
	recalcAfterChangeFunc  func(ctx context.Context, tx pgx.Tx, coorID int) error
	findBySesiAndCoorCodeF func(ctx context.Context, tx pgx.Tx, sesiCode string, coorCode string) (*model.Coordinator, error)
	findByIdFunc           func(ctx context.Context, tx pgx.Tx, id int) (*model.Coordinator, error)
}

func (f *fakeCoordinatorRepository) Save(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error {
	return f.saveFunc(ctx, tx, coordinator)
}

func (f *fakeCoordinatorRepository) FindAllSummary(ctx context.Context) ([]*model.CoordinatorSummary, error) {
	return f.findAllSummaryFunc(ctx)
}

func (f *fakeCoordinatorRepository) FindBySesiIdSummary(ctx context.Context, sesiId int) ([]*model.CoordinatorSummary, error) {
	return f.findBySesiIdSummaryFn(ctx, sesiId)
}

func (f *fakeCoordinatorRepository) FindByIdSummary(ctx context.Context, id int) (*model.CoordinatorSummary, error) {
	return f.findByIdSummaryFunc(ctx, id)
}

func (f *fakeCoordinatorRepository) FindByIdReport(ctx context.Context, id int) (*model.CoordinatorReport, error) {
	return f.findByIdReportFunc(ctx, id)
}

func (f *fakeCoordinatorRepository) FindByIdDetail(ctx context.Context, id int) (*model.CoordinatorDetail, error) {
	if f.findByIdDetailFunc == nil {
		return nil, nil
	}
	return f.findByIdDetailFunc(ctx, id)
}

func (f *fakeCoordinatorRepository) Update(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error {
	return f.updateFunc(ctx, tx, coordinator)
}

func (f *fakeCoordinatorRepository) LockById(ctx context.Context, tx pgx.Tx, coorID int) error {
	if f.lockByIdFunc == nil {
		return nil
	}
	return f.lockByIdFunc(ctx, tx, coorID)
}

func (f *fakeCoordinatorRepository) RecalcStatus(ctx context.Context, tx pgx.Tx, coorID int) error {
	if f.recalcStatusFunc == nil {
		return nil
	}
	return f.recalcStatusFunc(ctx, tx, coorID)
}

func (f *fakeCoordinatorRepository) RecalcStatusAfterResultChange(ctx context.Context, tx pgx.Tx, coorID int) error {
	if f.recalcAfterChangeFunc == nil {
		return nil
	}
	return f.recalcAfterChangeFunc(ctx, tx, coorID)
}

func (f *fakeCoordinatorRepository) FindBySesiAndCoorCode(ctx context.Context, tx pgx.Tx, sesiCode string, coorCode string) (*model.Coordinator, error) {
	return f.findBySesiAndCoorCodeF(ctx, tx, sesiCode, coorCode)
}

func (f *fakeCoordinatorRepository) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Coordinator, error) {
	if f.findByIdFunc == nil {
		return nil, nil
	}
	return f.findByIdFunc(ctx, tx, id)
}

type fakeProductRepository struct {
	saveFunc              func(ctx context.Context, tx pgx.Tx, product *model.Product) error
	updateFunc            func(ctx context.Context, tx pgx.Tx, product *model.Product) error
	deleteFunc            func(ctx context.Context, tx pgx.Tx, id int) error
	clearFunc             func(ctx context.Context, tx pgx.Tx, sesiId int) error
	countBySesiFunc       func(ctx context.Context, tx pgx.Tx, sesiId int) (int, error)
	findAllInPageSearchFn func(ctx context.Context, sesiId int, limit int, offset int, search string) ([]*model.Product, int, error)
	findAllForImportFunc  func(ctx context.Context, tx pgx.Tx, sesiId int) ([]*model.Product, error)
	findByIdFn            func(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error)
	findByPLUFn           func(ctx context.Context, tx pgx.Tx, sesiId int, plu string) (*model.Product, error)
	findByBarcodeFn       func(ctx context.Context, tx pgx.Tx, sesiId int, code string) (*model.Product, error)
	findByCodesOrPLUsFn   func(ctx context.Context, tx pgx.Tx, sesiId int, codes []string, plus []string) ([]*model.Product, error)
}

func (f *fakeProductRepository) Save(ctx context.Context, tx pgx.Tx, product *model.Product) error {
	return f.saveFunc(ctx, tx, product)
}

func (f *fakeProductRepository) Update(ctx context.Context, tx pgx.Tx, product *model.Product) error {
	return f.updateFunc(ctx, tx, product)
}

func (f *fakeProductRepository) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	return f.deleteFunc(ctx, tx, id)
}

func (f *fakeProductRepository) Clear(ctx context.Context, tx pgx.Tx, sesiId int) error {
	return f.clearFunc(ctx, tx, sesiId)
}

func (f *fakeProductRepository) CountBySesi(ctx context.Context, tx pgx.Tx, sesiId int) (int, error) {
	if f.countBySesiFunc == nil {
		return 0, nil
	}
	return f.countBySesiFunc(ctx, tx, sesiId)
}

func (f *fakeProductRepository) FindAllForImport(ctx context.Context, tx pgx.Tx, sesiId int) ([]*model.Product, error) {
	return f.findAllForImportFunc(ctx, tx, sesiId)
}

func (f *fakeProductRepository) FindByPLU(ctx context.Context, tx pgx.Tx, sesiId int, plu string) (*model.Product, error) {
	return f.findByPLUFn(ctx, tx, sesiId, plu)
}

func (f *fakeProductRepository) FindByBarcode(ctx context.Context, tx pgx.Tx, sesiId int, code string) (*model.Product, error) {
	return f.findByBarcodeFn(ctx, tx, sesiId, code)
}

func (f *fakeProductRepository) FindByBarcodesOrPLUs(ctx context.Context, tx pgx.Tx, sesiId int, codes []string, plus []string) ([]*model.Product, error) {
	if f.findByCodesOrPLUsFn == nil {
		return nil, nil
	}
	return f.findByCodesOrPLUsFn(ctx, tx, sesiId, codes, plus)
}

func (f *fakeProductRepository) FindAllInPageSearch(ctx context.Context, sesiId int, limit int, offset int, search string) ([]*model.Product, int, error) {
	return f.findAllInPageSearchFn(ctx, sesiId, limit, offset, search)
}

func (f *fakeProductRepository) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error) {
	return f.findByIdFn(ctx, tx, id)
}

type fakeRackRepository struct {
	findAllFunc            func(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackState, error)
	findProgressFunc       func(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error)
	saveFunc               func(ctx context.Context, tx pgx.Tx, rack *model.Rack) error
	claimForUploadFunc     func(ctx context.Context, tx pgx.Tx, uploadID int, rakIDs []int32, coorID int) (int64, error)
	setSnapshotFunc        func(ctx context.Context, tx pgx.Tx, rakID int, snapshot string) error
	findByIdFunc           func(ctx context.Context, tx pgx.Tx, rakID int) (*model.RackState, error)
	findByIdForUpdateFunc  func(ctx context.Context, tx pgx.Tx, rakID int) (*model.RackState, error)
	countByCoordinatorFunc func(ctx context.Context, tx pgx.Tx, coorID int) (int, error)
	deleteFunc             func(ctx context.Context, tx pgx.Tx, rakID int) error
}

func (f *fakeRackRepository) FindAll(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackState, error) {
	return f.findAllFunc(ctx, coordinatorId, sessionId)
}

func (f *fakeRackRepository) FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error) {
	return f.findProgressFunc(ctx, coordinatorId, sessionId)
}

func (f *fakeRackRepository) Save(ctx context.Context, tx pgx.Tx, rack *model.Rack) error {
	return f.saveFunc(ctx, tx, rack)
}

func (f *fakeRackRepository) ClaimForUpload(ctx context.Context, tx pgx.Tx, uploadID int, rakIDs []int32, coorID int) (int64, error) {
	if f.claimForUploadFunc == nil {
		return int64(len(rakIDs)), nil
	}
	return f.claimForUploadFunc(ctx, tx, uploadID, rakIDs, coorID)
}

func (f *fakeRackRepository) SetSnapshot(ctx context.Context, tx pgx.Tx, rakID int, snapshot string) error {
	if f.setSnapshotFunc == nil {
		return nil
	}
	return f.setSnapshotFunc(ctx, tx, rakID, snapshot)
}

func (f *fakeRackRepository) FindById(ctx context.Context, tx pgx.Tx, rakID int) (*model.RackState, error) {
	if f.findByIdFunc == nil {
		return nil, nil
	}
	return f.findByIdFunc(ctx, tx, rakID)
}

func (f *fakeRackRepository) FindByIdForUpdate(ctx context.Context, tx pgx.Tx, rakID int) (*model.RackState, error) {
	if f.findByIdForUpdateFunc == nil {
		return nil, nil
	}
	return f.findByIdForUpdateFunc(ctx, tx, rakID)
}

func (f *fakeRackRepository) CountByCoordinator(ctx context.Context, tx pgx.Tx, coorID int) (int, error) {
	if f.countByCoordinatorFunc == nil {
		return 0, nil
	}
	return f.countByCoordinatorFunc(ctx, tx, coorID)
}

func (f *fakeRackRepository) Delete(ctx context.Context, tx pgx.Tx, rakID int) error {
	if f.deleteFunc == nil {
		return nil
	}
	return f.deleteFunc(ctx, tx, rakID)
}

type fakeUploadRepository struct {
	saveFunc           func(ctx context.Context, tx pgx.Tx, upload *model.Upload) error
	findAllSummaryFunc func(ctx context.Context, coordinatorId int) ([]*model.UploadSummary, error)
	findDetailFunc     func(ctx context.Context, uploadID int) (*model.UploadDetail, error)
	findSummaryFunc    func(ctx context.Context, coordinatorId int) ([]*model.UploadReportRow, error)
}

func (f *fakeUploadRepository) Save(ctx context.Context, tx pgx.Tx, upload *model.Upload) error {
	if f.saveFunc == nil {
		upload.UploadID = 1
		return nil
	}
	return f.saveFunc(ctx, tx, upload)
}

func (f *fakeUploadRepository) FindAllSummary(ctx context.Context, coordinatorId int) ([]*model.UploadSummary, error) {
	if f.findAllSummaryFunc == nil {
		return nil, nil
	}
	return f.findAllSummaryFunc(ctx, coordinatorId)
}

func (f *fakeUploadRepository) FindDetail(ctx context.Context, uploadID int) (*model.UploadDetail, error) {
	if f.findDetailFunc == nil {
		return nil, nil
	}
	return f.findDetailFunc(ctx, uploadID)
}

func (f *fakeUploadRepository) FindSummaryByCoordinator(ctx context.Context, coordinatorId int) ([]*model.UploadReportRow, error) {
	if f.findSummaryFunc == nil {
		return nil, nil
	}
	return f.findSummaryFunc(ctx, coordinatorId)
}

type fakeAuditRepository struct {
	saveFunc        func(ctx context.Context, tx pgx.Tx, audit *model.StockOpnameAudit) error
	findAllFunc     func(ctx context.Context, filter repository.AuditFilter) ([]*model.StockOpnameAudit, int, error)
	findBatchesFunc func(ctx context.Context, filter repository.AuditFilter) ([]*model.StockOpnameAuditBatch, int, error)
}

func (f *fakeAuditRepository) Save(ctx context.Context, tx pgx.Tx, audit *model.StockOpnameAudit) error {
	if f.saveFunc == nil {
		return nil
	}
	return f.saveFunc(ctx, tx, audit)
}

func (f *fakeAuditRepository) FindAll(ctx context.Context, filter repository.AuditFilter) ([]*model.StockOpnameAudit, int, error) {
	if f.findAllFunc == nil {
		return nil, 0, nil
	}
	return f.findAllFunc(ctx, filter)
}

func (f *fakeAuditRepository) FindBatches(ctx context.Context, filter repository.AuditFilter) ([]*model.StockOpnameAuditBatch, int, error) {
	if f.findBatchesFunc == nil {
		return nil, 0, nil
	}
	return f.findBatchesFunc(ctx, filter)
}

type fakeSesiRepository struct {
	findByIdFn            func(ctx context.Context, id int) (*model.Sesi, error)
	findByIdForShareFn    func(ctx context.Context, tx pgx.Tx, id int) (*model.Sesi, error)
	findByIdForUpdateFn   func(ctx context.Context, tx pgx.Tx, id int) (*model.Sesi, error)
	findAllInPageSearchFn func(ctx context.Context, limit int, offset int, search string) ([]*model.Sesi, int, error)
	saveFunc              func(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error
	updateFunc            func(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error
	deleteFunc            func(ctx context.Context, tx pgx.Tx, id int) error
}

func (f *fakeSesiRepository) FindById(ctx context.Context, id int) (*model.Sesi, error) {
	return f.findByIdFn(ctx, id)
}

func (f *fakeSesiRepository) FindByIdForShare(ctx context.Context, tx pgx.Tx, id int) (*model.Sesi, error) {
	if f.findByIdForShareFn == nil {
		return nil, nil
	}
	return f.findByIdForShareFn(ctx, tx, id)
}

func (f *fakeSesiRepository) FindByIdForUpdate(ctx context.Context, tx pgx.Tx, id int) (*model.Sesi, error) {
	if f.findByIdForUpdateFn == nil {
		return nil, nil
	}
	return f.findByIdForUpdateFn(ctx, tx, id)
}

func (f *fakeSesiRepository) FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Sesi, int, error) {
	return f.findAllInPageSearchFn(ctx, limit, offset, search)
}

func (f *fakeSesiRepository) Save(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error {
	return f.saveFunc(ctx, tx, sesi)
}

func (f *fakeSesiRepository) Update(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error {
	return f.updateFunc(ctx, tx, sesi)
}

func (f *fakeSesiRepository) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	return f.deleteFunc(ctx, tx, id)
}

type fakeStockOpnameRepository struct {
	saveManyFunc                func(ctx context.Context, tx pgx.Tx, stockOpnames []*model.StockOpname) error
	createFunc                  func(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	findByIdFn                  func(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameSummary, error)
	findByIdForUpdateFn         func(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameAuditData, error)
	findAllFunc                 func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int, rakId *int) ([]*model.StockOpnameSummary, int, error)
	updateFunc                  func(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	deleteFunc                  func(ctx context.Context, tx pgx.Tx, id int) error
	findAllForExportFunc        func(ctx context.Context, sesiId *int, coorId *int) ([]*model.StockOpnameExport, error)
	findSesiIdByRakFunc         func(ctx context.Context, tx pgx.Tx, rakID int) (int, error)
	findSesiIdByStockOpnameFunc func(ctx context.Context, tx pgx.Tx, stockOpnameID int) (int, error)
	findRackContextFunc         func(ctx context.Context, tx pgx.Tx, rakID int) (*model.StockOpnameRackContext, error)
	findForMoveFunc             func(ctx context.Context, tx pgx.Tx, rakID int, ids []int) ([]*model.StockOpnameAuditData, error)
	findByRakAndPLUFunc         func(ctx context.Context, tx pgx.Tx, rakID int, plu string) (*model.StockOpnameAuditData, error)
	upsertForMoveFunc           func(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) (*model.StockOpnameAuditData, error)
	countBySesiFunc             func(ctx context.Context, tx pgx.Tx, sesiId int) (int, error)
	countByRakFunc              func(ctx context.Context, tx pgx.Tx, rakID int) (int, error)
	countForQtyFactorFunc       func(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) (int, error)
	findForQtyFactorFunc        func(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) ([]*model.StockOpnameAuditData, error)
	updateQtyFactorFunc         func(ctx context.Context, tx pgx.Tx, id int, factor int) error
	findSellSummaryFunc         func(ctx context.Context, sessionID int, groupBy string) ([]*model.SellSummaryGroup, error)
}

func (f *fakeStockOpnameRepository) Create(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
	if f.createFunc == nil {
		return nil
	}
	return f.createFunc(ctx, tx, stockOpname)
}

func (f *fakeStockOpnameRepository) SaveMany(ctx context.Context, tx pgx.Tx, stockOpnames []*model.StockOpname) error {
	if f.saveManyFunc == nil {
		return nil
	}
	return f.saveManyFunc(ctx, tx, stockOpnames)
}

func (f *fakeStockOpnameRepository) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameSummary, error) {
	return f.findByIdFn(ctx, tx, id)
}

func (f *fakeStockOpnameRepository) FindAll(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int, rakId *int) ([]*model.StockOpnameSummary, int, error) {
	return f.findAllFunc(ctx, limit, offset, search, sesiId, coorId, rakId)
}

func (f *fakeStockOpnameRepository) Update(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
	return f.updateFunc(ctx, tx, stockOpname)
}

func (f *fakeStockOpnameRepository) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	return f.deleteFunc(ctx, tx, id)
}

func (f *fakeStockOpnameRepository) FindByIdForUpdate(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameAuditData, error) {
	if f.findByIdForUpdateFn == nil {
		return nil, nil
	}
	return f.findByIdForUpdateFn(ctx, tx, id)
}

func (f *fakeStockOpnameRepository) FindRackContextForUpdate(ctx context.Context, tx pgx.Tx, rakID int) (*model.StockOpnameRackContext, error) {
	if f.findRackContextFunc == nil {
		return nil, nil
	}
	return f.findRackContextFunc(ctx, tx, rakID)
}

func (f *fakeStockOpnameRepository) FindForMove(ctx context.Context, tx pgx.Tx, rakID int, ids []int) ([]*model.StockOpnameAuditData, error) {
	if f.findForMoveFunc == nil {
		return nil, nil
	}
	return f.findForMoveFunc(ctx, tx, rakID, ids)
}

func (f *fakeStockOpnameRepository) FindByRakAndPLUForUpdate(ctx context.Context, tx pgx.Tx, rakID int, plu string) (*model.StockOpnameAuditData, error) {
	if f.findByRakAndPLUFunc == nil {
		return nil, nil
	}
	return f.findByRakAndPLUFunc(ctx, tx, rakID, plu)
}

func (f *fakeStockOpnameRepository) UpsertForMove(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) (*model.StockOpnameAuditData, error) {
	if f.upsertForMoveFunc == nil {
		return &model.StockOpnameAuditData{
			StockOpnameID:     stockOpname.StockOpnameID,
			RakID:             stockOpname.StockOpnameRakID,
			PLU:               stockOpname.SoProductPLU,
			Name:              stockOpname.SoProductName,
			Barcode:           stockOpname.SoBarcode,
			Quantity:          stockOpname.StockOpnameQuantity,
			BuyPrice:          stockOpname.SoBuyPrice,
			SellPrice:         stockOpname.SoSellPrice,
			QtyFactor:         stockOpname.SoQtyFactor,
			EffectiveQuantity: stockOpname.StockOpnameQuantity * float64(stockOpname.SoQtyFactor),
		}, nil
	}
	return f.upsertForMoveFunc(ctx, tx, stockOpname)
}

func (f *fakeStockOpnameRepository) FindAllForExport(ctx context.Context, sesiId *int, coorId *int) ([]*model.StockOpnameExport, error) {
	return f.findAllForExportFunc(ctx, sesiId, coorId)
}

func (f *fakeStockOpnameRepository) FindSesiIdByRak(ctx context.Context, tx pgx.Tx, rakID int) (int, error) {
	if f.findSesiIdByRakFunc == nil {
		return 0, nil
	}
	return f.findSesiIdByRakFunc(ctx, tx, rakID)
}

func (f *fakeStockOpnameRepository) FindSesiIdByStockOpnameID(ctx context.Context, tx pgx.Tx, stockOpnameID int) (int, error) {
	if f.findSesiIdByStockOpnameFunc == nil {
		return 0, nil
	}
	return f.findSesiIdByStockOpnameFunc(ctx, tx, stockOpnameID)
}

func (f *fakeStockOpnameRepository) CountBySesi(ctx context.Context, tx pgx.Tx, sesiId int) (int, error) {
	if f.countBySesiFunc == nil {
		return 0, nil
	}
	return f.countBySesiFunc(ctx, tx, sesiId)
}

func (f *fakeStockOpnameRepository) CountByRak(ctx context.Context, tx pgx.Tx, rakID int) (int, error) {
	if f.countByRakFunc == nil {
		return 0, nil
	}
	return f.countByRakFunc(ctx, tx, rakID)
}

func (f *fakeStockOpnameRepository) CountForQtyFactorUpdate(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) (int, error) {
	if f.countForQtyFactorFunc == nil {
		return 0, nil
	}
	return f.countForQtyFactorFunc(ctx, tx, sessionID, scope, scopeID, resultIDs, filter)
}

func (f *fakeStockOpnameRepository) FindForQtyFactorUpdate(ctx context.Context, tx pgx.Tx, sessionID int, scope string, scopeID *int, resultIDs []int, filter *model.QtyFactorFilter) ([]*model.StockOpnameAuditData, error) {
	if f.findForQtyFactorFunc == nil {
		return nil, nil
	}
	return f.findForQtyFactorFunc(ctx, tx, sessionID, scope, scopeID, resultIDs, filter)
}

func (f *fakeStockOpnameRepository) UpdateQtyFactor(ctx context.Context, tx pgx.Tx, id int, factor int) error {
	if f.updateQtyFactorFunc == nil {
		return nil
	}
	return f.updateQtyFactorFunc(ctx, tx, id, factor)
}

func (f *fakeStockOpnameRepository) FindSellSummary(ctx context.Context, sessionID int, groupBy string) ([]*model.SellSummaryGroup, error) {
	if f.findSellSummaryFunc == nil {
		return nil, nil
	}
	return f.findSellSummaryFunc(ctx, sessionID, groupBy)
}

type fakeSesiService struct {
	findByIdFn func(ctx context.Context, id int) (dto.SesiResponse, error)
}

func (f *fakeSesiService) Create(ctx context.Context, req dto.CreateSesiRequest) (dto.SesiResponse, error) {
	return dto.SesiResponse{}, nil
}

func (f *fakeSesiService) Update(ctx context.Context, id int, req dto.UpdateSesiRequest) error {
	return nil
}

func (f *fakeSesiService) Delete(ctx context.Context, id int) error { return nil }

func (f *fakeSesiService) FindById(ctx context.Context, id int) (dto.SesiResponse, error) {
	return f.findByIdFn(ctx, id)
}

func (f *fakeSesiService) FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.SesiResponse, error) {
	return nil, nil
}

type fakeCoordinatorService struct {
	findAllSummaryFunc  func(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error)
	findByIdSummaryFunc func(ctx context.Context, id int) (*dto.CoordinatorResponse, error)
	findByIdDetailFunc  func(ctx context.Context, id int) (*dto.CoordinatorDetailResponse, error)
	findByIdReportFunc  func(ctx context.Context, id int) (*dto.CoordinatorReportResponse, error)
}

func (f *fakeCoordinatorService) Create(ctx context.Context, req dto.CreateCoordinatorRequest) (dto.CoordinatorResponse, error) {
	return dto.CoordinatorResponse{}, nil
}

func (f *fakeCoordinatorService) FindByIdSummary(ctx context.Context, id int) (*dto.CoordinatorResponse, error) {
	return f.findByIdSummaryFunc(ctx, id)
}

func (f *fakeCoordinatorService) FindByIdDetail(ctx context.Context, id int) (*dto.CoordinatorDetailResponse, error) {
	if f.findByIdDetailFunc == nil {
		return nil, nil
	}
	return f.findByIdDetailFunc(ctx, id)
}

func (f *fakeCoordinatorService) FindByIdReport(ctx context.Context, id int) (*dto.CoordinatorReportResponse, error) {
	return f.findByIdReportFunc(ctx, id)
}

func (f *fakeCoordinatorService) FindAllSummary(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error) {
	return f.findAllSummaryFunc(ctx, sesiId)
}

func (f *fakeCoordinatorService) JoinByCoordinatorQR(ctx context.Context, req dto.JoinByCoordinatorQRRequest) (dto.CoordinatorJoinResponse, error) {
	return dto.CoordinatorJoinResponse{}, nil
}

func (f *fakeCoordinatorService) Update(ctx context.Context, id int, req *dto.UpdateCoordinatorRequest) error {
	return nil
}

type fakeRackService struct {
	findAllSummaryFunc func(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error)
	findProgressFunc   func(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error)
	deleteFunc         func(ctx context.Context, rakID int) error
}

func (f *fakeRackService) FindAllSummary(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error) {
	if f.findAllSummaryFunc == nil {
		return nil, nil
	}
	return f.findAllSummaryFunc(ctx, coordinatorId, sessionId)
}

func (f *fakeRackService) FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error) {
	if f.findProgressFunc == nil {
		return nil, nil
	}
	return f.findProgressFunc(ctx, coordinatorId, sessionId)
}

func (f *fakeRackService) DeleteRack(ctx context.Context, rakID int) error {
	if f.deleteFunc == nil {
		return nil
	}
	return f.deleteFunc(ctx, rakID)
}

type fakeUploadService struct {
	createFunc         func(ctx context.Context, req dto.CreateUploadRequest) (dto.UploadResponse, error)
	findAllSummaryFunc func(ctx context.Context, coordinatorId int) ([]dto.UploadResponse, error)
	findDetailFunc     func(ctx context.Context, uploadID int) (dto.UploadDetailResponse, error)
	findSummaryFunc    func(ctx context.Context, coordinatorId int) ([]*model.UploadReportRow, error)
}

func (f *fakeUploadService) Create(ctx context.Context, req dto.CreateUploadRequest) (dto.UploadResponse, error) {
	if f.createFunc == nil {
		return dto.UploadResponse{}, nil
	}
	return f.createFunc(ctx, req)
}

func (f *fakeUploadService) FindAllSummary(ctx context.Context, coordinatorId int) ([]dto.UploadResponse, error) {
	if f.findAllSummaryFunc == nil {
		return nil, nil
	}
	return f.findAllSummaryFunc(ctx, coordinatorId)
}

func (f *fakeUploadService) FindDetail(ctx context.Context, uploadID int) (dto.UploadDetailResponse, error) {
	if f.findDetailFunc == nil {
		return dto.UploadDetailResponse{}, nil
	}
	return f.findDetailFunc(ctx, uploadID)
}

func (f *fakeUploadService) FindSummaryByCoordinator(ctx context.Context, coordinatorId int) ([]*model.UploadReportRow, error) {
	if f.findSummaryFunc == nil {
		return nil, nil
	}
	return f.findSummaryFunc(ctx, coordinatorId)
}

type fakeStockOpnameService struct {
	findAllFunc          func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int, rakId *int) ([]dto.StockOpnameResponse, error)
	findAllForExportFunc func(ctx context.Context, sesiId *int, coorId *int) ([]dto.StockOpnameExportResponse, error)
	setQtyFactorFunc     func(ctx context.Context, sessionID int, req dto.SetQtyFactorRequest) (dto.SetQtyFactorResponse, error)
	findSellSummaryFunc  func(ctx context.Context, sessionID int, groupBy string) (dto.SellSummaryResponse, error)
}

func (f *fakeStockOpnameService) Create(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	return dto.StockOpnameResponse{}, nil
}

func (f *fakeStockOpnameService) Update(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	return dto.StockOpnameResponse{}, nil
}

func (f *fakeStockOpnameService) Delete(ctx context.Context, id int, actor string) error { return nil }

func (f *fakeStockOpnameService) Move(ctx context.Context, req dto.MoveStockOpnameRequest) (dto.MoveStockOpnameResponse, error) {
	return dto.MoveStockOpnameResponse{}, nil
}

func (f *fakeStockOpnameService) FindById(ctx context.Context, id int) (dto.StockOpnameResponse, error) {
	return dto.StockOpnameResponse{}, nil
}

func (f *fakeStockOpnameService) FindAll(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int, rakId *int) ([]dto.StockOpnameResponse, error) {
	return f.findAllFunc(ctx, pagination, search, coorId, sesiId, rakId)
}

func (f *fakeStockOpnameService) FindAllForExport(ctx context.Context, sesiId *int, coorId *int) ([]dto.StockOpnameExportResponse, error) {
	return f.findAllForExportFunc(ctx, sesiId, coorId)
}

func (f *fakeStockOpnameService) SetQtyFactor(ctx context.Context, sessionID int, req dto.SetQtyFactorRequest) (dto.SetQtyFactorResponse, error) {
	if f.setQtyFactorFunc == nil {
		return dto.SetQtyFactorResponse{}, nil
	}
	return f.setQtyFactorFunc(ctx, sessionID, req)
}

func (f *fakeStockOpnameService) FindSellSummary(ctx context.Context, sessionID int, groupBy string) (dto.SellSummaryResponse, error) {
	if f.findSellSummaryFunc == nil {
		return dto.SellSummaryResponse{}, nil
	}
	return f.findSellSummaryFunc(ctx, sessionID, groupBy)
}
