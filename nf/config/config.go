package config

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
	"time"
)

func init() {
	getSystemConfig()
}

// ConfigEntity 表示需要全局调用的实体对象
type ConfigEntity struct {
	config *viper.Viper
}

var (
	instance *ConfigEntity
	once     sync.Once
)

// GetInstance 返回 ConfigDataEntity 的单例实例
func GetInstance() *ConfigEntity {
	once.Do(func() {
		instance = &ConfigEntity{
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
		fmt.Errorf("读取config.toml配置文件失败: %s \n", err)
		return nil
	}

	v.WatchConfig()

	return v
}
func (c *ConfigEntity) GetConfig() *viper.Viper {
	return GetInstance().config
}

// Get 获取interface{}类型配置
func (c *ConfigEntity) Get(key string, defaultValue ...interface{}) interface{} {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.Get(key)
}

// GetBool 获取bool类型配置
func (c *ConfigEntity) GetBool(key string, defaultValue ...bool) bool {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetBool(key)
}

// GetFloat64 获取float64类型配置
func (c *ConfigEntity) GetFloat64(key string, defaultValue ...float64) float64 {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetFloat64(key)
}

// GetInt 获取int类型配置
func (c *ConfigEntity) GetInt(key string, defaultValue ...int) int {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetInt(key)
}

// GetIntSlice 获取[]int类型配置
func (c *ConfigEntity) GetIntSlice(key string, defaultValue ...[]int) []int {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetIntSlice(key)
}

// GetString 获取string类型配置
func (c *ConfigEntity) GetString(key string, defaultValue ...string) string {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetString(key)
}

// GetStringMap 获取map[string]interface{}类型配置
func (c *ConfigEntity) GetStringMap(key string, defaultValue ...map[string]interface{}) map[string]interface{} {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetStringMap(key)
}

// GetStringMapString 获取map[string]string类型配置
func (c *ConfigEntity) GetStringMapString(key string, defaultValue ...map[string]string) map[string]string {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetStringMapString(key)
}

// GetStringSlice 获取[]string类型配置
func (c *ConfigEntity) GetStringSlice(key string, defaultValue ...[]string) []string {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetStringSlice(key)
}

// GetTime 获取time.Time类型配置
func (c *ConfigEntity) GetTime(key string, defaultValue ...time.Time) time.Time {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetTime(key)
}

// GetDuration 获取time.Duration类型配置
func (c *ConfigEntity) GetDuration(key string, defaultValue ...time.Duration) time.Duration {
	if len(defaultValue) > 0 {
		c.config.SetDefault(key, defaultValue[0])
	}
	return c.config.GetDuration(key)
}
