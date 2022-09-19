package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
)

type RateHTTPHandler struct {
	service ports.RateService
}

func NewRateHTTPHandler(service ports.RateService) *RateHTTPHandler {
	return &RateHTTPHandler{
		service: service,
	}
}

func (h *RateHTTPHandler) Allow() http.HandlerFunc {
	type allowRequest struct {
		Namespace string `json:"namespace"`
		Resource  string `json:"resource"`
		Tokens    int64  `json:"tokens"`
	}

	type allowResult struct {
		WaitTime int64 `json:"wait_time"`
		OK       bool  `json:"ok"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req allowRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		waitTime, ok, err := h.service.Allow(r.Context(), req.Namespace, req.Resource, req.Tokens)
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
					OK:       ok,
				},
			},
		)
	}
}
