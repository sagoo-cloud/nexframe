package gvar

import (
	"encoding/json"
	"github.com/sagoo-cloud/nexframe/utils/deepcopy"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/sagoo-cloud/nexframe/utils/convert"
)

var varPool = sync.Pool{
	New: func() interface{} {
		return &Var{}
	},
}

type Var struct {
	value interface{} // Underlying value.
	safe  bool        // Concurrent safe or not.
}

// New 函数保持不变
func New(value interface{}, safe ...bool) *Var {
	v := &Var{}
	v.value = value
	if len(safe) > 0 {
		v.safe = safe[0]
	}
	return v
}

// Set 方法需要修复以正确处理 interface{} 类型
func (v *Var) Set(value interface{}) (old interface{}) {
	if v.safe {
		for {
			oldPtr := (*unsafe.Pointer)(unsafe.Pointer(&v.value))
			old = *(*interface{})(atomic.LoadPointer(oldPtr))
			if atomic.CompareAndSwapPointer(
				oldPtr,
				*(*unsafe.Pointer)(unsafe.Pointer(&old)),
				*(*unsafe.Pointer)(unsafe.Pointer(&value)),
			) {
				return old
			}
		}
	} else {
		old = v.value
		v.value = value
		return old
	}
}

// Val 方法需要修改以确保并发安全
func (v *Var) Val() interface{} {
	if v.safe {
		return *(*interface{})(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&v.value))))
	}
	return v.value
}

// Int 方法保持不变
func (v *Var) Int() int {
	val := v.Val()
	if i, ok := val.(int); ok {
		return i
	}
	return 0
}

// Release 释放 Var 对象回到对象池
func (v *Var) Release() {
	if v != nil {
		v.value = nil
		v.safe = false
		varPool.Put(v)
	}
}

// Clone 创建当前 Var 的浅拷贝
func (v *Var) Clone() *Var {
	if v == nil {
		return nil
	}
	return New(v.Val(), v.safe)
}

// Interface Val 的别名
func (v *Var) Interface() interface{} {
	return v.Val()
}

// Bytes 将 Var 转换为 []byte
func (v *Var) Bytes() []byte {
	if v == nil {
		return nil
	}
	switch val := v.Val().(type) {
	case string:
		return []byte(val)
	case []byte:
		return val
	default:
		return []byte(v.String())
	}
}

// String 将 Var 转换为字符串
func (v *Var) String() string {
	if v == nil {
		return ""
	}
	return convert.String(v.Val())
}

// Bool 将 Var 转换为 bool
func (v *Var) Bool() bool {
	if v == nil {
		return false
	}
	return convert.Bool(v.Val())
}

// Int8 将 Var 转换为 int8
func (v *Var) Int8() int8 {
	if v == nil {
		return 0
	}
	return convert.Int8(v.Val())
}

// Int16 将 Var 转换为 int16
func (v *Var) Int16() int16 {
	if v == nil {
		return 0
	}
	return convert.Int16(v.Val())
}

// Int32 将 Var 转换为 int32
func (v *Var) Int32() int32 {
	if v == nil {
		return 0
	}
	return convert.Int32(v.Val())
}

// Int64 将 Var 转换为 int64
func (v *Var) Int64() int64 {
	if v == nil {
		return 0
	}
	return convert.Int64(v.Val())
}

// Uint 将 Var 转换为 uint
func (v *Var) Uint() uint {
	return convert.Uint(v.Val())
}

// Uint8 将 Var 转换为 uint8
func (v *Var) Uint8() uint8 {
	return convert.Uint8(v.Val())
}

// Uint16 将 Var 转换为 uint16
func (v *Var) Uint16() uint16 {
	if v == nil {
		return 0
	}
	return convert.Uint16(v.Val())
}

// Uint32 将 Var 转换为 uint32
func (v *Var) Uint32() uint32 {
	if v == nil {
		return 0
	}
	return convert.Uint32(v.Val())
}

// Uint64 将 Var 转换为 uint64
func (v *Var) Uint64() uint64 {
	if v == nil {
		return 0
	}
	return convert.Uint64(v.Val())
}

// Float32 将 Var 转换为 float32
func (v *Var) Float32() float32 {
	if v == nil {
		return 0
	}
	return convert.Float32(v.Val())
}

// Float64 将 Var 转换为 float64
func (v *Var) Float64() float64 {
	if v == nil {
		return 0
	}
	return convert.Float64(v.Val())
}

// Time 将 Var 转换为 time.Time
func (v *Var) Time(format ...string) time.Time {
	if v == nil {
		return time.Time{}
	}
	switch val := v.Val().(type) {
	case time.Time:
		return val
	case int64:
		return time.Unix(val, 0)
	case string:
		if len(format) > 0 {
			t, _ := time.ParseInLocation(format[0], val, time.Local)
			return t
		}
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", val, time.Local)
		return t
	default:
		return time.Time{}
	}
}

// Duration 将 Var 转换为 time.Duration
func (v *Var) Duration() time.Duration {
	if v == nil {
		return 0
	}
	switch val := v.Val().(type) {
	case time.Duration:
		return val
	case string:
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	case int64:
		return time.Duration(val) * time.Second
	case float64:
		return time.Duration(val * float64(time.Second))
	}
	return time.Duration(v.Int64()) * time.Second
}

// MarshalJSON 实现 json.Marshal 的 MarshalJSON 接口
func (v *Var) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val())
}

// UnmarshalJSON 实现 json.Unmarshal 的 UnmarshalJSON 接口
func (v *Var) UnmarshalJSON(b []byte) error {
	var i interface{}
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	v.Set(convertNumberToAppropriateType(i))
	return nil
}

func convertNumberToAppropriateType(i interface{}) interface{} {
	switch v := i.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{})
		for k, val := range v {
			m[k] = convertNumberToAppropriateType(val)
		}
		return m
	case []interface{}:
		for i, val := range v {
			v[i] = convertNumberToAppropriateType(val)
		}
	case float64:
		if float64(int64(v)) == v {
			return int64(v)
		}
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			if float64(int64(f)) == f {
				return int64(f)
			}
			return f
		}
	}
	return i
}

// UnmarshalValue 实现接口，为 Var 设置任意类型的值
func (v *Var) UnmarshalValue(value interface{}) error {
	v.Set(value)
	return nil
}

// DeepCopy 实现当前类型深拷贝的接口
func (v *Var) DeepCopy() interface{} {
	if v == nil {
		return nil
	}
	return New(deepcopy.Copy(v.Val()), v.safe)
}
