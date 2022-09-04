package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
)

type RateQuotaHTTPHandler struct {
	service ports.RateQuotaService
}

func NewRateQuotaHTTPHandler(service ports.RateQuotaService) *RateQuotaHTTPHandler {
	return &RateQuotaHTTPHandler{
		service: service,
	}
}

func (h *RateQuotaHTTPHandler) Allow() http.HandlerFunc {
	type allowRequest struct {
		Namespace string `json:"namespace"`
		Resource  string `json:"resource"`
		Tokens    int64  `json:"tokens"`
	}

	type allowResult struct {
		WaitTime int64 `json:"wait_time"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req allowRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		waitTime, err := h.service.Allow(r.Context(), req.Namespace, req.Resource, req.Tokens)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			response{
				Status: StatusOK,
				Msg:    MsgOK,
				Result: allowResult{
					WaitTime: waitTime.Nanoseconds(),
				},
			},
		)
	}
}
