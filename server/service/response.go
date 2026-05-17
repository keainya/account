package service

type Response struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func DefaultResponse() Response {
	return Response{
		Code: 0,
		Msg:  "ok",
		Data: nil,
	}
}
