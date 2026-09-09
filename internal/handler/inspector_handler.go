package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type InspectorHandler interface {
	GetInspectors(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
