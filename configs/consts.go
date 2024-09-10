package configs

// 缓存配置
const (
	CacheRedisMemoryCacheSize = "redis.memoryCacheSize"
)

// web服务配置
const (
	ServerName              = "server.name"
	ServerAddress           = "server.address"
	ServerHTTPSAddress      = "server.httpsAddress"
	ServerHTTPSCertPath     = "server.httpsCertPath"
	ServerHTTPSKeyPath      = "server.httpsKeyPath"
	ServerReadTimeout       = "server.readTimeout"
	ServerWriteTimeout      = "server.writeTimeout"
	ServerIdleTimeout       = "server.idleTimeout"
	ServerMaxHeaderBytes    = "server.maxHeaderBytes"
	ServerKeepAlive         = "server.keepAlive"
	ServerServerAgent       = "server.serverAgent"
	ServerIndexFiles        = "server.indexFiles"
	ServerIndexFolder       = "server.indexFolder"
	ServerServerRoot        = "server.serverRoot"
	ServerSearchPaths       = "server.searchPaths"
	ServerFileServerEnabled = "server.fileServerEnabled"
	ServerPProfEnabled      = "server.pprofEnabled"
	ServerPProfPattern      = "server.pprofPattern"

	ServerCookieMaxAge = "server.cookie.maxAge"
	ServerCookiePath   = "server.cookie.path"
	ServerCookieDomain = "server.cookie.domain"

	ServerSessionIdName       = "server.session.idName"
	ServerSessionMaxAge       = "server.session.maxAge"
	ServerSessionCookieMaxAge = "server.session.cookieMaxAge"
	ServerSessionCookieOutput = "server.session.cookieOutput"
)

// token配置
const (
	TokenKey = "token.key"
	TokenExp = "token.exp"
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

// 消息队列配置
const (
	QueueInterval    = "queue.interval"
	QueuePrefix      = "queue.prefix"
	QueueListen      = "queue.listen"
	QueueConcurrency = "queue.concurrency"
)

// mqtt配置
const (
	MqttHost         = "mqtt.host"
	MattUsername     = "mqtt.username"
	MqttPassword     = "mqtt.password"
	MqttClientID     = "mqtt.client_id"
	MqttParallel     = "mqtt.parallel"
	MqttSubscribeQos = "mqtt.subscribe_qos"
	MqttPublishQos   = "mqtt.publish_qos"
)
