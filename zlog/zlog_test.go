package zlog

import "testing"

func TestLoggerImpl(t *testing.T) {
	logger := GetLogger()

	// 使用默认配置
	logger.Info("This is an info message with default config")

	// 修改配置
	newConfig := LogConfig{
		Level:   "debug",
		Pattern: "development",
		Output:  "file",
		LogRotate: LogRotate{
			Filename: "app.log",
			MaxSize:  50,
		},
	}
	logger.SetConfig(newConfig)

	// 使用新配置
	logger.Debug("This is a debug message with new config")

	if logger.IsInfoEnabled() {
		// 只有在 Info 级别启用时才执行
		logger.Info("This is another info message")
	}
}
