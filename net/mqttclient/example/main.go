package main

import (
	"context"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"sync"
	"time"

	mqtt "github.com/sagoo-cloud/nexframe/net/mqttclient"
)

// CustomLogger 实现自定义日志记录器
type CustomLogger struct {
	logger *log.Logger
	mu     sync.Mutex
}

// NewCustomLogger 创建新的日志记录器
func NewCustomLogger() *CustomLogger {
	return &CustomLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *CustomLogger) Debug(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[DEBUG] "+format, v...)
}

func (l *CustomLogger) Info(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[INFO] "+format, v...)
}

func (l *CustomLogger) Warn(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[WARN] "+format, v...)
}

func (l *CustomLogger) Error(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[ERROR] "+format, v...)
}

// ExampleUsage 展示MQTT客户端的基本使用
func ExampleUsage() {
	// 创建配置
	conf := mqtt.Config{
		Server:    "tcp://localhost:1883",
		Username:  "user",
		Password:  "pass",
		Logger:    NewCustomLogger(),
		LogLevel:  mqtt.LogLevelInfo, // 设置日志级别为INFO
		QueueSize: 100,               // 设置消息队列大小
	}

	// 创建上下文
	ctx := context.Background()

	// 创建MQTT客户端
	client, err := mqtt.NewClient(ctx, conf)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// 创建消息处理器
	handler := mqtt.Handler{
		Topic: "test/topic",
		Qos:   1,
		Handle: func(c paho.Client, msg paho.Message) {
			fmt.Printf("Received message on topic %s: %s\n", msg.Topic(), string(msg.Payload()))
		},
	}

	// 注册处理器
	if err := client.RegisterHandler(handler); err != nil {
		panic(err)
	}

	// 发布消息
	if err := client.Publish("test/topic", 1, []byte("Hello, MQTT!")); err != nil {
		panic(err)
	}

	// 等待一段时间以接收消息
	time.Sleep(time.Second * 5)
}

// ExampleWithTLS 展示带TLS配置的使用
func ExampleWithTLS() {
	conf := mqtt.Config{
		Server:      "ssl://localhost:8883",
		CAFile:      "/path/to/ca.crt",
		CertFile:    "/path/to/client.crt",
		CertKeyFile: "/path/to/client.key",
		Logger:      NewCustomLogger(),
		LogLevel:    mqtt.LogLevelDebug, // 设置更详细的日志级别
	}

	ctx := context.Background()
	client, err := mqtt.NewClient(ctx, conf)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// ... 其他操作
}

// ExampleWithReconnect 展示自动重连功能
func ExampleWithReconnect() {
	conf := mqtt.Config{
		Server:               "tcp://localhost:1883",
		MaxReconnectInterval: time.Second * 5, // 设置重连间隔
		Logger:               NewCustomLogger(),
		LogLevel:             mqtt.LogLevelWarn, // 只记录警告和错误
	}

	ctx := context.Background()
	client, err := mqtt.NewClient(ctx, conf)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// 监控连接状态
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !client.IsConnected() {
					fmt.Println("Client disconnected, waiting for reconnect...")
				}
			}
		}
	}()

	// ... 其他操作
}

// ExampleWithContext 展示使用上下文控制
func ExampleWithContext() {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	conf := mqtt.Config{
		Server:   "tcp://localhost:1883",
		Logger:   NewCustomLogger(),
		LogLevel: mqtt.LogLevelInfo,
	}

	client, err := mqtt.NewClient(ctx, conf)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// 使用上下文控制操作超时
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Operation timed out")
		}
	}()

	// ... 其他操作
}

func main() {
	// 运行示例
	ExampleUsage()
}
