package cache

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/configs"
	"sync"
	"testing"
	"time"
)

var testConfig = &configs.CacheConfig{
	MemoryCacheSize: 10 * 1024 * 1024, // 10MB
	RedisConfig: configs.RedisConfig{
		Addr:        "localhost:6379",
		Db:          0,
		RedisPrefix: "SagooCache:",
	},
}

func TestSetAndGet(t *testing.T) {
	cm := NewCacheManager(testConfig)

	// 测试 Set
	err := cm.Set("test_key", []byte("test_value"), time.Minute)
	if err != nil {
		t.Errorf("Failed to set cache: %v", err)
	}

	// 测试 Get
	value, exists, err := cm.Get("test_key")
	if err != nil {
		t.Errorf("Failed to get cache: %v", err)
	}
	if !exists {
		t.Error("Cache key does not exist")
	}
	if string(value) != "test_value" {
		t.Errorf("Cache value mismatch. Expected 'test_value', got '%s'", string(value))
	}

	// 测试不存在的键
	_, exists, _ = cm.Get("non_existent_key")
	if exists {
		t.Error("Non-existent key should not exist in cache")
	}
}

func TestDelete(t *testing.T) {
	cm := NewCacheManager(testConfig)

	// 设置一个值
	cm.Set("delete_key", []byte("delete_value"), time.Minute)

	// 删除该值
	err := cm.Delete("delete_key")
	if err != nil {
		t.Errorf("Failed to delete cache: %v", err)
	}

	// 尝试获取已删除的值
	_, exists, _ := cm.Get("delete_key")
	if exists {
		t.Error("Deleted key should not exist in cache")
	}
}

func TestTTL(t *testing.T) {
	cm := NewCacheManager(testConfig)

	// 设置一个 1 秒后过期的缓存项
	err := cm.Set("ttl_key", []byte("ttl_value"), time.Second)
	if err != nil {
		t.Errorf("Failed to set cache with TTL: %v", err)
	}

	// 立即获取，应该存在
	_, exists, _ := cm.Get("ttl_key")
	if !exists {
		t.Error("Cache key should exist immediately after setting")
	}

	// 等待 2 秒
	time.Sleep(2 * time.Second)

	// 再次获取，应该已经过期
	_, exists, _ = cm.Get("ttl_key")
	if exists {
		t.Error("Cache key should have expired")
	}
}

func TestConcurrency(t *testing.T) {
	cm := NewCacheManager(testConfig)
	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent_key_%d", i)
			value := []byte(fmt.Sprintf("concurrent_value_%d", i))

			// Set
			err := cm.Set(key, value, time.Minute)
			if err != nil {
				t.Errorf("Concurrent set failed: %v", err)
			}

			// Get
			retrievedValue, exists, err := cm.Get(key)
			if err != nil {
				t.Errorf("Concurrent get failed: %v", err)
			}
			if !exists {
				t.Errorf("Concurrent key %s should exist", key)
			}
			if string(retrievedValue) != string(value) {
				t.Errorf("Concurrent value mismatch for key %s", key)
			}
		}(i)
	}

	wg.Wait()
}

func TestErrorHandling(t *testing.T) {
	// 创建一个带有错误的 Redis 地址的配置
	badConfig := &configs.CacheConfig{
		MemoryCacheSize: 10 * 1024 * 1024,
		RedisConfig: configs.RedisConfig{
			Addr:        "localhost:6379",
			Db:          0,
			RedisPrefix: "SagooCache:",
		},
	}

	cm := NewCacheManager(badConfig)

	// 测试 Set 操作的错误处理
	err := cm.Set("error_key", []byte("error_value"), time.Minute)
	if err == nil {
		t.Error("Expected an error when setting to invalid Redis, but got nil")
	}

	// 测试 Get 操作的错误处理
	_, _, err = cm.Get("error_key")
	if err == nil {
		t.Error("Expected an error when getting from invalid Redis, but got nil")
	}
}

func TestPrewarmCache(t *testing.T) {
	cm := NewCacheManager(testConfig)

	// 准备预热数据
	prewarmKeys := []string{"prewarm_1", "prewarm_2", "prewarm_3"}
	for _, key := range prewarmKeys {
		cm.redisCache.Set(key, []byte("prewarm_value_"+key), time.Minute)
	}

	// 执行预热
	err := cm.PrewarmCache(prewarmKeys)
	if err != nil {
		t.Errorf("Failed to prewarm cache: %v", err)
	}

	// 验证预热结果
	for _, key := range prewarmKeys {
		value, exists, err := cm.Get(key)
		if err != nil {
			t.Errorf("Failed to get prewarm key %s: %v", key, err)
		}
		if !exists {
			t.Errorf("Prewarm key %s should exist in cache", key)
		}
		if string(value) != "prewarm_value_"+key {
			t.Errorf("Prewarm value mismatch for key %s", key)
		}
	}
}

func TestPrewarmCacheFromPattern(t *testing.T) {
	cm := NewCacheManager(testConfig)

	// 准备预热数据
	prewarmKeys := []string{"pattern_1", "pattern_2", "pattern_3"}
	for _, key := range prewarmKeys {
		cm.redisCache.Set(key, []byte("pattern_value_"+key), time.Minute)
	}

	// 执行预热
	err := cm.PrewarmCacheFromPattern("pattern_*")
	if err != nil {
		t.Errorf("Failed to prewarm cache from pattern: %v", err)
	}

	// 验证预热结果
	for _, key := range prewarmKeys {
		value, exists, err := cm.Get(key)
		if err != nil {
			t.Errorf("Failed to get pattern key %s: %v", key, err)
		}
		if !exists {
			t.Errorf("Pattern key %s should exist in cache", key)
		}
		if string(value) != "pattern_value_"+key {
			t.Errorf("Pattern value mismatch for key %s", key)
		}
	}
}
