package redisqueue

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"sync"

	"github.com/sagoo-cloud/nexframe/queue"
)

var (
	queuesMu sync.RWMutex
	queues   = make(map[string]queue.Queue)
)

// RedisQueue 实现了基于Redis的队列
type RedisQueue struct {
	client redis.UniversalClient
}

// newRedisQueue 创建一个新的 RedisQueue 实例
func newRedisQueue(diName string) queue.Queue {
	client := redisdb.DB().GetClient() // 假设这个函数存在并返回正确的 Redis 客户端
	return &RedisQueue{client: client}
}

// GetRedisQueue 获取 RedisQueue 实例（单例模式）
func GetRedisQueue(diName string) queue.Queue {
	queuesMu.RLock()
	q, ok := queues[diName]
	queuesMu.RUnlock()
	if ok {
		return q
	}

	q = newRedisQueue(diName)
	queuesMu.Lock()
	queues[diName] = q
	queuesMu.Unlock()
	return q
}

// Enqueue 实现了 Queue 接口的 Enqueue 方法
func (m *RedisQueue) Enqueue(ctx context.Context, key string, message string) (bool, error) {
	_, err := m.client.RPush(ctx, key, message).Result()
	return err == nil, err
}

// Dequeue 实现了 Queue 接口的 Dequeue 方法
func (m *RedisQueue) Dequeue(ctx context.Context, key string) (message string, tag string, token string, dequeueCount int64, err error) {
	message, err = m.client.LPop(ctx, key).Result()
	if err == redis.Nil {
		err = nil
		message = ""
	}
	return message, "", "", 1, err // Redis 不支持 tag, token, 出队次数固定为1
}

// AckMsg 实现了 Queue 接口的 AckMsg 方法
func (m *RedisQueue) AckMsg(ctx context.Context, key string, token string) (bool, error) {
	return true, nil // Redis 不需要确认消息
}

// BatchEnqueue 实现了 Queue 接口的 BatchEnqueue 方法
func (m *RedisQueue) BatchEnqueue(ctx context.Context, key string, messages []string) (bool, error) {
	if len(messages) == 0 {
		return false, errors.New("messages is empty")
	}
	_, err := m.client.RPush(ctx, key, stringSliceToInterface(messages)...).Result()
	return err == nil, err
}

// stringSliceToInterface 将字符串切片转换为空接口切片
func stringSliceToInterface(arr []string) []interface{} {
	result := make([]interface{}, len(arr))
	for i, v := range arr {
		result[i] = v
	}
	return result
}

func init() {
	queue.Register(queue.DriverTypeRedis, GetRedisQueue)
}
