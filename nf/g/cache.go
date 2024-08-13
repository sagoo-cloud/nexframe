package g

import (
	"github.com/sagoo-cloud/nexframe/cache"
	"sync"
)

// Cache 获取全局缓存管理器
func Cache() *cache.CacheManager {
	return GetCacheInstance()
}

var (
	instanceCache *cache.CacheManager
	once          sync.Once
)

// GetCacheInstance 返回 ConfigDataEntity 的单例实例
func GetCacheInstance() *cache.CacheManager {
	once.Do(func() {
		config := &cache.Config{
			MemoryCacheSize: 10 * 1024 * 1024, // 10MB
			RedisAddr:       "localhost:6379",
			RedisPassword:   "",
			RedisDB:         0,
			RedisPrefix:     "",
		}
		instanceCache = cache.NewCacheManager(config)
	})
	return instanceCache
}
