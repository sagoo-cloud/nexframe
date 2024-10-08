package commons

import (
	"context"
)

// Handler 通用接口
type Handler interface {
	ServeHandle(ctx context.Context, request interface{}) (interface{}, error)
}

type CommHandler struct {
	Handler Handler
}

func (s *CommHandler) Handle(ctx context.Context, req interface{}) (interface{}, error) {
	rsp, err := s.Handler.ServeHandle(ctx, req)
	if err != nil {
		return nil, err
	}
	return rsp, err
}

// 该接口的实现是为了 cronjob
func (s *CommHandler) Run() {
	ctx := context.Background()
	req := make(map[string]interface{})
	_, _ = s.Handler.ServeHandle(ctx, req)
}
