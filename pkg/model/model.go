package model

// HTTPReply uniform response structure
type HTTPReply struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
