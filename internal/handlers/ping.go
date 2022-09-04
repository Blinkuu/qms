package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
)

type PingHTTPHandler struct {
	service ports.PingService
}

func NewPingHTTPHandler(service ports.PingService) *PingHTTPHandler {
	return &PingHTTPHandler{
		service: service,
	}
}

func (h *PingHTTPHandler) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		result := h.service.Ping()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response{Status: StatusOK, Msg: MsgOK, Result: result})
	}
}
