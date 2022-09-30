package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/ports"
	"github.com/Blinkuu/qms/pkg/dto"
)

type MemberlistHTTPHandler struct {
	service ports.MemberlistService
}

func NewMemberlistHTTPHandler(service ports.MemberlistService) *MemberlistHTTPHandler {
	return &MemberlistHTTPHandler{
		service: service,
	}
}

func (h *MemberlistHTTPHandler) Memberlist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		members, err := h.service.Members(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			dto.NewOKResponseBody(
				dto.MemberlistResponseBody{
					Members: members,
				},
			),
		)
	}
}
