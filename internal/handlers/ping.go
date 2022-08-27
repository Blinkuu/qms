package handlers

import (
	"encoding/json"
	"github.com/Blinkuu/qms/internal/core/ports"
	"net/http"
)

const (
	StatusOK = 1001
)

const (
	MsgOK = "ok"
)

type response struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result,omitempty"`
}

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
