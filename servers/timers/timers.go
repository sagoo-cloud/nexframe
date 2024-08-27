package timers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// Server 代表定时器服务器
type Server struct {
	handlers map[string]*service // 存储所有注册的服务
	logger   *log.Logger         // 日志记录器
	mu       sync.RWMutex        // 用于保护 handlers 的并发访问
	wg       sync.WaitGroup      // 用于等待所有 goroutine 完成
	ctx      context.Context     // 用于控制所有服务的生命周期
	cancel   context.CancelFunc  // 用于取消 ctx
}

// service 代表单个定时服务
type service struct {
	freq    time.Duration                                                      // 服务执行频率
	handler func(context.Context, map[string]interface{}) (interface{}, error) // 服务处理函数
	params  map[string]interface{}                                             // 服务参数
}

// NewServer 创建一个新的 Server 实例
func NewServer(logger *log.Logger) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		handlers: make(map[string]*service),
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Register 注册一个新的定时服务
func (s *Server) Register(name string, freq time.Duration, handler func(context.Context, map[string]interface{}) (interface{}, error), params map[string]interface{}) error {
	if name == "" || freq <= 0 || handler == nil {
		return errors.New("无效的参数")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.handlers[name]; exists {
		return fmt.Errorf("服务 %s 已经注册", name)
	}

	s.handlers[name] = &service{
		freq:    freq,
		handler: handler,
		params:  params,
	}
	return nil
}

// Run 启动所有注册的服务
func (s *Server) Run() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for name, srv := range s.handlers {
		s.wg.Add(1)
		go s.runService(name, srv)
	}

	return nil
}

// runService 运行单个服务
func (s *Server) runService(name string, srv *service) {
	defer s.wg.Done()
	ticker := time.NewTicker(srv.freq)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			func() {
				ctx, cancel := context.WithTimeout(s.ctx, srv.freq)
				defer cancel()

				resp, err := srv.handler(ctx, srv.params)
				if err != nil {
					s.logger.Printf("服务 %s 发生错误: %v", name, err)
				} else {
					s.logger.Printf("服务 %s 完成: %v", name, resp)
				}
			}()
		}
	}
}

// Close 停止所有服务并等待它们完成
func (s *Server) Close() error {
	s.cancel()
	s.wg.Wait()
	return nil
}
