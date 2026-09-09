package httpio

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	status  int
	body    any
	cookies []*http.Cookie
	header  http.Header
}

func JSON(status int, body any) *Response {
	return &Response{status: status, body: body, header: http.Header{}}
}

func NoContent() *Response {
	return &Response{status: http.StatusNoContent, header: http.Header{}}
}

func (r *Response) WithCookie(c *http.Cookie) *Response {
	r.cookies = append(r.cookies, c)
	return r
}

func (r *Response) WithHeader(key, value string) *Response {
	r.header.Set(key, value)
	return r
}

func (r *Response) write(w http.ResponseWriter) {
	for key, values := range r.header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	for _, c := range r.cookies {
		http.SetCookie(w, c)
	}

	if r.body == nil {
		w.WriteHeader(r.status)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(r.status)
	_ = json.NewEncoder(w).Encode(r.body)
}
