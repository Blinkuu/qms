package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
	"github.com/Blinkuu/qms/internal/core/services/rate"
	"github.com/Blinkuu/qms/pkg/dto"
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
	return func(w http.ResponseWriter, r *http.Request) {
		var allowRequestBody dto.AllowRequestBody
		err := json.NewDecoder(r.Body).Decode(&allowRequestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		waitTime, ok, err := h.service.Allow(r.Context(), allowRequestBody.Namespace, allowRequestBody.Resource, allowRequestBody.Tokens)
		if err != nil {
			switch {
			case errors.Is(err, rate.ErrNotFound):
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(
					dto.NewResponseBody(
						dto.StatusAllowNotFound,
						err.Error(),
						dto.AllowResponseBody{},
					),
				)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(
				dto.AllowResponseBody{
					WaitTime: waitTime.Nanoseconds(),
					OK:       ok,
				},
			),
		)
	}
}
