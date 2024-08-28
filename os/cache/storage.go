package cache

import "time"

// CacheStorage 定义缓存存储接口
type CacheStorage interface {
	Set(key string, value []byte, ttl time.Duration) error
	Get(key string) ([]byte, bool, error)
	Delete(key string) error
}

// DistributedCache 定义分布式缓存接口
type DistributedCache interface {
	CacheStorage
	Publish(channel string, message []byte) error
	Subscribe(channel string) (<-chan []byte, error)
}

// ErrorHandler 自定义错误处理接口
type ErrorHandler interface {
	HandleError(err error)
}
