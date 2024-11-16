// Package g 提供全局变量和初始化功能。
package g

import (
	"context"
	"github.com/robfig/cron/v3"
	"github.com/sagoo-cloud/nexframe/configs"
	"github.com/sagoo-cloud/nexframe/database"
	"github.com/sagoo-cloud/nexframe/database/redisdb"
	"github.com/sagoo-cloud/nexframe/os/cache"
	"github.com/sagoo-cloud/nexframe/os/zlog"
	"github.com/sagoo-cloud/nexframe/utils/valid"
	"gorm.io/gorm"
	"sync"
)

var (
	// once 确保初始化过程只执行一次
	once sync.Once
	// Cache 全局缓存管理器实例
	Cache *cache.CacheManager
	// Cfg 全局配置实体实例
	Cfg *configs.ConfigEntity
	// Log 全局日志记录器实例
	Log zlog.Logger
	// Cron 全局定时任务管理器
	Cron *cron.Cron
	// DB 全局数据库连接实例
	DB *gorm.DB
	// Validator 全局验证器实例
	Validator *valid.Validator
	// RedisDB 全局Redis数据库管理器
	RedisDB *redisdb.RedisManager
)

// Init 初始化全局变量和服务。
func Init(ctx context.Context) error {
	var err error
	once.Do(func() {
		// 初始化配置
		Cfg = configs.GetInstance()

		// 初始化Redis数据库
		RedisDB = redisdb.DB()

		// 初始化缓存管理器
		config := configs.LoadCacheConfig()
		Cache = cache.NewCacheManager(config)

		// 初始化日志
		Log = zlog.NewLogger()

		// 初始化定时任务
		Cron = cron.New()

		// 初始化数据库
		dbConfig := configs.LoadDatabaseConfig()
		dbm, _ := database.GetGormDB(dbConfig) // 获取gorm db
		DB = dbm.GetDB()

		// 初始化验证器
		Validator = valid.New()

	})
	return err
}
