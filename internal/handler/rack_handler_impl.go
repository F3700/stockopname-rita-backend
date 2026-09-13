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

func NewRackHandler(rackService service.RackService) *RackHandlerImpl {
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
	inspectorIdStr := req.URL.Query().Get("inspectorId")
	coordinatorIdStr := req.URL.Query().Get("coordinatorId")
	sessionIdStr := req.URL.Query().Get("sessionId")

	var inspectorId, coordinatorId, sessionId *int

	if inspectorIdStr != "" {
		parsedInspectorId, err := strconv.Atoi(inspectorIdStr)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid inspectorId parameter", err)
			return
		}
		inspectorId = &parsedInspectorId
	}

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

	racks, err := r.RackService.FindAllSummary(req.Context(), inspectorId, coordinatorId, sessionId)
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

func (r *RackHandlerImpl) CreateRack(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var rackRequest dto.CreateRackRequest
	if err := helper.ReadFromRequestBody(req, &rackRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	rack, err := r.RackService.CreateRack(req.Context(), rackRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Rack created successfully",
		Data:    rack,
	}

	if err := helper.ResponseJson(writer, http.StatusCreated, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Error occurred while sending response", err)
		return
	}
}
