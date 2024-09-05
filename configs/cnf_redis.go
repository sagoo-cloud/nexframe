package configs

import (
	"time"
)

type RedisConfig struct {
	Mode               string
	SentinelMasterName string
	Addr               string
	Username           string
	Password           string
	Db                 int
	MaxActive          int
	MaxIdle            int
	IdleTimeout        time.Duration
	RedisPrefix        string
	DataCacheConfig    RedisDataCacheConfig
}

type RedisDataCacheConfig struct {
	PoolSize           int
	RecordDuration     string
	RecordLimit        int
	PipelineBufferSize int
}

func LoadRedisConfig() *RedisConfig {
	config := &RedisConfig{
		Mode:               EnvString(RedisMode, "single"), // single, sentinel, cluster
		SentinelMasterName: EnvString(RedisSentinelMasterName, "sagoo-master"),
		Addr:               EnvString(RedisAddr, "127.0.0.1:6937"),
		Username:           EnvString(RedisUsername, "default"),
		Password:           EnvString(RedisPassword, "password"),
		Db:                 EnvInt(RedisDb, 0),
		MaxActive:          EnvInt(RedisMaxActive, 50),
		MaxIdle:            EnvInt(RedisMaxIdle, 5),
		IdleTimeout:        time.Duration(EnvInt(RedisIdleTimeout, 10)) * time.Second,
		RedisPrefix:        EnvString(RedisPrefix, "sagooiot:"),
		DataCacheConfig: RedisDataCacheConfig{
			PoolSize:           EnvInt(RedisDataCachePoolSize, 500),
			RecordDuration:     EnvString(RedisDataCacheRecordDuration, "10m"),
			RecordLimit:        EnvInt(RedisDataCacheRecordLimit, 1000),
			PipelineBufferSize: EnvInt(RedisDataCachePipelineBufferSize, 3),
		},
	}
	return config
}
