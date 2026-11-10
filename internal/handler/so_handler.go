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
	MoveStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	SetQtyFactor(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetSellSummary(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
