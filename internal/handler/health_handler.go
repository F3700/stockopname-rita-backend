package handler

import (
	"context"
	"net/http"
	"time"

	"stockopname-rita-backend/internal/helper"

	"github.com/julienschmidt/httprouter"
)

// Pinger is the part of the database pool a readiness check needs.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler separates liveness from readiness. A probe that only asks "is
// the process up" must not restart the container when the database blips, while
// a deploy script does want to know whether the API can actually serve.
type HealthHandler struct {
	DB      Pinger
	Timeout time.Duration
}

func NewHealthHandler(db Pinger, timeout time.Duration) *HealthHandler {
	return &HealthHandler{DB: db, Timeout: timeout}
}

type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db,omitempty"`
}

// Live implements liveness: the process answers, nothing else is checked.
func (h *HealthHandler) Live(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	if err := helper.ResponseJson(writer, http.StatusOK, healthResponse{Status: "ok"}); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

// Ready implements readiness: it fails when the database cannot be reached.
func (h *HealthHandler) Ready(writer http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	ctx, cancel := context.WithTimeout(req.Context(), h.Timeout)
	defer cancel()

	if err := h.DB.Ping(ctx); err != nil {
		helper.WriteError(writer, http.StatusServiceUnavailable, "Database is not reachable", err)
		return
	}

	if err := helper.ResponseJson(writer, http.StatusOK, healthResponse{Status: "ready", DB: "ok"}); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
	}
}
