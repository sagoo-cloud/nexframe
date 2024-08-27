package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sagoo-cloud/nexframe/configs"
	"github.com/sagoo-cloud/nexframe/database/redisdb"
	"strings"
	"time"
)

var ctx = context.Background()

// RedisCache 结构体，实现 Redis 缓存
type RedisCache struct {
	client redis.UniversalClient
	stats  *CacheStats
	prefix string
	dbname string
}

// NewRedisCache 创建 RedisCache 实例
func NewRedisCache(config *configs.CacheConfig) *RedisCache {

	client := redisdb.DB().GetClient()

	// 启用键空间通知
	client.ConfigSet(ctx, "notify-keyspace-events", "Ex")

	return &RedisCache{
		client: client,
		stats:  NewCacheStats(),
		prefix: config.RedisPrefix,
		dbname: redisdb.DB().GetDbname(),
	}
}

// keyWithPrefix 给键添加前缀
func (rc *RedisCache) keyWithPrefix(key string) string {
	return rc.prefix + key
}

// Set 设置数据到 Redis
func (rc *RedisCache) Set(key string, value []byte, ttl time.Duration) error {
	rc.stats.IncrementRequestCount()
	return rc.client.Set(ctx, rc.keyWithPrefix(key), value, ttl).Err()
}

// Get 从 Redis 获取数据
func (rc *RedisCache) Get(key string) ([]byte, bool, error) {
	rc.stats.IncrementRequestCount()
	value, err := rc.client.Get(ctx, rc.keyWithPrefix(key)).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	rc.stats.IncrementHitCount()
	return value, true, nil
}

// Delete 从 Redis 删除数据
func (rc *RedisCache) Delete(key string) error {
	return rc.client.Del(ctx, rc.keyWithPrefix(key)).Err()
}

// GetStats 获取 Redis 缓存的统计信息
func (rc *RedisCache) GetStats() (float64, float64) {
	return rc.stats.GetStats()
}

// SubscribeExpiryEvents 订阅 Redis 键过期事件
func (rc *RedisCache) SubscribeExpiryEvents(callback func(string)) {
	dbname := "0"
	if rc.dbname != "" {
		dbname = rc.dbname
	}
	pubsub := rc.client.PSubscribe(ctx, fmt.Sprintf("__keyevent@%s__:expired", dbname))

	go func() {
		for msg := range pubsub.Channel() {
			key := strings.TrimPrefix(msg.Payload, rc.prefix)
			callback(key)
		}
	}()
}

// Keys 获取Redis中指定前缀的key列表
func (rc *RedisCache) Keys(prefix string) ([]string, error) {
	var pattern string
	if prefix == "" {
		pattern = rc.keyWithPrefix("*")
	} else {
		pattern = rc.keyWithPrefix(prefix + "*")
	}

	keys, err := rc.client.Keys(context.Background(), pattern).Result()
	if err != nil {
		return nil, err
	}

	// 移除前缀
	for i, key := range keys {
		keys[i] = strings.TrimPrefix(key, rc.prefix)
	}

	return keys, nil
}
