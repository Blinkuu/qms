package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/pkg/dto"
)

type HealthHTTPHandler struct {
	services []services.Service
}

func NewHealthHTTPHandler(services ...services.Service) *HealthHTTPHandler {
	return &HealthHTTPHandler{
		services: services,
	}
}

func (h *HealthHTTPHandler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		healthy := true
		for _, service := range h.services {
			if service.State() == services.Failed {
				healthy = false

				break
			}
		}

		w.Header().Set("Content-Type", "application/json")

		if !healthy {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(
				dto.NewResponseBody(
					dto.StatusInternalError,
					"not healthy",
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
