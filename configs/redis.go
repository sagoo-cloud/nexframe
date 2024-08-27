package configs

import (
	"time"
)

type RedisConfig struct {
	Mode               string
	SentinelMasterName string
	Addr               string
	Auth               string
	Db                 int
	MaxActive          int
	MaxIdle            int
	IdleTimeout        time.Duration
}

func LoadRedisConfig() *RedisConfig {
	config := &RedisConfig{
		Mode:               EnvString("redis.mode", "single"), // single, sentinel, cluster
		SentinelMasterName: EnvString("redis.sentinelMasterName", "sagoo-master"),
		Addr:               EnvString("redis.uri", "127.0.0.1:6937"),
		Auth:               EnvString("redis.auth", "password"),
		Db:                 EnvInt("redis.db", 0),
		MaxActive:          EnvInt("redis.maxActive", 50),
		MaxIdle:            EnvInt("redis.maxIdle", 5),
		IdleTimeout:        time.Duration(EnvInt("redis.timeout", 10)) * time.Second,
	}
	return config
}
