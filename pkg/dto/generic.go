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

func NewResponseBody[T any](status int, msg string, result T) ResponseBody[T] {
	return ResponseBody[T]{
		Status: status,
		Msg:    msg,
		Result: result,
	}
}
