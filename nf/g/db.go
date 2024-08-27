package g

import (
	"github.com/sagoo-cloud/nexframe/configs"
	"github.com/sagoo-cloud/nexframe/database"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func init() {
	defer func() {
		if r := recover(); r != nil {
			//log.Printf("Error: %s , %s", "Failed to obtain Gorm DB", r)
			//os.Exit(1)
		}
	}()

	dbConfig := configs.LoadDatabaseConfig()
	// 获取gorm db
	dbm, _ := database.GetGormDB(dbConfig)
	DB = dbm.GetDB()
}
