package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type CoordinatorHandler interface {
	CreateCoordinator(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetCoordinators(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetCoordinatorById(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	JoinByCoordinatorQR(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	UpdateCoordinator(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
