package cache

import (
	"sync/atomic"
	"time"
)

// CacheStats 结构体，用于记录缓存统计信息
type CacheStats struct {
	requestCount uint64
	hitCount     uint64
	startTime    time.Time
}

// NewCacheStats 创建 CacheStats 实例
func NewCacheStats() *CacheStats {
	return &CacheStats{
		startTime: time.Now(),
	}
}

// IncrementRequestCount 增加请求计数
func (cs *CacheStats) IncrementRequestCount() {
	atomic.AddUint64(&cs.requestCount, 1)
}

// IncrementHitCount 增加命中计数
func (cs *CacheStats) IncrementHitCount() {
	atomic.AddUint64(&cs.hitCount, 1)
}

// GetStats 获取统计信息
func (cs *CacheStats) GetStats() (requestQPS, hitQPS float64) {
	elapsed := time.Since(cs.startTime).Seconds()
	requestCount := atomic.LoadUint64(&cs.requestCount)
	hitCount := atomic.LoadUint64(&cs.hitCount)

	requestQPS = float64(requestCount) / elapsed
	hitQPS = float64(hitCount) / elapsed
	return
}
