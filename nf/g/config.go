package g

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
)

func Cfg() *viper.Viper {
	return GetInstance().config
}

func init() {
	getSystemConfig()
}

// ConfigDataEntity 表示需要全局调用的实体对象
type ConfigDataEntity struct {
	config *viper.Viper
}

var (
	instance *ConfigDataEntity
	once     sync.Once
)

// GetInstance 返回 ConfigDataEntity 的单例实例
func GetInstance() *ConfigDataEntity {
	once.Do(func() {
		instance = &ConfigDataEntity{
			config: getSystemConfig(),
		}
	})
	return instance
}

func getSystemConfig() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("../../config")

	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("读取swa toml配置文件失败: %s \n", err))
	}

	v.WatchConfig()

	return v
}
