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
	StockOpnameRequest.Actor = helper.ActorFromRequest(req)

	StockOpnameResponse, err := s.StockOpnameService.Create(req.Context(), StockOpnameRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
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

	if err := s.StockOpnameService.Delete(req.Context(), id, helper.ActorFromRequest(req)); err != nil {
		helper.WriteServiceError(writer, err)
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

	if err := helper.ValidatePage(page); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query parameter", err)
		return
	}

	if err := helper.ValidateLimit(limit); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query parameter", err)
		return
	}

	search := req.URL.Query().Get("search")

	sesiIdStr := req.URL.Query().Get("sessionId")
	coorIdStr := req.URL.Query().Get("coordinatorId")
	rakIdStr := req.URL.Query().Get("rakId")

	var sesiId, coorId, rakId *int

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

	if rakIdStr != "" {
		parsedRakId, err := strconv.Atoi(rakIdStr)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid rakId parameter", err)
			return
		}
		rakId = &parsedRakId
	}

	pagination := &dto.Pagination{
		Page:  page,
		Limit: limit,
	}

	stockOpnames, err := s.StockOpnameService.FindAll(req.Context(), pagination, search, coorId, sesiId, rakId)
	if err != nil {
		helper.WriteServiceError(writer, err)
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
		helper.WriteServiceError(writer, err)
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
	soUpdateRequest.Actor = helper.ActorFromRequest(req)

	stockOpnameResponse, err := s.StockOpnameService.Update(req.Context(), id, soUpdateRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
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

// MoveStockOpname implements [StockOpnameHandler].
func (s *StockOpnameHandlerImpl) MoveStockOpname(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	moveRequest := dto.MoveStockOpnameRequest{}
	if err := helper.ReadFromRequestBody(req, &moveRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	moveRequest.Actor = helper.ActorFromRequest(req)

	moveResponse, err := s.StockOpnameService.Move(req.Context(), moveRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Stock opname results moved successfully",
		Data:    moveResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}

// SetQtyFactor implements PATCH /stockopname/sessions/:id/qty-factor.
func (s *StockOpnameHandlerImpl) SetQtyFactor(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid session ID", err)
		return
	}

	var setRequest dto.SetQtyFactorRequest
	if err := helper.ReadFromRequestBody(req, &setRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	setRequest.Actor = helper.ActorFromRequest(req)

	result, err := s.StockOpnameService.SetQtyFactor(req.Context(), id, setRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	if err := helper.ResponseJson(writer, http.StatusOK, dto.Response{
		Message: "Qty factor updated successfully",
		Data:    result,
	}); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
	}
}

// GetSellSummary implements GET /stockopname/sessions/:id/sell-summary.
func (s *StockOpnameHandlerImpl) GetSellSummary(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid session ID", err)
		return
	}
	groupBy := req.URL.Query().Get("groupBy")
	if groupBy == "" {
		helper.WriteError(writer, http.StatusBadRequest, "groupBy parameter is required", nil)
		return
	}

	result, err := s.StockOpnameService.FindSellSummary(req.Context(), id, groupBy)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	if err := helper.ResponseJson(writer, http.StatusOK, dto.Response{
		Message: "Sell summary retrieved successfully",
		Data:    result,
	}); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
	}
}
