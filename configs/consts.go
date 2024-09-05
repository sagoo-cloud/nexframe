package configs

// 缓存配置
const (
	CacheRedisMemoryCacheSize = "redis.memoryCacheSize"
)

// redis配置
const (
	RedisMode               = "redis.mode"
	RedisSentinelMasterName = "redis.sentinelMasterName"
	RedisAddr               = "redis.addr"
	RedisUsername           = "redis.username"
	RedisPassword           = "redis.password"
	RedisDb                 = "redis.db"
	RedisMaxActive          = "redis.maxActive"
	RedisMaxIdle            = "redis.maxIdle"
	RedisIdleTimeout        = "redis.idleTimeout"
	RedisPrefix             = "redis.prefix"

	RedisDataCachePoolSize           = "dataCache.poolSize"
	RedisDataCacheRecordDuration     = "dataCache.recordDuration"
	RedisDataCacheRecordLimit        = "dataCache.recordLimit"
	RedisDataCachePipelineBufferSize = "dataCache.pipelineBufferSize"
)

// 数据库配置
const (
	DatabaseDriver       = "database.driver"
	DatabaseHost         = "database.host"
	DatabasePort         = "database.port"
	DatabaseUserName     = "database.username"
	DatabasePassword     = "database.password"
	DatabaseDbName       = "database.dbName"
	DatabaseConfig       = "database.config"
	DatabaseMaxIdleConns = "database.maxIdleConns"
	DatabaseMaxOpenConns = "database.maxOpenConns"
	DatabaseShowSQL      = "database.showSql"
)

// 日志配置
const (
	LogLevel            = "log.level"
	LogPattern          = "log.pattern"
	LogOutput           = "log.output"
	LogRotateFile       = "log.rotate.file"
	LogRotateMaxSize    = "log.rotate.maxSize"
	LogRotateMaxBackups = "log.rotate.maxBackups"
	LogRotateMaxAge     = "log.rotate.maxAge"
	LogRotateCompress   = "log.rotate.compress"
)
