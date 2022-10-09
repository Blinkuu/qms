package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
	"github.com/Blinkuu/qms/pkg/dto"
)

type ProxyHTTPHandler struct {
	service ports.ProxyService
}

func NewProxyHTTPHandler(service ports.ProxyService) *ProxyHTTPHandler {
	return &ProxyHTTPHandler{
		service: service,
	}
}

func (h *ProxyHTTPHandler) Allow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AllowRequestBody
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
			dto.NewOKResponseBody(
				dto.AllowResponseBody{
					WaitTime: waitTime.Nanoseconds(),
					OK:       ok,
				},
			),
		)
	}
}

func (h *ProxyHTTPHandler) Alloc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AllocRequestBody
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		remainingTokens, currentVersion, ok, err := h.service.Alloc(r.Context(), req.Namespace, req.Resource, req.Tokens, req.Version)
		if err != nil {
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

func (h *ProxyHTTPHandler) Free() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.FreeRequestBody
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		remainingTokens, currentVersion, ok, err := h.service.Free(r.Context(), req.Namespace, req.Resource, req.Tokens, req.Version)
		if err != nil {
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
