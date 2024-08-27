package websockets

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/sagoo-cloud/nexframe/contracts"
	"github.com/sagoo-cloud/nexframe/servers/commons"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	handlers     map[string]*commons.CommHandler
	handlersMu   sync.RWMutex
	upgrader     websocket.Upgrader
	maxConns     int
	activeConns  int32
	activeConnMu sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	logger       *slog.Logger
}

type ServerOption func(*Server)

func WithMaxConnections(n int) ServerOption {
	return func(s *Server) {
		s.maxConns = n
	}
}

func WithLogger(logger *slog.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

func NewServer(opts ...ServerOption) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		handlers: make(map[string]*commons.CommHandler),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // You might want to implement a more secure check
			},
		},
		maxConns: 1000, // Default max connections
		ctx:      ctx,
		cancel:   cancel,
		logger:   slog.Default(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Register(name string, handler *commons.CommHandler) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.handlers[name] = handler
}

func (s *Server) Serve(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.wsHandler)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-s.ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Server shutdown error", "error", err)
		}
	}()

	s.logger.Info("WebSocket Server starting", "address", addr)
	return srv.ListenAndServe()
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	s.activeConnMu.Lock()
	if s.activeConns >= int32(s.maxConns) {
		s.activeConnMu.Unlock()
		http.Error(w, "Too many connections", http.StatusServiceUnavailable)
		return
	}
	s.activeConns++
	s.activeConnMu.Unlock()

	defer func() {
		s.activeConnMu.Lock()
		s.activeConns--
		s.activeConnMu.Unlock()
	}()

	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed", "error", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket read error", "error", err)
			}
			break
		}

		if err := s.handleMessage(c, mt, message); err != nil {
			s.logger.Error("Message handling error", "error", err)
			break
		}
	}
}

func (s *Server) handleMessage(c *websocket.Conn, mt int, message []byte) error {
	payload := &contracts.Payload{}
	if err := json.Unmarshal(message, payload); err != nil {
		return s.writeError(c, mt, "Invalid JSON payload")
	}

	s.handlersMu.RLock()
	handler, exists := s.handlers[payload.Route]
	s.handlersMu.RUnlock()

	if !exists {
		return s.writeError(c, mt, "Handler not found")
	}

	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	response, err := handler.Handle(ctx, payload.Params)
	if err != nil {
		return s.writeError(c, mt, err.Error())
	}

	return s.writeJSON(c, mt, response)
}

func (s *Server) writeError(c *websocket.Conn, mt int, message string) error {
	resp := contracts.ResponseFailed(errors.New(message))
	return s.writeJSON(c, mt, resp)
}

func (s *Server) writeJSON(c *websocket.Conn, mt int, v interface{}) error {
	w, err := c.NextWriter(mt)
	if err != nil {
		return err
	}
	defer w.Close()

	return json.NewEncoder(w).Encode(v)
}

func (s *Server) Close() error {
	s.cancel()
	return nil
}
