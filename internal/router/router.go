package router

import (
	"net/http"

	apidocs "stockopname-rita-backend/docs"
	"stockopname-rita-backend/internal/handler"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/repository"
	"stockopname-rita-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

const (
	// DefaultMaxExportRows bounds how many rows one report export may load.
	DefaultMaxExportRows = repository.DefaultExportLimit
	// DefaultMaxConcurrentExport keeps several large reports from being built
	// at the same time; each one is rendered fully in memory.
	DefaultMaxConcurrentExport = 2
)

type routerOptions struct {
	maxExportRows       int
	maxConcurrentExport int
}

// Option tunes how the router assembles its handlers.
type Option func(*routerOptions)

// WithExportLimits overrides the export row cap and the concurrency limit.
// Non-positive values disable the corresponding limit.
func WithExportLimits(maxRows, maxConcurrent int) Option {
	return func(o *routerOptions) {
		o.maxExportRows = maxRows
		o.maxConcurrentExport = maxConcurrent
	}
}

// exportGuard bounds how many reports are generated at once.
type exportGuard struct {
	sem chan struct{}
}

func newExportGuard(maxConcurrent int) *exportGuard {
	if maxConcurrent <= 0 {
		return &exportGuard{}
	}
	return &exportGuard{sem: make(chan struct{}, maxConcurrent)}
}

func (g *exportGuard) wrap(next httprouter.Handle) httprouter.Handle {
	if g.sem == nil {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		select {
		case g.sem <- struct{}{}:
			defer func() { <-g.sem }()
			next(w, r, params)
		default:
			// Fail fast instead of queueing: a queue would only move the memory
			// pressure into the HTTP layer.
			helper.WriteError(w, http.StatusServiceUnavailable,
				"Laporan sedang dibuat, coba lagi sebentar.", nil)
		}
	}
}

func NewRouter(validate *validator.Validate, pool *pgxpool.Pool, opts ...Option) *httprouter.Router {
	options := routerOptions{
		maxExportRows:       DefaultMaxExportRows,
		maxConcurrentExport: DefaultMaxConcurrentExport,
	}
	for _, opt := range opts {
		opt(&options)
	}

	sesiRepository := repository.NewSesiRepository(pool)
	coordinatorRepository := repository.NewCoordinatorRepository(pool)
	productRepository := repository.NewProductRepository(pool)
	barcodeRepository := repository.NewBarcodeRepository(pool)
	rackRepository := repository.NewRackRepository(pool)
	uploadRepository := repository.NewUploadRepository(pool)
	stockOpnameRepository := repository.NewStockOpnameRepository(pool, repository.WithExportLimit(options.maxExportRows))
	auditRepository := repository.NewAuditRepository(pool)

	productService := service.NewProductService(productRepository, barcodeRepository, sesiRepository, pool, validate)
	importService := service.NewImportService(productRepository, barcodeRepository, sesiRepository, pool)
	productHandler := handler.NewProductHandler(productService, importService)

	coordinatorService := service.NewCoordinatorService(coordinatorRepository, sesiRepository, productRepository, rackRepository, pool)
	coordinatorHandler := handler.NewCoordinatorHandler(coordinatorService)

	sesiService := service.NewSesiService(sesiRepository, coordinatorRepository, productRepository, stockOpnameRepository, rackRepository, pool, validate)
	sesiHandler := handler.NewSesiHandler(sesiService)

	rackService := service.NewRackService(rackRepository, coordinatorRepository, sesiRepository, stockOpnameRepository, pool)
	rackHandler := handler.NewRackHandler(rackService)

	uploadService := service.NewUploadService(uploadRepository, rackRepository, coordinatorRepository, sesiRepository, productRepository, stockOpnameRepository, pool, validate)
	uploadHandler := handler.NewUploadHandler(uploadService)

	stockOpnameService := service.NewStockOpnameService(stockOpnameRepository, productRepository, sesiRepository, auditRepository, coordinatorRepository, pool, validate)
	stockOpnameHandler := handler.NewStockOpnameHandler(stockOpnameService)
	auditService := service.NewAuditService(auditRepository)
	auditHandler := handler.NewAuditHandler(auditService)
	reportService := service.NewReportService(sesiService, coordinatorService, uploadService, stockOpnameService)
	reportHandler := handler.NewReportHandler(reportService)

	// Report exports are rendered in memory, so they are the only routes that
	// get a concurrency cap.
	exports := newExportGuard(options.maxConcurrentExport)

	router := httprouter.New()

	// Health check for deploy verification.
	router.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})

	// API docs.
	apidocs.RegisterRoutes(router)

	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products", productHandler.GetProducts)
	router.POST("/products/import", productHandler.ImportProducts)
	router.POST("/products/clear", productHandler.ClearProducts)
	router.PATCH("/products/:id", productHandler.UpdateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)

	router.GET("/stockopname/coordinators", coordinatorHandler.GetCoordinators)
	router.POST("/stockopname/coordinators", coordinatorHandler.CreateCoordinator)
	router.GET("/stockopname/coordinators/:id", coordinatorHandler.GetCoordinatorById)
	router.GET("/stockopname/coordinators/:id/pdf", exports.wrap(reportHandler.CoordinatorPDF))
	router.GET("/stockopname/coordinators/:id/uploads/pdf", exports.wrap(reportHandler.CoordinatorUploadSummaryPDF))
	router.POST("/stockopname/coordinators/join-by-qr", coordinatorHandler.JoinByCoordinatorQR)
	router.PATCH("/stockopname/coordinators/:id", coordinatorHandler.UpdateCoordinator)

	router.POST("/stockopname/sessions", sesiHandler.CreateSesi)
	router.GET("/stockopname/sessions", sesiHandler.GetAllSesi)
	router.GET("/stockopname/sessions/:id", sesiHandler.GetSesiById)
	router.PATCH("/stockopname/sessions/:id/qty-factor", stockOpnameHandler.SetQtyFactor)
	router.GET("/stockopname/sessions/:id/sell-summary", stockOpnameHandler.GetSellSummary)
	router.GET("/stockopname/sessions/:id/pdf", exports.wrap(reportHandler.SessionPDF))
	router.GET("/stockopname/sessions/:id/excel", exports.wrap(reportHandler.SessionExcel))
	router.GET("/stockopname/sessions/:id/dbf", exports.wrap(reportHandler.SessionDBF))
	router.PATCH("/stockopname/sessions/:id", sesiHandler.UpdateSesi)
	router.DELETE("/stockopname/sessions/:id", sesiHandler.DeleteSesi)

	router.GET("/stockopname/racks/progress", rackHandler.GetRackProgress)
	router.GET("/stockopname/racks", rackHandler.GetRacks)
	router.DELETE("/stockopname/racks/:id", rackHandler.DeleteRack)

	router.POST("/stockopname/uploads", uploadHandler.CreateUpload)
	router.GET("/stockopname/uploads", uploadHandler.GetUploads)
	router.GET("/stockopname/uploads/:id/pdf", exports.wrap(reportHandler.UploadPDF))

	router.GET("/stockopname/results", stockOpnameHandler.GetAllStockOpname)
	router.GET("/stockopname/results/:id", stockOpnameHandler.GetStockOpnameById)
	router.POST("/stockopname/results", stockOpnameHandler.CreateStockOpname)
	router.POST("/stockopname/results/move", stockOpnameHandler.MoveStockOpname)
	router.PATCH("/stockopname/results/:id", stockOpnameHandler.UpdateStockOpname)
	router.DELETE("/stockopname/results/:id", stockOpnameHandler.DeleteStockOpname)

	router.GET("/stockopname/audits", auditHandler.GetAudits)
	router.GET("/stockopname/audits/batches", auditHandler.GetAuditBatches)
	router.GET("/stockopname/audits/timeline", auditHandler.GetAuditTimeline)

	return router
}
