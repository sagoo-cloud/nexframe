package cache

// Config 结构体用于存储缓存系统的配置
type Config struct {
	MemoryCacheSize int    // 内存缓存大小（字节）
	RedisAddr       string // Redis 地址
	RedisPassword   string // Redis 密码
	RedisDB         int    // Redis 数据库编号
	RedisPrefix     string // Redis 键前缀
}
