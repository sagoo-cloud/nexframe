package g

import "github.com/sagoo-cloud/nexframe/redisx"

// RedisDB 获取redis实例
func RedisDB() *redisx.RedisManager {
	return redisx.DB()
}
