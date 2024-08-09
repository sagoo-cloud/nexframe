package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Context interface {
	context.Context
	Request() *http.Request
	Response() http.ResponseWriter
	JSON(code int, v interface{}) error
	Middleware(h Handler) Handler
	Result(code int, v interface{}) error
	Reset(res http.ResponseWriter, req *http.Request)
}

type wrapper struct {
	req *http.Request
	res http.ResponseWriter
}

func (c *wrapper) Request() *http.Request        { return c.req }
func (c *wrapper) Response() http.ResponseWriter { return c.res }
func (c *wrapper) JSON(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/json")
	c.res.WriteHeader(code)
	return json.NewEncoder(c.res).Encode(v)
}
func (c *wrapper) Reset(res http.ResponseWriter, req *http.Request) {
	c.res = res
	c.req = req
}
func (c *wrapper) Deadline() (time.Time, bool) {
	if c.req == nil {
		return time.Time{}, false
	}
	return c.req.Context().Deadline()
}

func (c *wrapper) Done() <-chan struct{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Done()
}

func (c *wrapper) Err() error {
	if c.req == nil {
		return context.Canceled
	}
	return c.req.Context().Err()
}

func (c *wrapper) Value(key interface{}) interface{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Value(key)
}

func (c *wrapper) Result(code int, v interface{}) error {
	c.res.WriteHeader(code)
	return json.NewEncoder(c.res).Encode(map[string]interface{}{
		"code": code,
		"data": v,
	})
}

func (c *wrapper) Middleware(h Handler) Handler {
	return Chain()(h)
}
