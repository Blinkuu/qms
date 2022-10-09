package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
	"github.com/Blinkuu/qms/pkg/dto"
)

type RaftHTTPHandler struct {
	service ports.RaftService
}

func NewRaftHTTPHandler(service ports.RaftService) *RaftHTTPHandler {
	return &RaftHTTPHandler{
		service: service,
	}
}

func (h *RaftHTTPHandler) Join() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var joinRequestBody dto.JoinRequestBody
		err := json.NewDecoder(r.Body).Decode(&joinRequestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		alreadyMember, err := h.service.Join(r.Context(), joinRequestBody.ReplicaID, joinRequestBody.RaftAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(
				dto.JoinResponseBody{
					AlreadyMember: alreadyMember,
				},
			),
		)
	}
}

func (h *RaftHTTPHandler) Exit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var exitRequestBody dto.ExitRequestBody
		err := json.NewDecoder(r.Body).Decode(&exitRequestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.service.Exit(r.Context(), exitRequestBody.ReplicaID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(
				dto.ExitResponseBody{},
			),
		)
	}
}
