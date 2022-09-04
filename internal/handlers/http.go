package handlers

const (
	StatusOK = 1001
)

const (
	MsgOK = "ok"
)

type response struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result,omitempty"`
}
