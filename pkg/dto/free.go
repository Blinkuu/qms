package dto

type FreeRequestBody struct {
	Namespace string `json:"namespace"`
	Resource  string `json:"resource"`
	Tokens    int64  `json:"tokens"`
	Version   int64  `json:"version"`
}

type FreeResponseBody struct {
	RemainingTokens int64 `json:"remaining_tokens"`
	CurrentVersion  int64 `json:"current_version"`
	OK              bool  `json:"ok"`
}
