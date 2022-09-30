package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
	"github.com/Blinkuu/qms/pkg/dto"
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
	return func(w http.ResponseWriter, r *http.Request) {
		msg := h.service.Ping(r.Context())

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(
				dto.PingResponseBody{
					Msg: msg,
				},
			),
		)
	}
}
