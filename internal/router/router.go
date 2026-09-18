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
	deletedProductRepository := repository.NewDeletedProductRepository(pool)
	deletedProductService := service.NewDeletedProductService(deletedProductRepository)
	deletedProductHandler := handler.NewDeletedProductHandler(deletedProductService)

	productRepository := repository.NewProductRepository(pool)
	productService := service.NewProductService(productRepository, deletedProductRepository, pool, validate)
	productHandler := handler.NewProductHandler(productService)

	departmentRepository := repository.NewDepartmentRepositoryImpl(pool)
	departmentService := service.NewDepartmentService(departmentRepository, pool, validate)
	departmentHandler := handler.NewDepartmentHandler(departmentService)

	categoryRepository := repository.NewCategoryRepositoryImpl(pool)
	categoryService := service.NewCategoryService(categoryRepository, pool, validate)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	coordinatorRepository := repository.NewCoordinatorRepository(pool)
	coordinatorService := service.NewCoordinatorService(coordinatorRepository, pool)
	coordinatorHandler := handler.NewCoordinatorHandler(coordinatorService)

	sesiRepository := repository.NewSesiRepository(pool)
	sesiService := service.NewSesiService(sesiRepository, coordinatorRepository, pool, validate)
	sesiHandler := handler.NewSesiHandler(sesiService)

	rackRepository := repository.NewRackRepository(pool)
	rackService := service.NewRackService(rackRepository, pool)
	rackHandler := handler.NewRackHandler(rackService)

	inspectorRepository := repository.NewInspectorRepository(pool)
	inspectorService := service.NewInspectorService(inspectorRepository, coordinatorRepository, rackRepository, pool)
	inspectorHandler := handler.NewInspectorHandler(inspectorService)

	stockOpnameRepository := repository.NewStockOpnameRepository(pool)
	stockOpnameService := service.NewStockOpnameService(stockOpnameRepository, pool, validate)
	stockOpnameHandler := handler.NewStockOpnameHandler(stockOpnameService)
	reportService := service.NewReportService(sesiService, coordinatorService, inspectorService, stockOpnameService)
	reportHandler := handler.NewReportHandler(reportService)

	router := httprouter.New()

	// API docs.
	apidocs.RegisterRoutes(router)

	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products", productHandler.GetProducts)
	router.GET("/products/last-session", productHandler.GetProductsLastSession)
	router.GET("/products/sync", productHandler.SyncProducts)
	router.PATCH("/products/:id", productHandler.UpdateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)

	router.POST("/departments", departmentHandler.CreateDepartment)
	router.GET("/departments", departmentHandler.GetDepartments)
	router.PATCH("/departments/:id", departmentHandler.UpdateDepartment)
	router.DELETE("/departments/:id", departmentHandler.DeleteDepartment)

	router.POST("/categories", categoryHandler.CreateCategory)
	router.GET("/categories", categoryHandler.GetCategories)
	router.PATCH("/categories/:id", categoryHandler.UpdateCategory)
	router.DELETE("/categories/:id", categoryHandler.DeleteCategory)

	router.GET("/deleted/products", deletedProductHandler.GetDeletedProducts)
	router.DELETE("/deleted/products", deletedProductHandler.DeleteDeletedProducts)

	router.GET("/stockopname/coordinators", coordinatorHandler.GetCoordinators)
	router.GET("/stockopname/coordinators/:id", coordinatorHandler.GetCoordinatorById)
	router.GET("/stockopname/coordinators/:id/pdf", reportHandler.CoordinatorPDF)
	router.PATCH("/stockopname/coordinators/:id", coordinatorHandler.UpdateCoordinator)

	router.POST("/stockopname/sessions", sesiHandler.CreateSesi)
	router.GET("/stockopname/sessions", sesiHandler.GetAllSesi)
	router.GET("/stockopname/sessions/:id", sesiHandler.GetSesiById)
	router.GET("/stockopname/sessions/:id/pdf", reportHandler.SessionPDF)
	router.GET("/stockopname/sessions/:id/excel", reportHandler.SessionExcel)
	router.GET("/stockopname/sessions/:id/dbf", reportHandler.SessionDBF)
	router.PATCH("/stockopname/sessions/:id", sesiHandler.UpdateSesi)
	router.DELETE("/stockopname/sessions/:id", sesiHandler.DeleteSesi)

	router.GET("/stockopname/inspectors", inspectorHandler.GetInspectors)
	router.POST("/stockopname/inspectors", inspectorHandler.CreateInspector)
	router.POST("/stockopname/inspectors/join-by-coordinator-qr", inspectorHandler.CreateInspectorByCoordinatorQR)

	router.GET("/stockopname/racks/progress", rackHandler.GetRackProgress)
	router.GET("/stockopname/racks", rackHandler.GetRacks)
	router.POST("/stockopname/racks", rackHandler.CreateRack)

	router.GET("/stockopname/results", stockOpnameHandler.GetAllStockOpname)
	router.GET("/stockopname/results/:id", stockOpnameHandler.GetStockOpnameById)
	router.POST("/stockopname/results", stockOpnameHandler.CreateStockOpname)
	router.PATCH("/stockopname/results/:id", stockOpnameHandler.UpdateStockOpname)
	router.DELETE("/stockopname/results/:id", stockOpnameHandler.DeleteStockOpname)
	router.POST("/stockopname/results/racks", stockOpnameHandler.CreateStockOpnameByRack)

	return router
}
