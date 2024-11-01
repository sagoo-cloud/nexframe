package mqttclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	// LogLevelNone 不输出日志
	LogLevelNone LogLevel = iota
	// LogLevelError 只输出错误日志
	LogLevelError
	// LogLevelWarn 输出警告和错误日志
	LogLevelWarn
	// LogLevelInfo 输出信息、警告和错误日志
	LogLevelInfo
	// LogLevelDebug 输出所有日志
	LogLevelDebug
)

// LogLevelToInt 将 LogLevel 枚举转换为 int
func LogLevelToInt(level LogLevel) int {
	return int(level)
}

// IntToLogLevel 将 int 转换为 LogLevel 枚举
func IntToLogLevel(level int) LogLevel {
	switch level {
	case 0:
		return LogLevelNone
	case 1:
		return LogLevelError
	case 2:
		return LogLevelWarn
	case 3:
		return LogLevelInfo
	case 4:
		return LogLevelDebug
	default:
		// 如果给定的 int 值不匹配任何 LogLevel 枚举，可以选择默认值或者返回错误
		return LogLevelNone
	}
}

// Logger 定义日志接口
type Logger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

// defaultLogger 实现默认的空日志记录器
type defaultLogger struct{}

func (l *defaultLogger) Debug(format string, v ...interface{}) {}
func (l *defaultLogger) Info(format string, v ...interface{})  {}
func (l *defaultLogger) Warn(format string, v ...interface{})  {}
func (l *defaultLogger) Error(format string, v ...interface{}) {}

// 定义常量和错误
const (
	DefaultMaxReconnectInterval = time.Second * 60
	DefaultQueueSize            = 100
	randomChars                 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	ErrNilHandler         = errors.New("handler cannot be nil")
	ErrClientClosed       = errors.New("client is closed")
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrSubscriptionFailed = errors.New("subscription failed")
	ErrConnectionFailed   = errors.New("connection failed")
	ErrPublishFailed      = errors.New("publish failed")
	ErrUnsubscribeFailed  = errors.New("unsubscribe failed")
)

// Handler 定义订阅数据处理器
type Handler struct {
	Topic  string              // 订阅主题
	Qos    byte                // 订阅QoS级别
	Handle paho.MessageHandler // 消息处理函数
}

// Config 定义客户端配置
type Config struct {
	Server               string        // MQTT broker地址
	Username             string        // 用户名
	Password             string        // 密码
	MaxReconnectInterval time.Duration // 重连间隔
	QOS                  uint8         // 服务质量等级
	CleanSession         bool          // 清理会话标志
	ClientID             string        // 客户端ID
	CAFile               string        // CA证书文件
	CertFile             string        // 客户端证书
	CertKeyFile          string        // 客户端密钥
	Logger               Logger        // 日志记录器
	LogLevel             LogLevel      // 日志级别
	QueueSize            int           // 消息队列大小
}

// Client 实现MQTT客户端
type Client struct {
	sync.RWMutex
	client        paho.Client
	msgHandlerMap map[string]Handler
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	closed        bool
	logger        Logger
	logLevel      LogLevel
	config        Config
}

// NewClient 创建新的MQTT客户端实例
func NewClient(ctx context.Context, conf Config) (*Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := validateConfig(&conf); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	ctx, cancel := context.WithCancel(ctx)
	client := &Client{
		msgHandlerMap: make(map[string]Handler),
		ctx:           ctx,
		cancel:        cancel,
		logger:        conf.Logger,
		logLevel:      conf.LogLevel,
		config:        conf,
	}

	// 设置默认logger
	if client.logger == nil {
		client.logger = &defaultLogger{}
	}

	// 设置客户端选项
	opts := paho.NewClientOptions()
	opts.AddBroker(conf.Server)
	opts.SetUsername(conf.Username)
	opts.SetPassword(conf.Password)
	opts.SetCleanSession(conf.CleanSession)
	opts.SetClientID(client.getClientID(conf.ClientID))
	opts.SetOnConnectHandler(client.onConnected)
	opts.SetConnectionLostHandler(client.onConnectionLost)
	opts.SetMaxReconnectInterval(client.getReconnectInterval(conf.MaxReconnectInterval))
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(30 * time.Second)

	if conf.QueueSize > 0 {
		opts.SetMessageChannelDepth(uint(conf.QueueSize))
	} else {
		opts.SetMessageChannelDepth(DefaultQueueSize)
	}

	// 配置TLS
	if tlsConfig, err := newTLSConfig(conf.CAFile, conf.CertFile, conf.CertKeyFile); err != nil {
		client.log(LogLevelError, "TLS configuration failed: %v", err)
		return nil, fmt.Errorf("tls config error: %w", err)
	} else if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}

	client.client = paho.NewClient(opts)

	// 建立连接
	if err := client.connect(ctx); err != nil {
		cancel()
		return nil, err
	}

	client.log(LogLevelInfo, "MQTT client initialized successfully")
	return client, nil
}

// GetClient 获取底层的paho.Client
func (c *Client) GetClient() paho.Client {
	return c.client
}

// RegisterHandler 注册消息处理器
func (c *Client) RegisterHandler(handler Handler) error {
	if handler.Handle == nil {
		return ErrNilHandler
	}

	c.Lock()
	defer c.Unlock()

	if c.closed {
		return ErrClientClosed
	}

	c.log(LogLevelDebug, "Registering handler for topic: %s", handler.Topic)

	if err := c.subscribeHandler(handler); err != nil {
		return err
	}

	c.msgHandlerMap[handler.Topic] = handler
	c.log(LogLevelInfo, "Handler registered successfully for topic: %s", handler.Topic)
	return nil
}

// UnregisterHandler 注销消息处理器
func (c *Client) UnregisterHandler(topic string) error {
	c.Lock()
	defer c.Unlock()

	if c.closed {
		return ErrClientClosed
	}

	c.log(LogLevelDebug, "Unregistering handler for topic: %s", topic)

	if token := c.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		c.log(LogLevelError, "Failed to unsubscribe from topic %s: %v", topic, token.Error())
		return fmt.Errorf("%w: %v", ErrUnsubscribeFailed, token.Error())
	}

	delete(c.msgHandlerMap, topic)
	c.log(LogLevelInfo, "Handler unregistered successfully for topic: %s", topic)
	return nil
}

// GetHandlerByTopic 获取指定主题的处理器
func (c *Client) GetHandlerByTopic(topic string) (Handler, bool) {
	c.RLock()
	defer c.RUnlock()
	handler, exists := c.msgHandlerMap[topic]
	return handler, exists
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	c.Lock()
	if c.closed {
		c.Unlock()
		return nil
	}
	c.closed = true
	c.Unlock()

	c.log(LogLevelInfo, "Closing MQTT client")

	// 取消所有订阅
	for topic := range c.msgHandlerMap {
		if token := c.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			c.log(LogLevelWarn, "Failed to unsubscribe topic %s: %v", topic, token.Error())
		}
	}

	// 清理资源
	c.cancel()
	c.client.Disconnect(1000)
	c.wg.Wait()

	c.log(LogLevelInfo, "MQTT client closed successfully")
	return nil
}

// Publish 发布消息
func (c *Client) Publish(topic string, qos byte, data []byte) error {
	if c.closed {
		return ErrClientClosed
	}

	c.log(LogLevelDebug, "Publishing message to topic: %s", topic)

	token := c.client.Publish(topic, qos, false, data)
	if token.Wait() && token.Error() != nil {
		c.log(LogLevelError, "Failed to publish to topic %s: %v", topic, token.Error())
		return fmt.Errorf("%w: %v", ErrPublishFailed, token.Error())
	}

	c.log(LogLevelDebug, "Message published successfully to topic: %s", topic)
	return nil
}

// SetLogger 设置新的日志记录器
func (c *Client) SetLogger(logger Logger) {
	c.Lock()
	defer c.Unlock()
	if logger != nil {
		c.logger = logger
	}
}

// SetLogLevel 设置日志级别
func (c *Client) SetLogLevel(level LogLevel) {
	c.Lock()
	defer c.Unlock()
	c.logLevel = level
}

// IsConnected 检查客户端是否已连接
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

// 内部方法

func (c *Client) log(level LogLevel, format string, v ...interface{}) {
	if c.logLevel >= level {
		switch level {
		case LogLevelDebug:
			c.logger.Debug(format, v...)
		case LogLevelInfo:
			c.logger.Info(format, v...)
		case LogLevelWarn:
			c.logger.Warn(format, v...)
		case LogLevelError:
			c.logger.Error(format, v...)
		}
	}
}

func (c *Client) onConnected(client paho.Client) {
	c.RLock()
	defer c.RUnlock()

	c.log(LogLevelInfo, "Connected to MQTT broker")

	// 重新订阅所有主题
	for _, handler := range c.msgHandlerMap {
		if err := c.subscribeHandler(handler); err != nil {
			c.log(LogLevelError, "Failed to resubscribe topic %s: %v", handler.Topic, err)
		}
	}
}

func (c *Client) onConnectionLost(client paho.Client, err error) {
	c.log(LogLevelWarn, "Connection lost: %v", err)
}

func (c *Client) connect(ctx context.Context) error {
	for {
		token := c.client.Connect()
		if token.Wait() && token.Error() != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
				c.log(LogLevelWarn, "Connection failed, retrying: %v", token.Error())
				continue
			}
		}
		c.log(LogLevelInfo, "Successfully connected to MQTT broker")
		return nil
	}
}

func (c *Client) subscribeHandler(handler Handler) error {
	token := c.client.Subscribe(handler.Topic, handler.Qos, handler.Handle)
	if token.Wait() && token.Error() != nil {
		c.log(LogLevelError, "Subscribe failed for topic %s: %v", handler.Topic, token.Error())
		return fmt.Errorf("%w: %v", ErrSubscriptionFailed, token.Error())
	}

	if result, ok := token.(*paho.SubscribeToken).Result()[handler.Topic]; ok && result == 128 {
		c.log(LogLevelError, "ACL error for topic %s", handler.Topic)
		return fmt.Errorf("ACL error for topic %s", handler.Topic)
	}

	c.log(LogLevelDebug, "Successfully subscribed to topic %s", handler.Topic)
	return nil
}

// 工具方法

func (c *Client) getClientID(configID string) string {
	if configID != "" {
		return configID
	}
	return fmt.Sprintf("rulego/%s", randomString(8))
}

func (c *Client) getReconnectInterval(configInterval time.Duration) time.Duration {
	if configInterval <= 0 {
		return DefaultMaxReconnectInterval
	}
	return configInterval
}

func validateConfig(conf *Config) error {
	if conf.Server == "" {
		return errors.New("server address is required")
	}

	if conf.QueueSize < 0 {
		conf.QueueSize = DefaultQueueSize
	}

	return nil
}

func newTLSConfig(caFile, certFile, certKeyFile string) (*tls.Config, error) {
	if caFile == "" && certFile == "" && certKeyFile == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if caFile != "" {
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file: %w", err)
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to append CA certificate")
		}
		tlsConfig.RootCAs = certPool
	}

	if certFile != "" && certKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, certKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = randomChars[rand.Intn(len(randomChars))]
	}
	return string(b)
}
