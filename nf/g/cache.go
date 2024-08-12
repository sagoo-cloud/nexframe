package g

import "github.com/sagoo-cloud/nexframe/cache"

func Cache() *cache.CacheManager {

	config := &cache.Config{
		MemoryCacheSize: 10 * 1024 * 1024, // 10MB
		RedisAddr:       "localhost:6379",
		RedisPassword:   "",
		RedisDB:         0,
		RedisPrefix:     "",
	}
	return cache.NewCacheManager(config)
}
