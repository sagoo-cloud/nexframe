package signals

import (
	"context"
	"errors"
	"sync"
	"time"
)

// AsyncSignal 是实现 Signal 接口的结构体。
// 这是默认实现。它提供与 SyncSignal 相同的功能，
// 但监听器是在单独的 goroutine 中调用的。
// 这意味着所有监听器都是异步调用的。但是，该方法
// 在返回之前等待所有监听器完成。如果你不想
// 等待监听器完成，你可以在单独的 goroutine 中调用 Emit 方法。
type AsyncSignal[T any] struct {
	BaseSignal[T]
	mu sync.Mutex
}

// Emit 以异步方式通知所有信号订阅者并传递负载。
//
// 如果上下文有截止日期或可取消属性，监听器必须遵守它。
// 这意味着当上下文被取消时，监听器应停止处理。
// 在发射时，它在单独的 goroutine 中调用监听器，所以监听器是异步调用的。
// 然而，它在返回之前等待所有监听器完成。如果你不想
// 等待监听器完成，你可以在单独的 goroutine 中调用 Emit 方法。
// 另外，你必须知道 Emit 不保证发射值的类型安全。
//
// 示例:
//
//	signal := signals.New[string]()
//	signal.AddListener(func(ctx context.Context, payload string) {
//		// 监听器实现
//		// ...
//	})
//
//	signal.Emit(context.Background(), "Hello, world!")
func (s *AsyncSignal[T]) Emit(ctx context.Context, payload T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(s.subscribers))

	for _, sub := range s.subscribers {
		if sub.listener == nil {
			continue
		}

		wg.Add(1)
		go func(listener SignalListener[T]) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errChan <- errors.New("listener panicked")
				}
			}()

			listenerCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer cancel()

			done := make(chan struct{})
			go func() {
				listener(listenerCtx, payload)
				close(done)
			}()

			select {
			case <-listenerCtx.Done():
				errChan <- listenerCtx.Err()
			case <-done:
				// 监听器正常完成
			}
		}(sub.listener)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 等待所有监听器完成或上下文取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		if err != nil {
			return err
		}
	case <-time.After(150 * time.Millisecond):
		return errors.New("emit timeout")
	}

	return nil
}

// AddListener 添加了并发安全的实现
func (s *AsyncSignal[T]) AddListener(listener SignalListener[T], key ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.BaseSignal.AddListener(listener, key...)
}

// RemoveListener 添加了并发安全的实现
func (s *AsyncSignal[T]) RemoveListener(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.BaseSignal.RemoveListener(key)
}

// Reset 添加了并发安全的实现
func (s *AsyncSignal[T]) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BaseSignal.Reset()
}

// Len 添加了并发安全的实现
func (s *AsyncSignal[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.BaseSignal.Len()
}

// IsEmpty 添加了并发安全的实现
func (s *AsyncSignal[T]) IsEmpty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.BaseSignal.IsEmpty()
}
