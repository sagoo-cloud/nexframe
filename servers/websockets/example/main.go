package main

import (
	"context"
	"fmt"
	"github.com/sagoo-cloud/nexframe/servers/commons"
	"github.com/sagoo-cloud/nexframe/servers/websockets"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// EchoHandler 实现了 commons.Handler 接口
type EchoHandler struct{}

func (h *EchoHandler) ServeHandle(ctx context.Context, request interface{}) (interface{}, error) {
	// 简单地将接收到的消息作为响应返回
	return map[string]interface{}{"echo": request}, nil
}

func main() {
	// 创建日志记录器
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// 创建 WebSocket 服务器
	server := websockets.NewServer(
		websockets.WithLogger(logger),
		websockets.WithMaxConnections(100),
	)

	// 注册 echo 处理程序
	server.Register("echo", &commons.CommHandler{Handler: &EchoHandler{}})

	// 创建一个通道来接收操作系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 在一个新的 goroutine 中启动服务器
	go func() {
		if err := server.Serve(":8080"); err != nil {
			logger.Error("Server error", "error", err)
		}
	}()

	fmt.Println("WebSocket server is running on http://localhost:8080")
	fmt.Println("Use Ctrl+C to stop the server")

	// 等待中断信号
	<-sigChan

	// 关闭服务器
	if err := server.Close(); err != nil {
		logger.Error("Error closing server", "error", err)
	}

	fmt.Println("Server stopped")
}
