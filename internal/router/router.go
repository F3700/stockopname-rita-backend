package router

import (
	apidocs "stockopname-rita-backend/docs"
	"stockopname-rita-backend/internal/handler"
	"stockopname-rita-backend/internal/repository"
	"stockopname-rita-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func NewRouter(validate *validator.Validate, pool *pgxpool.Pool) *httprouter.Router {
	sesiRepository := repository.NewSesiRepository(pool)
	coordinatorRepository := repository.NewCoordinatorRepository(pool)
	productRepository := repository.NewProductRepository(pool)
	barcodeRepository := repository.NewBarcodeRepository(pool)
	rackRepository := repository.NewRackRepository(pool)
	uploadRepository := repository.NewUploadRepository(pool)
	stockOpnameRepository := repository.NewStockOpnameRepository(pool)

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

	stockOpnameService := service.NewStockOpnameService(stockOpnameRepository, productRepository, sesiRepository, pool, validate)
	stockOpnameHandler := handler.NewStockOpnameHandler(stockOpnameService)
	reportService := service.NewReportService(sesiService, coordinatorService, uploadService, stockOpnameService)
	reportHandler := handler.NewReportHandler(reportService)

	router := httprouter.New()

	// API docs.
	apidocs.RegisterRoutes(router)

	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products", productHandler.GetProducts)
	router.POST("/products/import", productHandler.ImportProducts)
	router.POST("/products/clear", productHandler.ClearProducts)
	router.PATCH("/products/:id", productHandler.UpdateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)

	router.GET("/stockopname/coordinators", coordinatorHandler.GetCoordinators)
	router.GET("/stockopname/coordinators/:id", coordinatorHandler.GetCoordinatorById)
	router.GET("/stockopname/coordinators/:id/pdf", reportHandler.CoordinatorPDF)
	router.GET("/stockopname/coordinators/:id/uploads/pdf", reportHandler.CoordinatorUploadSummaryPDF)
	router.POST("/stockopname/coordinators/join-by-qr", coordinatorHandler.JoinByCoordinatorQR)
	router.PATCH("/stockopname/coordinators/:id", coordinatorHandler.UpdateCoordinator)

	router.POST("/stockopname/sessions", sesiHandler.CreateSesi)
	router.GET("/stockopname/sessions", sesiHandler.GetAllSesi)
	router.GET("/stockopname/sessions/:id", sesiHandler.GetSesiById)
	router.GET("/stockopname/sessions/:id/pdf", reportHandler.SessionPDF)
	router.GET("/stockopname/sessions/:id/excel", reportHandler.SessionExcel)
	router.GET("/stockopname/sessions/:id/dbf", reportHandler.SessionDBF)
	router.PATCH("/stockopname/sessions/:id", sesiHandler.UpdateSesi)
	router.DELETE("/stockopname/sessions/:id", sesiHandler.DeleteSesi)

	router.GET("/stockopname/racks/progress", rackHandler.GetRackProgress)
	router.GET("/stockopname/racks", rackHandler.GetRacks)
	router.DELETE("/stockopname/racks/:id", rackHandler.DeleteRack)

	router.POST("/stockopname/uploads", uploadHandler.CreateUpload)
	router.GET("/stockopname/uploads", uploadHandler.GetUploads)
	router.GET("/stockopname/uploads/:id/pdf", reportHandler.UploadPDF)

	router.GET("/stockopname/results", stockOpnameHandler.GetAllStockOpname)
	router.GET("/stockopname/results/:id", stockOpnameHandler.GetStockOpnameById)
	router.POST("/stockopname/results", stockOpnameHandler.CreateStockOpname)
	router.PATCH("/stockopname/results/:id", stockOpnameHandler.UpdateStockOpname)
	router.DELETE("/stockopname/results/:id", stockOpnameHandler.DeleteStockOpname)

	return router
}
