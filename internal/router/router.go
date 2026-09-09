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

	coordinatorRepository := repository.NewCoordinatorRepository()
	coordinatorService := service.NewCoordinatorService(coordinatorRepository, pool)
	coordinatorHandler := handler.NewCoordinatorHandler(coordinatorService)

	router := httprouter.New()

	// API docs.
	apidocs.RegisterRoutes(router)

	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products", productHandler.GetProducts)
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
	router.PATCH("/stockopname/coordinators/:id", coordinatorHandler.UpdateCoordinator)

	return router
}
