package redisx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sagoo-cloud/nexframe/nf/configs"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	DeviceDataCachePrefix = "deviceCacheData:" // 设备数据缓存前缀
	maxRetries            = 10                 // 最大重试次数
	retryDelay            = 1 * time.Second    // 重试延迟时间
	exponentialBackoff    = 2                  // 指数退避倍数

	ModeSingle   = "single"
	ModeCluster  = "cluster"
	ModeSentinel = "sentinel"
)

// RedisManager 管理Redis操作和连接池
type RedisManager struct {
	client             redis.UniversalClient
	recordDuration     time.Duration // 记录保持时间
	recordLimit        int64         // 记录限制数量
	pipelineBufferSize int           // 管道缓冲大小
	pipelineCounter    int           // 管道计数器
	dbname             string
}

// redisOptions Redis配置选项
type redisOptions struct {
	Mode               string // Redis模式
	SentinelMasterName string // Sentinel模式下的主节点名称
	Addr               string // Redis服务器地址
	DB                 int    // Redis数据库
	UserName           string // Redis用户名
	Password           string // Redis密码
	PoolSize           int    // 连接池大小
	RecordDuration     string // 记录的有效时间
	RecordLimit        int64  // 记录的条数限制
	PipelineBufferSize int    //管道缓冲大小
}

var (
	managerInstance *RedisManager
	once            sync.Once
)

// DataProcessor 是一个处理新数据的函数类型
type DataProcessor func(data string)

// DB 用于全局获取RedisManager实例
func DB() *RedisManager {
	once.Do(func() {
		cfg := configs.GetInstance()
		// 从配置文件获取redis的ip以及db
		mode := cfg.GetString("redis.default.mode", "single")
		sentinelMasterName := cfg.GetString("redis.default.sentinelMasterName", "sagoo-master")
		address := cfg.GetString("redis.default.address", "localhost:6379")
		db := cfg.GetInt("redis.default.db", 0)
		userName := cfg.GetString("redis.default.user", "default")
		password := cfg.GetString("redis.default.pass", "")
		poolSize := cfg.GetInt("system.deviceCacheData.poolSize", 500)
		recordDuration := cfg.GetString("system.deviceCacheData.recordDuration", "10m")
		recordLimit := cfg.GetInt("system.deviceCacheData.recordLimit", 1000)
		pipelineBufferSize := cfg.GetInt("system.deviceCacheData.pipelineBufferSize", 3)

		// 从配置文件获取redis的ip以及db
		options := redisOptions{
			Mode:               mode,
			SentinelMasterName: sentinelMasterName,
			Addr:               address,
			DB:                 db,
			UserName:           userName,                   // 如果需要用户名
			Password:           password,                   // 如果需要密码
			PoolSize:           poolSize,                   // 设置连接池大小
			RecordDuration:     recordDuration,             // 记录的有效时间
			RecordLimit:        convert.Int64(recordLimit), // 记录的条数限制
			PipelineBufferSize: pipelineBufferSize,         // 将这个值加入到配置结构中

		}
		managerInstance = getRedisManager(options)
	})
	return managerInstance

}

// / getRedisManager 根据配置选项创建RedisManager实例
func getRedisManager(options redisOptions) *RedisManager {
	recordDuration, err := time.ParseDuration(options.RecordDuration)
	if err != nil {
		log.Println("Invalid RecordDuration format, setting to default: 10m")
		recordDuration = 10 * time.Minute
	}

	client := createRedisClient(options)
	if client == nil {
		log.Println(fmt.Sprintf("Failed to connect to Redis after %d retries", maxRetries))
		return nil
	}
	// 动态开启键空间通知
	err = client.ConfigSet(context.Background(), "notify-keyspace-events", "Ex").Err()
	if err != nil {
		fmt.Println("Error setting Redis config:", err)
	}
	return &RedisManager{
		client:             client,
		recordDuration:     recordDuration,
		recordLimit:        options.RecordLimit,
		pipelineBufferSize: options.PipelineBufferSize, // 确保这个值从配置正确赋值
		pipelineCounter:    0,
		dbname:             convert.String(options.DB),
	}
}

// createRedisClient 创建并尝试连接Redis客户端
func createRedisClient(options redisOptions) (client redis.UniversalClient) {
	var err error
	delay := retryDelay
	Adders := strings.Split(options.Addr, ",")
	for i := 0; i < maxRetries; i++ {
		// 根据不同的模式创建Redis客户端
		switch options.Mode {
		case ModeSingle:
			client = redis.NewClient(&redis.Options{
				Addr:     options.Addr,
				Username: options.UserName,
				Password: options.Password,
				DB:       options.DB,
			})
		case ModeCluster:
			client = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    Adders,
				Username: options.UserName,
				Password: options.Password,
			})
		case ModeSentinel:
			client = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:    options.SentinelMasterName,
				SentinelAddrs: Adders,
				Username:      options.UserName,
				Password:      options.Password,
				DB:            options.DB,
			})
		default:
			log.Println(fmt.Sprintf("unsupported Redis mode: %s", options.Mode))

			return nil
		}

		if _, err = client.Ping(context.Background()).Result(); err == nil {
			return client
		}
		log.Println(fmt.Sprintf("Failed to connect to Redis: %v. Retrying in %v...", err, delay))
		time.Sleep(delay)
		delay *= time.Duration(exponentialBackoff)
	}
	log.Println(fmt.Sprintf("Unable to connect to Redis after %d attempts.", maxRetries))
	return nil
}

// GetClient 获取Redis客户端
func (r *RedisManager) GetClient() redis.UniversalClient {
	return r.client
}

func (r *RedisManager) GetDbname() string {
	return r.dbname
}

// InsertBatchData 批量插入数据到Redis
func (r *RedisManager) InsertBatchData(ctx context.Context, key string, data []interface{}) error {
	serializedValue, err := json.Marshal(data)
	if err != nil {
		return err
	}

	fullKey := DeviceDataCachePrefix + key
	pipe := r.client.Pipeline()
	pipe.LPush(ctx, fullKey, serializedValue)
	pipe.LTrim(ctx, fullKey, 0, r.recordLimit-1)

	r.pipelineCounter += 2 // 每次插入操作和修剪操作计算为两个命令

	// 检查是否已经设置过期时间
	ttl, err := r.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return err
	}
	if ttl == -1 { // -1 表示没有设置过期时间
		pipe.Expire(ctx, fullKey, r.recordDuration)
		r.pipelineCounter++
	}

	// 判断是否达到管道缓冲大小，执行并重置计数器
	if r.pipelineCounter >= r.pipelineBufferSize {
		_, err = pipe.Exec(ctx)
		if err != nil {
			return err
		}
		r.pipelineCounter = 0 // 重置管道命令计数器
	}

	return nil
}

// InsertData 插入单条数据到Redis
func (r *RedisManager) InsertData(ctx context.Context, key string, data interface{}, buffer bool) (err error) {
	serializedValue, err := json.Marshal(data)
	if err != nil {
		return
	}

	fullKey := DeviceDataCachePrefix + key
	pipe := r.client.Pipeline()
	pipe.LPush(ctx, fullKey, serializedValue)
	pipe.LTrim(ctx, fullKey, 0, r.recordLimit-1)
	r.pipelineCounter += 2 // 每次插入操作和修剪操作计算为两个命令

	// 设置过期时间，仅当还没有设置时
	ttl, err := r.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return
	}
	if ttl == -1 {
		pipe.Expire(ctx, fullKey, r.recordDuration)
	}

	// 判断是否需要缓冲处理
	if buffer {
		// 判断是否达到管道缓冲大小，或者有未执行的命令，执行并重置计数器
		if r.pipelineCounter >= r.pipelineBufferSize || r.pipelineCounter > 0 {
			_, err = pipe.Exec(ctx)
			if err != nil {
				return
			}
			r.pipelineCounter = 0 // 重置管道命令计数器
		}
	} else {
		// 直接执行管道命令
		_, err = pipe.Exec(ctx)
		if err != nil {
			return
		}
		r.pipelineCounter = 0 // 重置管道命令计数器
	}
	return
}

// GetData 获取最新的数据
func (r *RedisManager) GetData(ctx context.Context, key string) ([]string, error) {
	return r.client.LRange(ctx, DeviceDataCachePrefix+key, 0, r.recordLimit-1).Result()
}

// GetDataByLatest 获取最新的一条数据
func (r *RedisManager) GetDataByLatest(ctx context.Context, key string) (string, error) {
	//获取列表长度
	length := r.client.LLen(ctx, DeviceDataCachePrefix+key)
	return r.client.LIndex(ctx, DeviceDataCachePrefix+key, convert.Int64(length)).Result()
}

// GetDataByPage 按分页获取数据，增加按字段内容搜索和时间区间搜索
func (r *RedisManager) GetDataByPage(ctx context.Context, deviceKey string, pageNum, pageSize int, types, dateRange []string) (res []string, total, currentPage int, err error) {
	listName := DeviceDataCachePrefix + deviceKey

	// 先获取全部数据
	allData, err := r.client.LRange(ctx, listName, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
		return
	}

	if dateRange == nil {
		dateRange = make([]string, 2)
		dateRange[0] = "1970-01-01"
		dateRange[1] = time.Now().Format("2006-01-02")
	}

	startDate, endDate := parseDateRange(dateRange)

	var tmpDataList []string
	for _, item := range allData {
		if matchesTypes(item, types) && inDateRange(item, startDate, endDate) {
			tmpDataList = append(tmpDataList, item)
		}
	}

	total = len(tmpDataList)
	if pageNum <= 0 {
		pageNum = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}
	currentPage = pageNum
	res, _ = getPage(tmpDataList, pageNum, pageSize)

	return
}

// getPage 进行分页
func getPage(data []string, pageNum, pageSize int) ([]string, int) {
	if pageNum < 1 || pageSize < 1 {
		return []string{}, 0
	}
	// 计算总页数
	total := len(data) / pageSize
	if len(data)%pageSize != 0 {
		total += 1
	}

	// 检查页码是否超出范围
	if pageNum > total {
		return []string{}, total
	}

	// 计算分页的起始位置和结束位置
	start := (pageNum - 1) * pageSize
	end := start + pageSize

	// 调整结束位置以防止越界
	if end > len(data) {
		end = len(data)
	}

	// 截取分页数据
	res := data[start:end]

	return res, total
}

// matchesTypes 检查数据是否匹配指定的类型
func matchesTypes(data string, types []string) bool {
	for _, t := range types {
		if strings.Contains(data, t) {
			return true
		}
	}
	return len(types) == 0
}

// inDateRange 检查数据是否在指定的日期范围内
func inDateRange(data string, startDate, endDate time.Time) bool {
	// 这里假设数据中包含可解析的日期格式
	// 实际应根据数据格式进行调整
	// 例如: "2021-02-01 15:04:05 - Some data"
	dateStr := strings.Split(data, " - ")[0]
	dataDate, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return true
	}

	return (startDate.IsZero() || dataDate.After(startDate)) && (endDate.IsZero() || dataDate.Before(endDate))
}

// parseDateRange 解析日期范围
func parseDateRange(dateRange []string) (startDate, endDate time.Time) {
	if len(dateRange) >= 2 {
		startDate, _ = time.Parse("2006-01-02", dateRange[0])
		endDate, _ = time.Parse("2006-01-02", dateRange[1])
	}
	return
}

// ListenForNewData 监听指定的 Redis key，对新数据执行处理函数，interval为轮询间隔
func (r *RedisManager) ListenForNewData(ctx context.Context, key string, processor DataProcessor, interval time.Duration) {
	var lastCheckedSize int64 = 0
	for {
		select {
		case <-ctx.Done():
			//fmt.Println("监听结束")
			return
		default:
			// 获取当前 list 的大小
			currentSize, err := r.client.LLen(ctx, DeviceDataCachePrefix+key).Result()
			if err != nil {
				time.Sleep(time.Second) // 简单的错误恢复
				continue
			}

			if currentSize > lastCheckedSize {
				// 只检查自上次以来的新元素
				newElements, err := r.client.LRange(ctx, DeviceDataCachePrefix+key, 0, currentSize-lastCheckedSize-1).Result()
				if err != nil {
					fmt.Println("获取新元素失败:", err)
					time.Sleep(time.Second) // 简单的错误恢复
					continue
				}

				// 使用传入的处理函数处理新元素
				for _, element := range newElements {
					processor(element)
				}

				// 更新上次检查的位置
				lastCheckedSize = currentSize
			}

			// 休眠以减少轮询频率
			time.Sleep(interval)
		}
	}
}
