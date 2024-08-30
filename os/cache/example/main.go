package main

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/configs"
	"github.com/sagoo-cloud/nexframe/os/cache"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

func main() {
	// 启用 pprof 进行性能分析
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// 创建配置
	config := &configs.CacheConfig{
		MemoryCacheSize: 10 * 1024 * 1024, // 10MB
		RedisConfig: configs.RedisConfig{
			Addr:        "localhost:6379",
			Db:          0,
			RedisPrefix: "SagooCache:",
		},
	}

	// 创建缓存管理器
	cacheManager := cache.NewCacheManager(config)

	// 设置自定义错误处理器
	cacheManager.WithErrorHandler(&CustomErrorHandler{})

	// 预热缓存
	err := cacheManager.PrewarmCacheFromPattern("user:*")
	if err != nil {
		log.Printf("Error prewarming cache: %v", err)
	} else {
		log.Println("Cache prewarming completed")
	}

	// 执行高并发测试
	performBenchmark(cacheManager)

	// 打印统计数据
	printStats(cacheManager)

	// 保持程序运行，以便观察键过期事件
	select {}
}

// 执行高并发测试
func performBenchmark(cacheManager *cache.CacheManager) {
	var wg sync.WaitGroup
	numRequests := 10000         // 增加请求数量
	duration := time.Second * 15 // 运行15秒

	startTime := time.Now()
	endTime := startTime.Add(duration)

	for time.Now().Before(endTime) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			key := fmt.Sprintf("user:%d", rand.Intn(numRequests))
			value := []byte(fmt.Sprintf("user_data_%s", key))

			// 随机执行 Set 或 Get 操作
			if rand.Float32() < 0.3 { // 30% 的概率执行 Set
				err := cacheManager.Set(key, value, time.Second*10)
				if err != nil {
					log.Printf("Set error: %v", err)
				}
			} else { // 70% 的概率执行 Get
				data, exists, err := cacheManager.Get(key)
				if err != nil {
					log.Printf("Get error: %v", err)
				} else if !exists {
					log.Printf("Data not found in cache for key: %s", key)
				} else {
					log.Printf("Data for key %s: %s", key, string(data))
				}
			}
		}()
	}

	wg.Wait()
	benchmarkDuration := time.Since(startTime)
	fmt.Printf("Benchmark completed in %v\n", benchmarkDuration)
}

// 打印统计数据
func printStats(cacheManager *cache.CacheManager) {
	memoryRequestQPS, memoryHitQPS, redisRequestQPS, redisHitQPS := cacheManager.GetStats()
	fmt.Printf("Memory Cache - Request QPS: %.2f, Hit QPS: %.2f\n", memoryRequestQPS, memoryHitQPS)
	fmt.Printf("Redis Cache - Request QPS: %.2f, Hit QPS: %.2f\n", redisRequestQPS, redisHitQPS)
}

// 自定义错误处理器
type CustomErrorHandler struct{}

func (h *CustomErrorHandler) HandleError(err error) {
	log.Printf("Cache error: %v", err)
}
