package appctx

import (
	"encoding/json"
	"sync"
	"time"
)

var (
	rsp    *Response
	oneRsp sync.Once
)

type Response struct {
	Code      int         `json:"code,omitempty"`
	Status    bool        `json:"status,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
	Errors    interface{} `json:"errors,omitempty"`
}

func (r *Response) WithCode(code int) *Response {
	r.Code = code
	return r
}

func (r *Response) WithStatus(status bool) *Response {
	r.Status = status

	return r
}

func (r *Response) WithData(data interface{}) *Response {
	r.Data = data
	return r
}

func (r *Response) WithErrors(errors interface{}) *Response {
	r.Errors = errors
	return r
}

func (r *Response) Byte() []byte {
	bytes, err := json.Marshal(r)
	if err != nil {
		return nil
	}

	return bytes
}

func NewResponse() *Response {
	oneRsp.Do(func() {
		rsp = &Response{
			Timestamp: time.Now(),
		}
	})

	x := *rsp

	return &x
}
