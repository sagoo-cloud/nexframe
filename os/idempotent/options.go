package idempotent

import (
	"github.com/redis/go-redis/v9"
	"github.com/sagoo-cloud/nexframe/database/redisdb"
)

// Options 配置选项结构体
type Options struct {
	redis  redis.UniversalClient
	prefix string
	expire int
}

// WithRedis 设置Redis客户端
func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd != nil {
			getOptionsOrSetDefault(options).redis = rd
		}
	}
}

// WithPrefix 设置键前缀
func WithPrefix(prefix string) func(*Options) {
	return func(options *Options) {
		if prefix != "" {
			getOptionsOrSetDefault(options).prefix = prefix
		}
	}
}

// WithExpire 设置过期时间（分钟）
func WithExpire(min int) func(*Options) {
	return func(options *Options) {
		if min > 0 {
			getOptionsOrSetDefault(options).expire = min
		}
	}
}

// getOptionsOrSetDefault 获取选项或设置默认值
func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			prefix: "idempotent",
			expire: 60,
		}
	}
	if options.redis == nil {
		options.redis = redisdb.DB().GetClient()
	}
	return options
}
