package convert

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// String 将任意类型转换为字符串
func String(any interface{}) string {
	if any == nil {
		return ""
	}

	// 使用类型断言处理常见类型
	switch v := any.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case time.Time:
		if v.IsZero() {
			return ""
		}
		return v.Format(time.RFC3339)
	case fmt.Stringer:
		return v.String()
	case error:
		return v.Error()
	}

	// 使用反射处理其他类型
	rv := reflect.ValueOf(any)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return ""
		}
		return String(rv.Elem().Interface())
	case reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		if rv.IsNil() {
			return ""
		}
	}

	// 最后尝试使用 JSON 编码
	jsonContent, err := json.Marshal(any)
	if err != nil {
		return fmt.Sprint(any)
	}
	return string(jsonContent)
}
