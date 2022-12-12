package dto

const (
	StatusAllowNotFound = 1002
)

type AllowRequestBody struct {
	Namespace string `json:"namespace"`
	Resource  string `json:"resource"`
	Tokens    int64  `json:"tokens"`
}

type AllowResponseBody struct {
	WaitTime int64 `json:"wait_time"`
	OK       bool  `json:"ok"`
}
