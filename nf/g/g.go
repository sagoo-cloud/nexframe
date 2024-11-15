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
	once      sync.Once
	Cache     *cache.CacheManager
	Cfg       *configs.ConfigEntity
	Log       zlog.Logger
	Cron      *cron.Cron
	DB        *gorm.DB
	Validator *valid.Validator
	RedisDB   *redisdb.RedisManager
)

func Init(ctx context.Context) error {
	var err error
	once.Do(func() {
		// 初始化配置
		Cfg = configs.GetInstance()

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

		// 初始化 基于Redis数据库
		RedisDB = redisdb.DB()
	})
	return err
}
