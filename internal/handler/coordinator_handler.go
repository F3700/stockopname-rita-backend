package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type CoordinatorHandler interface {
	GetCoordinators(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetCoordinatorById(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	CancelCoordinator(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	CompleteCoordinator(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
