package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type RackHandlerImpl struct {
	RackService service.RackService
}

func NewRackHandler(rackService service.RackService) RackHandler {
	return &RackHandlerImpl{
		RackService: rackService,
	}
}

// GetRackProgress implements [RackHandler].
func (r *RackHandlerImpl) GetRackProgress(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
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

	progress, err := r.RackService.FindProgress(req.Context(), coorId, sesiId)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Rack progress retrieved successfully",
		Data:    progress,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Error occurred while sending response", err)
		return
	}
}

// GetRacks implements [RackHandler].
func (r *RackHandlerImpl) GetRacks(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	coordinatorIdStr := req.URL.Query().Get("coordinatorId")
	sessionIdStr := req.URL.Query().Get("sessionId")

	var coordinatorId, sessionId *int

	if coordinatorIdStr != "" {
		parsedCoordinatorId, err := strconv.Atoi(coordinatorIdStr)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid coordinatorId parameter", err)
			return
		}
		coordinatorId = &parsedCoordinatorId
	}

	if sessionIdStr != "" {
		parsedSessionId, err := strconv.Atoi(sessionIdStr)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid sessionId parameter", err)
			return
		}
		sessionId = &parsedSessionId
	}

	racks, err := r.RackService.FindAllSummary(req.Context(), coordinatorId, sessionId)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Racks retrieved successfully",
		Data:    racks,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Error occurred while sending response", err)
		return
	}
}

// DeleteRack implements [RackHandler].
func (r *RackHandlerImpl) DeleteRack(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid rack ID", err)
		return
	}

	if err := r.RackService.DeleteRack(req.Context(), id); err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Rack deleted successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Error occurred while sending response", err)
		return
	}
}
