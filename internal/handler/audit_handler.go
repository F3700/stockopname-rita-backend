package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/repository"
	"stockopname-rita-backend/internal/service"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type AuditHandler struct {
	AuditService service.AuditService
}

func NewAuditHandler(auditService service.AuditService) *AuditHandler {
	return &AuditHandler{AuditService: auditService}
}

// GetAudits implements GET /stockopname/audits.
func (h *AuditHandler) GetAudits(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
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

	filter := repository.AuditFilter{}
	if value := req.URL.Query().Get("sessionId"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid sessionId parameter", err)
			return
		}
		filter.SesiID = &parsed
	}
	if value := req.URL.Query().Get("coordinatorId"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid coordinatorId parameter", err)
			return
		}
		filter.CoordinatorID = &parsed
	}
	if value := req.URL.Query().Get("resultId"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid resultId parameter", err)
			return
		}
		filter.ResultID = &parsed
	}

	pagination := &dto.Pagination{Page: page, Limit: limit}
	audits, err := h.AuditService.FindAll(req.Context(), filter, pagination)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.ResponsePagination{
		Message:    "Audits retrieved successfully",
		Data:       audits,
		Pagination: pagination,
	}
	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
		return
	}
}
