package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

type SesiHandlerImpl struct {
	SesiService service.SesiService
}

func NewSesiHandler(sesiService service.SesiService) SesiHandler {
	return &SesiHandlerImpl{
		SesiService: sesiService,
	}
}

// CreateSesi implements [SesiHandler].
func (s *SesiHandlerImpl) CreateSesi(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	sesiCreateRequest := dto.CreateSesiRequest{}
	if err := helper.ReadFromRequestBody(req, &sesiCreateRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	sesiResponse, err := s.SesiService.Create(req.Context(), sesiCreateRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Sesi created successfully",
		Data:    sesiResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusCreated, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

// DeleteSesi implements [SesiHandler].
func (s *SesiHandlerImpl) DeleteSesi(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid parameter", err)
		return
	}

	if err := s.SesiService.Delete(req.Context(), id); err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Sesi deleted successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

// GetAllSesi implements [SesiHandler].
func (s *SesiHandlerImpl) GetAllSesi(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
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

	sesiResponses, err := s.SesiService.FindAll(req.Context(), pagination, search)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.ResponsePagination{
		Message:    "Sesi retrieved successfully",
		Data:       sesiResponses,
		Pagination: pagination,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

// GetSesiById implements [SesiHandler].
func (s *SesiHandlerImpl) GetSesiById(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid parameter", err)
		return
	}

	sesiResponse, err := s.SesiService.FindById(req.Context(), id)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Sesi retrieved successfully",
		Data:    sesiResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

// UpdateSesi implements [SesiHandler].
func (s *SesiHandlerImpl) UpdateSesi(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid parameter", err)
		return
	}

	sesiUpdateRequest := dto.UpdateSesiRequest{}
	if err := helper.ReadFromRequestBody(req, &sesiUpdateRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.SesiService.Update(req.Context(), id, sesiUpdateRequest); err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Sesi updated successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}
