package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type RackHandler interface {
	GetRackProgress(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetRacks(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
