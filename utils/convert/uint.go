package convert

import (
	"github.com/sagoo-cloud/nexframe/encoding/gbinary"
	"math"
	"reflect"
	"strconv"
)

// uintConverter 是一个通用的无符号整数转换函数类型
type uintConverter func(uint64) uint64

// uintConvert 是一个通用的无符号整数转换函数
func uintConvert[T uint | uint8 | uint16 | uint32 | uint64](any interface{}, converter uintConverter) T {
	if any == nil {
		return 0
	}
	return T(converter(Uint64(any)))
}

// Uint 将任意类型转换为 uint
func Uint(any interface{}) uint {
	return uintConvert[uint](any, func(v uint64) uint64 { return v })
}

// Uint8 将任意类型转换为 uint8，超过 255 的值会被截断为 255
func Uint8(any interface{}) uint8 {
	return uintConvert[uint8](any, func(v uint64) uint64 {
		if v > 255 {
			return 255
		}
		return v
	})
}

// Uint16 将任意类型转换为 uint16，超过 65535 的值会被截断为 65535
func Uint16(any interface{}) uint16 {
	return uintConvert[uint16](any, func(v uint64) uint64 {
		if v > 65535 {
			return 65535
		}
		return v
	})
}

// Uint32 将任意类型转换为 uint32，超过 4294967295 的值会被截断为 4294967295
func Uint32(any interface{}) uint32 {
	return uintConvert[uint32](any, func(v uint64) uint64 {
		if v > 4294967295 {
			return 4294967295
		}
		return v
	})
}

// Uint64 将任意类型转换为 uint64
func Uint64(any interface{}) uint64 {
	if any == nil {
		return 0
	}
	switch value := any.(type) {
	case int:
		return uint64(value)
	case int8:
		return uint64(uint8(value))
	case int16:
		return uint64(uint16(value))
	case int32:
		return uint64(uint32(value))
	case int64:
		return uint64(value)
	case uint:
		return uint64(value)
	case uint8:
		return uint64(value)
	case uint16:
		return uint64(value)
	case uint32:
		return uint64(value)
	case uint64:
		return value
	case float32:
		if value < 0 {
			return 0
		}
		return uint64(value)
	case float64:
		if value < 0 {
			return 0
		}
		return uint64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	case []byte:
		return gbinary.DecodeToUint64(value)
	case string:
		return parseString(value)
	default:
		v := reflect.ValueOf(any)
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return 0
			}
			v = v.Elem()
		}
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return uint64(v.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return v.Uint()
		case reflect.Float32, reflect.Float64:
			f := v.Float()
			if f < 0 {
				return 0
			}
			return uint64(f)
		default:
			return parseString(String(any))
		}
	}
}

// parseString 尝试将字符串解析为 uint64
func parseString(s string) uint64 {
	// 十六进制
	if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		if v, err := strconv.ParseUint(s[2:], 16, 64); err == nil {
			return v
		}
	}
	// 十进制
	if v, err := strconv.ParseUint(s, 10, 64); err == nil {
		return v
	}
	// 浮点数
	if v, err := strconv.ParseFloat(s, 64); err == nil && !math.IsNaN(v) {
		if v < 0 {
			return 0
		}
		return uint64(v)
	}
	return 0
}
