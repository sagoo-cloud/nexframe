package signals

import (
	"context"
	"errors"
	"sync"
)

// SyncSignal 是实现 Signal 接口的结构体。
// 它提供了一种同步方式来通知所有信号订阅者。
// 类型参数 `T` 是任意类型的占位符。
type SyncSignal[T any] struct {
	BaseSignal[T]
	mu sync.RWMutex // 添加读写锁以保证并发安全
}

// Emit 以同步方式通知所有信号订阅者并传递负载。
//
// 负载的类型与 SyncSignal 的类型参数 `T` 相同。
// 该方法遍历 SyncSignal 的订阅者切片，
// 对于每个订阅者，它调用订阅者的监听器函数，
// 传递上下文和负载。
// 如果上下文有截止日期或可取消属性，监听器必须遵守它。
// 这意味着当上下文被取消时，监听器应停止处理。
// 与 AsyncSignal 的 Emit 方法不同，此方法不会在单独的 goroutine 中调用监听器，
// 因此监听器是同步调用的，一个接一个。
//
// 示例:
//
//	signal := signals.NewSync[string]()
//	signal.AddListener(func(ctx context.Context, payload string) {
//		// 监听器实现
//		// ...
//	})
//
//	err := signal.Emit(context.Background(), "Hello, world!")
//	if err != nil {
//		// 处理错误
//	}
func (s *SyncSignal[T]) Emit(ctx context.Context, payload T) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var emitErr error
	for _, sub := range s.subscribers {
		if sub.listener != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						emitErr = errors.New("listener panicked")
					}
				}()
				select {
				case <-ctx.Done():
					emitErr = ctx.Err()
					return
				default:
					sub.listener(ctx, payload)
				}
			}()
			if emitErr != nil {
				break
			}
		}
	}

	if emitErr != nil {
		return emitErr
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// AddListener 添加了并发安全的实现
func (s *SyncSignal[T]) AddListener(listener SignalListener[T], key ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.BaseSignal.AddListener(listener, key...)
}

// RemoveListener 添加了并发安全的实现
func (s *SyncSignal[T]) RemoveListener(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.BaseSignal.RemoveListener(key)
}

// Reset 添加了并发安全的实现
func (s *SyncSignal[T]) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BaseSignal.Reset()
}

// Len 添加了并发安全的实现
func (s *SyncSignal[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BaseSignal.Len()
}

// IsEmpty 添加了并发安全的实现
func (s *SyncSignal[T]) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BaseSignal.IsEmpty()
}
