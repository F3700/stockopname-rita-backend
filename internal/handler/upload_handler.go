package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type UploadHandler interface {
	CreateUpload(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
	GetUploads(writer http.ResponseWriter, req *http.Request, params httprouter.Params)
}

type UploadHandlerImpl struct {
	UploadService service.UploadService
}

func NewUploadHandler(uploadService service.UploadService) UploadHandler {
	return &UploadHandlerImpl{UploadService: uploadService}
}

// CreateUpload implements [UploadHandler].
func (h *UploadHandlerImpl) CreateUpload(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var uploadRequest dto.CreateUploadRequest
	if err := helper.ReadFromRequestBody(req, &uploadRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	upload, err := h.UploadService.Create(req.Context(), uploadRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Upload created successfully",
		Data:    upload,
	}
	if err := helper.ResponseJson(writer, http.StatusCreated, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Error occurred while sending response", err)
		return
	}
}

// GetUploads implements [UploadHandler].
func (h *UploadHandlerImpl) GetUploads(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	coordinatorIdStr := req.URL.Query().Get("coordinatorId")
	if coordinatorIdStr == "" {
		helper.WriteError(writer, http.StatusBadRequest, "coordinatorId parameter is required", nil)
		return
	}

	coordinatorId, err := strconv.Atoi(coordinatorIdStr)
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid coordinatorId parameter", err)
		return
	}

	uploads, err := h.UploadService.FindAllSummary(req.Context(), coordinatorId)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Uploads retrieved successfully",
		Data:    uploads,
	}
	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Error occurred while sending response", err)
		return
	}
}
