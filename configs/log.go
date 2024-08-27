package configs

func LoadLogConfig() *LogConfig {
	config := &LogConfig{
		Level:   EnvString("LOG_LEVEL", "debug"),
		Pattern: EnvString("LOG_PATTERN", ""),
		Output:  EnvString("LOG_OUTPUT", ""),
		LogRotate: LogRotate{
			Filename:   EnvString("LOG_ROTATE_FILE", ""),
			MaxSize:    EnvInt("LOG_ROTATE_MAX_SIZE", ""),
			MaxBackups: EnvInt("LOG_ROTATE_MAX_BACKUPS", ""),
			MaxAge:     EnvInt("LOG_ROTATE_MAX_AGE", ""),
			Compress:   EnvBool("LOG_ROTATE_COMPRESS", true),
		},
	}

	return config
}
