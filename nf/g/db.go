package g

import (
	"github.com/sagoo-cloud/nexframe/database"
	"github.com/sagoo-cloud/nexframe/nf/config"
	"gorm.io/gorm"
	"log"
	"os"
)

var (
	DB *gorm.DB
)

func init() {
	defer func() {
		if r := recover(); r != nil {
			// 自定义错误信息，隐藏详细路径
			log.Printf("Error: %s , %s", "Failed to obtain Gorm DB", r)
			os.Exit(1)
		}
	}()

	config := getDefaultConfig()
	// 获取gorm db
	dbm, _ := database.GetGormDB(config)
	DB = dbm.GetDB()
}

func getDefaultConfig() config.ModGormDb {
	var config config.ModSwaConfig
	if err := Cfg().GetConfig().Unmarshal(&config); err != nil {
		Log().Error("数据库配置文件解析失败", error.Error(err))
	}
	return config.GormDB
}
