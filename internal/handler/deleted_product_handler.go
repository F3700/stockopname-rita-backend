package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type DeletedProductHandler interface {
	DeleteDeletedProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetDeletedProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
