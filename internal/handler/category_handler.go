package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type CategoryHandler interface {
	GetCategories(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	CreateCategory(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	UpdateCategory(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	DeleteCategory(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
