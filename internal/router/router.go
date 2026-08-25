package router

import (
	"stockopname-rita-backend/internal/handler"
	"stockopname-rita-backend/internal/repository"
	"stockopname-rita-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func NewRouter(validate *validator.Validate, pool *pgxpool.Pool) *httprouter.Router {
	productRepository := repository.NewProductRepository(pool)
	productService := service.NewProductService(productRepository, pool, validate)
	handler := handler.NewProductHandler(productService)

	router := httprouter.New()

	router.POST("/products", handler.CreateProduct)
	router.GET("/products", handler.GetProducts)
	router.PATCH("/products/:id", handler.UpdateProduct)
	router.DELETE("/products/:id", handler.DeleteProduct)

	return router
}
