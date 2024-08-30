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
	RedisPrefix        string
}

func LoadRedisConfig() *RedisConfig {
	config := &RedisConfig{
		Mode:               EnvString(RedisMode, "single"), // single, sentinel, cluster
		SentinelMasterName: EnvString(RedisSentinelMasterName, "sagoo-master"),
		Addr:               EnvString(RedisAddr, "127.0.0.1:6937"),
		Auth:               EnvString(RedisAuth, "password"),
		Db:                 EnvInt(RedisDb, 0),
		MaxActive:          EnvInt(RedisMaxActive, 50),
		MaxIdle:            EnvInt(RedisMaxIdle, 5),
		IdleTimeout:        time.Duration(EnvInt(RedisIdleTimeout, 10)) * time.Second,
		RedisPrefix:        EnvString(RedisPrefix, "sagooiot:"),
	}
	return config
}
