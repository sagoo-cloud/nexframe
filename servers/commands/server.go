package commands

import (
	"context"
	"errors"
	"github.com/sagoo-cloud/nexframe/command/args"
	"github.com/sagoo-cloud/nexframe/servers/commons"
	"log/slog"
)

type Server struct {
	handlers map[string]*commons.CommHandler
	Logger   *slog.Logger
}

func NewServer() *Server {
	s := &Server{
		Logger:   slog.Default(),
		handlers: make(map[string]*commons.CommHandler),
	}
	return s
}

func (s *Server) Register(name string, handler *commons.CommHandler) {
	s.handlers[name] = handler

}

func (s *Server) Serve() error {
	if args.Cmd != "" {
		handler, isExist := s.handlers[args.Cmd]
		if isExist == false {
			return errors.New("handler not exist")
		}
		ctx := context.Background()
		response, err := handler.Handle(ctx, args.Args)
		if err != nil {
			return err
		}
		s.Logger.Info("response:", response)
	}
	return nil
}
func (s *Server) Close() {

}
