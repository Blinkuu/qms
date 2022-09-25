package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Blinkuu/qms/internal/core/domain/cloud"
	"github.com/Blinkuu/qms/internal/core/ports"
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
	type memberlistResult struct {
		Members []*cloud.Instance `json:"members"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		members, err := h.service.Members()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(
			response{
				Status: StatusOK,
				Msg:    MsgOK,
				Result: memberlistResult{
					Members: members,
				},
			},
		)
	}
}
