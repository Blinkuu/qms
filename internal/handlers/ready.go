package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/pkg/dto"
)

type ReadyHTTPHandler struct {
	services []services.Service
}

func NewReadyHTTPHandler(services ...services.Service) *ReadyHTTPHandler {
	return &ReadyHTTPHandler{
		services: services,
	}
}

func (h *ReadyHTTPHandler) Ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ready := true
		for _, service := range h.services {
			if service.State() != services.Running {
				ready = false

				break
			}
		}

		w.Header().Set("Content-Type", "application/json")

		if !ready {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(
				dto.NewResponseBody(
					dto.StatusInternalError,
					"not ready",
					"",
				),
			)

			return
		}

		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(""),
		)
	}
}
