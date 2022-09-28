package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
	"github.com/Blinkuu/qms/pkg/dto"
)

type AllocHTTPHandler struct {
	service ports.AllocService
}

func NewAllocHTTPHandler(service ports.AllocService) *AllocHTTPHandler {
	return &AllocHTTPHandler{
		service: service,
	}
}

func (h *AllocHTTPHandler) Alloc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var allocRequestBody dto.AllocRequestBody
		err := json.NewDecoder(r.Body).Decode(&allocRequestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		remainingTokens, ok, err := h.service.Alloc(r.Context(), allocRequestBody.Namespace, allocRequestBody.Resource, allocRequestBody.Tokens)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(
				dto.AllocResponseBody{
					RemainingTokens: remainingTokens,
					OK:              ok,
				},
			),
		)
	}
}

func (h *AllocHTTPHandler) Free() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var freeRequestBody dto.FreeRequestBody
		err := json.NewDecoder(r.Body).Decode(&freeRequestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		remainingTokens, ok, err := h.service.Free(r.Context(), freeRequestBody.Namespace, freeRequestBody.Resource, freeRequestBody.Tokens)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(
				dto.FreeResponseBody{
					RemainingTokens: remainingTokens,
					OK:              ok,
				},
			),
		)
	}
}
