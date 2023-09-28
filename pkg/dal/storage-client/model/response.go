package model

// 对外数据模型
type Response struct {
	Code    int         `json:"code" `
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
