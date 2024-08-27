package g

import "github.com/sagoo-cloud/nexframe/database/redisdb"

// RedisDB 获取redis实例
func RedisDB() *redisdb.RedisManager {
	return redisdb.DB()
}
