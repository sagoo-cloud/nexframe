package signals

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSyncSignal(t *testing.T) {
	t.Run("Add and emit to listeners", func(t *testing.T) {
		signal := NewSync[int]()
		result := make(chan int, 2)

		signal.AddListener(func(ctx context.Context, payload int) {
			result <- payload
		})

		signal.AddListener(func(ctx context.Context, payload int) {
			result <- payload * 2
		})

		err := signal.Emit(context.Background(), 5)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(result))
		}

		if <-result != 5 {
			t.Errorf("Expected 5, got %d", <-result)
		}

		if <-result != 10 {
			t.Errorf("Expected 10, got %d", <-result)
		}
	})

	t.Run("Remove listener", func(t *testing.T) {
		signal := NewSync[int]()
		result := make(chan int, 2)

		key := "testKey"
		signal.AddListener(func(ctx context.Context, payload int) {
			result <- payload
		}, key)

		signal.AddListener(func(ctx context.Context, payload int) {
			result <- payload * 2
		})

		signal.RemoveListener(key)

		err := signal.Emit(context.Background(), 5)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result))
		}

		if <-result != 10 {
			t.Errorf("Expected 10, got %d", <-result)
		}
	})

	t.Run("Reset signal", func(t *testing.T) {
		signal := NewSync[int]()
		signal.AddListener(func(ctx context.Context, payload int) {})
		signal.AddListener(func(ctx context.Context, payload int) {})

		signal.Reset()

		if !signal.IsEmpty() {
			t.Errorf("Expected signal to be empty after reset")
		}
	})
}

func TestAsyncSignal(t *testing.T) {
	t.Run("Add and emit to listeners", func(t *testing.T) {
		signal := New[int]()
		result := make(chan int, 2)
		var wg sync.WaitGroup
		wg.Add(2)

		signal.AddListener(func(ctx context.Context, payload int) {
			defer wg.Done()
			result <- payload
		})

		signal.AddListener(func(ctx context.Context, payload int) {
			defer wg.Done()
			result <- payload * 2
		})

		err := signal.Emit(context.Background(), 5)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		wg.Wait()

		if len(result) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(result))
		}

		results := []int{<-result, <-result}
		if !(results[0] == 5 && results[1] == 10) && !(results[0] == 10 && results[1] == 5) {
			t.Errorf("Expected 5 and 10 in any order, got %v", results)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		signal := New[int]()
		result := make(chan int, 1)

		signal.AddListener(func(ctx context.Context, payload int) {
			select {
			case <-ctx.Done():
				result <- -1
			case <-time.After(200 * time.Millisecond):
				result <- payload
			}
		})

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel() // 在短暂延迟后取消上下文
		}()

		err := signal.Emit(ctx, 5)
		if err == nil {
			t.Fatalf("Expected error due to cancelled context, got nil")
		}

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

			if signal.Len() != 0 {
				t.Errorf("Expected initial length 0, got %d", signal.Len())
			}

			if !signal.IsEmpty() {
				t.Errorf("Expected signal to be initially empty")
			}

			signal.AddListener(func(ctx context.Context, payload int) {})

			if signal.Len() != 1 {
				t.Errorf("Expected length 1 after adding listener, got %d", signal.Len())
			}

			if signal.IsEmpty() {
				t.Errorf("Expected signal not to be empty after adding listener")
			}

			signal.Reset()

			if !signal.IsEmpty() {
				t.Errorf("Expected signal to be empty after reset")
			}
		})
	}
}

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

			signal.AddListener(func(ctx context.Context, payload int) {
				fmt.Printf("Listener for %s about to panic\n", tc.name)
				panic("Listener panic")
			})

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
