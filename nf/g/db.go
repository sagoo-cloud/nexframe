package g

import (
	"github.com/sagoo-cloud/nexframe/database"
	"github.com/sagoo-cloud/nexframe/nf/config"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func init() {
	config := getDefaultConfig()
	// 获取gorm db
	dbm, err := database.GetGormDB(config)
	if err != nil {
		Log().Error("获取gorm db失败", error.Error(err))
	}
	DB = dbm.GetDB()
}

func getDefaultConfig() config.ModGormDb {
	var config config.ModSwaConfig
	if err := Cfg().Unmarshal(&config); err != nil {
		Log().Error("数据库配置文件解析失败", error.Error(err))
	}
	return config.GormDB
}
