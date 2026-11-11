package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/repository"
	"stockopname-rita-backend/internal/service"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func isUUID(value string) bool {
	return uuidPattern.MatchString(value)
}

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
	if value := req.URL.Query().Get("batchId"); value != "" {
		if !isUUID(value) {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid batchId parameter", nil)
			return
		}
		filter.BatchID = &value
	}
	filter.WithoutBatch = req.URL.Query().Get("withoutBatch") == "true"

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

// GetAuditBatches implements GET /stockopname/audits/batches.
func (h *AuditHandler) GetAuditBatches(writer http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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
	batches, err := h.AuditService.FindBatches(req.Context(), filter, pagination)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.ResponsePagination{
		Message:    "Audit batches retrieved successfully",
		Data:       batches,
		Pagination: pagination,
	}
	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
	}
}

var auditTimelineActions = map[string]struct{}{
	"CREATE": {},
	"UPDATE": {},
	"DELETE": {},
	"MOVE":   {},
}

// auditTimelineFilter reads the shared audit filters plus the optional action
// list (comma separated or repeated query parameter).
func auditTimelineFilter(req *http.Request) (repository.AuditFilter, error) {
	var filter repository.AuditFilter

	if value := req.URL.Query().Get("sessionId"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return filter, err
		}
		filter.SesiID = &parsed
	}
	if value := req.URL.Query().Get("coordinatorId"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return filter, err
		}
		filter.CoordinatorID = &parsed
	}
	if value := req.URL.Query().Get("resultId"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return filter, err
		}
		filter.ResultID = &parsed
	}

	seen := make(map[string]struct{})
	for _, value := range req.URL.Query()["action"] {
		for _, part := range strings.Split(value, ",") {
			action := strings.ToUpper(strings.TrimSpace(part))
			if action == "" {
				continue
			}
			if _, ok := auditTimelineActions[action]; !ok {
				return filter, fmt.Errorf("unknown action %q", action)
			}
			if _, duplicate := seen[action]; duplicate {
				continue
			}
			seen[action] = struct{}{}
			filter.Actions = append(filter.Actions, action)
		}
	}

	return filter, nil
}

// GetAuditTimeline implements GET /stockopname/audits/timeline.
func (h *AuditHandler) GetAuditTimeline(writer http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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

	filter, err := auditTimelineFilter(req)
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query parameter", err)
		return
	}

	pagination := &dto.Pagination{Page: page, Limit: limit}
	items, err := h.AuditService.FindTimeline(req.Context(), filter, pagination)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.ResponsePagination{
		Message:    "Audit timeline retrieved successfully",
		Data:       items,
		Pagination: pagination,
	}
	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to respond with JSON", err)
	}
}
