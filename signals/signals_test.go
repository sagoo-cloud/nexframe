package signals

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestSyncSignal 测试同步信号的功能
func TestSyncSignal(t *testing.T) {
	// 测试添加监听器并向其发送信号
	t.Run("Add and emit to listeners", func(t *testing.T) {
		signal := NewSync[int]()
		result := make(chan int, 2)

		// 添加第一个监听器，直接返回接收到的值
		signal.AddListener(func(ctx context.Context, payload int) {
			result <- payload
		})

		// 添加第二个监听器，返回接收到的值的两倍
		signal.AddListener(func(ctx context.Context, payload int) {
			result <- payload * 2
		})

		// 发送信号
		err := signal.Emit(context.Background(), 5)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// 验证结果数量
		if len(result) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(result))
		}

		// 验证第一个结果
		if <-result != 5 {
			t.Errorf("Expected 5, got %d", <-result)
		}

		// 验证第二个结果
		if <-result != 10 {
			t.Errorf("Expected 10, got %d", <-result)
		}
	})

	// 测试移除监听器
	t.Run("Remove listener", func(t *testing.T) {
		signal := NewSync[int]()
		result := make(chan int, 2)

		// 添加第一个带键的监听器
		key := "testKey"
		signal.AddListener(func(ctx context.Context, payload int) {
			result <- payload
		}, key)

		// 添加第二个不带键的监听器
		signal.AddListener(func(ctx context.Context, payload int) {
			result <- payload * 2
		})

		// 移除第一个监听器
		signal.RemoveListener(key)

		// 发送信号
		err := signal.Emit(context.Background(), 5)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// 验证结果数量
		if len(result) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result))
		}

		// 验证结果
		if <-result != 10 {
			t.Errorf("Expected 10, got %d", <-result)
		}
	})

	// 测试重置信号
	t.Run("Reset signal", func(t *testing.T) {
		signal := NewSync[int]()
		// 添加两个监听器
		signal.AddListener(func(ctx context.Context, payload int) {})
		signal.AddListener(func(ctx context.Context, payload int) {})

		// 重置信号
		signal.Reset()

		// 验证信号是否为空
		if !signal.IsEmpty() {
			t.Errorf("Expected signal to be empty after reset")
		}
	})
}

// TestAsyncSignal 测试异步信号的功能
func TestAsyncSignal(t *testing.T) {
	// 测试添加监听器并向其发送信号
	t.Run("Add and emit to listeners", func(t *testing.T) {
		signal := New[int]()
		result := make(chan int, 2)
		var wg sync.WaitGroup
		wg.Add(2)

		// 添加第一个监听器
		signal.AddListener(func(ctx context.Context, payload int) {
			defer wg.Done()
			result <- payload
		})

		// 添加第二个监听器
		signal.AddListener(func(ctx context.Context, payload int) {
			defer wg.Done()
			result <- payload * 2
		})

		// 发送信号
		err := signal.Emit(context.Background(), 5)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// 等待所有监听器完成
		wg.Wait()

		// 验证结果数量
		if len(result) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(result))
		}

		// 验证结果，考虑到异步执行可能导致的顺序不确定性
		results := []int{<-result, <-result}
		if !(results[0] == 5 && results[1] == 10) && !(results[0] == 10 && results[1] == 5) {
			t.Errorf("Expected 5 and 10 in any order, got %v", results)
		}
	})

	// 测试上下文取消
	t.Run("Context cancellation", func(t *testing.T) {
		signal := New[int]()
		result := make(chan int, 1)

		// 添加一个带有延迟的监听器
		signal.AddListener(func(ctx context.Context, payload int) {
			select {
			case <-ctx.Done():
				result <- -1
			case <-time.After(200 * time.Millisecond):
				result <- payload
			}
		})

		// 创建一个可取消的上下文
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel() // 在短暂延迟后取消上下文
		}()

		// 发送信号
		err := signal.Emit(ctx, 5)
		if err == nil {
			t.Fatalf("Expected error due to cancelled context, got nil")
		}

		// 验证结果
		select {
		case r := <-result:
			if r != -1 {
				t.Errorf("Expected -1 (cancelled), got %d", r)
			}
		case <-time.After(300 * time.Millisecond):
			t.Errorf("Listener didn't respond in time")
		}
	})
}

// TestSignalInterface 测试信号接口的通用功能
func TestSignalInterface(t *testing.T) {
	testCases := []struct {
		name   string
		signal func() Signal[int]
	}{
		{"SyncSignal", NewSync[int]},
		{"AsyncSignal", New[int]},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signal := tc.signal()

			// 测试初始状态
			if signal.Len() != 0 {
				t.Errorf("Expected initial length 0, got %d", signal.Len())
			}

			if !signal.IsEmpty() {
				t.Errorf("Expected signal to be initially empty")
			}

			// 测试添加监听器
			signal.AddListener(func(ctx context.Context, payload int) {})

			if signal.Len() != 1 {
				t.Errorf("Expected length 1 after adding listener, got %d", signal.Len())
			}

			if signal.IsEmpty() {
				t.Errorf("Expected signal not to be empty after adding listener")
			}

			// 测试重置
			signal.Reset()

			if !signal.IsEmpty() {
				t.Errorf("Expected signal to be empty after reset")
			}
		})
	}
}

// TestErrorHandling 测试错误处理机制
func TestErrorHandling(t *testing.T) {
	testCases := []struct {
		name   string
		signal func() Signal[int]
	}{
		{"SyncSignal", NewSync[int]},
		{"AsyncSignal", New[int]},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signal := tc.signal()

			// 添加一个会引发 panic 的监听器
			signal.AddListener(func(ctx context.Context, payload int) {
				fmt.Printf("Listener for %s about to panic\n", tc.name)
				panic("Listener panic")
			})

			// 发送信号并检查是否正确处理了 panic
			fmt.Printf("Emitting signal for %s\n", tc.name)
			err := signal.Emit(context.Background(), 5)
			fmt.Printf("Emit for %s completed with error: %v\n", tc.name, err)

			if err == nil {
				t.Errorf("Expected error due to panic in listener, got nil")
			} else if err.Error() != "listener panicked" {
				t.Errorf("Expected 'listener panicked' error, got: %v", err)
			}
		})
	}
}
