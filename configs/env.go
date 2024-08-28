package configs

import (
	"github.com/sagoo-cloud/nexframe/os/command/args"
	"strings"
)

func Env(key string, value ...interface{}) interface{} {
	mode := args.Mode
	modeKey := strings.Join([]string{mode, key}, ".")
	if cfg.IsSet(modeKey) {
		return cfg.Get(modeKey)
	} else if cfg.IsSet(key) {
		return cfg.Get(key)
	} else {
		return value[0]
	}
}

func EnvString(key string, value ...interface{}) string {
	mode := args.Mode
	modeKey := strings.Join([]string{mode, key}, ".")
	var ret string
	if cfg.IsSet(modeKey) {
		ret = cfg.GetString(modeKey)
	} else if cfg.IsSet(key) {
		ret = cfg.GetString(key)
	} else {
		ret = value[0].(string)
	}
	return ret
}
func EnvInt(key string, value ...interface{}) int {
	mode := args.Mode
	modeKey := strings.Join([]string{mode, key}, ".")
	var ret int
	if cfg.IsSet(modeKey) {
		ret = cfg.GetInt(modeKey)
	} else if cfg.IsSet(key) {
		ret = cfg.GetInt(key)
	} else {
		ret = value[0].(int)
	}
	return ret
}
func EnvBool(key string, value ...interface{}) bool {
	mode := args.Mode
	modeKey := strings.Join([]string{mode, key}, ".")
	var ret bool
	if cfg.IsSet(modeKey) {
		ret = cfg.GetBool(modeKey)
	} else if cfg.IsSet(key) {
		ret = cfg.GetBool(key)
	} else {
		ret = value[0].(bool)
	}
	return ret
}
func EnvStringSlice(key string, value ...interface{}) []string {
	mode := args.Mode
	modeKey := strings.Join([]string{mode, key}, ".")
	var ret []string
	if cfg.IsSet(modeKey) {
		ret = cfg.GetStringSlice(modeKey)
	} else if cfg.IsSet(key) {
		ret = cfg.GetStringSlice(key)
	} else {
		ret = value[0].([]string)
	}
	return ret
}
