package nx

import (
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type Options struct {
	redis    redis.UniversalClient
	key      string        // Redis 缓存密钥，默认 nx.lock
	expire   int           // 密钥过期时间，默认为 60 秒，避免死锁，不应设置太长
	retry    int           // 重试次数，默认 10 次
	interval time.Duration // 重试间隔，默认 25 毫秒
}

// Validate 校验 Options 结构体的参数是否合法
func (o *Options) Validate() error {
	if o.redis == nil {
		return errors.New("redis client must not be nil")
	}
	if o.key == "" {
		return errors.New("redis key must not be empty")
	}
	if o.expire <= 0 {
		return errors.New("expire time must be greater than zero")
	}
	if o.retry < 0 {
		return errors.New("retry count must not be negative")
	}
	if o.interval <= 0 {
		return errors.New("interval must be greater than zero")
	}
	return nil
}

// getOptionsOrSetDefault 获取 Options 结构体的指针，如果参数为 nil，则返回默认值
func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			key:      "nx.lock",
			expire:   60,
			retry:    10,
			interval: 25 * time.Millisecond,
		}
	}
	return options
}

// WithRedis 设置 Redis 客户端
func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		options.redis = rd
	}
}

// WithKey 设置 Redis 缓存密钥
func WithKey(key string) func(*Options) {
	return func(options *Options) {
		options.key = key
	}
}

// WithExpire 设置 Redis 缓存密钥过期时间
func WithExpire(seconds int) func(*Options) {
	return func(options *Options) {
		options.expire = seconds
	}
}

// WithRetry 设置重试次数
func WithRetry(count int) func(*Options) {
	return func(options *Options) {
		options.retry = count
	}
}

// WithInterval 设置重试间隔时间
func WithInterval(interval time.Duration) func(*Options) {
	return func(options *Options) {
		options.interval = interval
	}
}
