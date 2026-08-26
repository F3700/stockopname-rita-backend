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
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to create department", err)
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
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to delete department", err)
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
	departments, err := d.DepartmentService.FindAll(req.Context())
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to retrieve departments", err)
		return
	}

	response := dto.Response{
		Message: "Departments retrieved successfully",
		Data:    departments,
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
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to update department", err)
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
