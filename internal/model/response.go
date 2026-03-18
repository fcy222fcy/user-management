package model

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`    //状态码
	Message string      `json:"message"` //提示信息
	Data    interface{} `json:"data"`    //数据内容
}

func JsonResponse(w http.ResponseWriter, code int, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	res := Response{
		Code:    code,
		Message: msg,
		Data:    data,
	}
	json.NewEncoder(w).Encode(res)
}
