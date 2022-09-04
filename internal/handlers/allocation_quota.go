package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
)

type AllocationQuotaHTTPHandler struct {
	service ports.AllocationQuotaService
}

func NewAllocationQuotaHTTPHandler(service ports.AllocationQuotaService) *AllocationQuotaHTTPHandler {
	return &AllocationQuotaHTTPHandler{
		service: service,
	}
}

func (h *AllocationQuotaHTTPHandler) Alloc() http.HandlerFunc {
	type allocRequest struct {
		Namespace string `json:"namespace"`
		Resource  string `json:"resource"`
		Tokens    int64  `json:"tokens"`
	}

	type allocResult struct {
		RemainingTokens int64 `json:"remaining_tokens"`
		OK              bool  `json:"ok"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req allocRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		remainingTokens, ok, err := h.service.Alloc(r.Context(), req.Namespace, req.Resource, req.Tokens)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			response{
				Status: StatusOK,
				Msg:    MsgOK,
				Result: allocResult{
					RemainingTokens: remainingTokens,
					OK:              ok,
				},
			},
		)
	}
}

func (h *AllocationQuotaHTTPHandler) Free() http.HandlerFunc {
	type freeRequest struct {
		Namespace string `json:"namespace"`
		Resource  string `json:"resource"`
		Tokens    int64  `json:"tokens"`
	}

	type freeResult struct {
		RemainingTokens int64 `json:"remaining_tokens"`
		OK              bool  `json:"ok"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req freeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		remainingTokens, ok, err := h.service.Free(r.Context(), req.Namespace, req.Resource, req.Tokens)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			response{
				Status: StatusOK,
				Msg:    MsgOK,
				Result: freeResult{
					RemainingTokens: remainingTokens,
					OK:              ok,
				},
			},
		)
	}
}
