package database

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/configs"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DBConfig 数据库配置结构体
type DBConfig struct {
	Driver string
	DSN    string
	Config *gorm.Config
}

// DBManager 数据库管理器
type DBManager struct {
	db   *gorm.DB
	mu   sync.RWMutex
	conf DBConfig
}

var (
	instance *DBManager
	once     sync.Once
)

// GetDBManager 获取数据库管理器单例
func GetDBManager() *DBManager {
	once.Do(func() {
		instance = &DBManager{}
	})
	return instance
}

// InitDB 初始化数据库连接
func (m *DBManager) InitDB(config DBConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.conf = config
	db, err := connectDatabase(config)
	if err != nil {
		return err
	}

	m.db = db
	return nil
}

// GetDB 获取数据库连接
func (m *DBManager) GetDB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.db
}

// connectDatabase 连接数据库
func connectDatabase(config DBConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch config.Driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(config.DSN), config.Config)
	case "postgres":
		db, err = gorm.Open(postgres.Open(config.DSN), config.Config)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.DSN), config.Config)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", config.Driver)
	}

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// ReconnectIfNeeded 检查连接并在需要时重连
func (m *DBManager) ReconnectIfNeeded() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sqlDB, err := m.db.DB()
	if err != nil {
		return m.reconnect()
	}

	if err := sqlDB.Ping(); err != nil {
		return m.reconnect()
	}

	return nil
}

// reconnect 重新连接数据库
func (m *DBManager) reconnect() error {
	db, err := connectDatabase(m.conf)
	if err != nil {
		return err
	}

	m.db = db
	return nil
}

// StartHealthCheck 启动健康检查
func (m *DBManager) StartHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := m.ReconnectIfNeeded(); err != nil {
				fmt.Printf("数据库重连失败: %v\n", err)
			}
		}
	}()
}

// GetGormDB 获取gorm.DB
func GetGormDB(dbConfig *configs.ModGormDb) (dbm *DBManager, err error) {
	manager := GetDBManager()

	// 如果数据库管理器未初始化，则进行初始化
	if manager.GetDB() == nil {
		dsnStr := dbConfig.Dsn
		if dbConfig.Dsn == "" {
			dsnStr = SetDsn(dbConfig)
		}

		//数据库日志配置
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold（慢速SQL阈值）
				LogLevel:                  logger.Info, // Log level（日志级别）
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				ParameterizedQueries:      true,        // Don't include params in the SQL log（忽略记录器的ErrRecordNotFound错误）
				Colorful:                  true,        // Disable color(禁用颜色)设置彩色打印
			},
		)

		cnf := DBConfig{
			Driver: dbConfig.Driver,
			DSN:    dsnStr,
			Config: &gorm.Config{
				Logger:                                   newLogger,
				DisableForeignKeyConstraintWhenMigrating: true,
				PrepareStmt:                              true,
				SkipDefaultTransaction:                   true,
			},
		}
		if err := manager.InitDB(cnf); err != nil {
			return nil, err
		}

		// 启动健康检查
		manager.StartHealthCheck(time.Minute * 5)
	}

	return manager, nil
}
func SetDsn(m *configs.ModGormDb) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		m.Username,
		m.Password,
		m.Host,
		m.Port,
		m.Dbname,
		m.Config,
	)
}
