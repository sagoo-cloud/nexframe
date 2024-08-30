package main

import (
	"github.com/sagoo-cloud/nexframe/configs"
	"github.com/sagoo-cloud/nexframe/os/zlog"
	"time"
)

func logSomething(logger zlog.Logger) {
	logger.Info("This is logged from logSomething function")
}

func main() {
	// 获取全局日志实例
	logger := zlog.GetLogger()

	// 使用默认配置（生产模式）
	logger.Info("Starting application with default configuration")
	logger.Debug("This debug message should not appear in production mode")

	// 修改日志配置为开发模式
	devConfig := configs.LogConfig{
		Level:   "trace",
		Pattern: "development",
		Output:  "stdout",
	}
	logger.SetConfig(devConfig)

	// 记录不同级别的日志
	logger.Trace("Trace level message")
	logger.Debug("Debug level message")
	logger.Info("Info level message")
	logger.Warn("Warning level message")
	logger.Error("Error level message")

	// 从另一个函数记录日志
	logSomething(logger)

	// 记录带有额外参数的结构化日志
	logger.Info("User action", "user_id", 12345, "action", "login", "timestamp", time.Now().Unix())

	// 切换回生产模式
	prodConfig := configs.LogConfig{
		Level:   "info",
		Pattern: "production",
		Output:  "stdout",
	}
	logger.SetConfig(prodConfig)

	logger.Info("Switched back to production mode")
	logger.Debug("This debug message should not appear in production mode")
	logger.Info("This info message should appear in production mode")
}
