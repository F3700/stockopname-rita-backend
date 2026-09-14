package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type InspectorHandlerImpl struct {
	InspectorService service.InspectorService
}

func NewInspectorHandler(inspectorService service.InspectorService) InspectorHandler {
	return &InspectorHandlerImpl{
		InspectorService: inspectorService,
	}
}

// GetInspectors implements [InspectorHandler].
func (i *InspectorHandlerImpl) GetInspectors(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	coorIdStr := req.URL.Query().Get("coordinatorId")
	var coorId *int

	if coorIdStr != "" {
		parsedCoorId, err := strconv.Atoi(coorIdStr)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid coordinatorId parameter", err)
			return
		}
		coorId = &parsedCoorId
	}

	inspectors, err := i.InspectorService.FindAllSummary(req.Context(), coorId)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Inspectors retrieved successfully",
		Data:    inspectors,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}

func (i *InspectorHandlerImpl) CreateInspector(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var inspectorRequest dto.InspectorRequest
	if err := helper.ReadFromRequestBody(req, &inspectorRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	inspector, err := i.InspectorService.CreateInspector(req.Context(), inspectorRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Inspector created successfully",
		Data:    inspector,
	}

	if err := helper.ResponseJson(writer, http.StatusCreated, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}
