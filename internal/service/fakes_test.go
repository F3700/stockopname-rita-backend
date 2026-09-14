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

var _ repository.CategoryRepository = (*fakeCategoryRepository)(nil)
var _ repository.CoordinatorRepository = (*fakeCoordinatorRepository)(nil)
var _ repository.DeletedProductRepository = (*fakeDeletedProductRepository)(nil)
var _ repository.DepartmentRepository = (*fakeDepartmentRepository)(nil)
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

type fakeCategoryRepository struct {
	saveFunc              func(ctx context.Context, tx pgx.Tx, category *model.Category) error
	findAllFunc           func(ctx context.Context) ([]*model.Category, error)
	findAllInPageSearchFn func(ctx context.Context, limit int, offset int, search string) ([]*model.Category, int, error)
	findByIdFn            func(ctx context.Context, tx pgx.Tx, id int) (*model.Category, error)
	updateFunc            func(ctx context.Context, tx pgx.Tx, category *model.Category) error
	deleteFunc            func(ctx context.Context, tx pgx.Tx, id int) error
}

func (f *fakeCategoryRepository) Save(ctx context.Context, tx pgx.Tx, category *model.Category) error {
	return f.saveFunc(ctx, tx, category)
}

func (f *fakeCategoryRepository) FindAll(ctx context.Context) ([]*model.Category, error) {
	return f.findAllFunc(ctx)
}

func (f *fakeCategoryRepository) FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Category, int, error) {
	return f.findAllInPageSearchFn(ctx, limit, offset, search)
}

func (f *fakeCategoryRepository) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Category, error) {
	return f.findByIdFn(ctx, tx, id)
}

func (f *fakeCategoryRepository) Update(ctx context.Context, tx pgx.Tx, category *model.Category) error {
	return f.updateFunc(ctx, tx, category)
}

func (f *fakeCategoryRepository) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	return f.deleteFunc(ctx, tx, id)
}

type fakeCoordinatorRepository struct {
	saveFunc               func(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error
	findAllSummaryFunc     func(ctx context.Context) ([]*model.CoordinatorSummary, error)
	findBySesiIdSummaryFn  func(ctx context.Context, sesiId int) ([]*model.CoordinatorSummary, error)
	findByIdSummaryFunc    func(ctx context.Context, id int) (*model.CoordinatorSummary, error)
	findByIdReportFunc     func(ctx context.Context, id int) (*model.CoordinatorReport, error)
	updateFunc             func(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error
	findBySesiAndCoorCodeF func(ctx context.Context, tx pgx.Tx, sesiCode string, coorCode string) (*model.Coordinator, error)
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

func (f *fakeCoordinatorRepository) Update(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error {
	return f.updateFunc(ctx, tx, coordinator)
}

func (f *fakeCoordinatorRepository) FindBySesiAndCoorCode(ctx context.Context, tx pgx.Tx, sesiCode string, coorCode string) (*model.Coordinator, error) {
	return f.findBySesiAndCoorCodeF(ctx, tx, sesiCode, coorCode)
}

type fakeDeletedProductRepository struct {
	saveFunc                func(ctx context.Context, tx pgx.Tx, id int) error
	deleteFunc              func(ctx context.Context) error
	findAllFunc             func(ctx context.Context) ([]*model.DeletedProduct, error)
	findAllUpdatedAfterFunc func(ctx context.Context, date time.Time) ([]*model.DeletedProduct, error)
}

func (f *fakeDeletedProductRepository) Save(ctx context.Context, tx pgx.Tx, id int) error {
	return f.saveFunc(ctx, tx, id)
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

type fakeDepartmentRepository struct {
	saveFunc              func(ctx context.Context, tx pgx.Tx, department *model.Department) error
	updateFunc            func(ctx context.Context, tx pgx.Tx, department *model.Department) error
	deleteFunc            func(ctx context.Context, tx pgx.Tx, id int) error
	findAllFunc           func(ctx context.Context) ([]*model.Department, error)
	findAllInPageSearchFn func(ctx context.Context, limit int, offset int, search string) ([]*model.Department, int, error)
	findByIdFn            func(ctx context.Context, tx pgx.Tx, id int) (*model.Department, error)
}

func (f *fakeDepartmentRepository) Save(ctx context.Context, tx pgx.Tx, department *model.Department) error {
	return f.saveFunc(ctx, tx, department)
}

func (f *fakeDepartmentRepository) Update(ctx context.Context, tx pgx.Tx, department *model.Department) error {
	return f.updateFunc(ctx, tx, department)
}

func (f *fakeDepartmentRepository) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	return f.deleteFunc(ctx, tx, id)
}

func (f *fakeDepartmentRepository) FindAll(ctx context.Context) ([]*model.Department, error) {
	return f.findAllFunc(ctx)
}

func (f *fakeDepartmentRepository) FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Department, int, error) {
	return f.findAllInPageSearchFn(ctx, limit, offset, search)
}

func (f *fakeDepartmentRepository) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Department, error) {
	return f.findByIdFn(ctx, tx, id)
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
	findAllFunc             func(ctx context.Context) ([]*model.Product, error)
	findAllInPageSearchFn   func(ctx context.Context, limit int, offset int, search string) ([]*model.Product, int, error)
	findAllUpdatedAfterFunc func(ctx context.Context, date time.Time, limit int, offset int) ([]*model.Product, int, error)
	findAllLastSessionFunc  func(ctx context.Context, limit int) ([]*model.ProductLastSession, error)
	findByIdFn              func(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error)
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

func (f *fakeProductRepository) FindAll(ctx context.Context) ([]*model.Product, error) {
	return f.findAllFunc(ctx)
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
	saveFunc    func(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	findByIdFn  func(ctx context.Context, tx pgx.Tx, id int) (*model.StockOpnameSummary, error)
	findAllFunc func(ctx context.Context, limit int, offset int, search string, sesiId *int, coorId *int) ([]*model.StockOpnameSummary, int, error)
	updateFunc  func(ctx context.Context, tx pgx.Tx, stockOpname *model.StockOpname) error
	deleteFunc  func(ctx context.Context, tx pgx.Tx, id int) error
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
	findByIdReportFunc  func(ctx context.Context, id int) (*dto.CoordinatorReportResponse, error)
}

func (f *fakeCoordinatorService) FindByIdSummary(ctx context.Context, id int) (*dto.CoordinatorResponse, error) {
	return f.findByIdSummaryFunc(ctx, id)
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
