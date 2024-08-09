package http

import (
	"context"
	"encoding/json"
	"github.com/ServiceWeaver/weaver"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

type HandlerFunc func(ctx Context) error

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc func(http.ResponseWriter, *http.Request, interface{}) error

func DefaultResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	type ms struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	switch e := v.(type) {
	case error:
		return json.NewEncoder(w).Encode(ms{
			Code:    50,
			Message: e.Error(),
		})
	}
	return json.NewEncoder(w).Encode(ms{
		Code:    0,
		Message: "OK",
		Data:    v,
	})
}

type Server struct {
	router *mux.Router
	pool   sync.Pool
	ene    EncodeResponseFunc
}

func NewServer(ctx context.Context) *Server {
	r := mux.NewRouter()
	server := &Server{
		router: r,
		ene:    DefaultResponseEncoder,
	}
	server.pool.New = func() any {
		return &wrapper{}
	}
	return server
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Handle(method, path string, handler HandlerFunc, label string) {
	next := weaver.InstrumentHandlerFunc(label, func(writer http.ResponseWriter, request *http.Request) {
		ctx := s.pool.Get().(Context)
		ctx.Reset(writer, request)
		ctx.Reset(nil, nil)
		s.pool.Put(ctx)
	})
	s.router.Handle(path, next)
}

func (s *Server) HandlerFunc(path string, handler HandlerFunc) *mux.Route {
	next := http.Handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := s.pool.Get().(Context)
		ctx.Reset(writer, request)
		if err := ctx.Middleware(func(ctx Context) error {
			return handler(ctx)
		})(ctx); err != nil {
			_ = s.ene(ctx.Response(), ctx.Request(), err)
		}
		ctx.Reset(nil, nil)
		s.pool.Put(ctx)
	}))
	return s.router.Handle(path, next)
}
