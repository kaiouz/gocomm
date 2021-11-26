package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// ParseCode 解析结果
func ParseCode(res *http.Response, v interface{}) error {
	if res.StatusCode != 200 {
		return fmt.Errorf("http status error, %v:%v", res.StatusCode, res.Status)
	}

	defer res.Body.Close()

	rc := ResultCode{Data: v}
	err := json.NewDecoder(res.Body).Decode(&rc)

	if err != nil {
		return fmt.Errorf("http json result parse error, %w", err)
	}

	if rc.Code != 0 {
		return fmt.Errorf(rc.Msg)
	}

	return nil
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

// 异常恢复中间件
func PanicRecover(log *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("request error, url: %v\n%+v", rq.URL, err)

				// To avoid 'superfluous response.WriteHeader call' error
				if rw.Header().Get("Content-Type") == "" {
					rw.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(rw, rq)
	})
}

// GetJson
func GetJson(url string, params map[string]string, ret interface{}) error {
	return GetJsonWithContext(context.Background(), url, params, ret)
}

// GetJsonWithContext
func GetJsonWithContext(ctx context.Context, url string, params map[string]string, ret interface{}) error {
	var err error

	var req *http.Request

	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil); err == nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
		var res *http.Response
		if res, err = http.DefaultClient.Do(req); err == nil {
			err = ParseCode(res, ret)
		}
	}

	return err
}

// Download
func Download(url string, writer io.Writer) error {
	return DownloadWithContext(context.Background(), url, writer)
}

// DownloadWithContext
func DownloadWithContext(ctx context.Context, url string, writer io.Writer) error {
	var err error

	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil); err == nil {
		var res *http.Response
		if res, err = http.DefaultClient.Do(req); err == nil {
			if res.StatusCode == 200 {
				defer res.Body.Close()
				_, err = io.Copy(writer, res.Body)
			} else {
				err = fmt.Errorf("http error, code=%v\n", res.StatusCode)
			}
		}
	}

	return err
}

// PostJson
func PostJson(url string, data, ret interface{}) error {
	return PostJsonWithContext(context.Background(), url, data, ret)
}

// PostJsonWithContext
func PostJsonWithContext(ctx context.Context, url string, data, ret interface{}) error {
	var err error
	var content []byte

	if content, err = json.Marshal(data); err == nil {
		var req *http.Request
		if req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(content)); err == nil {
			req.Header.Set("Content-Type", "application/json")
			var res *http.Response
			if res, err = http.DefaultClient.Do(req); err == nil {
				err = ParseCode(res, ret)
			}
		}
	}

	return err
}
