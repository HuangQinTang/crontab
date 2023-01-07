package common

import (
	"encoding/json"
	"log"
	"net/http"
)

type Errno int

const (
	Success Errno = 0
	Fail    Errno = 1
)

// Response http应答
type Response struct {
	Errno      Errno       `json:"errno"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data"`
	httpResp   http.ResponseWriter
	httpStatus int
	httpHeader map[string]string
}

type SetResponse func(*Response)

// SetResp 设置响应状态码，消息，响应数据
func SetResp(errno Errno, msg string, data interface{}) SetResponse {
	return func(resp *Response) {
		resp.Errno = errno
		resp.Msg = msg
		resp.Data = data
	}
}

// SetHttpCode 设置http状态吗
func SetHttpCode(code int) SetResponse {
	return func(resp *Response) {
		resp.httpStatus = code
	}
}

func SetHttpHeader(header map[string]string) SetResponse {
	return func(resp *Response) {
		resp.httpHeader = header
	}
}

func NewResponse(httpResp http.ResponseWriter, funs ...SetResponse) *Response { // todo 好像没必要弄成这样诶
	r := new(Response)
	r.httpResp = httpResp
	for _, f := range funs {
		f(r)
	}
	return r
}

func ReturnOkJson(httpResp http.ResponseWriter, data interface{}) {
	NewResponse(httpResp, SetResp(Success, "请求成功", data)).ReturnJson()
}

func ReturnFailJson(httpResp http.ResponseWriter, err error) {
	NewResponse(httpResp, SetResp(Fail, "请求失败，"+err.Error(), nil)).ReturnJson()
}

// ReturnJson json序列化，响应http请求
func (r *Response) ReturnJson() { // todo 嫌麻烦，这里错误不返回，后续可以记录日志
	var (
		resp []byte
		err  error
	)

	// 设置响应头
	for attr, value := range r.httpHeader {
		r.httpResp.Header().Set(attr, value)
	}

	// 设置状态码
	if r.httpStatus != 0 {
		r.httpResp.WriteHeader(r.httpStatus)
	} else { //未设置状态吗，默认 Errno = 0, 响应 http.StatusOK。 否则响应 http.StatusBadRequest。
		if r.Errno == 0 {
			r.httpResp.WriteHeader(http.StatusOK)
		} else {
			r.httpResp.WriteHeader(http.StatusBadRequest)
		}
	}

	// 设置响应体
	if r.Data == nil { //数据为空时，避免返回nil，返回 {}
		r.Data = struct{}{}
	}
	if resp, err = json.Marshal(r); err != nil {
		log.Println(err.Error())
	}
	if _, err = r.httpResp.Write(resp); err != nil {
		log.Println(err.Error())
	}
	return
}
