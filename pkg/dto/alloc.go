package dto

type AllocRequestBody struct {
	Namespace string `json:"namespace"`
	Resource  string `json:"resource"`
	Tokens    int64  `json:"tokens"`
}

type AllocResponseBody struct {
	RemainingTokens int64 `json:"remaining_tokens"`
	OK              bool  `json:"ok"`
}
