package configs

func LoadLogConfig() LogConfig {
	config := LogConfig{
		Level:   EnvString(LogLevel, "debug"),
		Pattern: EnvString(LogPattern, "prod"),
		Output:  EnvString(LogOutput, "stdout"),
		LogRotate: LogRotate{
			Filename:   EnvString(LogRotateFile, "app.log"),
			MaxSize:    EnvInt(LogRotateMaxSize, 100),
			MaxBackups: EnvInt(LogRotateMaxBackups, 3),
			MaxAge:     EnvInt(LogRotateMaxAge, 7),
			Compress:   EnvBool(LogRotateCompress, true),
		},
	}

	return config
}
