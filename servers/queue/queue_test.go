package queue_test

import (
	"context"
	"fmt"
	"github.com/sagoo-cloud/nexframe/servers/queue"
	"github.com/sagoo-cloud/nexframe/servers/queue/redisqueue"
	"github.com/sagoo-cloud/nexframe/servers/queue/rocketqueue"
	"sync"
	"testing"
)

type QueueTestSuite struct {
	q queue.Queue
}

func (suite *QueueTestSuite) TestEnqueueDequeue(t *testing.T) {
	ctx := context.Background()
	key := "test_queue"
	message := "test_message"

	// 测试入队
	ok, err := suite.q.Enqueue(ctx, key, message)
	if err != nil {
		t.Errorf("Enqueue failed: %v", err)
	}
	if !ok {
		t.Error("Enqueue returned false")
	}

	// 测试出队
	receivedMsg, _, token, dequeueCount, err := suite.q.Dequeue(ctx, key)
	if err != nil {
		t.Errorf("Dequeue failed: %v", err)
	}
	if receivedMsg != message {
		t.Errorf("Expected message %s, got %s", message, receivedMsg)
	}
	if dequeueCount != 1 {
		t.Errorf("Expected dequeue count 1, got %d", dequeueCount)
	}

	// 确认消息
	ok, err = suite.q.AckMsg(ctx, key, token)
	if err != nil {
		t.Errorf("AckMsg failed: %v", err)
	}
	if !ok {
		t.Error("AckMsg returned false")
	}

	// 测试空队列
	_, _, _, _, err = suite.q.Dequeue(ctx, key)
	if err == nil {
		t.Error("Expected error when dequeueing from empty queue, got nil")
	}
}

func (suite *QueueTestSuite) TestBatchEnqueue(t *testing.T) {
	ctx := context.Background()
	key := "test_batch_queue"
	messages := []string{"msg1", "msg2", "msg3"}

	// 测试批量入队
	ok, err := suite.q.BatchEnqueue(ctx, key, messages)
	if err != nil {
		t.Errorf("BatchEnqueue failed: %v", err)
	}
	if !ok {
		t.Error("BatchEnqueue returned false")
	}

	// 验证批量入队的消息
	for _, expectedMsg := range messages {
		receivedMsg, _, _, _, err := suite.q.Dequeue(ctx, key)
		if err != nil {
			t.Errorf("Dequeue failed: %v", err)
		}
		if receivedMsg != expectedMsg {
			t.Errorf("Expected message %s, got %s", expectedMsg, receivedMsg)
		}
	}

	// 确保队列为空
	_, _, _, _, err = suite.q.Dequeue(ctx, key)
	if err == nil {
		t.Error("Expected error when dequeueing from empty queue, got nil")
	}
}

func TestRedisQueue(t *testing.T) {
	redisQueue := redisqueue.GetRedisQueue("test")
	suite := &QueueTestSuite{q: redisQueue}

	t.Run("EnqueueDequeue", func(t *testing.T) {
		suite.TestEnqueueDequeue(t)
	})

	t.Run("BatchEnqueue", func(t *testing.T) {
		suite.TestBatchEnqueue(t)
	})
}

func TestRocketMQQueue(t *testing.T) {
	rocketQueue := rocketqueue.GetRocketQueue("test")
	suite := &QueueTestSuite{q: rocketQueue}

	t.Run("EnqueueDequeue", func(t *testing.T) {
		suite.TestEnqueueDequeue(t)
	})

	t.Run("BatchEnqueue", func(t *testing.T) {
		suite.TestBatchEnqueue(t)
	})
}

func TestConcurrency(t *testing.T) {
	redisQueue := redisqueue.GetRedisQueue("test")
	ctx := context.Background()
	key := "test_concurrency_queue"

	// 并发入队
	concurrency := 100
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := redisQueue.Enqueue(ctx, key, fmt.Sprintf("msg%d", i))
			if err != nil {
				t.Errorf("Concurrent Enqueue failed: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// 并发出队
	receivedMsgs := make(map[string]bool)
	var mu sync.Mutex
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg, _, _, _, err := redisQueue.Dequeue(ctx, key)
			if err != nil {
				t.Errorf("Concurrent Dequeue failed: %v", err)
			}
			mu.Lock()
			receivedMsgs[msg] = true
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(receivedMsgs) != concurrency {
		t.Errorf("Expected %d unique messages, got %d", concurrency, len(receivedMsgs))
	}
}
