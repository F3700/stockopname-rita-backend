package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

type ReportHandler struct{ Service service.ReportService }

func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{Service: reportService}
}

func (h *ReportHandler) SessionPDF(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid session ID", err)
		return
	}
	data, filename, err := h.Service.SessionPDF(req.Context(), id)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}
	writePDF(writer, filename, data)
}

func (h *ReportHandler) CoordinatorPDF(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid coordinator ID", err)
		return
	}
	data, filename, err := h.Service.CoordinatorPDF(req.Context(), id)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}
	writePDF(writer, filename, data)
}

func writePDF(writer http.ResponseWriter, filename string, data []byte) {
	writer.Header().Set("Content-Type", "application/pdf")
	writer.Header().Set("Content-Disposition", `inline; filename="`+filename+`"`)
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(data)
}
