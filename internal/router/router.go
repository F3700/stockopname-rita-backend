package router

import (
	"stockopname-rita-backend/internal/handler"

	"github.com/julienschmidt/httprouter"
)

func NewRouter() *httprouter.Router {
	router := httprouter.New()

	router.GET("/products", handler.GetProducts)

	return router
}
