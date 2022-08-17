package handlers

import (
	"encoding/json"
	"github.com/Blinkuu/qms/internal/core/ports"
	"net/http"
)

type QuotaHTTPHandler struct {
	service ports.QuotaService
}

func NewQuotaHTTPHandler(service ports.QuotaService) *QuotaHTTPHandler {
	return &QuotaHTTPHandler{
		service: service,
	}
}

func (h *QuotaHTTPHandler) Allow() http.HandlerFunc {
	type allowRequest struct {
		Namespace string `json:"namespace"`
		Resource  string `json:"resource"`
		Weight    int64  `json:"weight"`
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

		waitTime, err := h.service.Allow(req.Namespace, req.Resource, req.Weight)
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
