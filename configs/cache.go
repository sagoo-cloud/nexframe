package configs

import (
	"time"
)

// CacheConfig 结构体用于存储缓存系统的配置
type CacheConfig struct {
	MemoryCacheSize int           // 内存缓存大小（字节）
	RedisAddr       string        // Redis 地址
	RedisPassword   string        // Redis 密码
	RedisDB         int           // Redis 数据库编号
	RedisPrefix     string        // Redis 键前缀
	MaxActive       int           // Redis 最大活动连接数
	MaxIdle         int           // Redis 最大空闲连接数
	Timeout         time.Duration // Redis 连接超时时间
}

func LoadCacheConfig() *CacheConfig {
	config := &CacheConfig{
		MemoryCacheSize: EnvInt("redis.memoryCacheSize", 10*1024*1024), //10MB
		RedisAddr:       EnvString("redis.addr", "127.0.0.1:6937"),
		RedisDB:         EnvInt("redis.db", 0),
		RedisPrefix:     EnvString("redis.prefix", "sagooiot"),
		MaxActive:       EnvInt("redis.maxActive", 50),
		MaxIdle:         EnvInt("redis.maxIdle", 5),
		Timeout:         time.Duration(EnvInt("redis.timeout", 10)) * time.Second,
	}
	return config
}
