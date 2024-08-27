package rocketqueue

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/sagoo-cloud/nexframe/servers/queue"
)

var (
	queuesMu  sync.RWMutex
	queues    = make(map[string]queue.Queue)
	clientsMu sync.RWMutex
	clients   = make(map[string]*RocketMQClient)
)

// RocketMQClient 封装了 RocketMQ 的生产者和消费者
type RocketMQClient struct {
	Producer rocketmq.Producer
	Consumer rocketmq.PushConsumer
}

// RocketQueue 实现了基于RocketMQ的队列
type RocketQueue struct {
	client       *RocketMQClient
	msgChan      chan *primitive.MessageExt
	stopChan     chan struct{}
	producerOnce sync.Once
	consumerOnce sync.Once
}

// getRocketMq 获取或创建 RocketMQ 客户端
func getRocketMq(diName string) *RocketMQClient {
	clientsMu.RLock()
	client, ok := clients[diName]
	clientsMu.RUnlock()
	if ok {
		return client
	}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	// 双重检查锁定
	if client, ok = clients[diName]; ok {
		return client
	}

	// 创建新的 RocketMQ 客户端
	producerInstance, err := producer.NewDefaultProducer(
		producer.WithNameServer([]string{"127.0.0.1:9876"}), // 使用实际的 NameServer 地址
		producer.WithRetry(2),
		producer.WithGroupName("ProducerGroupName"), // 使用实际的生产者组名
	)
	if err != nil {
		slog.Error("failed to create producer", "error", err)
		return nil
	}

	consumerInstance, err := consumer.NewPushConsumer(
		consumer.WithNameServer([]string{"127.0.0.1:9876"}), // 使用实际的 NameServer 地址
		consumer.WithGroupName("ConsumerGroupName"),         // 使用实际的消费者组名
	)
	if err != nil {
		slog.Error("failed to create consumer", "error", err)
		return nil
	}

	client = &RocketMQClient{
		Producer: producerInstance,
		Consumer: consumerInstance,
	}

	clients[diName] = client
	return client
}

// newRocketQueue 创建一个新的 RocketQueue 实例
func newRocketQueue(diName string) queue.Queue {
	client := getRocketMq(diName)
	if client == nil {
		return nil
	}
	m := &RocketQueue{
		client:   client,
		msgChan:  make(chan *primitive.MessageExt, 100),
		stopChan: make(chan struct{}),
	}
	return m
}

// GetRocketQueue 获取 RocketQueue 实例（单例模式）
func GetRocketQueue(diName string) queue.Queue {
	queuesMu.RLock()
	q, ok := queues[diName]
	queuesMu.RUnlock()
	if ok {
		return q
	}

	q = newRocketQueue(diName)
	if q == nil {
		return nil
	}
	queuesMu.Lock()
	queues[diName] = q
	queuesMu.Unlock()
	return q
}

// Enqueue 实现了 Queue 接口的 Enqueue 方法
func (m *RocketQueue) Enqueue(ctx context.Context, key string, message string) (bool, error) {
	err := m.initProducer(ctx)
	if err != nil {
		return false, err
	}

	msg := &primitive.Message{
		Topic: key,
		Body:  []byte(message),
	}

	res, err := m.client.Producer.SendSync(ctx, msg)
	if err != nil {
		return false, err
	}

	slog.Info("Enqueue", "message", message, "msgID", res.MsgID)
	return true, nil
}

// Dequeue 实现了 Queue 接口的 Dequeue 方法
func (m *RocketQueue) Dequeue(ctx context.Context, key string) (message string, tag string, token string, dequeueCount int64, err error) {
	err = m.initConsumer(ctx, key, "", 5)
	if err != nil {
		return
	}

	select {
	case msg, ok := <-m.msgChan:
		if !ok {
			return "", "", "", 0, nil
		}
		return string(msg.Body), msg.GetTags(), "", int64(msg.ReconsumeTimes), nil
	case <-ctx.Done():
		return "", "", "", 0, ctx.Err()
	}
}

// AckMsg 实现了 Queue 接口的 AckMsg 方法
func (m *RocketQueue) AckMsg(ctx context.Context, key string, token string) (bool, error) {
	// RocketMQ 自动确认消息，这里不需要实现
	return true, nil
}

// BatchEnqueue 实现了 Queue 接口的 BatchEnqueue 方法
func (m *RocketQueue) BatchEnqueue(ctx context.Context, key string, messages []string) (bool, error) {
	if len(messages) == 0 {
		return false, errors.New("messages is empty")
	}

	for _, message := range messages {
		ok, err := m.Enqueue(ctx, key, message)
		if !ok || err != nil {
			return false, err
		}
	}

	return true, nil
}

// initProducer 初始化生产者
func (m *RocketQueue) initProducer(ctx context.Context) error {
	var err error
	m.producerOnce.Do(func() {
		err = m.client.Producer.Start()
		if err != nil {
			slog.Error("RocketQueue:Producer:Start failed", "error", err)
		}
	})
	return err
}

// initConsumer 初始化消费者
func (m *RocketQueue) initConsumer(ctx context.Context, topic, messageTag string, num int) error {
	var err error
	m.consumerOnce.Do(func() {
		var selector consumer.MessageSelector
		if messageTag != "" {
			selector = consumer.MessageSelector{
				Type:       consumer.TAG,
				Expression: messageTag,
			}
		}

		err = m.client.Consumer.Subscribe(topic, selector, func(ctx context.Context, messages ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, msg := range messages {
				select {
				case m.msgChan <- msg:
				case <-ctx.Done():
					return consumer.ConsumeRetryLater, ctx.Err()
				}
			}
			return consumer.ConsumeSuccess, nil
		})

		if err != nil {
			slog.Error("RocketQueue:Subscribe failed", "error", err)
			return
		}

		err = m.client.Consumer.Start()
		if err != nil {
			slog.Error("RocketQueue:Start failed", "error", err)
			return
		}

		go m.handleSignals()
	})

	return err
}

// handleSignals 处理系统信号
func (m *RocketQueue) handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			m.shutdown()
			return
		case syscall.SIGHUP:
		default:
		}
	}
}

// shutdown 关闭队列
func (m *RocketQueue) shutdown() {
	close(m.stopChan)
	close(m.msgChan)
	_ = m.client.Consumer.Shutdown()
	_ = m.client.Producer.Shutdown()
}

func init() {
	queue.Register(queue.DriverTypeRocketMq, GetRocketQueue)
}
