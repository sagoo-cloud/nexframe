package cache

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/configs"
	"sync"
	"time"
)

// CacheManager 结构体，管理内存缓存和 Redis 缓存
type CacheManager struct {
	memoryCache *FreeCache
	redisCache  *RedisCache
	mu          sync.RWMutex
	errHandler  ErrorHandler
}

// NewCacheManager 创建 CacheManager 实例
func NewCacheManager(config *configs.CacheConfig) *CacheManager {
	cm := &CacheManager{
		memoryCache: NewFreeCache(config.MemoryCacheSize),
		redisCache:  NewRedisCache(config),
	}

	// 订阅 Redis 键过期事件
	cm.redisCache.SubscribeExpiryEvents(func(key string) {
		cm.mu.Lock()
		defer cm.mu.Unlock()
		cm.memoryCache.Delete(key)
	})

	return cm
}

// WithErrorHandler 设置错误处理器
func (cm *CacheManager) WithErrorHandler(handler ErrorHandler) *CacheManager {
	cm.errHandler = handler
	return cm
}

// handleError 处理错误
func (cm *CacheManager) handleError(err error) {
	if cm.errHandler != nil {
		cm.errHandler.HandleError(err)
	}
}

// Set 设置缓存数据
func (cm *CacheManager) Set(key string, value []byte, ttl time.Duration) error {
	// 首先设置到 Redis
	err := cm.redisCache.Set(key, value, ttl)
	if err != nil {
		cm.handleError(err)
		return err
	}

	// 然后更新内存缓存
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.memoryCache.Set(key, value, ttl)
}

// Get 获取缓存数据
func (cm *CacheManager) Get(key string) ([]byte, bool, error) {
	// 首先检查内存缓存
	cm.mu.RLock()
	value, exists := cm.memoryCache.Get(key)
	cm.mu.RUnlock()

	if exists {
		return value, true, nil
	}

	// 如果内存缓存中不存在，则从 Redis 获取，包含重试逻辑
	var redisErr error
	for retries := 0; retries < 3; retries++ {
		value, exists, redisErr = cm.redisCache.Get(key)
		if redisErr == nil {
			break
		}
		time.Sleep(time.Millisecond * 100 * time.Duration(retries+1))
	}

	if redisErr != nil {
		cm.handleError(redisErr)
		return nil, false, redisErr
	}

	if exists {
		// 将从 Redis 获取的数据更新到内存缓存
		cm.mu.Lock()
		cm.memoryCache.Set(key, value, 0) // 这里可以设置一个较短的 TTL
		cm.mu.Unlock()
	}

	return value, exists, nil
}

// Delete 删除缓存数据
func (cm *CacheManager) Delete(key string) error {
	// 首先从 Redis 删除
	err := cm.redisCache.Delete(key)
	if err != nil {
		cm.handleError(err)
		return err
	}

	// 然后从内存缓存删除
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.memoryCache.Delete(key)
}

// GetStats 获取缓存统计数据
func (cm *CacheManager) GetStats() (memoryRequestQPS, memoryHitQPS, redisRequestQPS, redisHitQPS float64) {
	memoryRequestQPS, memoryHitQPS = cm.memoryCache.GetStats()
	redisRequestQPS, redisHitQPS = cm.redisCache.GetStats()
	return
}

// PrewarmCache 预热缓存
func (cm *CacheManager) PrewarmCache(keys []string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(keys))

	// 使用工作池来并行处理预热
	workerCount := 10 // 可以根据需要调整
	jobs := make(chan string, len(keys))

	// 启动工作协程
	for i := 0; i < workerCount; i++ {
		go func() {
			for key := range jobs {
				value, exists, err := cm.redisCache.Get(key)
				if err != nil {
					errChan <- fmt.Errorf("error prewarm key %s: %v", key, err)
					wg.Done()
					continue
				}
				if exists {
					err = cm.memoryCache.Set(key, value, 0) // 可以设置一个默认的 TTL
					if err != nil {
						errChan <- fmt.Errorf("error setting prewarm key %s: %v", key, err)
					}
				}
				wg.Done()
			}
		}()
	}

	// 发送任务
	for _, key := range keys {
		wg.Add(1)
		jobs <- key
	}
	close(jobs)

	// 等待所有任务完成
	wg.Wait()
	close(errChan)

	// 收集错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during prewarming: %v", len(errors), errors)
	}

	return nil
}

// PrewarmCacheFromPattern 从 Redis 模式匹配预热缓存
func (cm *CacheManager) PrewarmCacheFromPattern(pattern string) error {
	keys, err := cm.redisCache.Keys(pattern)
	if err != nil {
		return fmt.Errorf("error getting keys from Redis: %v", err)
	}
	return cm.PrewarmCache(keys)
}

// Keys 获取缓存中指定前缀的key列表，如果prefix为空则返回所有key
func (cm *CacheManager) Keys(prefix string) (keys []interface{}, err error) {
	// 首先从Redis获取键
	redisKeys, err := cm.redisCache.Keys(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys from Redis: %w", err)
	}

	// 将Redis的键添加到结果中
	for _, key := range redisKeys {
		keys = append(keys, key)
	}

	// 去重
	uniqueKeys := make(map[interface{}]bool)
	var result []interface{}
	for _, key := range keys {
		if _, exists := uniqueKeys[key]; !exists {
			uniqueKeys[key] = true
			result = append(result, key)
		}
	}

	return result, nil
}
