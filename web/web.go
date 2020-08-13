package web

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ResultCode 接口返回结果
type ResultCode struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// WriteJSON 写入json到响应
func WriteJSON(w http.ResponseWriter, v interface{}) error {
	bytes, err := json.Marshal(v)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(bytes)

	return err
}

// OkCode 写入成功结果
func OkCode(w http.ResponseWriter, v interface{}) error {
	rc := ResultCode{Code: 0, Data: v}
	return WriteJSON(w, rc)
}

// ErrorCode 写入失败结果
func ErrorCode(w http.ResponseWriter, msg string) error {
	rc := &ResultCode{Code: -1, Msg: msg}
	return WriteJSON(w, rc)
}

// Cors 跨域中间件
func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		if origin := r.Header.Get("Origin"); origin != "" {
			header.Set("Access-Control-Allow-Origin", origin)

			if r.Method == "OPTIONS" {
				// Preflight request
				if allowMethod := r.Header.Get("Access-Control-Request-Method"); allowMethod != "" {
					header.Set("Access-Control-Allow-Methods", allowMethod)
				}
				if allowHeaders := r.Header["Access-Control-Request-Headers"]; len(allowHeaders) > 0 {
					header.Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
				}
				header.Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}