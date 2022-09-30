package dto

import (
	"github.com/Blinkuu/qms/internal/core/domain"
)

type MemberlistResponseBody struct {
	Members []domain.Instance `json:"members"`
}
