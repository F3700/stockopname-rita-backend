package router

import (
	"net/http"
	"stockopname-rita-backend/internal/handler"
	"stockopname-rita-backend/internal/repository"
	"stockopname-rita-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	v5 "github.com/swaggest/swgui/v5"
)

func NewRouter(validate *validator.Validate, pool *pgxpool.Pool) *httprouter.Router {
	productRepository := repository.NewProductRepository(pool)
	productService := service.NewProductService(productRepository, pool, validate)
	productHandler := handler.NewProductHandler(productService)

	departmentRepository := repository.NewDepartmentRepositoryImpl(pool)
	departmentService := service.NewDepartmentService(departmentRepository, pool, validate)
	departmentHandler := handler.NewDepartmentHandler(departmentService)

	router := httprouter.New()

	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products", productHandler.GetProducts)
	router.PATCH("/products/:id", productHandler.UpdateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)

	router.POST("/departments", departmentHandler.CreateDepartment)
	router.GET("/departments", departmentHandler.GetDepartments)
	router.PATCH("/departments/:id", departmentHandler.UpdateDepartment)
	router.DELETE("/departments/:id", departmentHandler.DeleteDepartment)

	//API DOCS
	router.GET("/docs/apispec.json", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.ServeFile(w, r, "./docs/apispec.json")
	})
	router.Handler(
		http.MethodGet,
		"/docs",
		v5.New(
			"Rita Stock Opname API",
			"/docs/apispec.json",
			"/docs/",
		),
	)

	return router
}
