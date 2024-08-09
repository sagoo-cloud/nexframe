package ratelimit

import (
	"context"
	"testing"
	"time"
)

// mockClock 是一个用于测试的模拟时钟
type mockClock struct {
	now time.Time
}

func (m *mockClock) Now() time.Time {
	return m.now
}

func (m *mockClock) Sleep(d time.Duration) {
	m.now = m.now.Add(d)
}

func TestNewBucket(t *testing.T) {
	ctx := context.Background()
	b, err := NewBucket(ctx, time.Second, 10)
	if err != nil {
		t.Fatalf("NewBucket failed: %v", err)
	}
	if b.capacity != 10 {
		t.Errorf("Expected capacity 10, got %d", b.capacity)
	}
	if b.fillInterval != time.Second {
		t.Errorf("Expected fill interval 1s, got %v", b.fillInterval)
	}
	if b.mode != ModeRelaxed {
		t.Errorf("Expected relaxed mode by default")
	}
}

func TestBucketTakeRelaxed(t *testing.T) {
	ctx := context.Background()
	fillInterval := time.Millisecond * 100
	b, _ := NewBucket(ctx, fillInterval, 10)

	// 应该能够立即获取10个令牌
	for i := 0; i < 10; i++ {
		err := b.Take(ctx, 1)
		if err != nil {
			t.Errorf("Failed to take token: %v", err)
		}
	}

	// 第11个令牌应该会有一个小的延迟
	start := time.Now()
	err := b.Take(ctx, 1)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Failed to take 11th token: %v", err)
	}

	// 在宽松模式下，我们期望有一个小的延迟，但不应该太长
	minExpectedDelay := fillInterval / 20 // 5ms (5% of fill interval)
	maxExpectedDelay := fillInterval / 2  // 50ms (50% of fill interval)
	if duration < minExpectedDelay {
		t.Errorf("Expected delay for 11th token, but it was too quick: %v. Expected at least %v", duration, minExpectedDelay)
	} else if duration > maxExpectedDelay {
		t.Errorf("Expected a small delay for 11th token, but it was too long: %v. Expected at most %v", duration, maxExpectedDelay)
	}
}
func TestBucketTakeStrict(t *testing.T) {
	ctx := context.Background()
	b, _ := NewBucket(ctx, time.Millisecond*100, 10, WithStrictMode())

	// 应该能够立即获取10个令牌
	for i := 0; i < 10; i++ {
		err := b.Take(ctx, 1)
		if err != nil {
			t.Errorf("Failed to take token: %v", err)
		}
	}

	// 第11个令牌应该会严格延迟约100ms
	start := time.Now()
	err := b.Take(ctx, 1)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Failed to take 11th token: %v", err)
	}

	// 添加这个时间容差检查
	if duration < time.Millisecond*95 || duration > time.Millisecond*105 {
		t.Errorf("Expected delay of ~100ms for 11th token in strict mode, got: %v", duration)
	}
}

func TestBucketAvailable(t *testing.T) {
	ctx := context.Background()
	b, _ := NewBucket(ctx, time.Millisecond*100, 10)

	if available := b.Available(); available != 10 {
		t.Errorf("Expected 10 available tokens, got %d", available)
	}

	b.Take(ctx, 5)

	if available := b.Available(); available != 5 {
		t.Errorf("Expected 5 available tokens, got %d", available)
	}
}

func TestBucketClose(t *testing.T) {
	ctx := context.Background()
	b, _ := NewBucket(ctx, time.Millisecond*100, 10)

	b.Close()

	// 尝试在关闭后获取令牌应该失败
	err := b.Take(ctx, 1)
	if err == nil {
		t.Errorf("Expected error when taking token from closed bucket")
	}
}

func TestBucketWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	b, _ := NewBucket(ctx, time.Millisecond*100, 1, WithStrictMode())

	// 首先获取唯一的令牌
	b.Take(ctx, 1)

	// 尝试获取另一个令牌应该因为上下文超时而失败
	err := b.Take(ctx, 1)
	if err == nil {
		t.Errorf("Expected context deadline exceeded error")
	}
}

func TestBucketRefill(t *testing.T) {
	ctx := context.Background()
	b, _ := NewBucket(ctx, time.Millisecond*100, 10)

	// 耗尽所有令牌
	for i := 0; i < 10; i++ {
		b.Take(ctx, 1)
	}

	// 等待足够的时间以重新填充一些令牌
	time.Sleep(time.Millisecond * 250)

	// 应该至少有2个新的令牌可用
	available := b.Available()
	if available < 2 {
		t.Errorf("Expected at least 2 tokens after refill, got %d", available)
	}
}
