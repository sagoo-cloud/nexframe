package configs

type LogConfig struct {
	Level     string    `mapstructure:"level" json:"level" yaml:"level"`
	Pattern   string    `mapstructure:"pattern" json:"pattern" yaml:"pattern"`
	Output    string    `mapstructure:"output" json:"output" yaml:"output"`
	LogRotate LogRotate `mapstructure:"logRotate" json:"logRotate" yaml:"logRotate"`
}

type LogRotate struct {
	Filename   string `mapstructure:"filename" json:"filename" yaml:"filename"`
	MaxSize    int    `mapstructure:"maxSize" json:"maxSize" yaml:"maxSize"`
	MaxBackups int    `mapstructure:"maxBackups" json:"maxBackups" yaml:"maxBackups"`
	MaxAge     int    `mapstructure:"maxAge" json:"maxAge" yaml:"maxAge"`
	Compress   bool   `mapstructure:"compress" json:"compress" yaml:"compress"`
}

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
