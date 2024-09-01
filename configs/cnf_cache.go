package configs

import (
	"time"
)

// CacheConfig 结构体用于存储缓存系统的配置
type CacheConfig struct {
	MemoryCacheSize int // 内存缓存大小（字节）
	RedisConfig
}

func LoadCacheConfig() *CacheConfig {
	config := &CacheConfig{
		MemoryCacheSize: EnvInt(CacheRedisMemoryCacheSize, 10*1024*1024), //10MB
		RedisConfig: RedisConfig{
			Mode:               EnvString(RedisMode, "single"), // single, sentinel, cluster
			SentinelMasterName: EnvString(RedisSentinelMasterName, "sagoo-master"),
			Addr:               EnvString(RedisAddr, "127.0.0.1:6937"),
			Auth:               EnvString(RedisAuth, "password"),
			Db:                 EnvInt(RedisDb, 0),
			MaxActive:          EnvInt(RedisMaxActive, 50),
			MaxIdle:            EnvInt(RedisMaxIdle, 5),
			IdleTimeout:        time.Duration(EnvInt(RedisIdleTimeout, 10)) * time.Second,
		},
	}
	return config
}
