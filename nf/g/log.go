package g

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sagoo-cloud/nexframe/nf/configs"
	"github.com/sagoo-cloud/nexframe/zlog"
)

var logger zlog.Logger

func init() {
	logger = zlog.GetLogger()
	newConfig := getLogConfig()
	logger.SetConfig(newConfig)
}
func Log(name ...string) zlog.Logger {
	return logger
}

func getLogConfig() zlog.LogConfig {
	var swaConfig configs.ModSwaConfig
	v := Cfg().GetConfig()
	if v == nil {
		return zlog.LogConfig{}
	}
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err := v.Unmarshal(&swaConfig); err != nil {
			fmt.Println("系统配置数据被修改", error.Error(err))
		}
	})

	if err := v.Unmarshal(&swaConfig); err != nil {
		fmt.Println("系统配置数据转换失败", error.Error(err))
	}

	return zlog.LogConfig{
		Level:   swaConfig.Log.Level,
		Pattern: swaConfig.Log.Pattern,
		Output:  swaConfig.Log.Output,
		LogRotate: zlog.LogRotate{
			Filename:   swaConfig.Log.LogRotate.Filename,
			MaxSize:    swaConfig.Log.LogRotate.MaxSize,
			MaxAge:     swaConfig.Log.LogRotate.MaxAge,
			MaxBackups: swaConfig.Log.LogRotate.MaxBackups,
			Compress:   swaConfig.Log.LogRotate.Compress,
		},
	}
}
