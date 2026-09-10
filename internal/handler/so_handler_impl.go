package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type StockOpnameHandlerImpl struct {
	StockOpnameService service.StockOpnameService
}

func NewStockOpnameHandler(stockOpnameService service.StockOpnameService) StockOpnameHandler {
	return &StockOpnameHandlerImpl{
		StockOpnameService: stockOpnameService,
	}
}

// CreateStockOpname implements [StockOpnameHandler].
func (s *StockOpnameHandlerImpl) CreateStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	StockOpnameRequest := dto.CreateStockOpnameRequest{}
	if err := helper.ReadFromRequestBody(req, &StockOpnameRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	StockOpnameResponse, err := s.StockOpnameService.Create(req.Context(), StockOpnameRequest)
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to create stock opname", err)
		return
	}

	response := dto.Response{
		Message: "Stock opname created successfully",
		Data:    StockOpnameResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusCreated, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

// DeleteStockOpname implements [StockOpnameHandler].
func (s *StockOpnameHandlerImpl) DeleteStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid parameter", err)
		return
	}

	if err := s.StockOpnameService.Delete(req.Context(), id); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to delete stock opname", err)
		return
	}

	response := dto.Response{
		Message: "Stock opname deleted successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

// GetAllStockOpname implements [StockOpnameHandler].
func (s *StockOpnameHandlerImpl) GetAllStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	page, err := helper.ParseIntQueryNotNull(req, "page")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query parameter", err)
		return
	}

	limit, err := helper.ParseIntQueryNotNull(req, "limit")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query parameter", err)
		return
	}

	search := req.URL.Query().Get("search")

	sesiIdStr := req.URL.Query().Get("sessionId")
	coorIdStr := req.URL.Query().Get("coordinatorId")

	var sesiId, coorId *int

	if sesiIdStr != "" {
		parsedSesiId, err := strconv.Atoi(sesiIdStr)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid sessionId parameter", err)
			return
		}
		sesiId = &parsedSesiId
	}

	if coorIdStr != "" {
		parsedCoorId, err := strconv.Atoi(coorIdStr)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid coordinatorId parameter", err)
			return
		}
		coorId = &parsedCoorId
	}

	pagination := &dto.Pagination{
		Page:  page,
		Limit: limit,
	}

	stockOpnames, err := s.StockOpnameService.FindAll(req.Context(), pagination, search, coorId, sesiId)
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to get all stock opname", err)
		return
	}

	response := dto.ResponsePagination{
		Message:    "Stock opname retrieved successfully",
		Data:       stockOpnames,
		Pagination: pagination,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

// GetStockOpnameById implements [StockOpnameHandler].
func (s *StockOpnameHandlerImpl) GetStockOpnameById(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid parameter", err)
		return
	}

	stockOpnameResponse, err := s.StockOpnameService.FindById(req.Context(), id)
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to get stock opname", err)
		return
	}

	response := dto.Response{
		Message: "Stock opname retrieved successfully",
		Data:    stockOpnameResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

// UpdateStockOpname implements [StockOpnameHandler].
func (s *StockOpnameHandlerImpl) UpdateStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid parameter", err)
		return
	}

	soUpdateRequest := dto.UpdateStockOpnameRequest{}
	if err := helper.ReadFromRequestBody(req, &soUpdateRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	stockOpnameResponse, err := s.StockOpnameService.Update(req.Context(), id, soUpdateRequest)
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to update stock opname", err)
		return
	}

	response := dto.Response{
		Message: "Stock opname updated successfully",
		Data:    stockOpnameResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

func (s *StockOpnameHandlerImpl) CreateStockOpnameByRack(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	soByRackRequest := dto.CreateStockOpnameByRackRequest{}
	if err := helper.ReadFromRequestBody(req, &soByRackRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.StockOpnameService.CreateByRack(req.Context(), soByRackRequest); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to create stock opname by rack", err)
		return
	}

	response := dto.Response{
		Message: "Stock opname by rack created successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}
