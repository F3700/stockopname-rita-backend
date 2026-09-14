package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type SesiHandler interface {
	CreateSesi(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	UpdateSesi(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	DeleteSesi(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetSesiById(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetAllSesi(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
