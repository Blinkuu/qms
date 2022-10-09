package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
	"github.com/Blinkuu/qms/internal/core/services/alloc"
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
		var req dto.AllocRequestBody
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		remainingTokens, currentVersion, ok, err := h.service.Alloc(r.Context(), req.Namespace, req.Resource, req.Tokens, req.Version)
		if err != nil {
			switch {
			case errors.Is(err, alloc.ErrInvalidVersion):
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(
					dto.NewResponseBody(
						dto.StatusAllocInvalidVersion,
						err.Error(),
						dto.AllocResponseBody{},
					),
				)
			default:
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(
				dto.AllocResponseBody{
					RemainingTokens: remainingTokens,
					CurrentVersion:  currentVersion,
					OK:              ok,
				},
			),
		)
	}
}

func (h *AllocHTTPHandler) Free() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.FreeRequestBody
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		remainingTokens, currentVersion, ok, err := h.service.Free(r.Context(), req.Namespace, req.Resource, req.Tokens, req.Version)
		if err != nil {
			switch {
			case errors.Is(err, alloc.ErrInvalidVersion):
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(
					dto.NewResponseBody(
						dto.StatusAllocInvalidVersion,
						err.Error(),
						dto.AllocResponseBody{},
					),
				)
			default:
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(
				dto.FreeResponseBody{
					RemainingTokens: remainingTokens,
					CurrentVersion:  currentVersion,
					OK:              ok,
				},
			),
		)
	}
}
