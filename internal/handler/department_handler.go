package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type DepartmentHandler interface {
	GetDepartments(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	CreateDepartment(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	UpdateDepartment(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	DeleteDepartment(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}
