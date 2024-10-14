package bloom

import (
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Options 定义了 Bloom 过滤器的配置选项
type Options struct {
	redis   redis.UniversalClient
	key     string
	expire  time.Duration
	hash    []func(str string) uint64
	timeout time.Duration
}

// DefaultOptions 返回默认的 Options 配置
func DefaultOptions() *Options {
	return &Options{
		key:     "bloom",
		expire:  5 * time.Minute,
		hash:    []func(str string) uint64{BKDRHash, SDBMHash, DJBHash},
		timeout: 3 * time.Second,
	}
}

// WithRedis 设置 Redis 客户端
func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(o *Options) {
		if rd != nil {
			o.redis = rd
		}
	}
}

// WithKey 设置 Bloom 过滤器在 Redis 中使用的键名
func WithKey(key string) func(*Options) {
	return func(o *Options) {
		if key != "" {
			o.key = key
		}
	}
}

// WithExpire 设置 Bloom 过滤器的过期时间
func WithExpire(d time.Duration) func(*Options) {
	return func(o *Options) {
		if d > 0 {
			o.expire = d
		}
	}
}

// WithHash 添加自定义的哈希函数
func WithHash(f ...func(str string) uint64) func(*Options) {
	return func(o *Options) {
		if len(f) > 0 {
			o.hash = append(o.hash, f...)
		}
	}
}

// WithTimeout 设置操作超时时间
func WithTimeout(d time.Duration) func(*Options) {
	return func(o *Options) {
		if d > 0 {
			o.timeout = d
		}
	}
}

// Validate 验证 Options 的配置是否有效
func (o *Options) Validate() error {
	if o.redis == nil {
		return errors.New("Redis client is not set")
	}
	if o.key == "" {
		return errors.New("Key is not set")
	}
	if len(o.hash) == 0 {
		return errors.New("No hash functions provided")
	}
	return nil
}

// getOptionsOrSetDefault 获取选项或设置默认值
func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return DefaultOptions()
	}
	return options
}
