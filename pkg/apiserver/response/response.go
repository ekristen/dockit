package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Response struct {
	Status bool        `json:"success"`
	Data   interface{} `json:"data"`
	Errors []string    `json:"errors,omitempty"`

	s bool
	w http.ResponseWriter
	r *http.Request
}

func New(w http.ResponseWriter, r *http.Request) *Response {
	return &Response{
		w: w,
		r: r,
		s: true,
	}
}
func (r *Response) Success() *Response {
	r.Status, r.s = true, true
	return r
}
func (r *Response) Failed() *Response {
	r.Status, r.s = false, false
	return r
}
func (r *Response) AddError(err error) *Response {
	r.Status, r.s = false, false
	r.Errors = append(r.Errors, err.Error())
	return r
}
func (r *Response) AddData(data interface{}) *Response {
	r.Status, r.s = true, true
	r.Data = data
	return r
}
func (r *Response) Send(code int) {
	r.w.WriteHeader(code)
	r.w.Header().Add("content-type", "application/json")
	if err := json.NewEncoder(r.w).Encode(r); err != nil {
		r.w.WriteHeader(500)
		r.w.Write([]byte(`{"success": false, "errors":[{"message":""}]}`))
	}
}

func ReadAllDecode(in io.Reader) (r *Response, err error) {
	r = &Response{}
	if err = json.NewDecoder(in).Decode(r); err != nil {
		fmt.Println("error", err)
		return nil, err
	}

	return r, err
}
