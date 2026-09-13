package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type CoordinatorHandlerImpl struct {
	CoordinatorService service.CoordinatorService
}

func NewCoordinatorHandler(coordinatorService service.CoordinatorService) CoordinatorHandler {
	return &CoordinatorHandlerImpl{
		CoordinatorService: coordinatorService,
	}
}

func (c *CoordinatorHandlerImpl) UpdateCoordinator(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid coordinator ID", err)
		return
	}

	var updateReq dto.UpdateCoordinatorRequest
	if err := helper.ReadFromRequestBody(req, &updateReq); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := c.CoordinatorService.Update(req.Context(), id, &updateReq); err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Coordinator updated successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}

// GetCoordinatorById implements [CoordinatorHandler].
func (c *CoordinatorHandlerImpl) GetCoordinatorById(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid coordinator ID", err)
		return
	}

	coordinator, err := c.CoordinatorService.FindByIdSummary(req.Context(), id)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Coordinator retrieved successfully",
		Data:    coordinator,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}

// GetCoordinators implements [CoordinatorHandler].
func (c *CoordinatorHandlerImpl) GetCoordinators(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	sesiIdStr := req.URL.Query().Get("sessionId")
	var sesiID *int

	if sesiIdStr != "" {
		parsedSesiID, err := strconv.Atoi(sesiIdStr)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid sesi ID", err)
			return
		}
		sesiID = &parsedSesiID
	}

	coordinators, err := c.CoordinatorService.FindAllSummary(req.Context(), sesiID)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Coordinators retrieved successfully",
		Data:    coordinators,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}
