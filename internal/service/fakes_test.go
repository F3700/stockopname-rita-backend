package service

import (
	"context"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

var _ repository.BarcodeRepository = (*fakeBarcodeRepository)(nil)
var _ repository.CoordinatorRepository = (*fakeCoordinatorRepository)(nil)
var _ repository.DeletedProductRepository = (*fakeDeletedProductRepository)(nil)
var _ repository.InspectorRepository = (*fakeInspectorRepository)(nil)
var _ repository.ProductRepository = (*fakeProductRepository)(nil)
var _ repository.RackRepository = (*fakeRackRepository)(nil)
var _ repository.SesiRepository = (*fakeSesiRepository)(nil)
var _ repository.StockOpnameRepository = (*fakeStockOpnameRepository)(nil)

var _ SesiService = (*fakeSesiService)(nil)
var _ CoordinatorService = (*fakeCoordinatorService)(nil)
var _ InspectorService = (*fakeInspectorService)(nil)
var _ StockOpnameService = (*fakeStockOpnameService)(nil)

func newTestValidator(t *testing.T) *validator.Validate {
	t.Helper()
	return validator.New()
}

type fakeBarcodeRepository struct {
	saveBatchFunc func(ctx context.Context, tx pgx.Tx, productID int, codes []string) error
	replaceFunc   func(ctx context.Context, tx pgx.Tx, productID int, codes []string) error
}

func (f *fakeBarcodeRepository) SaveBatch(ctx context.Context, tx pgx.Tx, productID int, codes []string) error {
	if f.saveBatchFunc == nil {
		return nil
	}
	return f.saveBatchFunc(ctx, tx, productID, codes)
}

func (f *fakeBarcodeRepository) Replace(ctx context.Context, tx pgx.Tx, productID int, codes []string) error {
	if f.replaceFunc == nil {
		return nil
	}
	return f.replaceFunc(ctx, tx, productID, codes)
}

type fakeCoordinatorRepository struct {
	saveFunc               func(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error
	findAllSummaryFunc     func(ctx context.Context) ([]*model.CoordinatorSummary, error)
	findBySesiIdSummaryFn  func(ctx context.Context, sesiId int) ([]*model.CoordinatorSummary, error)
	findByIdSummaryFunc    func(ctx context.Context, id int) (*model.CoordinatorSummary, error)
	findByIdReportFunc     func(ctx context.Context, id int) (*model.CoordinatorReport, error)
	findByIdDetailFunc     func(ctx context.Context, id int) (*model.CoordinatorDetail, error)
	updateFunc             func(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error
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

func (f *fakeCoordinatorRepository) FindBySesiAndCoorCode(ctx context.Context, tx pgx.Tx, sesiCode string, coorCode string) (*model.Coordinator, error) {
	return f.findBySesiAndCoorCodeF(ctx, tx, sesiCode, coorCode)
}

func (f *fakeCoordinatorRepository) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Coordinator, error) {
	if f.findByIdFunc == nil {
		return nil, nil
	}
	return f.findByIdFunc(ctx, tx, id)
}

type fakeDeletedProductRepository struct {
	saveFunc                func(ctx context.Context, tx pgx.Tx, plu string) error
	deleteFunc              func(ctx context.Context) error
	findAllFunc             func(ctx context.Context) ([]*model.DeletedProduct, error)
	findAllUpdatedAfterFunc func(ctx context.Context, date time.Time) ([]*model.DeletedProduct, error)
}

func (f *fakeDeletedProductRepository) Save(ctx context.Context, tx pgx.Tx, plu string) error {
	return f.saveFunc(ctx, tx, plu)
}

func (f *fakeDeletedProductRepository) Delete(ctx context.Context) error {
	return f.deleteFunc(ctx)
}

func (f *fakeDeletedProductRepository) FindAll(ctx context.Context) ([]*model.DeletedProduct, error) {
	return f.findAllFunc(ctx)
}

func (f *fakeDeletedProductRepository) FindAllUpdatedAfter(ctx context.Context, date time.Time) ([]*model.DeletedProduct, error) {
	return f.findAllUpdatedAfterFunc(ctx, date)
}

type fakeInspectorRepository struct {
	findAllFunc      func(ctx context.Context) ([]*model.InspectorSummary, error)
	findByCoorIdFunc func(ctx context.Context, coordinatorId int) ([]*model.InspectorSummary, error)
	saveFunc         func(ctx context.Context, tx pgx.Tx, inspector *model.Inspector) error
}

func (f *fakeInspectorRepository) FindAll(ctx context.Context) ([]*model.InspectorSummary, error) {
	return f.findAllFunc(ctx)
}

func (f *fakeInspectorRepository) FindByCoorId(ctx context.Context, coordinatorId int) ([]*model.InspectorSummary, error) {
	return f.findByCoorIdFunc(ctx, coordinatorId)
}

func (f *fakeInspectorRepository) Save(ctx context.Context, tx pgx.Tx, inspector *model.Inspector) error {
	return f.saveFunc(ctx, tx, inspector)
}

type fakeProductRepository struct {
	saveFunc                func(ctx context.Context, tx pgx.Tx, product *model.Product) error
	updateFunc              func(ctx context.Context, tx pgx.Tx, product *model.Product) error
	deleteFunc              func(ctx context.Context, tx pgx.Tx, id int) error
	clearFunc               func(ctx context.Context, tx pgx.Tx) error
	findAllInPageSearchFn   func(ctx context.Context, limit int, offset int, search string) ([]*model.Product, int, error)
	findAllUpdatedAfterFunc func(ctx context.Context, date time.Time, limit int, offset int) ([]*model.Product, int, error)
	findAllLastSessionFunc  func(ctx context.Context, limit int) ([]*model.ProductLastSession, error)
	findAllForImportFunc    func(ctx context.Context, tx pgx.Tx) ([]*model.Product, error)
	findByIdFn              func(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error)
	findByPLUFn             func(ctx context.Context, tx pgx.Tx, plu string) (*model.Product, error)
	findByBarcodeFn         func(ctx context.Context, tx pgx.Tx, code string) (*model.Product, error)
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

func (f *fakeProductRepository) Clear(ctx context.Context, tx pgx.Tx) error {
	return f.clearFunc(ctx, tx)
}

func (f *fakeProductRepository) FindAllForImport(ctx context.Context, tx pgx.Tx) ([]*model.Product, error) {
	return f.findAllForImportFunc(ctx, tx)
}

func (f *fakeProductRepository) FindByPLU(ctx context.Context, tx pgx.Tx, plu string) (*model.Product, error) {
	return f.findByPLUFn(ctx, tx, plu)
}

func (f *fakeProductRepository) FindByBarcode(ctx context.Context, tx pgx.Tx, code string) (*model.Product, error) {
	return f.findByBarcodeFn(ctx, tx, code)
}

func (f *fakeProductRepository) FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Product, int, error) {
	return f.findAllInPageSearchFn(ctx, limit, offset, search)
}

func (f *fakeProductRepository) FindAllUpdatedAfter(ctx context.Context, date time.Time, limit int, offset int) ([]*model.Product, int, error) {
	return f.findAllUpdatedAfterFunc(ctx, date, limit, offset)
}

func (f *fakeProductRepository) FindAllLastSession(ctx context.Context, limit int) ([]*model.ProductLastSession, error) {
	return f.findAllLastSessionFunc(ctx, limit)
}

func (f *fakeProductRepository) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error) {
	return f.findByIdFn(ctx, tx, id)
}

type fakeRackRepository struct {
	findAllFunc      func(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]*model.Rack, error)
	findProgressFunc func(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error)
	saveFunc         func(ctx context.Context, tx pgx.Tx, rack *model.Rack) error
}

func (f *fakeRackRepository) FindAll(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]*model.Rack, error) {
	return f.findAllFunc(ctx, inspectorId, coordinatorId, sessionId)
}

func (f *fakeRackRepository) FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]*model.RackProgress, error) {
	return f.findProgressFunc(ctx, coordinatorId, sessionId)
}

func (f *fakeRackRepository) Save(ctx context.Context, tx pgx.Tx, rack *model.Rack) error {
	return f.saveFunc(ctx, tx, rack)
}

type fakeSesiRepository struct {
	findByIdFn            func(ctx context.Context, id int) (*model.Sesi, error)
	findAllInPageSearchFn func(ctx context.Context, limit int, offset int, search string) ([]*model.Sesi, int, error)
	saveFunc              func(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error
	updateFunc            func(ctx context.Context, tx pgx.Tx, sesi *model.Sesi) error
	deleteFunc            func(ctx context.Context, tx pgx.Tx, id int) error
}

func (f *fakeSesiRepository) FindById(ctx context.Context, id int) (*model.Sesi, error) {
	return f.findByIdFn(ctx, id)
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
	saveFunc             func(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	findByIdFn           func(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameSummary, error)
	findAllFunc          func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error)
	updateFunc           func(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	deleteFunc           func(ctx context.Context, tx pgx.Tx, id int) error
	findAllForExportFunc func(ctx context.Context, sesiId int) ([]*model.StockOpnameExport, error)
}

func (f *fakeStockOpnameRepository) Save(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
	return f.saveFunc(ctx, tx, stockOpname)
}

func (f *fakeStockOpnameRepository) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameSummary, error) {
	return f.findByIdFn(ctx, tx, id)
}

func (f *fakeStockOpnameRepository) FindAll(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error) {
	return f.findAllFunc(ctx, limit, offset, search, sesiId, coorId)
}

func (f *fakeStockOpnameRepository) Update(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error {
	return f.updateFunc(ctx, tx, stockOpname)
}

func (f *fakeStockOpnameRepository) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	return f.deleteFunc(ctx, tx, id)
}

func (f *fakeStockOpnameRepository) FindAllForExport(ctx context.Context, sesiId int) ([]*model.StockOpnameExport, error) {
	return f.findAllForExportFunc(ctx, sesiId)
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

func (f *fakeCoordinatorService) Update(ctx context.Context, id int, req *dto.UpdateCoordinatorRequest) error {
	return nil
}

type fakeInspectorService struct {
	findAllSummaryFunc func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error)
}

func (f *fakeInspectorService) FindAllSummary(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
	return f.findAllSummaryFunc(ctx, coorId)
}

func (f *fakeInspectorService) CreateInspector(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error) {
	return dto.InspectorJoinResponse{}, nil
}

func (f *fakeInspectorService) CreateInspectorByCoordinatorQR(ctx context.Context, request dto.InspectorJoinByCoordinatorQRRequest) (dto.InspectorJoinResponse, error) {
	return dto.InspectorJoinResponse{}, nil
}

type fakeStockOpnameService struct {
	findAllFunc          func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error)
	findAllForExportFunc func(ctx context.Context, sesiId int) ([]dto.StockOpnameExportResponse, error)
}

func (f *fakeStockOpnameService) Create(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	return dto.StockOpnameResponse{}, nil
}

func (f *fakeStockOpnameService) Update(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	return dto.StockOpnameResponse{}, nil
}

func (f *fakeStockOpnameService) Delete(ctx context.Context, id int) error { return nil }

func (f *fakeStockOpnameService) FindById(ctx context.Context, id int) (dto.StockOpnameResponse, error) {
	return dto.StockOpnameResponse{}, nil
}

func (f *fakeStockOpnameService) FindAll(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
	return f.findAllFunc(ctx, pagination, search, coorId, sesiId)
}

func (f *fakeStockOpnameService) CreateByRack(ctx context.Context, req dto.CreateStockOpnameByRackRequest) error {
	return nil
}

func (f *fakeStockOpnameService) FindAllForExport(ctx context.Context, sesiId int) ([]dto.StockOpnameExportResponse, error) {
	return f.findAllForExportFunc(ctx, sesiId)
}
