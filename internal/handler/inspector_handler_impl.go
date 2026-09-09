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

func NewInspectorHandler(inspectorService service.InspectorService) *InspectorHandlerImpl {
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
			http.Error(writer, "Invalid coordinatorId parameter", http.StatusBadRequest)
			return
		}
		coorId = &parsedCoorId
	}

	inspectors, err := i.InspectorService.FindAllSummary(req.Context(), coorId)
	if err != nil {
		http.Error(writer, "Failed to retrieve inspectors", http.StatusInternalServerError)
		return
	}

	response := dto.Response{
		Message: "Inspectors retrieved successfully",
		Data:    inspectors,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		http.Error(writer, "Failed to send response", http.StatusInternalServerError)
		return
	}
}
