package dto

type FreeRequestBody struct {
	Namespace string `json:"namespace"`
	Resource  string `json:"resource"`
	Tokens    int64  `json:"tokens"`
}

type FreeResponseBody struct {
	RemainingTokens int64 `json:"remaining_tokens"`
	OK              bool  `json:"ok"`
}
