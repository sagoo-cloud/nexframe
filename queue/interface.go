package queue

import "context"

// Queue 定义队列驱动接口，所有队列驱动都需要实现以下接口
type Queue interface {
	// Enqueue 单消息入队
	// key: 队列名称
	// message: 消息内容
	// 返回值：
	// - bool: 是否成功
	// - error: 错误信息
	Enqueue(ctx context.Context, key string, message string) (bool, error)

	// Dequeue 单消息出队
	// key: 队列名称
	// 返回值：
	// - message: 消息内容，消息不存在时返回空字符串
	// - tag: 消息标签
	// - token: 消息唯一标识，用于确认消息接收
	// - dequeueCount: 出队消费次数
	// - error: 错误信息
	Dequeue(ctx context.Context, key string) (message string, tag string, token string, dequeueCount int64, err error)

	// AckMsg 确认接收消息
	// key: 队列名称
	// token: 消息唯一标识
	// 返回值：
	// - bool: 是否成功
	// - error: 错误信息
	AckMsg(ctx context.Context, key string, token string) (bool, error)

	// BatchEnqueue 批量入队
	// key: 队列名称
	// messages: 消息列表
	// 返回值：
	// - bool: 是否成功
	// - error: 错误信息
	BatchEnqueue(ctx context.Context, key string, messages []string) (bool, error)
}
