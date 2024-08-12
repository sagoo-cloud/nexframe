package cache

import (
	"github.com/coocood/freecache"
	"time"
)

// FreeCache 结构体，实现内存缓存
type FreeCache struct {
	cache *freecache.Cache
	stats *CacheStats
}

// NewFreeCache 创建 FreeCache 实例
func NewFreeCache(size int) *FreeCache {
	return &FreeCache{
		cache: freecache.NewCache(size),
		stats: NewCacheStats(),
	}
}

// Get 从内存缓存中获取数据
func (fc *FreeCache) Get(key string) ([]byte, bool) {
	fc.stats.IncrementRequestCount()
	value, err := fc.cache.Get([]byte(key))
	if err != nil {
		return nil, false
	}
	fc.stats.IncrementHitCount()
	return value, true
}

// Set 设置数据到内存缓存
func (fc *FreeCache) Set(key string, value []byte, ttl time.Duration) error {
	return fc.cache.Set([]byte(key), value, int(ttl.Seconds()))
}

// Delete 从内存缓存中删除数据
func (fc *FreeCache) Delete(key string) error {
	fc.cache.Del([]byte(key))
	return nil
}

// GetStats 获取内存缓存的统计信息
func (fc *FreeCache) GetStats() (float64, float64) {
	return fc.stats.GetStats()
}
