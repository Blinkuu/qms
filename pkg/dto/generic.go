package dto

const (
	StatusOK = 1001
)

const (
	MsgOK = "ok"
)

type ResponseBody[T any] struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Result T      `json:"result,omitempty"`
}

func NewOKResponseBody[T any](result T) ResponseBody[T] {
	return ResponseBody[T]{
		Status: StatusOK,
		Msg:    MsgOK,
		Result: result,
	}
}
