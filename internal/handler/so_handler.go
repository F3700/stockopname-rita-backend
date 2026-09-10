package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type StockOpnameHandler interface {
	GetAllStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetStockOpnameById(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	CreateStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	UpdateStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	DeleteStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	CreateStockOpnameByRack(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
