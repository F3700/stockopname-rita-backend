package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

type DepartmentHandlerImpl struct {
	DepartmentService service.DepartmentService
}

func NewDepartmentHandler(departmentService service.DepartmentService) DepartmentHandler {
	return &DepartmentHandlerImpl{
		DepartmentService: departmentService,
	}
}

// CreateDepartment implements [DepartmentHandler].
func (d *DepartmentHandlerImpl) CreateDepartment(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	departmentCreateRequest := dto.DepartmentCreateRequest{}

	if err := helper.ReadFromRequestBody(req, &departmentCreateRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	departmentCreateResponse, err := d.DepartmentService.Create(req.Context(), departmentCreateRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Department created successfully",
		Data:    departmentCreateResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusCreated, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}

// DeleteDepartment implements [DepartmentHandler].
func (d *DepartmentHandlerImpl) DeleteDepartment(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid department ID", err)
		return
	}

	if err := d.DepartmentService.Delete(req.Context(), id); err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Department deleted successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}

// GetDepartments implements [DepartmentHandler].
func (d *DepartmentHandlerImpl) GetDepartments(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	page, err := helper.ParseIntQueryNotNull(req, "page")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query page parameter", err)
		return
	}

	limit, err := helper.ParseIntQueryNotNull(req, "limit")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query limit parameter", err)
		return
	}

	search := req.URL.Query().Get("search")

	pagination := &dto.Pagination{
		Page:  page,
		Limit: limit,
	}

	departments, err := d.DepartmentService.FindAll(req.Context(), pagination, search)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.ResponsePagination{
		Message:    "Departments retrieved successfully",
		Data:       departments,
		Pagination: pagination,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}

// UpdateDepartment implements [DepartmentHandler].
func (d *DepartmentHandlerImpl) UpdateDepartment(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid department ID", err)
		return
	}

	departmentUpdateRequest := dto.DepartmentUpdateRequest{}
	if err := helper.ReadFromRequestBody(req, &departmentUpdateRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	departmentUpdateResponse, err := d.DepartmentService.Update(req.Context(), id, departmentUpdateRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Department updated successfully",
		Data:    departmentUpdateResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}
