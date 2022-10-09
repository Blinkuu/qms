package dto

const (
	StatusAllocInvalidVersion = 1002
)

type AllocRequestBody struct {
	Namespace string `json:"namespace"`
	Resource  string `json:"resource"`
	Tokens    int64  `json:"tokens"`
	Version   int64  `json:"version"`
}

type AllocResponseBody struct {
	RemainingTokens int64 `json:"remaining_tokens"`
	CurrentVersion  int64 `json:"current_version"`
	OK              bool  `json:"ok"`
}
