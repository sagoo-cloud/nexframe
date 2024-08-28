package g

import (
	"github.com/sagoo-cloud/nexframe/configs"
	"github.com/sagoo-cloud/nexframe/os/cache"
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
		config := configs.LoadCacheConfig()
		instanceCache = cache.NewCacheManager(config)
	})
	return instanceCache
}
