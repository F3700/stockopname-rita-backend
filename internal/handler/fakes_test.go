package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/service"
)

var _ service.CoordinatorService = (*fakeCoordinatorService)(nil)
var _ service.DeletedProductService = (*fakeDeletedProductService)(nil)
var _ service.InspectorService = (*fakeInspectorService)(nil)
var _ service.ProductService = (*fakeProductService)(nil)
var _ service.ImportService = (*fakeImportService)(nil)
var _ service.RackService = (*fakeRackService)(nil)
var _ service.SesiService = (*fakeSesiService)(nil)
var _ service.StockOpnameService = (*fakeStockOpnameService)(nil)
var _ service.ReportService = (*fakeReportService)(nil)

type fakeCoordinatorService struct {
	findByIdSummaryFunc func(ctx context.Context, id int) (*dto.CoordinatorResponse, error)
	findByIdDetailFunc  func(ctx context.Context, id int) (*dto.CoordinatorDetailResponse, error)
	findByIdReportFunc  func(ctx context.Context, id int) (*dto.CoordinatorReportResponse, error)
	findAllSummaryFunc  func(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error)
	updateFunc          func(ctx context.Context, id int, req *dto.UpdateCoordinatorRequest) error
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
	return f.updateFunc(ctx, id, req)
}

type fakeDeletedProductService struct {
	findAllFunc func(ctx context.Context, date *time.Time) ([]dto.DeletedProductResponse, error)
	deleteFunc  func(ctx context.Context) error
}

func (f *fakeDeletedProductService) FindAll(ctx context.Context, date *time.Time) ([]dto.DeletedProductResponse, error) {
	return f.findAllFunc(ctx, date)
}

func (f *fakeDeletedProductService) Delete(ctx context.Context) error {
	return f.deleteFunc(ctx)
}

type fakeInspectorService struct {
	findAllSummaryFunc        func(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error)
	createFunc                func(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error)
	createByCoordinatorQRFunc func(ctx context.Context, request dto.InspectorJoinByCoordinatorQRRequest) (dto.InspectorJoinResponse, error)
}

func (f *fakeInspectorService) FindAllSummary(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
	return f.findAllSummaryFunc(ctx, coorId)
}

func (f *fakeInspectorService) CreateInspector(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error) {
	return f.createFunc(ctx, request)
}

func (f *fakeInspectorService) CreateInspectorByCoordinatorQR(ctx context.Context, request dto.InspectorJoinByCoordinatorQRRequest) (dto.InspectorJoinResponse, error) {
	if f.createByCoordinatorQRFunc == nil {
		return dto.InspectorJoinResponse{}, nil
	}
	return f.createByCoordinatorQRFunc(ctx, request)
}

type fakeProductService struct {
	createFunc            func(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error)
	updateFunc            func(ctx context.Context, id int, req dto.ProductUpdateRequest) (dto.ProductResponse, error)
	deleteFunc            func(ctx context.Context, id int) error
	clearFunc             func(ctx context.Context) error
	findAllFunc           func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.ProductResponse, error)
	findAllUpdatedAfterFn func(ctx context.Context, pagination *dto.Pagination, date time.Time) ([]dto.ProductResponse, error)
	findAllLastSessionFn  func(ctx context.Context, limit int) ([]dto.ProductLastSessionResponse, error)
}

func (f *fakeProductService) Clear(ctx context.Context) error {
	return f.clearFunc(ctx)
}

type fakeImportService struct {
	importFunc func(ctx context.Context, produkData []byte, barcodeData []byte, dryRun bool) (dto.ImportResultResponse, error)
}

func (f *fakeImportService) Import(ctx context.Context, produkData []byte, barcodeData []byte, dryRun bool) (dto.ImportResultResponse, error) {
	return f.importFunc(ctx, produkData, barcodeData, dryRun)
}

func (f *fakeProductService) Create(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error) {
	return f.createFunc(ctx, req)
}

func (f *fakeProductService) Update(ctx context.Context, id int, req dto.ProductUpdateRequest) (dto.ProductResponse, error) {
	return f.updateFunc(ctx, id, req)
}

func (f *fakeProductService) Delete(ctx context.Context, id int) error {
	return f.deleteFunc(ctx, id)
}

func (f *fakeProductService) FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.ProductResponse, error) {
	return f.findAllFunc(ctx, pagination, search)
}

func (f *fakeProductService) FindAllUpdatedAfter(ctx context.Context, pagination *dto.Pagination, date time.Time) ([]dto.ProductResponse, error) {
	return f.findAllUpdatedAfterFn(ctx, pagination, date)
}

func (f *fakeProductService) FindAllLastSession(ctx context.Context, limit int) ([]dto.ProductLastSessionResponse, error) {
	return f.findAllLastSessionFn(ctx, limit)
}

type fakeRackService struct {
	findAllSummaryFunc func(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error)
	findProgressFunc   func(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error)
	createFunc         func(ctx context.Context, request dto.CreateRackRequest) (dto.RackResponse, error)
}

func (f *fakeRackService) FindAllSummary(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error) {
	return f.findAllSummaryFunc(ctx, inspectorId, coordinatorId, sessionId)
}

func (f *fakeRackService) FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error) {
	return f.findProgressFunc(ctx, coordinatorId, sessionId)
}

func (f *fakeRackService) CreateRack(ctx context.Context, request dto.CreateRackRequest) (dto.RackResponse, error) {
	return f.createFunc(ctx, request)
}

type fakeSesiService struct {
	createFunc  func(ctx context.Context, req dto.CreateSesiRequest) (dto.SesiResponse, error)
	updateFunc  func(ctx context.Context, id int, req dto.UpdateSesiRequest) error
	deleteFunc  func(ctx context.Context, id int) error
	findByIdFn  func(ctx context.Context, id int) (dto.SesiResponse, error)
	findAllFunc func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.SesiResponse, error)
}

func (f *fakeSesiService) Create(ctx context.Context, req dto.CreateSesiRequest) (dto.SesiResponse, error) {
	return f.createFunc(ctx, req)
}

func (f *fakeSesiService) Update(ctx context.Context, id int, req dto.UpdateSesiRequest) error {
	return f.updateFunc(ctx, id, req)
}

func (f *fakeSesiService) Delete(ctx context.Context, id int) error {
	return f.deleteFunc(ctx, id)
}

func (f *fakeSesiService) FindById(ctx context.Context, id int) (dto.SesiResponse, error) {
	return f.findByIdFn(ctx, id)
}

func (f *fakeSesiService) FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.SesiResponse, error) {
	return f.findAllFunc(ctx, pagination, search)
}

type fakeStockOpnameService struct {
	createFunc           func(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error)
	updateFunc           func(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error)
	deleteFunc           func(ctx context.Context, id int) error
	findByIdFn           func(ctx context.Context, id int) (dto.StockOpnameResponse, error)
	findAllFunc          func(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error)
	createByRackFn       func(ctx context.Context, req dto.CreateStockOpnameByRackRequest) error
	findAllForExportFunc func(ctx context.Context, sesiId int) ([]dto.StockOpnameExportResponse, error)
}

func (f *fakeStockOpnameService) FindAllForExport(ctx context.Context, sesiId int) ([]dto.StockOpnameExportResponse, error) {
	if f.findAllForExportFunc == nil {
		return nil, nil
	}
	return f.findAllForExportFunc(ctx, sesiId)
}

func (f *fakeStockOpnameService) Create(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	return f.createFunc(ctx, req)
}

func (f *fakeStockOpnameService) Update(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	return f.updateFunc(ctx, id, req)
}

func (f *fakeStockOpnameService) Delete(ctx context.Context, id int) error {
	return f.deleteFunc(ctx, id)
}

func (f *fakeStockOpnameService) FindById(ctx context.Context, id int) (dto.StockOpnameResponse, error) {
	return f.findByIdFn(ctx, id)
}

func (f *fakeStockOpnameService) FindAll(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
	return f.findAllFunc(ctx, pagination, search, coorId, sesiId)
}

func (f *fakeStockOpnameService) CreateByRack(ctx context.Context, req dto.CreateStockOpnameByRackRequest) error {
	return f.createByRackFn(ctx, req)
}

type fakeReportService struct {
	sessionPDFFunc     func(ctx context.Context, id int) ([]byte, string, error)
	coordinatorPDFFunc func(ctx context.Context, id int) ([]byte, string, error)
	sessionExcelFunc   func(ctx context.Context, id int) ([]byte, string, error)
	sessionDBFFunc     func(ctx context.Context, id int) ([]byte, string, error)
}

func (f *fakeReportService) SessionPDF(ctx context.Context, id int) ([]byte, string, error) {
	return f.sessionPDFFunc(ctx, id)
}

func (f *fakeReportService) CoordinatorPDF(ctx context.Context, id int) ([]byte, string, error) {
	return f.coordinatorPDFFunc(ctx, id)
}

func (f *fakeReportService) SessionExcel(ctx context.Context, id int) ([]byte, string, error) {
	return f.sessionExcelFunc(ctx, id)
}

func (f *fakeReportService) SessionDBF(ctx context.Context, id int) ([]byte, string, error) {
	return f.sessionDBFFunc(ctx, id)
}

func newJSONRequest(t *testing.T, method, target, body string) *http.Request {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func paramsWithID(id string) httprouter.Params {
	return httprouter.Params{httprouter.Param{Key: "id", Value: id}}
}

func decodeMessage(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var res struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	return res.Message
}
