// Package ratelimit 提供了一个灵活的令牌桶实现，用于速率限制。
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// 确保 Bucket 实现了 RateLimiter 接口
var _ RateLimiter = (*Bucket)(nil)

// LimitMode 定义速率限制的模式
type LimitMode int

const (
	// ModeRelaxed 允许在高负载下稍微超过速率限制
	ModeRelaxed LimitMode = iota
	// ModeStrict 严格执行速率限制
	ModeStrict
)

// RateLimiter 接口定义了速率限制器的基本操作
type RateLimiter interface {
	Take(ctx context.Context, count int64) error
	Close()
	Available() int64
}

// Bucket 表示一个灵活的速率限制令牌桶
type Bucket struct {
	mu           sync.Mutex
	capacity     int64
	tokens       int64
	fillInterval time.Duration
	lastFillTime time.Time
	mode         LimitMode
	metrics      *bucketMetrics
	ctx          context.Context
	cancelFunc   context.CancelFunc
	tokenChan    chan struct{}
	closed       bool
}

// BucketOption 定义了创建Bucket时的选项
type BucketOption func(*Bucket)

// WithStrictMode 设置严格模式的选项
func WithStrictMode() BucketOption {
	return func(b *Bucket) {
		b.mode = ModeStrict
	}
}

// NewBucket 创建并返回一个新的令牌桶
func NewBucket(ctx context.Context, fillInterval time.Duration, capacity int64, opts ...BucketOption) (*Bucket, error) {
	if fillInterval <= 0 || capacity <= 0 {
		return nil, fmt.Errorf("invalid parameters: fillInterval and capacity must be positive")
	}

	ctx, cancel := context.WithCancel(ctx)
	b := &Bucket{
		capacity:     capacity,
		tokens:       capacity,
		fillInterval: fillInterval,
		lastFillTime: time.Now(),
		mode:         ModeRelaxed, // 默认为宽松模式
		metrics:      newBucketMetrics(),
		ctx:          ctx,
		cancelFunc:   cancel,
		tokenChan:    make(chan struct{}, capacity),
	}

	for _, opt := range opts {
		opt(b)
	}

	go b.fillLoop()

	return b, nil
}

// fillLoop 持续填充令牌桶
func (b *Bucket) fillLoop() {
	ticker := time.NewTicker(b.fillInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.mu.Lock()
			if b.tokens < b.capacity {
				b.tokens++
				b.metrics.tokensAdded.Add(1)
				select {
				case b.tokenChan <- struct{}{}:
				default:
					// 通道已满，丢弃多余的令牌
				}
			}
			b.mu.Unlock()
		}
	}
}

// Take 尝试从桶中获取指定数量的令牌
func (b *Bucket) Take(ctx context.Context, count int64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("bucket is closed")
	}

	now := time.Now()
	elapsed := now.Sub(b.lastFillTime)
	tokensToAdd := int64(elapsed / b.fillInterval)

	if tokensToAdd > 0 {
		b.tokens = min(b.capacity, b.tokens+tokensToAdd)
		b.lastFillTime = now.Add(-elapsed % b.fillInterval) // 保留未用完的时间
	}

	if b.tokens < count {
		if b.mode == ModeStrict {
			waitTime := b.fillInterval * time.Duration(count-b.tokens)
			timer := time.NewTimer(waitTime)
			defer timer.Stop()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				b.tokens = count
			}
		} else {
			// 在宽松模式下，我们添加一个小的、一致的延迟
			waitTime := b.fillInterval / 10 // 使用填充间隔的 1/10 作为延迟
			time.Sleep(waitTime)            // 使用 Sleep 而不是 timer，以确保一致的延迟
			// 继续执行，即使没有足够的令牌
			b.tokens = maxNum(b.tokens-count, -count) // 允许令牌数变为负数，但有限制
		}
	} else {
		b.tokens -= count
	}

	return nil
}

func maxNum(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// takeStrict 严格模式下的Take操作
func (b *Bucket) takeStrict(ctx context.Context, count int64, span trace.Span) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.tokens < count {
		waitTime := b.timeUntilTokens(count)
		timer := time.NewTimer(waitTime)
		defer timer.Stop()

		b.mu.Unlock()
		select {
		case <-ctx.Done():
			span.SetAttributes(attribute.String("result", "context_cancelled"))
			b.mu.Lock()
			return ctx.Err()
		case <-timer.C:
			b.mu.Lock()
		}
	}

	if b.tokens < count {
		span.SetAttributes(attribute.String("result", "insufficient_tokens"))
		return fmt.Errorf("insufficient tokens")
	}

	b.tokens -= count
	b.metrics.tokensTaken.Add(count)
	span.SetAttributes(attribute.String("result", "success"))
	return nil
}

// takeRelaxed 宽松模式下的Take操作
func (b *Bucket) takeRelaxed(ctx context.Context, count int64, span trace.Span) error {
	for i := int64(0); i < count; i++ {
		select {
		case <-ctx.Done():
			span.SetAttributes(attribute.String("result", "context_cancelled"))
			return ctx.Err()
		case <-b.tokenChan:
			b.metrics.tokensTaken.Add(1)
		case <-time.After(b.fillInterval):
			// 允许稍微超过限制
			b.metrics.tokensTaken.Add(1)
		}
	}

	span.SetAttributes(attribute.String("result", "success"))
	return nil
}

// timeUntilTokens 计算直到有足够令牌可用需要等待的时间
func (b *Bucket) timeUntilTokens(count int64) time.Duration {
	if b.tokens >= count {
		return 0
	}
	tokensNeeded := count - b.tokens
	return b.fillInterval * time.Duration(tokensNeeded)
}

// Close 关闭令牌桶，停止填充
func (b *Bucket) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.closed = true
		b.cancelFunc()
	}
}

// Available 返回当前可用的令牌数量
func (b *Bucket) Available() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.tokens
}

// bucketMetrics 用于跟踪令牌桶的指标
type bucketMetrics struct {
	tokensTaken *safeCounter
	tokensAdded *safeCounter
}

func newBucketMetrics() *bucketMetrics {
	return &bucketMetrics{
		tokensTaken: newSafeCounter(),
		tokensAdded: newSafeCounter(),
	}
}

// safeCounter 是一个线程安全的计数器
type safeCounter struct {
	mu    sync.Mutex
	value int64
}

func newSafeCounter() *safeCounter {
	return &safeCounter{}
}

func (sc *safeCounter) Add(n int64) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.value += n
}

func (sc *safeCounter) Get() int64 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.value
}
