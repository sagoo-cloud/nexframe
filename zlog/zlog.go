package zlog

import (
	"fmt"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"strings"
	"sync"
)

const (
	development = "dev"
	production  = "prod"
)

// LogConfig holds the configuration for the logger
type LogConfig struct {
	Level     string    `yaml:"Level"`
	Pattern   string    `yaml:"Pattern"`
	Output    string    `yaml:"Output"`
	LogRotate LogRotate `yaml:"LogRotate"`
}

// LogRotate holds the configuration for log rotation
type LogRotate struct {
	Filename   string `yaml:"Filename"`
	MaxSize    int    `yaml:"MaxSize"`
	MaxBackups int    `yaml:"MaxBackups"`
	MaxAge     int    `yaml:"MaxAge"`
	Compress   bool   `yaml:"Compress"`
}

var DefaultConfig = LogConfig{
	Level:   "debug",
	Pattern: production,
	Output:  "stdout",
	LogRotate: LogRotate{
		Filename:   "app.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	},
}

// Logger interface defines the logging methods
type Logger interface {
	Trace(message string, args ...interface{})
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
	Fatal(message string, args ...interface{})
	Panic(message string, args ...interface{})

	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	SetConfig(config LogConfig)
	GetConfig() LogConfig
	IsTraceEnabled() bool
	IsDebugEnabled() bool
	IsInfoEnabled() bool
	IsWarnEnabled() bool
	IsErrorEnabled() bool
	IsFatalEnabled() bool
	IsPanicEnabled() bool
}

type loggerImpl struct {
	zl             zerolog.Logger
	config         LogConfig
	isTraceEnabled bool
	isDebugEnabled bool
	isInfoEnabled  bool
	isWarnEnabled  bool
	isErrorEnabled bool
	isFatalEnabled bool
	isPanicEnabled bool
	output         io.Writer
	httpWriter     *LogHTTPWriter
}

var (
	globalLogger *loggerImpl
	once         sync.Once
)

// GetLogger returns the global logger instance
func GetLogger() Logger {
	once.Do(func() {
		globalLogger = &loggerImpl{
			config: DefaultConfig,
		}
		globalLogger.init()
	})
	return globalLogger
}
func (l *loggerImpl) init() {
	level, _ := zerolog.ParseLevel(strings.ToLower(l.config.Level))
	zerolog.SetGlobalLevel(level)
	l.updateEnabledFlags(level)

	var newOutput io.Writer
	if l.config.Output == "stdout" {
		newOutput = os.Stdout
	} else {
		newOutput = &lumberjack.Logger{
			Filename:   l.config.LogRotate.Filename,
			MaxSize:    l.config.LogRotate.MaxSize,
			MaxBackups: l.config.LogRotate.MaxBackups,
			MaxAge:     l.config.LogRotate.MaxAge,
			Compress:   l.config.LogRotate.Compress,
		}
	}

	if l.config.Pattern == development {
		output := zerolog.ConsoleWriter{
			Out:        newOutput,
			TimeFormat: "2006-01-02 15:04:05",
			NoColor:    false,
		}

		l.zl = zerolog.New(output).With().Timestamp().Logger()
	} else {
		l.zl = zerolog.New(newOutput).With().Timestamp().Logger()
	}
}
func (l *loggerImpl) SetConfig(config LogConfig) {
	configChanged := l.config != config
	l.config = config
	if configChanged {
		l.init()
		if l.config.Pattern == development && l.isDebugEnabled {
			l.Debug("Logger configuration updated",
				"level", l.config.Level,
				"output", l.config.Output,
				"pattern", l.config.Pattern)
		}
	}
}
func (l *loggerImpl) updateEnabledFlags(level zerolog.Level) {
	l.isTraceEnabled = level <= zerolog.TraceLevel
	l.isDebugEnabled = level <= zerolog.DebugLevel
	l.isInfoEnabled = level <= zerolog.InfoLevel
	l.isWarnEnabled = level <= zerolog.WarnLevel
	l.isErrorEnabled = level <= zerolog.ErrorLevel
	l.isFatalEnabled = level <= zerolog.FatalLevel
	l.isPanicEnabled = level <= zerolog.PanicLevel
}

func (l *loggerImpl) GetConfig() LogConfig {
	return l.config
}

func (l *loggerImpl) log(level zerolog.Level, message string, args ...interface{}) {
	var event *zerolog.Event
	event = l.zl.WithLevel(level)
	if len(args) > 0 {
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				event = event.Interface(args[i].(string), args[i+1])
			}
		}
	}
	event.Msg(message)
}
func (l *loggerImpl) Trace(message string, args ...interface{}) {
	if l.isTraceEnabled {
		l.log(zerolog.TraceLevel, message, args...)
	}
}

func (l *loggerImpl) Debug(message string, args ...interface{}) {
	if l.isDebugEnabled {
		l.log(zerolog.DebugLevel, message, args...)
	}
}

func (l *loggerImpl) Info(message string, args ...interface{}) {
	if l.isInfoEnabled {
		l.log(zerolog.InfoLevel, message, args...)
	}
}

func (l *loggerImpl) Warn(message string, args ...interface{}) {
	if l.isWarnEnabled {
		l.log(zerolog.WarnLevel, message, args...)
	}
}

func (l *loggerImpl) Error(message string, args ...interface{}) {
	if l.isErrorEnabled {
		l.log(zerolog.ErrorLevel, message, args...)
	}
}

func (l *loggerImpl) Fatal(message string, args ...interface{}) {
	if l.isFatalEnabled {
		l.log(zerolog.FatalLevel, message, args...)
	}
}

func (l *loggerImpl) Panic(message string, args ...interface{}) {
	if l.isPanicEnabled {
		l.log(zerolog.PanicLevel, message, args...)
	}
}

func (l *loggerImpl) logf(level zerolog.Level, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.zl.WithLevel(level).Msg(message)
}

func (l *loggerImpl) Tracef(format string, args ...interface{}) {
	if l.isTraceEnabled {
		l.logf(zerolog.TraceLevel, format, args...)
	}
}

func (l *loggerImpl) Debugf(format string, args ...interface{}) {
	if l.isDebugEnabled {
		l.logf(zerolog.DebugLevel, format, args...)
	}
}

func (l *loggerImpl) Infof(format string, args ...interface{}) {
	if l.isInfoEnabled {
		l.logf(zerolog.InfoLevel, format, args...)
	}
}

func (l *loggerImpl) Warnf(format string, args ...interface{}) {
	if l.isWarnEnabled {
		l.logf(zerolog.WarnLevel, format, args...)
	}
}

func (l *loggerImpl) Errorf(format string, args ...interface{}) {
	if l.isErrorEnabled {
		l.logf(zerolog.ErrorLevel, format, args...)
	}
}

func (l *loggerImpl) Fatalf(format string, args ...interface{}) {
	if l.isFatalEnabled {
		l.logf(zerolog.FatalLevel, format, args...)
	}
}

func (l *loggerImpl) Panicf(format string, args ...interface{}) {
	if l.isPanicEnabled {
		l.logf(zerolog.PanicLevel, format, args...)
	}
}

func (l *loggerImpl) IsTraceEnabled() bool {
	return l.isTraceEnabled
}

func (l *loggerImpl) IsDebugEnabled() bool {
	return l.isDebugEnabled
}

func (l *loggerImpl) IsInfoEnabled() bool {
	return l.isInfoEnabled
}

func (l *loggerImpl) IsWarnEnabled() bool {
	return l.isWarnEnabled
}

func (l *loggerImpl) IsErrorEnabled() bool {
	return l.isErrorEnabled
}

func (l *loggerImpl) IsFatalEnabled() bool {
	return l.isFatalEnabled
}

func (l *loggerImpl) IsPanicEnabled() bool {
	return l.isPanicEnabled
}
