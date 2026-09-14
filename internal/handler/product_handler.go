package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ProductHandler interface {
	GetProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetProductsLastSession(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	SyncProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	CreateProduct(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	UpdateProduct(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	DeleteProduct(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
